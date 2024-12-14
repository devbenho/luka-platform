package services

import (
	"context"
	"log"

	"github.com/devbenho/luka-platform/internal/category/dtos"
	"github.com/devbenho/luka-platform/internal/category/models"
	"github.com/devbenho/luka-platform/internal/category/repositories"
	"github.com/devbenho/luka-platform/pkg/errors"
	"github.com/devbenho/luka-platform/pkg/validation"
	"github.com/go-playground/validator/v10"
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
	log.Println(`Context is `, ctx)
	if err := category.Validate(); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			validationErrorsResult := convertValidationErrors(validationErrors)
			return nil, validationErrorsResult
		}
		return nil, err
	}
	categoryResult, err := s.repo.CreateCategory(ctx, category.ToCategory())

	log.Print(2)
	if err != nil {
		return nil, err
	}

	return &dtos.CreateCategoryResponse{
		Category: *categoryResult,
	}, nil
}

func (s *CategoryService) GetCategoryByID(ctx context.Context, id string) (*models.Category, error) {
	category, err := s.repo.GetCategoryByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if category.DeletedAt != nil {
		return nil, errors.NewNotFoundError("category", id)
	}
	return category, nil
}

func (s *CategoryService) UpdateCategory(ctx context.Context, id string, category *dtos.UpdateCategoryRequest) (*models.Category, error) {
	if err := category.Validate(); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			validationErrorsResult := convertValidationErrors(validationErrors)
			return nil, validationErrorsResult
		}
		return nil, err
	}

	existingCategory, err := s.repo.GetCategoryByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existingCategory.DeletedAt != nil {
		return nil, errors.NewNotFoundError("category", id)
	}

	updatedCategory, _ := category.ToCategory()
	updatedCategory.ID = existingCategory.ID

	err = s.repo.UpdateCategory(ctx, id, &updatedCategory)
	if err != nil {
		return nil, err
	}

	return &updatedCategory, nil
}

func (s *CategoryService) DeleteCategory(ctx context.Context, id string) error {
	category, err := s.repo.GetCategoryByID(ctx, id)
	if err != nil {
		return err
	}
	if category.DeletedAt != nil {
		return errors.NewNotFoundError("category", id)
	}
	err = s.repo.UpdateCategory(ctx, id, category)
	if err != nil {
		return err
	}

	return nil
}

func convertValidationErrors(validationErrors validator.ValidationErrors) errors.ValidationErrors {
	var customErrors errors.ValidationErrors
	for _, err := range validationErrors {
		customErrors = append(customErrors, errors.NewValidationError(err.Field(), err.Tag()))
	}
	return customErrors
}
