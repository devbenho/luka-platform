package dtos

import (
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CreateOrderRequest struct {
	CustomerID      primitive.ObjectID       `json:"customerID" validate:"required"`
	Items           []CreateOrderItemRequest `json:"items" validate:"required,min=1,dive"`
	ShippingAddress string                   `json:"shippingAddress" validate:"required"`
	Notes           string                   `json:"notes"`
}

type CreateOrderItemRequest struct {
	ProductID primitive.ObjectID `json:"productID" validate:"required"`
	Quantity  int                `json:"quantity" validate:"required,gt=0"`
}

func (r *CreateOrderRequest) Validate() error {
	validator := validator.New()
	if err := validator.Struct(r); err != nil {
		return err
	}
	return nil
}
