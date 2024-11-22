package dtos

import (
	"github.com/devbenho/luka-platform/internal/category/models"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UpdateCategoryRequest struct {
	Name        string `json:"name" validate:"omitempty,min=3,max=50"`
	Slug        string `json:"slug" validate:"omitempty,slug"`
	Description string `json:"description" validate:"omitempty,max=200"`
	ParentID    string `json:"parent_id" validate:"omitempty"`
}

type UpdateCategoryResponse struct {
	Category models.Category `json:"category"`
}

func (r *UpdateCategoryRequest) ToCategory() (models.Category, error) {
	parentID, err := primitive.ObjectIDFromHex(r.ParentID)
	if err != nil && r.ParentID != "" {
		return models.Category{}, err
	}

	return models.Category{
		Name:        r.Name,
		Slug:        r.Slug,
		Description: r.Description,
		ParentID:    &parentID,
	}, nil
}

func (r *UpdateCategoryRequest) Validate() error {
	validate := validator.New()

	validate.RegisterValidation("slug", func(fl validator.FieldLevel) bool {
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
