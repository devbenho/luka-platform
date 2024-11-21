package models

import (
	"reflect"
	"time"

	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Store struct {
	ID               primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name             string             `json:"name" validate:"required"`
	Slug             string             `json:"slug" validate:"required"`
	OwnerId          primitive.ObjectID `json:"ownerId" validate:"required"`
	Location         Location           `json:"location" validate:"required"`
	Type             StoreType          `json:"type" validate:"required"`
	SocialMediaLinks map[string]string  `json:"social_media_links"`
	CreatedAt        time.Time          `json:"created_at"`
	UpdatedAt        time.Time          `json:"updated_at"`
	DeletedAt        *time.Time         `json:"deleted_at" default:"null"`
}

func (s *Store) Validate() error {
	validator := validator.New()
	if err := validator.Struct(s); err != nil {
		return err
	}
	return nil
}

func (s *Store) SetDefaults() {
	s.CreatedAt = time.Now()
	s.UpdatedAt = time.Now()

	if reflect.DeepEqual(s.Type, StoreType{}) {
		s.Type = StoreType{
			Online:  true,
			Offline: false,
		}
	}
	s.Slug = "store-" + s.ID.Hex()
}
