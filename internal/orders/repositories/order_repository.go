package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/devbenho/luka-platform/internal/orders/models"
	"github.com/devbenho/luka-platform/pkg/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type IOrderRepository interface {
	CreateOrder(ctx context.Context, order *models.Order) (*models.Order, error)
	UpdateOrder(ctx context.Context, id string, order *models.Order) error
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
		return fmt.Errorf("invalid order ID: %w", err)
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

func (r *OrderRepository) GetOrderByID(ctx context.Context, id string) (*models.Order, error) {
	var order models.Order
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid order ID: %w", err)
	}
	filter := bson.M{"_id": objID}
	if err := r.db.FindOne(ctx, "orders", filter, &order); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("order not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get order from db: %w", err)
	}
	return &order, nil
}

func (r *OrderRepository) ListOrders(ctx context.Context, customerID string) ([]models.Order, error) {
	var orders []models.Order
	custID, err := primitive.ObjectIDFromHex(customerID)
	if err != nil {
		return nil, err
	}
	filter := bson.M{"customer_id": custID}
	if err := r.db.Find(ctx, "orders", filter, &orders); err != nil {
		return nil, err
	}
	return orders, nil
}
