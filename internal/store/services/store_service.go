package services

import (
	"context"

	dtos "github.com/devbenho/luka-platform/internal/store/dtos"
	"github.com/devbenho/luka-platform/internal/store/models"
	"github.com/devbenho/luka-platform/internal/store/repositories"
	"github.com/devbenho/luka-platform/pkg/errors"
	"github.com/devbenho/luka-platform/pkg/validation"
)

type IStoreService interface {
	CreateStore(ctx context.Context, store *dtos.CreateStoreRequest) (*dtos.CreateStoreResponse, error)
	GetStoreByID(ctx context.Context, id string) (*models.Store, error)
	UpdateStore(ctx context.Context, id, userID string, store *dtos.UpdateStoreRequest) (*models.Store, error)
	DeleteStore(ctx context.Context, id, userID string) error
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
		return nil, err
	}

	storeResult, err := s.repo.CreateStore(ctx, store.ToStore())
	if err != nil {
		return nil, errors.Wrap(err, "creating store in database")
	}

	return &dtos.CreateStoreResponse{
		ID:   storeResult.ID.Hex(),
		Name: storeResult.Name,
		Slug: storeResult.Slug,
	}, nil
}

// GetStoreByID returns the store, or a not-found error when it does not exist
// or has been soft-deleted.
func (s *StoreService) GetStoreByID(ctx context.Context, id string) (*models.Store, error) {
	store, err := s.repo.GetStoreByID(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, "fetching store")
	}
	if store == nil || store.DeletedAt != nil {
		return nil, errors.NewNotFoundError("store", id)
	}
	return store, nil
}

// getOwnedStore returns the store if userID owns it.
func (s *StoreService) getOwnedStore(ctx context.Context, id, userID string) (*models.Store, error) {
	store, err := s.GetStoreByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if store.OwnerId.Hex() != userID {
		return nil, errors.NewForbiddenError("you can only modify your own store")
	}
	return store, nil
}

func (s *StoreService) UpdateStore(ctx context.Context, id, userID string, update *dtos.UpdateStoreRequest) (*models.Store, error) {
	if err := s.validator.ValidateStruct(update); err != nil {
		return nil, err
	}

	store, err := s.getOwnedStore(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	update.ApplyTo(store)
	if err = s.repo.UpdateStore(ctx, id, store); err != nil {
		return nil, errors.Wrap(err, "updating store in database")
	}

	return store, nil
}

func (s *StoreService) DeleteStore(ctx context.Context, id, userID string) error {
	if _, err := s.getOwnedStore(ctx, id, userID); err != nil {
		return err
	}
	if err := s.repo.DeleteStore(ctx, id); err != nil {
		return errors.Wrap(err, "deleting store")
	}
	return nil
}
