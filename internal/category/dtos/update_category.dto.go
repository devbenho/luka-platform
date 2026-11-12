package dtos

import (
	"github.com/devbenho/luka-platform/internal/category/models"
	"github.com/devbenho/luka-platform/pkg/errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UpdateCategoryRequest is a partial update: nil fields are left unchanged.
type UpdateCategoryRequest struct {
	Name        *string `json:"name" validate:"omitempty,min=3,max=50"`
	Slug        *string `json:"slug" validate:"omitempty,slug"`
	Description *string `json:"description" validate:"omitempty,max=200"`
	ParentID    *string `json:"parent_id" validate:"omitempty"`
}

type UpdateCategoryResponse struct {
	Category models.Category `json:"category"`
}

// ApplyTo copies the provided fields onto category. It returns a bad-request
// error if parent_id is not a valid ID or points at the category itself.
func (r *UpdateCategoryRequest) ApplyTo(category *models.Category) error {
	if r.ParentID != nil {
		parentID, err := primitive.ObjectIDFromHex(*r.ParentID)
		if err != nil {
			return errors.NewBadRequestError("invalid parent_id")
		}
		if parentID == category.ID {
			return errors.NewBadRequestError("a category cannot be its own parent")
		}
		category.ParentID = &parentID
	}
	if r.Name != nil {
		category.Name = *r.Name
	}
	if r.Slug != nil {
		category.Slug = *r.Slug
	}
	if r.Description != nil {
		category.Description = *r.Description
	}
	return nil
}
