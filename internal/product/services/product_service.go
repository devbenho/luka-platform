package services

import (
	"context"

	"github.com/devbenho/luka-platform/internal/product/dtos"
	"github.com/devbenho/luka-platform/internal/product/models"
	"github.com/devbenho/luka-platform/internal/product/repositories"
	storeRepo "github.com/devbenho/luka-platform/internal/store/repositories"
	"github.com/devbenho/luka-platform/pkg/errors"
	"github.com/devbenho/luka-platform/pkg/validation"
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
		return nil, err
	}

	store, err := s.storeRepo.GetStoreByID(ctx, product.StoreID.Hex())
	if err != nil {
		return nil, errors.Wrap(err, "fetching store")
	}
	if store == nil || store.DeletedAt != nil {
		return nil, errors.NewNotFoundError("store", product.StoreID.Hex())
	}

	productResult, err := s.repo.CreateProduct(ctx, product.ToProduct())
	if err != nil {
		return nil, errors.Wrap(err, "creating product in database")
	}

	return &dtos.CreateProductResponse{
		Product: productResult,
	}, nil
}

// GetProductByID returns the product, or a not-found error when it does not
// exist or has been soft-deleted.
func (s *ProductService) GetProductByID(ctx context.Context, id string) (*models.Product, error) {
	product, err := s.repo.GetProductByID(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, "fetching product")
	}
	if product == nil || product.DeletedAt != nil {
		return nil, errors.NewNotFoundError("product", id)
	}
	return product, nil
}

func (s *ProductService) UpdateProduct(ctx context.Context, id string, update *dtos.UpdateProductRequest) (*models.Product, error) {
	if err := s.validator.ValidateStruct(update); err != nil {
		return nil, err
	}

	product, err := s.GetProductByID(ctx, id)
	if err != nil {
		return nil, err
	}

	update.ApplyTo(product)
	if err := s.repo.UpdateProduct(ctx, id, product); err != nil {
		return nil, errors.Wrap(err, "updating product")
	}

	return product, nil
}

func (s *ProductService) DeleteProduct(ctx context.Context, id string) error {
	if _, err := s.GetProductByID(ctx, id); err != nil {
		return err
	}
	if err := s.repo.DeleteProduct(ctx, id); err != nil {
		return errors.Wrap(err, "soft deleting product")
	}
	return nil
}
