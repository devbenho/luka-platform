package services

import (
	"context"
	"fmt"
	"log"

	"github.com/devbenho/luka-platform/internal/dtos"
	"github.com/devbenho/luka-platform/internal/models"
	"github.com/devbenho/luka-platform/internal/repositories"
	"github.com/devbenho/luka-platform/internal/utils"
	"github.com/devbenho/luka-platform/pkg/errors"
	"github.com/devbenho/luka-platform/pkg/hasher"
	"github.com/devbenho/luka-platform/pkg/tokens"
	"github.com/devbenho/luka-platform/pkg/validation"
	"github.com/go-playground/validator/v10"
)

type IUserService interface {
	Register(ctx context.Context, dto *dtos.CreateUserRequest) (*dtos.CreateUserResponse, error)
	Login(ctx context.Context, dto *dtos.AuthDTO) (*dtos.AuthResponseDTO, error)
	GetUserByID(id string) (*dtos.UserResponseDTO, error)
	UpdateUser(id string, user *dtos.UpdateUserRequest) (*dtos.UserResponseDTO, error)
	DeleteUser(id string) error
	FindUser(login string) (*models.User, error)
}

type UserService struct {
	validator validation.Validator
	repo      repositories.IUserRepository
	token     tokens.TokenService
	hasher    hasher.Hasher
}

func (s *UserService) Login(ctx context.Context, dto *dtos.AuthDTO) (*dtos.AuthResponseDTO, error) {
	existUser, err := s.FindUser(dto.Login)
	if err != nil {
		return nil, err
	}
	if existUser == nil {
		return nil, &errors.NotFoundError{
			Entity: "user",
			Field:  "login key",
			Value:  dto.Login,
		}
	}

	if err := s.hasher.Compare(existUser.Password, dto.Password); err != nil {
		return nil, &errors.InvalidCredentialsError{}
	}

	payload := tokens.JWTPayload{
		Username: existUser.Username,
		Role:     existUser.Role,
	}

	token, err := s.token.GenerateToken(payload)
	if err != nil {
		return nil, err
	}

	return &dtos.AuthResponseDTO{
		Email: existUser.Email,
		Token: token.Token,
	}, nil
}

func (s *UserService) GetUserByID(id string) (*dtos.UserResponseDTO, error) {
	user, err := s.repo.GetUserByID(id)
	if err != nil {
		return nil, err
	}
	return &dtos.UserResponseDTO{
		ID:       user.ID.Hex(),
		Username: user.Username,
		Email:    user.Email,
		Role:     user.Role,
	}, nil
}

func (s *UserService) FindUser(login string) (*models.User, error) {
	isEmail := utils.IsEmailValid(login)
	var user *models.User
	var err error
	if isEmail {
		user, err = s.repo.GetUserByEmail(login)
		if err == nil {
			return user, nil
		}
	}
	user, err = s.repo.GetUserByUsername(login)

	if err != nil {
		return nil, &errors.NotFoundError{
			Entity: "user",
			Field:  "login key",
			Value:  login,
		}
	}

	return user, nil
}

func (s *UserService) UpdateUser(id string, user *dtos.UpdateUserRequest) (*dtos.UserResponseDTO, error) {
	existingUser, err := s.repo.GetUserByID(id)
	if err != nil {
		return nil, err
	}

	if user.Username != "" {
		existingUser.Username = user.Username
	}
	if user.Email != "" {
		existingUser.Email = user.Email
	}

	existingUser = user.ToUser()

	if err := s.repo.UpdateUser(id, existingUser); err != nil {
		return nil, err
	}

	return &dtos.UserResponseDTO{
		ID:       existingUser.ID.Hex(),
		Username: existingUser.Username,
		Email:    existingUser.Email,
		Role:     existingUser.Role,
	}, nil
}

func (s *UserService) DeleteUser(id string) error {
	return s.repo.DeleteUser(id)
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
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			validationErrorsResult := convertValidationErrors(validationErrors)
			return nil, validationErrorsResult
		}
		return nil, err
	}

	dto.Password, _ = s.hasher.Hash(dto.Password)
	user := dto.ToUser()

	if err := s.repo.CreateUser(user); err != nil {
		return nil, err
	}

	log.Println(`User created successfully`)
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

func convertValidationErrors(validationErrors validator.ValidationErrors) errors.ValidationErrors {
	var customErrors errors.ValidationErrors
	for _, e := range validationErrors {
		newError := errors.NewValidationError(e.Field(), e.Tag(), fmt.Sprintf("%v", e.Value()))
		customErrors = append(customErrors, newError)
	}

	return customErrors
}
