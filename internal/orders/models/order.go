package models

import (
	"time"

	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type OrderStatus string

const (
	OrderStatusPending    OrderStatus = "PENDING"
	OrderStatusProcessing OrderStatus = "PROCESSING"
	OrderStatusConfirmed  OrderStatus = "confirmed"
	OrderStatusShipped    OrderStatus = "SHIPPED"
	OrderStatusDelivered  OrderStatus = "DELIVERED"
	OrderStatusCancelled  OrderStatus = "CANCELLED"
)

type OrderItem struct {
	ProductID  primitive.ObjectID `bson:"productID" json:"product_id"`
	Quantity   int                `bson:"quantity" json:"quantity"`
	UnitPrice  float64            `bson:"unitPrice" json:"unit_price"`
	TotalPrice float64            `bson:"totalPrice" json:"total_price"`
}

type Order struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	CustomerID      primitive.ObjectID `bson:"customerID" json:"customer_id"`
	Items           []OrderItem        `bson:"items" json:"items"`
	Status          OrderStatus        `bson:"status" json:"status"`
	TotalAmount     float64            `bson:"totalAmount" json:"total_amount"`
	ShippingAddress string             `bson:"shippingAddress" json:"shipping_address"`
	Notes           string             `bson:"notes,omitempty" json:"notes,omitempty"`
	CreatedAt       time.Time          `bson:"createdAt" json:"created_at"`
	UpdatedAt       time.Time          `bson:"updatedAt" json:"updated_at"`
	DeletedAt       *time.Time         `bson:"deleted_at,omitempty" json:"deleted_at,omitempty"`
}

func (o *Order) Validate() error {
	validator := validator.New()
	if err := validator.Struct(o); err != nil {
		return err
	}
	return nil
}
