package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Inventory struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ProductID   primitive.ObjectID `bson:"product_id" json:"product_id" validate:"required"`
	WarehouseID primitive.ObjectID `bson:"warehouse_id" json:"warehouse_id" validate:"required"`
	StoreID     primitive.ObjectID `bson:"store_id" json:"store_id" validate:"required"`
	Quantity    int                `bson:"quantity" json:"quantity" validate:"gte=0"`
	Status      string             `bson:"status" json:"status" validate:"required,oneof=in_stock out_of_stock low_stock"`
	MinQuantity int                `bson:"min_quantity" json:"min_quantity" validate:"required,gte=0"`
	MaxQuantity int                `bson:"max_quantity" json:"max_quantity" validate:"required,gtfield=MinQuantity"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
	DeletedAt   *time.Time         `bson:"deleted_at,omitempty" json:"deleted_at,omitempty"`
}

// UpdateQuantity updates the inventory quantity and status
func (i *Inventory) UpdateQuantity(quantity int) {
	i.Quantity = quantity
	i.UpdateStatus()
}

// UpdateStatus updates the inventory status based on quantity thresholds
func (i *Inventory) UpdateStatus() {
	switch {
	case i.Quantity <= 0:
		i.Status = "out_of_stock"
	case i.Quantity <= i.MinQuantity:
		i.Status = "low_stock"
	default:
		i.Status = "in_stock"
	}
}
