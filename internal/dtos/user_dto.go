package dtos

import "github.com/devbenho/luka-platform/internal/models"

type UserDTO struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Role      string `json:"role"`
}

type UserResponseDTO struct {
	ID        string `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type CreateUserRequest struct {
	Username string `json:"username" validate:"required,min=3,max=20"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	Role     string `json:"role" validate:"required,oneof=buyer seller supplier"`

	// generate function to convert CreateUserRequest to User
}

func (c *CreateUserRequest) ToUser() *models.User {
	return &models.User{
		Username: c.Username,
		Email:    c.Email,
		Password: c.Password,
		Role:     c.Role,
	}
}

type CreateUserResponse struct {
	ID    string       `json:"id"`
	Token models.Token `json:"token"`
}

type GetUserResponse struct {
	User UserResponseDTO `json:"user"`
}

type UpdateUserRequest struct {
	Username string `json:"username" validate:"min=3,max=20"`
	Email    string `json:"email" validate:"email"`
	Role     string `json:"role" validate:"oneof=buyer seller supplier"`
}

func (u *UpdateUserRequest) ToUser() *models.User {
	return &models.User{
		Username: u.Username,
		Email:    u.Email,
		Role:     u.Role,
	}
}
