package dtos

import (
	"github.com/go-playground/validator/v10"
)

type LoginRequest struct {
	Login    string `json:"login" validate:"required"` // Can be either username or email
	Password string `json:"password" validate:"required,min=6"`
}

type AuthResponse struct {
	Email string `json:"email"`
	Token string `json:"token"`
}

func (a *LoginRequest) Validate() error {
	return validator.New().Struct(a)
}
