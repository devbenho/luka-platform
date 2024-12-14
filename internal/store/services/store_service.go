package services

import (
	"context"
	"fmt"

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
	if err := s.validator.ValidateStruct(store); err != nil {
		return nil, errors.Wrap(err, "validating store")
	}

	storeEntity := store.ToStore()
	storeResult, err := s.repo.CreateStore(ctx, storeEntity)
	if err != nil {
		return nil, errors.Wrap(err, "creating store in database")
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
		return nil, errors.Wrap(err, "fetching store")
	}
	if store.DeletedAt != nil {
		return nil, errors.NewNotFoundError("store", id)
	}
	return store, nil
}

func (s *StoreService) UpdateStore(ctx context.Context, id string, store *dtos.UpdateStoreRequest) (*models.Store, error) {
	if err := s.validator.ValidateStruct(store); err != nil {
		return nil, errors.Wrap(err, "validating store update")
	}

	existingStore, err := s.repo.GetStoreByID(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, "fetching existing store")
	}
	if existingStore.DeletedAt != nil {
		return nil, errors.NewNotFoundError("store", id)
	}

	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID != existingStore.OwnerId.String() {
		return nil, errors.NewUnauthorizedError("not authorized to update this store")
	}

	updatedStore := store.ToStore()
	updatedStore.ID = existingStore.ID
	updatedStore.OwnerId = existingStore.OwnerId

	if err = s.repo.UpdateStore(ctx, id, updatedStore); err != nil {
		return nil, errors.Wrap(err, "updating store in database")
	}

	return updatedStore, nil
}

func (s *StoreService) DeleteStore(ctx context.Context, id string) error {
	store, err := s.repo.GetStoreByID(ctx, id)
	if err != nil {
		return errors.Wrap(err, "fetching store")
	}
	if store.DeletedAt != nil {
		return errors.NewNotFoundError("store", id)
	}

	if err := s.repo.DeleteStore(ctx, id); err != nil {
		return errors.Wrap(err, "deleting store")
	}
	return nil
}

func convertValidationErrors(validationErrors validator.ValidationErrors) errors.ValidationErrors {
	var customErrors errors.ValidationErrors
	for _, e := range validationErrors {
		newError := errors.NewValidationError(
			e.Field(),
			fmt.Sprintf("validation failed for tag %s", e.Tag()),
		)
		customErrors = append(customErrors, newError)
	}
	return customErrors
}
