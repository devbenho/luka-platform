package dtos

import (
	"github.com/devbenho/luka-platform/internal/user/models"
)

type UserDTO struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}

type UserResponseDTO struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type GetUserResponse struct {
	User UserResponseDTO `json:"user"`
}

type UpdateUserRequest struct {
	Username *string `json:"username" validate:"min=3,max=20"`
	Email    *string `json:"email" validate:"email"`
	Role     *string `json:"role"`
}

func (u *UpdateUserRequest) ToUser() *models.User {
	return &models.User{
		Username: *u.Username,
		Email:    *u.Email,
		Role:     *u.Role,
	}
}
