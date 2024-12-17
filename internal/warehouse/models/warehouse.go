package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Warehouse struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name      string             `bson:"name" json:"name" validate:"required"`
	Location  Location           `bson:"location" json:"location" validate:"required"`
	StoreID   primitive.ObjectID `bson:"store_id" json:"store_id" validate:"required"`
	Status    string             `bson:"status" json:"status" validate:"required,oneof=active inactive"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
	DeletedAt *time.Time         `bson:"deleted_at,omitempty" json:"deleted_at,omitempty"`
}

type Location struct {
	Address    string  `bson:"address" json:"address" validate:"required"`
	City       string  `bson:"city" json:"city" validate:"required"`
	Country    string  `bson:"country" json:"country" validate:"required"`
	PostalCode string  `bson:"postal_code" json:"postal_code"`
	Latitude   float64 `bson:"latitude" json:"latitude" validate:"required"`
	Longitude  float64 `bson:"longitude" json:"longitude" validate:"required"`
}
