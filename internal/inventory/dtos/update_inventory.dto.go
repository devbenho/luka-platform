package dtos

import (
	"github.com/devbenho/luka-platform/internal/inventory/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UpdateInventoryRequest struct {
	ProductId *primitive.ObjectID `json:"product_id"`
	Quantity  *int                `json:"quantity,omitempty" validate:"omitempty,gte=0"`
	Status    *string             `json:"status,omitempty" validate:"omitempty,oneof=in_stock out_of_stock low_stock"`
}

func (dto *UpdateInventoryRequest) ToInventory() *models.Inventory {
	return &models.Inventory{
		ProductID: *dto.ProductId,
	}
}
