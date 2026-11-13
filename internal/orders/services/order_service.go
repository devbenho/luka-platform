package services

import (
	"context"
	stderrors "errors"
	"fmt"
	"time"

	"github.com/devbenho/luka-platform/internal/inventory/services"
	"github.com/devbenho/luka-platform/internal/orders/models"
	dtos "github.com/devbenho/luka-platform/internal/orders/order_dtos"
	"github.com/devbenho/luka-platform/internal/orders/repositories"
	productService "github.com/devbenho/luka-platform/internal/product/services"
	storeRepositories "github.com/devbenho/luka-platform/internal/store/repositories"
	storeServices "github.com/devbenho/luka-platform/internal/store/services"
	"github.com/devbenho/luka-platform/pkg/database"
	"github.com/devbenho/luka-platform/pkg/errors"
	"github.com/devbenho/luka-platform/pkg/validation"
	"go.mongodb.org/mongo-driver/mongo"
)

// IOrderService manages orders. Access rules:
//   - an order is visible to its customer and to its seller (the user who
//     owns the store of every product in the order);
//   - the customer may only cancel; the seller may make any valid transition.
type IOrderService interface {
	CreateOrder(ctx context.Context, dto dtos.CreateOrderRequest) (*models.Order, error)
	UpdateOrderStatus(ctx context.Context, id, userID string, status models.OrderStatus) error
	GetOrderByID(ctx context.Context, id, userID string) (*models.Order, error)
	ListOrders(ctx context.Context, customerID string) ([]models.Order, error)
}

type OrderService struct {
	repo             repositories.IOrderRepository
	inventoryService services.IInventoryService
	productService   productService.IProductService
	storeRepo        storeRepositories.IStoreRepository
	tx               database.Transactor
	validator        *validation.Validator
}

func NewOrderService(
	repo repositories.IOrderRepository,
	inventoryService services.IInventoryService,
	productService productService.IProductService,
	storeRepo storeRepositories.IStoreRepository,
	tx database.Transactor,
	validator *validation.Validator,
) *OrderService {
	return &OrderService{
		repo:             repo,
		inventoryService: inventoryService,
		productService:   productService,
		storeRepo:        storeRepo,
		tx:               tx,
		validator:        validator,
	}
}

// CreateOrder reserves stock for every item and saves the order in one
// transaction: if any item is out of stock or the save fails, no stock is
// taken and no order is written.
func (s *OrderService) CreateOrder(ctx context.Context, dto dtos.CreateOrderRequest) (*models.Order, error) {
	if err := s.validator.ValidateStruct(dto); err != nil {
		return nil, err
	}

	orderItems, totalAmount, err := s.prepareOrderItems(ctx, dto.Items)
	if err != nil {
		return nil, err
	}

	var createdOrder *models.Order
	err = s.tx.RunInTransaction(ctx, func(ctx context.Context) error {
		if err := s.reserveInventory(ctx, orderItems); err != nil {
			return err
		}
		order, err := s.repo.CreateOrder(ctx, s.buildOrder(dto, orderItems, totalAmount))
		if err != nil {
			return errors.Wrap(err, "creating order")
		}
		createdOrder = order
		return nil
	})
	if err != nil {
		return nil, err
	}
	return createdOrder, nil
}

// prepareOrderItems prices each requested item from the product catalogue.
// Stock is checked atomically when it is reserved, not here.
func (s *OrderService) prepareOrderItems(ctx context.Context, items []dtos.CreateOrderItemRequest) ([]models.OrderItem, float64, error) {
	var totalAmount float64
	orderItems := make([]models.OrderItem, len(items))

	for i, item := range items {
		product, err := s.productService.GetProductByID(ctx, item.ProductID.Hex())
		if err != nil {
			return nil, 0, err
		}

		itemTotal := float64(item.Quantity) * product.Price
		orderItems[i] = models.OrderItem{
			ProductID:  item.ProductID,
			Quantity:   item.Quantity,
			UnitPrice:  product.Price,
			TotalPrice: itemTotal,
		}
		totalAmount += itemTotal
	}

	return orderItems, totalAmount, nil
}

func (s *OrderService) buildOrder(dto dtos.CreateOrderRequest, items []models.OrderItem, totalAmount float64) *models.Order {
	return &models.Order{
		CustomerID:      dto.CustomerID,
		Items:           items,
		Status:          models.OrderStatusPending,
		TotalAmount:     totalAmount,
		ShippingAddress: dto.ShippingAddress,
		Notes:           dto.Notes,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
}

// reserveInventory reserves stock for every item, recording on each item the
// inventory record it came from. It must run inside a transaction, which
// undoes earlier reservations if a later one fails.
func (s *OrderService) reserveInventory(ctx context.Context, items []models.OrderItem) error {
	for i := range items {
		inventoryID, err := s.inventoryService.ReserveStock(ctx, items[i].ProductID, items[i].Quantity)
		if err != nil {
			return err
		}
		items[i].InventoryID = inventoryID
	}
	return nil
}

// releaseInventory returns the stock reserved for items.
func (s *OrderService) releaseInventory(ctx context.Context, items []models.OrderItem) error {
	for _, item := range items {
		if item.InventoryID.IsZero() {
			continue // order created before reservations were tracked
		}
		if err := s.inventoryService.ReleaseStock(ctx, item.InventoryID, item.Quantity); err != nil {
			return fmt.Errorf("releasing stock for product %s: %w", item.ProductID.Hex(), err)
		}
	}
	return nil
}

func (s *OrderService) UpdateOrderStatus(ctx context.Context, id, userID string, status models.OrderStatus) error {
	if !isValidStatus(status) {
		return errors.NewBadRequestError("invalid order status")
	}

	order, err := s.getOrder(ctx, id)
	if err != nil {
		return err
	}

	isSeller, err := s.isSeller(ctx, order, userID)
	if err != nil {
		return err
	}
	isCustomer := order.CustomerID.Hex() == userID
	switch {
	case isSeller:
	case isCustomer && status == models.OrderStatusCancelled:
	case isCustomer:
		return errors.NewForbiddenError("customers can only cancel their orders")
	default:
		return errors.NewForbiddenError("you cannot change this order")
	}

	if !isValidStatusTransition(order.Status, status) {
		return errors.NewBadRequestError("invalid status transition")
	}

	// The status change and (on cancel) the stock release commit together.
	return s.tx.RunInTransaction(ctx, func(ctx context.Context) error {
		// Only succeeds if nobody changed the status since we read it, so
		// concurrent requests cannot both cancel (and release stock twice).
		err := s.repo.UpdateOrderStatus(ctx, id, order.Status, status)
		if stderrors.Is(err, mongo.ErrNoDocuments) {
			return errors.NewConflictError("order status changed concurrently; retry")
		}
		if err != nil {
			return errors.Wrap(err, "updating order status")
		}
		if status == models.OrderStatusCancelled {
			if err := s.releaseInventory(ctx, order.Items); err != nil {
				return errors.Wrap(err, "releasing stock for cancelled order")
			}
		}
		return nil
	})
}

func isValidStatusTransition(from, to models.OrderStatus) bool {
	transitions := map[models.OrderStatus][]models.OrderStatus{
		models.OrderStatusPending:    {models.OrderStatusProcessing, models.OrderStatusCancelled},
		models.OrderStatusProcessing: {models.OrderStatusShipped, models.OrderStatusCancelled},
		models.OrderStatusShipped:    {models.OrderStatusDelivered},
		models.OrderStatusDelivered:  {},
		models.OrderStatusCancelled:  {},
	}

	allowedTransitions, exists := transitions[from]
	if !exists {
		return false
	}

	for _, allowed := range allowedTransitions {
		if to == allowed {
			return true
		}
	}
	return false
}

func isValidStatus(status models.OrderStatus) bool {
	validStatuses := []models.OrderStatus{
		models.OrderStatusPending,
		models.OrderStatusProcessing,
		models.OrderStatusShipped,
		models.OrderStatusDelivered,
		models.OrderStatusCancelled,
	}

	for _, s := range validStatuses {
		if status == s {
			return true
		}
	}
	return false
}

// GetOrderByID returns the order if userID is its customer or seller.
func (s *OrderService) GetOrderByID(ctx context.Context, id, userID string) (*models.Order, error) {
	order, err := s.getOrder(ctx, id)
	if err != nil {
		return nil, err
	}
	if order.CustomerID.Hex() == userID {
		return order, nil
	}
	isSeller, err := s.isSeller(ctx, order, userID)
	if err != nil {
		return nil, err
	}
	if !isSeller {
		return nil, errors.NewForbiddenError("you cannot view this order")
	}
	return order, nil
}

// isSeller reports whether userID owns the store of every product in order.
func (s *OrderService) isSeller(ctx context.Context, order *models.Order, userID string) (bool, error) {
	if len(order.Items) == 0 {
		return false, nil
	}
	for _, item := range order.Items {
		product, err := s.productService.GetProductByID(ctx, item.ProductID.Hex())
		if err != nil {
			var appErr *errors.AppError
			if stderrors.As(err, &appErr) && appErr.Type == errors.NotFoundErrorType {
				return false, nil
			}
			return false, err
		}
		err = storeServices.RequireStoreOwner(ctx, s.storeRepo, product.StoreID.Hex(), userID)
		if err != nil {
			var appErr *errors.AppError
			if stderrors.As(err, &appErr) && (appErr.Type == errors.ForbiddenType || appErr.Type == errors.NotFoundErrorType) {
				return false, nil
			}
			return false, err
		}
	}
	return true, nil
}

func (s *OrderService) getOrder(ctx context.Context, id string) (*models.Order, error) {
	if ctx.Err() != nil {
		return nil, errors.Wrap(ctx.Err(), "context cancelled")
	}

	if id == "" {
		return nil, errors.NewBadRequestError("order ID is required")
	}

	order, err := s.repo.GetOrderByID(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, "fetching order")
	}

	if order == nil {
		return nil, errors.NewNotFoundError("order", id)
	}

	return order, nil
}

func (s *OrderService) ListOrders(ctx context.Context, customerID string) ([]models.Order, error) {
	if ctx.Err() != nil {
		return nil, errors.Wrap(ctx.Err(), "context cancelled")
	}

	if customerID == "" {
		return nil, errors.NewBadRequestError("customer ID is required")
	}

	orders, err := s.repo.ListOrders(ctx, customerID)
	if err != nil {
		return nil, errors.Wrap(err, "listing orders")
	}

	return orders, nil
}
