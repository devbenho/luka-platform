package services

import (
	"context"
	"fmt"
	"log"
	"time"

	inventoryDTOs "github.com/devbenho/luka-platform/internal/inventory/dtos"
	"github.com/devbenho/luka-platform/internal/inventory/services"
	"github.com/devbenho/luka-platform/internal/orders/models"
	dtos "github.com/devbenho/luka-platform/internal/orders/order_dtos"
	"github.com/devbenho/luka-platform/internal/orders/repositories"
	productService "github.com/devbenho/luka-platform/internal/product/services"
	"github.com/devbenho/luka-platform/internal/utils"
	"github.com/devbenho/luka-platform/pkg/errors"
	"github.com/devbenho/luka-platform/pkg/validation"
	"github.com/go-playground/validator/v10"
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
	if err := s.validator.ValidateStruct(&dto); err != nil {
		return nil, errors.Wrap(err, "validating order request")
	}

	orderItems, totalAmount, err := s.prepareOrderItems(ctx, dto.Items)
	if err != nil {
		return nil, errors.Wrap(err, "preparing order items")
	}

	order := s.buildOrder(dto, orderItems, totalAmount)

	if err := s.reserveInventory(ctx, orderItems); err != nil {
		return nil, errors.Wrap(err, "reserving inventory")
	}

	createdOrder, err := s.repo.CreateOrder(ctx, order)
	if err != nil {
		s.rollbackInventory(ctx, orderItems)
		// Use NewError when we need to add metadata
		return nil, errors.NewError(
			errors.InternalServerType,
			500,
			fmt.Sprintf("creating order: %v", err),
			errors.WithMetadata(map[string]interface{}{
				"customer_id": dto.CustomerID,
				"items_count": len(dto.Items),
			}),
		)
	}

	return createdOrder, nil
}

func (s *OrderService) prepareOrderItems(ctx context.Context, items []dtos.CreateOrderItemRequest) ([]models.OrderItem, float64, error) {
	var totalAmount float64
	orderItems := make([]models.OrderItem, len(items))

	for i, item := range items {
		product, err := s.productService.GetProductByID(ctx, item.ProductID.Hex())
		if err != nil {
			return nil, 0, errors.NewNotFoundError("product", item.ProductID.Hex())
		}

		inventory, err := s.inventoryService.GetInventoryByID(ctx, item.ProductID.Hex())
		if err != nil {
			return nil, 0, errors.NewNotFoundError("inventory", item.ProductID.Hex())
		}

		if inventory.Quantity < item.Quantity {
			return nil, 0, errors.Wrapf(err, "product_id: %s", item.ProductID.Hex())
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

func (s *OrderService) reserveInventory(ctx context.Context, items []models.OrderItem) error {
	for _, item := range items {
		updateReq := inventoryDTOs.UpdateInventoryRequest{
			Quantity: utils.IntPtr(-item.Quantity),
		}
		if _, err := s.inventoryService.UpdateInventory(ctx, item.ProductID.Hex(), updateReq); err != nil {
			return errors.Wrap(err, "updating inventory quantity")
		}
	}
	return nil
}

func (s *OrderService) rollbackInventory(ctx context.Context, items []models.OrderItem) {
	for _, item := range items {
		updateReq := inventoryDTOs.UpdateInventoryRequest{
			Quantity: utils.IntPtr(item.Quantity),
		}
		if _, err := s.inventoryService.UpdateInventory(ctx, item.ProductID.String(), updateReq); err != nil {
			log.Printf("failed to rollback inventory: %v", err)
		}
	}
}

func (s *OrderService) UpdateOrderStatus(ctx context.Context, id string, status models.OrderStatus) error {
	if !isValidStatus(status) {
		return errors.NewBadRequestError("invalid order status")
	}

	order, err := s.GetOrderByID(ctx, id)
	if err != nil {
		return errors.Wrap(err, "fetching order")
	}

	if !isValidStatusTransition(order.Status, status) {
		return errors.NewBadRequestError("invalid status transition")
	}

	order.Status = status
	order.UpdatedAt = time.Now()

	if err := s.repo.UpdateOrder(ctx, id, order); err != nil {
		return errors.Wrap(err, "updating order status")
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
		// Remove the extra argument
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
		// Use NewError when we need to add metadata
		return nil, errors.NewError(
			errors.InternalServerType,
			500,
			fmt.Sprintf("listing orders: %v", err),
			errors.WithMetadata(map[string]interface{}{
				"customer_id": customerID,
			}),
		)
	}

	return orders, nil
}

func convertValidationErrors(validationErrors validator.ValidationErrors) errors.ValidationErrors {
	var customErrors errors.ValidationErrors
	for _, e := range validationErrors {
		newError := errors.NewValidationError(e.Field(), e.Tag())
		customErrors = append(customErrors, newError)
	}

	return customErrors
}
