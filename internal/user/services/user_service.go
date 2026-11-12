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

	existing, err := s.FindUserByEmailOrUsername(ctx, dto.Email, dto.Username)
	if err != nil {
		return nil, errors.Wrap(err, "checking for existing user")
	}
	if existing != nil {
		return nil, errors.NewError(errors.UserAlreadyExists, http.StatusBadRequest, "user already exists")
	}

	hashedPassword, err := s.hasher.Hash(dto.Password)
	if err != nil {
		return nil, errors.NewInternalError("hashing password", err)
	}

	user := dto.ToUser()
	user.Password = hashedPassword
	if err := s.repo.CreateUser(ctx, user); err != nil {
		return nil, errors.Wrap(err, "creating user")
	}

	payload := map[string]interface{}{
		"Id":       user.ID.Hex(),
		"Email":    user.Email,
		"Role":     user.Role,
		"Username": user.Username,
	}

	token, err := tokens.GenerateAccessToken(payload)
	if err != nil {
		return nil, errors.NewInternalError("generating access token", err)
	}

	return &dtos.CreateUserResponse{
		ID:    user.ID.Hex(),
		Token: token,
	}, nil
}

func (s *UserService) Login(ctx context.Context, dto *dtos.AuthDTO) (*dtos.AuthResponseDTO, error) {
	if err := s.validator.ValidateStruct(dto); err != nil {
		return nil, err
	}

	existUser, err := s.FindUser(ctx, dto.Login)
	if err != nil {
		return nil, errors.Wrap(err, "finding user")
	}
	// Same response for an unknown user and a wrong password, so the login
	// endpoint cannot be used to discover which accounts exist.
	if existUser == nil || s.hasher.Compare(existUser.Password, dto.Password) != nil {
		return nil, errors.NewError(errors.InvalidCredentials, http.StatusUnauthorized, "invalid credentials")
	}

	payload := map[string]interface{}{
		"Id":       existUser.ID.Hex(),
		"Email":    existUser.Email,
		"Role":     existUser.Role,
		"Username": existUser.Username,
	}

	token, err := tokens.GenerateAccessToken(payload)
	if err != nil {
		return nil, errors.NewInternalError("generating access token", err)
	}

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
		return nil, err
	}
	if err := s.ensureUnique(ctx, existingUser, user); err != nil {
		return nil, err
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

// FindUser looks a user up by email (when login is an email address) or by
// username. It returns (nil, nil) when no user matches.
func (s *UserService) FindUser(ctx context.Context, login string) (*models.User, error) {
	if utils.IsEmailValid(login) {
		user, err := s.repo.GetUserByEmail(ctx, login)
		if err != nil || user != nil {
			return user, err
		}
	}
	return s.repo.GetUserByUsername(ctx, login)
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

// ensureUnique rejects an update that would give the user a username or email
// already used by another account.
func (s *UserService) ensureUnique(ctx context.Context, current *models.User, update *dtos.UpdateUserRequest) error {
	if update.Username != nil && *update.Username != current.Username {
		other, err := s.repo.GetUserByUsername(ctx, *update.Username)
		if err != nil {
			return errors.Wrap(err, "checking username")
		}
		if other != nil {
			return errors.NewError(errors.UserAlreadyExists, http.StatusBadRequest, "username already taken")
		}
	}
	if update.Email != nil && *update.Email != current.Email {
		other, err := s.repo.GetUserByEmail(ctx, *update.Email)
		if err != nil {
			return errors.Wrap(err, "checking email")
		}
		if other != nil {
			return errors.NewError(errors.UserAlreadyExists, http.StatusBadRequest, "email already taken")
		}
	}
	return nil
}
