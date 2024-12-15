package dtos

import (
	"github.com/devbenho/luka-platform/internal/user/models"
	"github.com/go-playground/validator/v10"
)

type CreateUserRequest struct {
	Username string `json:"username" validate:"required,min=3,max=20"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	Role     string `json:"role" `
}

type CreateUserResponse struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`
	Role     string `json:"role"`
	Token    string `json:"token"`
}

// MODEL MAPPER
func (c *CreateUserRequest) ToUser() *models.User {
	return &models.User{
		Username: c.Username,
		Email:    c.Email,
		Password: c.Password,
		Role:     c.Role,
	}
}

// VALIDATOR
func (c *CreateUserRequest) Validate() error {
	return validator.New().Struct(c)
}
