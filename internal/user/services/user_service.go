package services

import (
	"context"
	"net/http"

	dtos "github.com/devbenho/luka-platform/internal/user/dtos/users"
	"github.com/devbenho/luka-platform/internal/user/models"
	"github.com/devbenho/luka-platform/internal/user/repositories"
	"github.com/devbenho/luka-platform/internal/utils"
	"github.com/devbenho/luka-platform/pkg/errors"
	"github.com/devbenho/luka-platform/pkg/hasher"
	"github.com/devbenho/luka-platform/pkg/tokens"
	"github.com/devbenho/luka-platform/pkg/validation"
)

type IUserService interface {
	Register(ctx context.Context, dto *dtos.CreateUserRequest) (*dtos.CreateUserResponse, error)
	Login(ctx context.Context, dto *dtos.AuthDTO) (*dtos.AuthResponseDTO, error)
	GetUserByID(ctx context.Context, id string) (*dtos.UserResponseDTO, error)
	UpdateUser(ctx context.Context, id string, user *dtos.UpdateUserRequest) (*dtos.UserResponseDTO, error)
	DeleteUser(ctx context.Context, id string) error
	FindUser(ctx context.Context, login string) (*models.User, error)
	FindUserByEmail(ctx context.Context, email string) (*models.User, error)
	FindUserByUsername(ctx context.Context, username string) (*models.User, error)
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
		if validationErrors, ok := err.(errors.ValidationErrors); ok {
			return nil, validationErrors
		}
		return nil, err
	}

	// validate if user already exists

	user, err := s.FindUserByEmailOrUsername(ctx, dto.Email, dto.Username)
	if err != nil {
		return nil, err
	}
	if user != nil {
		return nil, errors.NewError(errors.UserAlreadyExists, http.StatusBadRequest, "user already exists")
	}

	// hash password
	dto.Password, err = s.hasher.Hash(dto.Password)
	if err != nil {
		return nil, err
	}

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
		return nil, errors.Wrap(err, "finding user")
	}
	if existUser == nil {
		return nil, errors.NewNotFoundError("user", dto.Login)
	}

	if err := s.hasher.Compare(existUser.Password, dto.Password); err != nil {
		return nil, errors.NewError(errors.InvalidCredentials, http.StatusUnauthorized, "invalid credentials")
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
		return nil, errors.Wrap(err, "getting user by ID")
	}
	if user == nil {
		return nil, errors.NewNotFoundError("user", id)
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
		return nil, errors.Wrap(err, "fetching user")
	}
	if existingUser == nil {
		return nil, errors.NewNotFoundError("user", id)
	}

	if err := s.validator.ValidateStruct(user); err != nil {
		return nil, errors.Wrap(err, "validating update request")
	}

	utils.Copy(existingUser, user)
	if err := s.repo.UpdateUser(ctx, id, existingUser); err != nil {
		return nil, errors.Wrap(err, "updating user in database")
	}

	return &dtos.UserResponseDTO{
		ID:       existingUser.ID.Hex(),
		Username: existingUser.Username,
		Email:    existingUser.Email,
		Role:     existingUser.Role,
	}, nil
}

func (s *UserService) DeleteUser(ctx context.Context, id string) error {
	user, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		return errors.Wrap(err, "fetching user")
	}
	if user == nil {
		return errors.NewNotFoundError("user", id)
	}

	if user.Role == "admin" {
		return errors.NewError(errors.AdminCannotBeDeleted, http.StatusBadRequest, "admin cannot be deleted")
	}

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
		return nil, errors.NewNotFoundError("user", login)
	}

	return user, nil
}

func (s *UserService) FindUserByEmail(ctx context.Context, email string) (*models.User, error) {
	return s.repo.GetUserByEmail(ctx, email)
}

func (s *UserService) FindUserByUsername(ctx context.Context, username string) (*models.User, error) {
	return s.repo.GetUserByUsername(ctx, username)
}

func (s *UserService) FindUserByEmailOrUsername(ctx context.Context, email, username string) (*models.User, error) {
	user, err := s.FindUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if user != nil {
		return user, nil
	}
	return s.FindUserByUsername(ctx, username)
}
