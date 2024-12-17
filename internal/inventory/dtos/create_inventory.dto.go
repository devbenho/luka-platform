package dtos

import (
	"github.com/devbenho/luka-platform/internal/inventory/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CreateInventoryRequest struct {
	ProductID   primitive.ObjectID `json:"product_id" validate:"required"`
	WarehouseID primitive.ObjectID `json:"warehouse_id" validate:"required"`
	StoreID     primitive.ObjectID `json:"store_id" validate:"required"`
	Quantity    int                `json:"quantity" validate:"required,gte=0"`
	MinQuantity int                `json:"min_quantity" validate:"required,gte=0"`
	MaxQuantity int                `json:"max_quantity" validate:"required,gtfield=MinQuantity"`
}

func (r *CreateInventoryRequest) ToInventory() *models.Inventory {
	inventory := &models.Inventory{
		ProductID:   r.ProductID,
		WarehouseID: r.WarehouseID,
		StoreID:     r.StoreID,
		Quantity:    r.Quantity,
		MinQuantity: r.MinQuantity,
		MaxQuantity: r.MaxQuantity,
	}
	inventory.UpdateStatus()
	return inventory
}
