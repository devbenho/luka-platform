package models

import (
	"time"

	"github.com/devbenho/luka-platform/pkg/errors"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	Username  string             `bson:"username" validate:"required,min=3,max=20"`
	Email     string             `bson:"email" validate:"required,email"`
	Password  string             `bson:"password" validate:"required,min=6"`
	Role      string             `bson:"role" validate:"required,oneof=buyer seller supplier user"`
	CreatedAt time.Time          `bson:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at"`
	DeletedAt *time.Time         `bson:"deleted_at"`
}

func ValidateUser(user User) error {
	validate := validator.New()
	err := validate.Struct(user)
	if err == nil {
		return nil
	}

	var validationErrors errors.ValidationErrors
	for _, err := range err.(validator.ValidationErrors) {
		validationErrors = append(validationErrors, errors.ValidationError{
			Field: err.Field(),
			Tag:   err.Tag(),
		})
	}

	return validationErrors

}
