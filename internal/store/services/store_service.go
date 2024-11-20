package services

import (
	"context"
	"fmt"
	"log"

	dtos "github.com/devbenho/luka-platform/internal/store/dtos"
	"github.com/devbenho/luka-platform/internal/store/models"
	"github.com/devbenho/luka-platform/internal/store/repositories"
	"github.com/devbenho/luka-platform/pkg/errors"
	"github.com/devbenho/luka-platform/pkg/validation"
	"github.com/go-playground/validator/v10"
)

type IStoreService interface {
	CreateStore(ctx context.Context, store *dtos.CreateStoreRequest) (*dtos.CreateStoreResponse, error)
	GetStoreByID(ctx context.Context, id string) (*models.Store, error)
	UpdateStore(ctx context.Context, id string, store *dtos.UpdateStoreRequest) (*models.Store, error)
	DeleteStore(ctx context.Context, id string) error
}

type StoreService struct {
	repo      repositories.IStoreRepository
	validator *validation.Validator
}

func NewStoreService(repository repositories.IStoreRepository, validator *validation.Validator) IStoreService {
	return &StoreService{
		repo:      repository,
		validator: validator,
	}
}

func (s *StoreService) CreateStore(ctx context.Context, store *dtos.CreateStoreRequest) (*dtos.CreateStoreResponse, error) {
	log.Println(`Context is `, ctx)
	if err := store.Validate(); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			validationErrorsResult := convertValidationErrors(validationErrors)
			return nil, validationErrorsResult
		}
		return nil, err
	}
	log.Println(`The store entity is `, store.ToStore())
	storeResult, err := s.repo.CreateStore(ctx, store.ToStore())
	if err != nil {
		return nil, err
	}

	return &dtos.CreateStoreResponse{
		ID:   storeResult.ID.Hex(),
		Name: storeResult.Name,
		Slug: storeResult.Slug,
	}, nil
}

func (s *StoreService) GetStoreByID(ctx context.Context, id string) (*models.Store, error) {
	return nil, nil
}

func (s *StoreService) UpdateStore(ctx context.Context, id string, store *dtos.UpdateStoreRequest) (*models.Store, error) {
	return nil, nil
}

func (s *StoreService) DeleteStore(ctx context.Context, id string) error {
	return nil
}

func convertValidationErrors(validationErrors validator.ValidationErrors) errors.ValidationErrors {
	var customErrors errors.ValidationErrors
	for _, e := range validationErrors {
		newError := errors.NewValidationError(e.Field(), e.Tag(), fmt.Sprintf("%v", e.Value()))
		customErrors = append(customErrors, newError)
	}

	return customErrors
}