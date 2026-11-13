package dtos

import (
	"github.com/devbenho/luka-platform/internal/category/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CreateCategoryRequest struct {
	Name        string              `json:"name" validate:"required,min=3,max=50"`
	Slug        string              `json:"slug" validate:"required,slug"`
	Description string              `json:"description" validate:"max=200"`
	ParentID    *primitive.ObjectID `json:"parent_id" validate:"omitempty"`
}

type CreateCategoryResponse struct {
	Category models.Category `json:"category"`
}

func (r *CreateCategoryRequest) ToCategory() *models.Category {
	return &models.Category{
		Name:        r.Name,
		Slug:        r.Slug,
		Description: r.Description,
		ParentID:    r.ParentID,
	}
}
