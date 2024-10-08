package models

import (
	"github.com/devbenho/bazar-user-service/pkg/errors"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	Username  string             `bson:"username" validate:"required,min=3,max=20"`
	Email     string             `bson:"email" validate:"required,email"`
	Password  string             `bson:"password" validate:"required,min=6"`
	Role      string             `bson:"role" validate:"required,oneof=buyer seller supplier"`
	CreatedAt time.Time          `bson:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at"`
}

func ValidateUser(user User) error {
	validate := validator.New()
	err := validate.Struct(user)
	if err == nil {
		return nil // No validation errors
	}

	var validationErrors errors.ValidationErrors
	for _, err := range err.(validator.ValidationErrors) {
		validationErrors = append(validationErrors, errors.NewValidationError(err.Field(), err.Tag(), err.Value().(string)))
	}

	return validationErrors

}
