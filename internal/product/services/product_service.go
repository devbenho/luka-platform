package services

import (
	"context"
	"time"

	"github.com/devbenho/luka-platform/internal/product/dtos"
	"github.com/devbenho/luka-platform/internal/product/models"
	"github.com/devbenho/luka-platform/internal/product/repositories"
	storeRepo "github.com/devbenho/luka-platform/internal/store/repositories"
	"github.com/devbenho/luka-platform/internal/utils"
	"github.com/devbenho/luka-platform/pkg/errors"
	"github.com/devbenho/luka-platform/pkg/validation"
	"github.com/go-playground/validator/v10"
)

type IProductService interface {
	CreateProduct(ctx context.Context, product *dtos.CreateProductRequest) (*dtos.CreateProductResponse, error)
	GetProductByID(ctx context.Context, id string) (*models.Product, error)
	UpdateProduct(ctx context.Context, id string, product *dtos.UpdateProductRequest) (*models.Product, error)
	DeleteProduct(ctx context.Context, id string) error
}

type ProductService struct {
	repo      repositories.IProductRepository
	storeRepo storeRepo.IStoreRepository
	validator *validation.Validator
}

func NewProductService(repository repositories.IProductRepository, storeRepo storeRepo.IStoreRepository, validator *validation.Validator) IProductService {
	return &ProductService{
		repo:      repository,
		validator: validator,
		storeRepo: storeRepo,
	}
}

func (s *ProductService) CreateProduct(ctx context.Context, product *dtos.CreateProductRequest) (*dtos.CreateProductResponse, error) {
	if err := s.validator.ValidateStruct(product); err != nil {
		return nil, errors.Wrap(err, "validating product")
	}

	_, err := s.storeRepo.GetStoreByID(ctx, product.StoreID.Hex())
	if err != nil {
		return nil, errors.Wrap(err, "verifying store existence")
	}

	productResult, err := s.repo.CreateProduct(ctx, product.ToProduct())
	if err != nil {
		return nil, errors.Wrap(err, "creating product in database")
	}

	return &dtos.CreateProductResponse{
		Product: productResult,
	}, nil
}

func (s *ProductService) GetProductByID(ctx context.Context, id string) (*models.Product, error) {
	product, err := s.repo.GetProductByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if product.DeletedAt != nil {
		return nil, errors.NewNotFoundError(
			"product",
			id,
		)
	}
	return product, nil
}

func (s *ProductService) UpdateProduct(ctx context.Context, id string, product *dtos.UpdateProductRequest) (*models.Product, error) {
	if err := product.Validate(); err != nil {
		return nil, errors.Wrap(err, "validating product update")
	}

	existingProduct, err := s.repo.GetProductByID(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, "finding product")
	}
	if existingProduct.DeletedAt != nil {
		return nil, errors.NewNotFoundError("product", id)
	}

	utils.Copy(existingProduct, product)
	if err := s.repo.UpdateProduct(ctx, id, existingProduct); err != nil {
		return nil, errors.Wrap(err, "updating product")
	}

	return existingProduct, nil
}

func (s *ProductService) DeleteProduct(ctx context.Context, id string) error {
	product, err := s.repo.GetProductByID(ctx, id)
	if err != nil {
		return errors.Wrap(err, "finding product")
	}
	if product.DeletedAt != nil {
		return errors.NewNotFoundError("product", id)
	}

	now := time.Now()
	product.DeletedAt = &now
	if err := s.repo.UpdateProduct(ctx, id, product); err != nil {
		return errors.Wrap(err, "soft deleting product")
	}

	return nil
}

func convertValidationErrors(validationErrors validator.ValidationErrors) errors.ValidationErrors {
	var customErrors errors.ValidationErrors
	for _, e := range validationErrors {
		newError := errors.NewValidationError(e.Field(), e.Tag())
		customErrors = append(customErrors, newError)
	}

	return customErrors
}
