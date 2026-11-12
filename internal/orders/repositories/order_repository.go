package repositories

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/devbenho/luka-platform/internal/orders/models"
	"github.com/devbenho/luka-platform/pkg/database"
	apperrors "github.com/devbenho/luka-platform/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type IOrderRepository interface {
	CreateOrder(ctx context.Context, order *models.Order) (*models.Order, error)
	UpdateOrder(ctx context.Context, id string, order *models.Order) error
	// UpdateOrderStatus moves the order from status `from` to `to` only if it
	// is still in `from`; otherwise it returns mongo.ErrNoDocuments.
	UpdateOrderStatus(ctx context.Context, id string, from, to models.OrderStatus) error
	// GetOrderByID returns (nil, nil) when no order has the given ID.
	GetOrderByID(ctx context.Context, id string) (*models.Order, error)
	ListOrders(ctx context.Context, customerID string) ([]models.Order, error)
}

type OrderRepository struct {
	db database.IDatabase
}

func NewOrderRepository(db database.IDatabase) IOrderRepository {
	return &OrderRepository{
		db: db,
	}
}

func (r *OrderRepository) CreateOrder(ctx context.Context, order *models.Order) (*models.Order, error) {
	order.ID = primitive.NewObjectID()
	order.CreatedAt = time.Now()
	order.UpdatedAt = time.Now()
	err := r.db.Create(ctx, "orders", order)
	if err != nil {
		return nil, fmt.Errorf("failed to create order in db: %w", err)
	}
	return order, nil
}

func (r *OrderRepository) UpdateOrder(ctx context.Context, id string, order *models.Order) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return apperrors.NewBadRequestError("invalid order ID")
	}
	order.UpdatedAt = time.Now()
	filter := bson.M{"_id": objID}
	update := bson.M{"$set": order}
	err = r.db.Update(ctx, "orders", filter, update)
	if err != nil {
		return fmt.Errorf("failed to update order in db: %w", err)
	}
	return nil
}

func (r *OrderRepository) UpdateOrderStatus(ctx context.Context, id string, from, to models.OrderStatus) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return apperrors.NewBadRequestError("invalid order ID")
	}
	filter := bson.M{"_id": objID, "status": from}
	update := bson.M{"$set": bson.M{"status": to, "updatedAt": time.Now()}}
	return r.db.Update(ctx, "orders", filter, update)
}

func (r *OrderRepository) GetOrderByID(ctx context.Context, id string) (*models.Order, error) {
	var order models.Order
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, apperrors.NewBadRequestError("invalid order ID")
	}
	filter := bson.M{"_id": objID}
	if err := r.db.FindOne(ctx, "orders", filter, &order); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get order from db: %w", err)
	}
	return &order, nil
}

func (r *OrderRepository) ListOrders(ctx context.Context, customerID string) ([]models.Order, error) {
	var orders []models.Order
	custID, err := primitive.ObjectIDFromHex(customerID)
	if err != nil {
		return nil, apperrors.NewBadRequestError("invalid customer ID")
	}
	// Must match the bson tag on Order.CustomerID.
	filter := bson.M{"customerID": custID}
	if err := r.db.Find(ctx, "orders", filter, &orders); err != nil {
		return nil, err
	}
	if orders == nil {
		orders = []models.Order{}
	}
	return orders, nil
}
