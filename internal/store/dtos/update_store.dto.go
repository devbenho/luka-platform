package dtos

import (
	"github.com/devbenho/luka-platform/internal/store/models"
	"github.com/go-playground/validator/v10"
)

type UpdateStoreRequest struct {
	Name string `json:"name" validate:"required"`
}

func (u *UpdateStoreRequest) ToStore() *models.Store {
	return &models.Store{
		Name: u.Name,
	}
}

func (u *UpdateStoreRequest) Validate() error {
	validator := validator.New()
	if err := validator.Struct(u); err != nil {
		return err
	}
	return nil
}
