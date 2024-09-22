package services

import (
	"context"
	"github.com/devbenho/bazar-user-service/internal/dtos"
	"github.com/devbenho/bazar-user-service/internal/repositories"
	"github.com/devbenho/bazar-user-service/pkg/hasher"
	"github.com/devbenho/bazar-user-service/pkg/tokens"
	"github.com/devbenho/bazar-user-service/pkg/validation"
)

type IUserService interface {
	Register(ctx context.Context, dto *dtos.CreateUserRequest) (*dtos.CreateUserResponse, error)
	Login(ctx context.Context, dto *dtos.AuthDTO) (*dtos.AuthResponseDTO, error)
	GetUserByID(id string) (*dtos.UserResponseDTO, error)
	IsUserExists(login string) (bool, error)
	UpdateUser(id string, user *dtos.UpdateUserRequest) (*dtos.UserResponseDTO, error)
	DeleteUser(id string) error
}

type UserService struct {
	validator validation.Validator
	repo      repositories.IUserRepository
	token     tokens.TokenService
	hasher    hasher.Hasher
}

func (s *UserService) Login(ctx context.Context, dto *dtos.AuthDTO) (*dtos.AuthResponseDTO, error) {
	//TODO implement me
	panic("implement me")
}

func (s *UserService) GetUserByID(id string) (*dtos.UserResponseDTO, error) {
	//TODO implement me
	panic("implement me")
}

func (s *UserService) IsUserExists(login string) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (s *UserService) UpdateUser(id string, user *dtos.UpdateUserRequest) (*dtos.UserResponseDTO, error) {
	//TODO implement me
	panic("implement me")
}

func (s *UserService) DeleteUser(id string) error {
	//TODO implement me
	panic("implement me")
}

func NewUserService(
	validator *validation.Validator,
	token *tokens.TokenService,
	repo repositories.IUserRepository,
	hasher hasher.Hasher,
) *UserService {
	return &UserService{
		validator: *validator,
		repo:      repo,
		token:     *token,
		hasher:    hasher,
	}
}

func (s *UserService) Register(ctx context.Context, dto *dtos.CreateUserRequest) (*dtos.CreateUserResponse, error) {
	if err := s.validator.ValidateStruct(dto); err != nil {
		return nil, err
	}

	user := dto.ToUser()
	if err := s.repo.CreateUser(user); err != nil {
		return nil, err
	}
	payload := tokens.JWTPayload{
		Username: user.Username,
		Role:     user.Role,
	}
	service := s.token
	token, err := service.GenerateToken(payload)
	if err != nil {
		return nil, err
	}

	return &dtos.CreateUserResponse{
		ID:    user.ID.Hex(),
		Token: token,
	}, nil
}
