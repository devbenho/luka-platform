package repositories

import (
	"context"
	"time"

	"github.com/devbenho/luka-platform/internal/inventory/models"
	"github.com/devbenho/luka-platform/pkg/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type IInventoryRepository interface {
	CreateInventory(ctx context.Context, inventory *models.Inventory) (*models.Inventory, error)
	GetInventoryByID(ctx context.Context, id string) (*models.Inventory, error)
	UpdateInventory(ctx context.Context, id string, inventory *models.Inventory) error
	DeleteInventory(ctx context.Context, id string) error
	GetInventoryByWarehouse(ctx context.Context, warehouseID primitive.ObjectID) ([]*models.Inventory, error)
	GetProductInventoryAcrossWarehouses(ctx context.Context, productID primitive.ObjectID) ([]*models.Inventory, error)
	TransferInventory(ctx context.Context, fromWarehouseID, toWarehouseID primitive.ObjectID, productID primitive.ObjectID, quantity int) error
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

func (r *InventoryRepository) GetInventoryByWarehouse(ctx context.Context, warehouseID primitive.ObjectID) ([]*models.Inventory, error) {
	var inventories []*models.Inventory
	filter := bson.M{"warehouse_id": warehouseID, "deleted_at": nil}
	err := r.db.Find(ctx, "inventories", filter, &inventories)
	return inventories, err
}

func (r *InventoryRepository) GetProductInventoryAcrossWarehouses(ctx context.Context, productID primitive.ObjectID) ([]*models.Inventory, error) {
	var inventories []*models.Inventory
	filter := bson.M{"product_id": productID, "deleted_at": nil}
	err := r.db.Find(ctx, "inventories", filter, &inventories)
	return inventories, err
}

func (r *InventoryRepository) TransferInventory(ctx context.Context, fromWarehouseID, toWarehouseID primitive.ObjectID, productID primitive.ObjectID, quantity int) error {
	return r.db.WithTransaction(ctx, func(sessCtx mongo.SessionContext) error {
		// Deduct from source warehouse
		fromFilter := bson.M{
			"warehouse_id": fromWarehouseID,
			"product_id":   productID,
			"quantity":     bson.M{"$gte": quantity},
		}
		fromUpdate := bson.M{"$inc": bson.M{"quantity": -quantity}}

		// Add to destination warehouse
		toFilter := bson.M{
			"warehouse_id": toWarehouseID,
			"product_id":   productID,
		}
		toUpdate := bson.M{"$inc": bson.M{"quantity": quantity}}

		if err := r.db.Update(sessCtx, "inventories", fromFilter, fromUpdate); err != nil {
			return err
		}

		return r.db.Update(sessCtx, "inventories", toFilter, toUpdate)
	})
}
