package dtos

import (
	"github.com/devbenho/luka-platform/internal/store/models"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CreateStoreRequest struct {
	Name             string             `json:"name" validate:"required"`
	Slug             string             `json:"slug" validate:"required"`
	OwnerId          primitive.ObjectID `json:"ownerId" validate:"required"`
	Location         models.Location    `json:"location"`
	StoreType        models.StoreType   `json:"store_type"`
	SocialMediaLinks map[string]string  `json:"social_media_links"`
}

type CreateStoreResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

func (c *CreateStoreRequest) Validate() error {
	validator := validator.New()
	c.OwnerId = primitive.NewObjectID()
	if err := validator.Struct(c); err != nil {
		return err
	}
	return nil
}

func (c *CreateStoreRequest) ToStore() *models.Store {
	return &models.Store{
		Name:             c.Name,
		Slug:             c.Slug,
		OwnerId:          c.OwnerId,
		Location:         c.Location,
		SocialMediaLinks: c.SocialMediaLinks,
	}
}
