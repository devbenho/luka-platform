package models

import (
	"time"

	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Category struct {
	ID          primitive.ObjectID  `json:"id" bson:"_id"`
	Name        string              `json:"name" bson:"name"`
	Slug        string              `json:"slug" bson:"slug"`
	Description string              `json:"description" bson:"description"`
	ParentID    *primitive.ObjectID `json:"parent_id" bson:"parent_id"`
	CreatedAt   time.Time           `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at" bson:"updated_at"`
	DeletedAt   *time.Time          `json:"delete_at" bson:"delete_at"`
}

func (c *Category) Validate() error {
	validate := validator.New()
	if err := validate.Struct(c); err != nil {
		return err
	}
	return nil
}
