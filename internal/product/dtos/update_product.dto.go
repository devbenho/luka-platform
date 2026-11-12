package dtos

import (
	"github.com/devbenho/luka-platform/internal/product/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UpdateProductRequest is a partial update: nil fields are left unchanged.
// A product cannot be moved to another store.
type UpdateProductRequest struct {
	Name        *string               `json:"name" validate:"omitempty,min=1"`
	Description *string               `json:"description"`
	Price       *float64              `json:"price" validate:"omitempty,gt=0"`
	Categories  []*primitive.ObjectID `json:"categories"`
	Images      *[]string             `json:"images"`
}

type UpdateProductResponse struct {
	Product *models.Product `json:"product"`
}

// ApplyTo copies the fields that were provided onto product.
func (r *UpdateProductRequest) ApplyTo(product *models.Product) {
	if r.Name != nil {
		product.Name = *r.Name
	}
	if r.Description != nil {
		product.Description = *r.Description
	}
	if r.Price != nil {
		product.Price = *r.Price
	}
	if r.Categories != nil {
		product.Categories = r.Categories
	}
	if r.Images != nil {
		product.Images = *r.Images
	}
}
