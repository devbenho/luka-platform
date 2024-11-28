package models

import (
	"time"

	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Product struct {
	ID          primitive.ObjectID    `json:"id" bson:"_id"`
	Name        string                `json:"name" bson:"name"`
	Description string                `json:"description" bson:"description"`
	Price       float64               `json:"price" bson:"price"`
	StoreID     primitive.ObjectID    `json:"store_id" bson:"store_id"`
	Categories  []*primitive.ObjectID `json:"categories" bson:"categories"`
	Images      []string              `json:"images" bson:"images"`
	Qty         int                   `json:"qty" bson:"qty"`
	CreatedAt   time.Time             `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at" bson:"updated_at"`
	DeletedAt   *time.Time            `json:"deleted_at" bson:"deleted_at"`
}

func (p *Product) Validate() error {
	validator := validator.New()
	if err := validator.Struct(p); err != nil {
		return err
	}
	return nil
}
