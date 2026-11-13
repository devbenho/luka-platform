package dtos

import (
	"github.com/devbenho/luka-platform/internal/store/models"
	"github.com/devbenho/luka-platform/pkg/slug"
)

type UpdateStoreRequest struct {
	Name *string `json:"name" validate:"required,min=1"`
}

// ApplyTo renames the store and regenerates its slug.
func (u *UpdateStoreRequest) ApplyTo(store *models.Store) {
	store.Name = *u.Name
	store.Slug = slug.GenerateSlug(*u.Name)
}
