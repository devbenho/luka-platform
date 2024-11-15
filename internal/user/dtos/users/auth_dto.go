package dtos

type AuthDTO struct {
	Login    string `json:"login" validate:"required"` // Can be either username or email
	Password string `json:"password" validate:"required,min=6"`
}

type AuthResponseDTO struct {
	Email string `json:"email"`
	Token string `json:"token"`
}
