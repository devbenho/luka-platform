package services

import (
	"context"

	"github.com/devbenho/luka-platform/internal/category/dtos"
	"github.com/devbenho/luka-platform/internal/category/models"
	"github.com/devbenho/luka-platform/internal/category/repositories"
	"github.com/devbenho/luka-platform/pkg/errors"
	"github.com/devbenho/luka-platform/pkg/validation"
)

type ICategoryService interface {
	CreateCategory(ctx context.Context, category *dtos.CreateCategoryRequest) (*dtos.CreateCategoryResponse, error)
	GetCategoryByID(ctx context.Context, id string) (*models.Category, error)
	UpdateCategory(ctx context.Context, id string, category *dtos.UpdateCategoryRequest) (*models.Category, error)
	DeleteCategory(ctx context.Context, id string) error
}

type CategoryService struct {
	repo      repositories.ICategoryRepository
	validator *validation.Validator
}

func NewCategoryService(repository repositories.ICategoryRepository, validator *validation.Validator) ICategoryService {
	return &CategoryService{
		repo:      repository,
		validator: validator,
	}
}

func (s *CategoryService) CreateCategory(ctx context.Context, category *dtos.CreateCategoryRequest) (*dtos.CreateCategoryResponse, error) {
	if err := s.validator.ValidateStruct(category); err != nil {
		return nil, err
	}
	categoryResult, err := s.repo.CreateCategory(ctx, category.ToCategory())
	if err != nil {
		return nil, errors.Wrap(err, "creating category")
	}

	return &dtos.CreateCategoryResponse{
		Category: *categoryResult,
	}, nil
}

// GetCategoryByID returns the category, or a not-found error when it does not
// exist or has been soft-deleted.
func (s *CategoryService) GetCategoryByID(ctx context.Context, id string) (*models.Category, error) {
	category, err := s.repo.GetCategoryByID(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, "fetching category")
	}
	if category == nil || category.DeletedAt != nil {
		return nil, errors.NewNotFoundError("category", id)
	}
	return category, nil
}

func (s *CategoryService) UpdateCategory(ctx context.Context, id string, update *dtos.UpdateCategoryRequest) (*models.Category, error) {
	if err := s.validator.ValidateStruct(update); err != nil {
		return nil, err
	}

	category, err := s.GetCategoryByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := update.ApplyTo(category); err != nil {
		return nil, err
	}
	if err := s.repo.UpdateCategory(ctx, id, category); err != nil {
		return nil, errors.Wrap(err, "updating category")
	}

	return category, nil
}

func (s *CategoryService) DeleteCategory(ctx context.Context, id string) error {
	if _, err := s.GetCategoryByID(ctx, id); err != nil {
		return err
	}
	if err := s.repo.DeleteCategory(ctx, id); err != nil {
		return errors.Wrap(err, "deleting category")
	}
	return nil
}
