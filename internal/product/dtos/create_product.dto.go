package dtos

import (
	"github.com/devbenho/luka-platform/internal/product/models"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CreateProductRequest struct {
	Name        string                `json:"name" binding:"required"`
	Description string                `json:"description" binding:"required"`
	Price       float64               `json:"price" binding:"required"`
	StoreID     primitive.ObjectID    `json:"store_id" binding:"required"`
	Categories  []*primitive.ObjectID `json:"categories" binding:"required"`
	Images      []string              `json:"images" bson:"images" binding:"required"`
	Qty         int                   `json:"qty" bson:"qty" binding:"required"`
}

type CreateProductResponse struct {
	Product *models.Product `json:"product"`
}

func (r *CreateProductRequest) ToProduct() *models.Product {
	return &models.Product{
		Name:        r.Name,
		Description: r.Description,
		Price:       r.Price,
		StoreID:     r.StoreID,
		Categories:  r.Categories,
		Images:      r.Images,
		Qty:         r.Qty,
	}
}

func (r *CreateProductRequest) Validate() error {
	validator := validator.New()
	if err := validator.Struct(r); err != nil {
		return err
	}
	return nil
}
