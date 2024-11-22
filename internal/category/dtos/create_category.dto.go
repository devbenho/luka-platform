package dtos

import (
	"github.com/devbenho/luka-platform/internal/category/models"
	"github.com/go-playground/validator/v10"
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

func (r *CreateCategoryRequest) Validate() error {
	validate := validator.New()

	// Custom validation for slug format
	validate.RegisterValidation("slug", func(fl validator.FieldLevel) bool {
		// Add your slug validation logic here
		// For example, check if it only contains alphanumeric characters and hyphens
		slug := fl.Field().String()
		for _, char := range slug {
			if !(char >= 'a' && char <= 'z' || char >= '0' && char <= '9' || char == '-') {
				return false
			}
		}
		return true
	})

	if err := validate.Struct(r); err != nil {
		return err
	}
	return nil
}
