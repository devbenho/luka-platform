package services

import (
	"context"

	dtos "github.com/devbenho/luka-platform/internal/user/dtos/users"
	"github.com/devbenho/luka-platform/internal/user/repositories"
	"github.com/devbenho/luka-platform/pkg/database"
	apperror "github.com/devbenho/luka-platform/pkg/errors"
	"github.com/devbenho/luka-platform/pkg/hasher"
	"github.com/devbenho/luka-platform/pkg/tokens"
)

type IUserService interface {
	Register(ctx context.Context, dto *dtos.CreateUserRequest) (*dtos.CreateUserResponse, error)
}

type UserService struct {
	repo   repositories.IUserRepository
	token  tokens.TokenService
	hasher hasher.Hasher
}

func NewUserService(
	token *tokens.TokenService,
	repo repositories.IUserRepository,
	hasher hasher.Hasher,
) IUserService {
	return &UserService{
		repo:   repo,
		token:  *token,
		hasher: hasher,
	}
}

func (s *UserService) Register(ctx context.Context, dto *dtos.CreateUserRequest) (*dtos.CreateUserResponse, error) {

	if ok := s.repo.IsUserExists(ctx, dto.Email); ok {
		return nil, apperror.Conflict(apperror.ErrEmailAlreadyExist)
	}

	dto.Password, _ = s.hasher.Hash(dto.Password)

	user, err := s.repo.CreateUser(ctx, dto.ToUser())
	if err != nil {
		switch e := err.(type) {
		case *database.DBDuplicateError:
			if e.Field == "email" {
				return nil, apperror.Conflict(apperror.ErrEmailAlreadyExist)
			}
			return nil, apperror.Conflict(apperror.ErrorUsernameAlreadyExist)
		case *database.DBConnectionError:
			return nil, apperror.ServiceUnavailable(e)
		case *database.DBValidationError:
			return nil, apperror.BadRequest(e)
		case *database.DBInternalError:
			return nil, apperror.InternalServerError(e)
		default:
			return nil, apperror.InternalServerError(err)
		}
	}

	payload := map[string]interface{}{
		"Id":       user.ID.Hex(),
		"Email":    user.Email,
		"Role":     user.Role,
		"Username": user.Username,
	}

	token := tokens.GenerateAccessToken(payload)

	return &dtos.CreateUserResponse{
		ID:       user.ID.Hex(),
		Token:    token,
		Email:    user.Email,
		Username: user.Username,
		Role:     user.Role,
	}, nil
}
