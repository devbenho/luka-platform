package services

import (
	"context"
	stderrors "errors"
	"fmt"
	"log"
	"time"

	"github.com/devbenho/luka-platform/internal/inventory/services"
	"github.com/devbenho/luka-platform/internal/orders/models"
	dtos "github.com/devbenho/luka-platform/internal/orders/order_dtos"
	"github.com/devbenho/luka-platform/internal/orders/repositories"
	productService "github.com/devbenho/luka-platform/internal/product/services"
	"github.com/devbenho/luka-platform/pkg/errors"
	"github.com/devbenho/luka-platform/pkg/validation"
	"go.mongodb.org/mongo-driver/mongo"
)

type IOrderService interface {
	CreateOrder(ctx context.Context, dto dtos.CreateOrderRequest) (*models.Order, error)
	UpdateOrderStatus(ctx context.Context, id string, status models.OrderStatus) error
	GetOrderByID(ctx context.Context, id string) (*models.Order, error)
	ListOrders(ctx context.Context, customerID string) ([]models.Order, error)
}

type OrderService struct {
	repo             repositories.IOrderRepository
	inventoryService services.IInventoryService
	productService   productService.IProductService
	validator        *validation.Validator
}

func NewOrderService(
	repo repositories.IOrderRepository,
	inventoryService services.IInventoryService,
	productService productService.IProductService,
	validator *validation.Validator,
) *OrderService {
	return &OrderService{
		repo:             repo,
		inventoryService: inventoryService,
		productService:   productService,
		validator:        validator,
	}
}

func (s *OrderService) CreateOrder(ctx context.Context, dto dtos.CreateOrderRequest) (*models.Order, error) {
	if err := s.validator.ValidateStruct(dto); err != nil {
		return nil, err
	}

	orderItems, totalAmount, err := s.prepareOrderItems(ctx, dto.Items)
	if err != nil {
		return nil, err
	}

	if err := s.reserveInventory(ctx, orderItems); err != nil {
		return nil, err
	}

	order := s.buildOrder(dto, orderItems, totalAmount)
	createdOrder, err := s.repo.CreateOrder(ctx, order)
	if err != nil {
		message := fmt.Sprintf("creating order: %v", err)
		cause := err
		if rollbackErr := s.rollbackInventory(ctx, orderItems); rollbackErr != nil {
			// The HTTP layer returns only the message, so record which items
			// were left reserved for operators.
			log.Printf("order creation failed for customer %s and inventory rollback was incomplete: %v", dto.CustomerID.Hex(), rollbackErr)
			message += "; inventory rollback failed"
			cause = stderrors.Join(err, rollbackErr)
		}
		return nil, errors.NewError(
			errors.InternalServerType,
			500,
			message,
			errors.WithCause(cause),
			errors.WithMetadata(map[string]interface{}{
				"customer_id": dto.CustomerID,
				"items_count": len(dto.Items),
			}),
		)
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
// inventory record it came from. If any item cannot be reserved, the items
// already reserved are released before the error is returned.
func (s *OrderService) reserveInventory(ctx context.Context, items []models.OrderItem) error {
	for i := range items {
		inventoryID, err := s.inventoryService.ReserveStock(ctx, items[i].ProductID, items[i].Quantity)
		if err != nil {
			if rollbackErr := s.rollbackInventory(ctx, items[:i]); rollbackErr != nil {
				log.Printf("reserving stock failed and releasing earlier items was incomplete: %v", rollbackErr)
				return errors.NewInternalError("reserving inventory", stderrors.Join(err, rollbackErr))
			}
			return err
		}
		items[i].InventoryID = inventoryID
	}
	return nil
}

// rollbackInventory releases the stock reserved for items. It attempts every
// item and returns all failures joined, so a partial rollback is never
// silently lost.
func (s *OrderService) rollbackInventory(ctx context.Context, items []models.OrderItem) error {
	var errs []error
	for _, item := range items {
		if item.InventoryID.IsZero() {
			continue
		}
		if err := s.inventoryService.ReleaseStock(ctx, item.InventoryID, item.Quantity); err != nil {
			errs = append(errs, fmt.Errorf("releasing stock for product %s (inventory %s, quantity %d): %w",
				item.ProductID.Hex(), item.InventoryID.Hex(), item.Quantity, err))
		}
	}
	return stderrors.Join(errs...)
}

func (s *OrderService) UpdateOrderStatus(ctx context.Context, id string, status models.OrderStatus) error {
	if !isValidStatus(status) {
		return errors.NewBadRequestError("invalid order status")
	}

	order, err := s.GetOrderByID(ctx, id)
	if err != nil {
		return err
	}

	if !isValidStatusTransition(order.Status, status) {
		return errors.NewBadRequestError("invalid status transition")
	}

	// Only succeeds if nobody changed the status since we read it, so
	// concurrent requests cannot both cancel (and release stock twice).
	err = s.repo.UpdateOrderStatus(ctx, id, order.Status, status)
	if stderrors.Is(err, mongo.ErrNoDocuments) {
		return errors.NewConflictError("order status changed concurrently; retry")
	}
	if err != nil {
		return errors.Wrap(err, "updating order status")
	}

	if status == models.OrderStatusCancelled {
		if err := s.rollbackInventory(ctx, order.Items); err != nil {
			// The order is cancelled either way; surface the stock leak to
			// operators rather than failing a request that cannot be retried.
			log.Printf("order %s cancelled but releasing its stock was incomplete: %v", id, err)
		}
	}

	return nil
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

func (s *OrderService) GetOrderByID(ctx context.Context, id string) (*models.Order, error) {
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
