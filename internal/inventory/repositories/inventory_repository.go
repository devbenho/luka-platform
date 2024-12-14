package repositories

import (
	"context"
	"time"

	"github.com/devbenho/luka-platform/internal/inventory/models"
	"github.com/devbenho/luka-platform/pkg/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type IInventoryRepository interface {
	CreateInventory(ctx context.Context, inventory *models.Inventory) (*models.Inventory, error)
	GetInventoryByID(ctx context.Context, id string) (*models.Inventory, error)
	UpdateInventory(ctx context.Context, id string, inventory *models.Inventory) error
	DeleteInventory(ctx context.Context, id string) error
}

type InventoryRepository struct {
	db database.IDatabase
}

func NewInventoryRepository(db database.IDatabase) IInventoryRepository {
	return &InventoryRepository{
		db: db,
	}
}

func (r *InventoryRepository) CreateInventory(ctx context.Context, inventory *models.Inventory) (*models.Inventory, error) {
	inventory.ID = primitive.NewObjectID()
	inventory.CreatedAt = time.Now()
	inventory.UpdatedAt = time.Now()
	inventory.Status = "in_stock"
	if err := r.db.Create(ctx, "inventories", inventory); err != nil {
		return nil, err
	}
	return inventory, nil
}

func (r *InventoryRepository) GetInventoryByID(ctx context.Context, id string) (*models.Inventory, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	var inventory models.Inventory
	filter := bson.M{"_id": objID}
	if err := r.db.FindOne(ctx, "inventories", filter, &inventory); err != nil {
		return nil, err
	}
	return &inventory, nil
}

func (r *InventoryRepository) UpdateInventory(ctx context.Context, id string, inventory *models.Inventory) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	inventory.UpdatedAt = time.Now()
	filter := bson.M{"_id": objID}
	update := bson.M{"$set": inventory}
	return r.db.Update(ctx, "inventories", filter, update)
}

func (r *InventoryRepository) DeleteInventory(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	filter := bson.M{"_id": objID}
	return r.db.Delete(ctx, "inventories", filter)
}
