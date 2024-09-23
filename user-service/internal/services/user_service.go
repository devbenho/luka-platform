package services

import (
	"context"
	"errors"
	"github.com/devbenho/bazar-user-service/internal/dtos"
	"github.com/devbenho/bazar-user-service/internal/repositories"
	"github.com/devbenho/bazar-user-service/pkg/hasher"
	"github.com/devbenho/bazar-user-service/pkg/tokens"
	"github.com/devbenho/bazar-user-service/pkg/validation"
	"log"
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
	isExist, err := s.IsUserExists(dto.Login)
	if err != nil {
		return nil, err
	}
	if !isExist {
		return nil, errors.New("user not found")
	}

	user, err := s.repo.GetUserByUsername(dto.Login)
	if err != nil {
		user, err = s.repo.GetUserByEmail(dto.Login)
		if err != nil {
			return nil, err
		}
	}

	log.Println("user is ", user.Password)
	log.Println("dto is ", dto.Password)

	log.Print("res is ", err)
	if s.hasher.Compare(user.Password, dto.Password); err != nil {
		return nil, errors.New("invalid password")
	}

	payload := tokens.JWTPayload{
		Username: user.Username,
		Role:     user.Role,
	}

	token, err := s.token.GenerateToken(payload)
	if err != nil {
		return nil, err
	}

	return &dtos.AuthResponseDTO{
		Email: user.Email,
		Token: token.Token,
	}, nil

}

func (s *UserService) GetUserByID(id string) (*dtos.UserResponseDTO, error) {
	//TODO implement me
	panic("implement me")
}

func (s *UserService) IsUserExists(login string) (bool, error) {
	_, err := s.repo.GetUserByUsername(login)
	if err != nil {
		_, err = s.repo.GetUserByEmail(login)
		if err != nil {
			return false, nil
		}
	}
	return true, nil
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

	dto.Password, _ = s.hasher.Hash(dto.Password)
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
