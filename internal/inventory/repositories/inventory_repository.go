package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/devbenho/luka-platform/internal/inventory/models"
	"github.com/devbenho/luka-platform/pkg/database"
	apperrors "github.com/devbenho/luka-platform/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const collection = "inventories"

// ErrInsufficientStock is returned by ReserveStock when no inventory record
// for the product holds the requested quantity.
var ErrInsufficientStock = errors.New("insufficient stock")

type IInventoryRepository interface {
	CreateInventory(ctx context.Context, inventory *models.Inventory) (*models.Inventory, error)
	// GetInventoryByID returns (nil, nil) when no inventory has the given ID.
	GetInventoryByID(ctx context.Context, id string) (*models.Inventory, error)
	UpdateInventory(ctx context.Context, id string, inventory *models.Inventory) error
	DeleteInventory(ctx context.Context, id string) error
	// ReserveStock atomically takes quantity units of productID from a single
	// inventory record that has enough stock and returns that record's ID.
	ReserveStock(ctx context.Context, productID primitive.ObjectID, quantity int) (primitive.ObjectID, error)
	// ReleaseStock atomically returns quantity units to an inventory record.
	ReleaseStock(ctx context.Context, inventoryID primitive.ObjectID, quantity int) error
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

func parseInventoryID(id string) (primitive.ObjectID, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return primitive.NilObjectID, apperrors.NewBadRequestError("invalid inventory ID")
	}
	return objID, nil
}

func (r *InventoryRepository) CreateInventory(ctx context.Context, inventory *models.Inventory) (*models.Inventory, error) {
	inventory.ID = primitive.NewObjectID()
	inventory.CreatedAt = time.Now()
	inventory.UpdatedAt = time.Now()
	inventory.UpdateStatus()
	if err := r.db.Create(ctx, collection, inventory); err != nil {
		return nil, err
	}
	return inventory, nil
}

func (r *InventoryRepository) GetInventoryByID(ctx context.Context, id string) (*models.Inventory, error) {
	objID, err := parseInventoryID(id)
	if err != nil {
		return nil, err
	}
	var inventory models.Inventory
	filter := bson.M{"_id": objID}
	if err := r.db.FindOne(ctx, collection, filter, &inventory); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &inventory, nil
}

func (r *InventoryRepository) UpdateInventory(ctx context.Context, id string, inventory *models.Inventory) error {
	objID, err := parseInventoryID(id)
	if err != nil {
		return err
	}
	inventory.UpdatedAt = time.Now()
	filter := bson.M{"_id": objID}
	update := bson.M{"$set": inventory}
	return r.db.Update(ctx, collection, filter, update)
}

func (r *InventoryRepository) DeleteInventory(ctx context.Context, id string) error {
	objID, err := parseInventoryID(id)
	if err != nil {
		return err
	}
	filter := bson.M{"_id": objID}
	return r.db.Delete(ctx, collection, filter)
}

// adjustQuantity is an update pipeline that adds delta to quantity and
// recomputes status from the new quantity (same rules as
// Inventory.UpdateStatus), so both change in one atomic write.
func adjustQuantity(delta int) mongo.Pipeline {
	return mongo.Pipeline{
		{{Key: "$set", Value: bson.D{
			{Key: "quantity", Value: bson.D{{Key: "$add", Value: bson.A{"$quantity", delta}}}},
			{Key: "updated_at", Value: time.Now()},
		}}},
		{{Key: "$set", Value: bson.D{
			{Key: "status", Value: bson.D{{Key: "$switch", Value: bson.D{
				{Key: "branches", Value: bson.A{
					bson.D{{Key: "case", Value: bson.D{{Key: "$lte", Value: bson.A{"$quantity", 0}}}}, {Key: "then", Value: "out_of_stock"}},
					bson.D{{Key: "case", Value: bson.D{{Key: "$lte", Value: bson.A{"$quantity", "$min_quantity"}}}}, {Key: "then", Value: "low_stock"}},
				}},
				{Key: "default", Value: "in_stock"},
			}}}},
		}}},
	}
}

func (r *InventoryRepository) ReserveStock(ctx context.Context, productID primitive.ObjectID, quantity int) (primitive.ObjectID, error) {
	ctx, cancel := context.WithTimeout(ctx, database.DatabaseTimeout)
	defer cancel()

	filter := bson.M{
		"product_id": productID,
		"quantity":   bson.M{"$gte": quantity},
		"deleted_at": nil,
	}
	// Take from the record with the most stock.
	opts := options.FindOneAndUpdate().SetSort(bson.D{{Key: "quantity", Value: -1}})
	var reserved models.Inventory
	err := r.db.GetDB().Collection(collection).FindOneAndUpdate(ctx, filter, adjustQuantity(-quantity), opts).Decode(&reserved)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return primitive.NilObjectID, ErrInsufficientStock
	}
	if err != nil {
		return primitive.NilObjectID, err
	}
	return reserved.ID, nil
}

func (r *InventoryRepository) ReleaseStock(ctx context.Context, inventoryID primitive.ObjectID, quantity int) error {
	ctx, cancel := context.WithTimeout(ctx, database.DatabaseTimeout)
	defer cancel()

	result, err := r.db.GetDB().Collection(collection).UpdateOne(ctx, bson.M{"_id": inventoryID}, adjustQuantity(quantity))
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

func (r *InventoryRepository) GetInventoryByWarehouse(ctx context.Context, warehouseID primitive.ObjectID) ([]*models.Inventory, error) {
	var inventories []*models.Inventory
	filter := bson.M{"warehouse_id": warehouseID, "deleted_at": nil}
	err := r.db.Find(ctx, collection, filter, &inventories)
	return inventories, err
}

func (r *InventoryRepository) GetProductInventoryAcrossWarehouses(ctx context.Context, productID primitive.ObjectID) ([]*models.Inventory, error) {
	var inventories []*models.Inventory
	filter := bson.M{"product_id": productID, "deleted_at": nil}
	err := r.db.Find(ctx, collection, filter, &inventories)
	return inventories, err
}

func (r *InventoryRepository) TransferInventory(ctx context.Context, fromWarehouseID, toWarehouseID primitive.ObjectID, productID primitive.ObjectID, quantity int) error {
	return r.db.WithTransaction(ctx, func(sessCtx mongo.SessionContext) error {
		// Deduct from source warehouse. Update returns ErrNoDocuments when the
		// source lacks the stock, which aborts the transaction.
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

		if err := r.db.Update(sessCtx, collection, fromFilter, fromUpdate); err != nil {
			return err
		}

		return r.db.Update(sessCtx, collection, toFilter, toUpdate)
	})
}
