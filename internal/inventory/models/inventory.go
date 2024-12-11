package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Inventory struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ProductId primitive.ObjectID `bson:"product_id" json:"product_id" validate:"required"`
	Quantity  int                `bson:"quantity" json:"quantity" validate:"gte=0"`
	Status    string             `bson:"status" json:"status" validate:"required,oneof=in_stock out_of_stock low_stock"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
	DeletedAt *time.Time         `bson:"deleted_at,omitempty" json:"deleted_at,omitempty"`
}
