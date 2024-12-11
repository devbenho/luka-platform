package dtos

import (
	"github.com/devbenho/luka-platform/internal/inventory/models"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CreateInventoryRequest struct {
	ProductId primitive.ObjectID `json:"product_id" validate:"required"`
	Quantity  int                `json:"quantity" validate:"required,gte=0"`
	Status    string             `json:"status" validate:"required,oneof=in_stock out_of_stock low_stock"`
}

func (r *CreateInventoryRequest) ToInventory() *models.Inventory {
	return &models.Inventory{
		ProductId: r.ProductId,
		Quantity:  r.Quantity,
		Status:    r.Status,
	}
}

func (r *CreateInventoryRequest) Validate() error {
	validator := validator.New()
	if err := validator.Struct(r); err != nil {
		return err
	}
	return nil
}
