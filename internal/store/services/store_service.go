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
	store, err := s.repo.GetStoreByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if store.DeletedAt != nil {
		return nil, &errors.NotFoundError{Entity: "store", Field: "id", Value: id}
	}
	return store, nil
}

func (s *StoreService) UpdateStore(ctx context.Context, id string, store *dtos.UpdateStoreRequest) (*models.Store, error) {
	if err := store.Validate(); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			validationErrorsResult := convertValidationErrors(validationErrors)
			return nil, validationErrorsResult
		}
		return nil, err
	}

	existingStore, err := s.repo.GetStoreByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existingStore.DeletedAt != nil {
		return nil, &errors.NotFoundError{Entity: "store", Field: "id", Value: id}
	}

	updatedStore := store.ToStore()
	log.Println(`The updated store entity is `, updatedStore)
	updatedStore.ID = existingStore.ID
	updatedStore.OwnerId = existingStore.OwnerId

	err = s.repo.UpdateStore(ctx, id, updatedStore)
	if err != nil {
		return nil, err
	}

	err = s.repo.UpdateStore(ctx, id, updatedStore)
	if err != nil {
		return nil, err
	}

	return updatedStore, nil
}

func (s *StoreService) DeleteStore(ctx context.Context, id string) error {
	store, err := s.repo.GetStoreByID(ctx, id)
	log.Println(`The store entity is `, store)
	if err != nil {
		return err
	}
	if store.DeletedAt != nil {
		return &errors.NotFoundError{
			Entity: "store",
			Field:  "id",
			Value:  id,
		}
	}

	if err := s.repo.DeleteStore(ctx, id); err != nil {
		return err
	}
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
