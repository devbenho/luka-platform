package services

import (
	"context"
	"fmt"

	dtos "github.com/devbenho/luka-platform/internal/user/dtos/users"
	"github.com/devbenho/luka-platform/internal/user/models"
	"github.com/devbenho/luka-platform/internal/user/repositories"
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
	GetUserByID(ctx context.Context, id string) (*dtos.UserResponseDTO, error)
	UpdateUser(ctx context.Context, id string, user *dtos.UpdateUserRequest) (*dtos.UserResponseDTO, error)
	DeleteUser(ctx context.Context, id string) error
	FindUser(ctx context.Context, login string) (*models.User, error)
}

type UserService struct {
	validator validation.Validator
	repo      repositories.IUserRepository
	token     tokens.TokenService
	hasher    hasher.Hasher
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

	if err := s.repo.CreateUser(ctx, user); err != nil {
		return nil, err
	}

	payload := map[string]interface{}{
		"Id":       user.ID.Hex(),
		"Email":    user.Email,
		"Role":     user.Role,
		"Username": user.Username,
	}

	token := tokens.GenerateAccessToken(payload)

	return &dtos.CreateUserResponse{
		ID:    user.ID.Hex(),
		Token: token,
	}, nil
}

func (s *UserService) Login(ctx context.Context, dto *dtos.AuthDTO) (*dtos.AuthResponseDTO, error) {
	existUser, err := s.FindUser(ctx, dto.Login)
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

	payload := map[string]interface{}{
		"Id":    existUser.ID.Hex(),
		"Email": existUser.Email,
		"Role":  existUser.Role,
	}

	token := tokens.GenerateAccessToken(payload)

	return &dtos.AuthResponseDTO{
		Email: existUser.Email,
		Token: token,
	}, nil
}

func (s *UserService) GetUserByID(ctx context.Context, id string) (*dtos.UserResponseDTO, error) {
	user, err := s.repo.GetUserByID(ctx, id)
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

func (s *UserService) UpdateUser(ctx context.Context, id string, user *dtos.UpdateUserRequest) (*dtos.UserResponseDTO, error) {
	existingUser, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if user.Username != "" {
		existingUser.Username = user.Username
	}
	if user.Email != "" {
		existingUser.Email = user.Email
	}

	if err := s.repo.UpdateUser(ctx, id, existingUser); err != nil {
		return nil, err
	}

	return &dtos.UserResponseDTO{
		ID:       existingUser.ID.Hex(),
		Username: existingUser.Username,
		Email:    existingUser.Email,
		Role:     existingUser.Role,
	}, nil
}

func (s *UserService) DeleteUser(ctx context.Context, id string) error {
	return s.repo.DeleteUser(ctx, id)
}

func (s *UserService) FindUser(ctx context.Context, login string) (*models.User, error) {
	isEmail := utils.IsEmailValid(login)
	var user *models.User
	var err error
	if isEmail {
		user, err = s.repo.GetUserByEmail(ctx, login)
		if err == nil {
			return user, nil
		}
	}
	user, err = s.repo.GetUserByUsername(ctx, login)

	if err != nil {
		return nil, &errors.NotFoundError{
			Entity: "user",
			Field:  "login key",
			Value:  login,
		}
	}

	return user, nil
}

func convertValidationErrors(validationErrors validator.ValidationErrors) errors.ValidationErrors {
	var customErrors errors.ValidationErrors
	for _, e := range validationErrors {
		newError := errors.NewValidationError(e.Field(), e.Tag(), fmt.Sprintf("%v", e.Value()))
		customErrors = append(customErrors, newError)
	}

	return customErrors
}
