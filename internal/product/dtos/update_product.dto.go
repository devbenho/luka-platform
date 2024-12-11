package dtos

import (
	"github.com/devbenho/luka-platform/internal/product/models"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UpdateProductRequest struct {
	Name        *string               `json:"name"`
	Description *string               `json:"description"`
	Price       *float64              `json:"price"`
	StoreID     *primitive.ObjectID   `json:"store_id"`
	Categories  []*primitive.ObjectID `json:"categories"`
	Images      *[]string             `json:"images"`
}

type UpdateProductResponse struct {
	Product *models.Product `json:"product"`
}

func (r *UpdateProductRequest) Validate() error {
	validator := validator.New()
	if err := validator.Struct(r); err != nil {
		return err
	}
	return nil
}
