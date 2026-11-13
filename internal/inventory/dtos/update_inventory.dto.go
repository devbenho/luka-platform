package dtos

import (
	"github.com/devbenho/luka-platform/internal/inventory/models"
	"github.com/devbenho/luka-platform/pkg/errors"
)

// UpdateInventoryRequest is a partial update: nil fields are left unchanged.
// Status is always derived from the quantity and cannot be set directly, and
// an inventory record cannot be moved to another product.
type UpdateInventoryRequest struct {
	Quantity    *int `json:"quantity,omitempty" validate:"omitempty,gte=0"`
	MinQuantity *int `json:"min_quantity,omitempty" validate:"omitempty,gte=0"`
	MaxQuantity *int `json:"max_quantity,omitempty" validate:"omitempty,gte=0"`
}

// ApplyTo copies the provided fields onto inventory and recomputes its status.
func (dto *UpdateInventoryRequest) ApplyTo(inventory *models.Inventory) error {
	if dto.Quantity != nil {
		inventory.Quantity = *dto.Quantity
	}
	if dto.MinQuantity != nil {
		inventory.MinQuantity = *dto.MinQuantity
	}
	if dto.MaxQuantity != nil {
		inventory.MaxQuantity = *dto.MaxQuantity
	}
	if inventory.MaxQuantity <= inventory.MinQuantity {
		return errors.NewBadRequestError("max_quantity must be greater than min_quantity")
	}
	inventory.UpdateStatus()
	return nil
}
