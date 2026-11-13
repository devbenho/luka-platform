package services

import (
	"context"

	"github.com/devbenho/luka-platform/internal/product/dtos"
	"github.com/devbenho/luka-platform/internal/product/models"
	"github.com/devbenho/luka-platform/internal/product/repositories"
	storeRepo "github.com/devbenho/luka-platform/internal/store/repositories"
	storeServices "github.com/devbenho/luka-platform/internal/store/services"
	"github.com/devbenho/luka-platform/pkg/errors"
	"github.com/devbenho/luka-platform/pkg/validation"
)

type IProductService interface {
	// CreateProduct, UpdateProduct and DeleteProduct require userID to own
	// the product's store.
	CreateProduct(ctx context.Context, userID string, product *dtos.CreateProductRequest) (*dtos.CreateProductResponse, error)
	GetProductByID(ctx context.Context, id string) (*models.Product, error)
	UpdateProduct(ctx context.Context, id, userID string, product *dtos.UpdateProductRequest) (*models.Product, error)
	DeleteProduct(ctx context.Context, id, userID string) error
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

func (s *ProductService) CreateProduct(ctx context.Context, userID string, product *dtos.CreateProductRequest) (*dtos.CreateProductResponse, error) {
	if err := s.validator.ValidateStruct(product); err != nil {
		return nil, err
	}

	if err := storeServices.RequireStoreOwner(ctx, s.storeRepo, product.StoreID.Hex(), userID); err != nil {
		return nil, err
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

func (s *ProductService) UpdateProduct(ctx context.Context, id, userID string, update *dtos.UpdateProductRequest) (*models.Product, error) {
	if err := s.validator.ValidateStruct(update); err != nil {
		return nil, err
	}

	product, err := s.getOwnedProduct(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	update.ApplyTo(product)
	if err := s.repo.UpdateProduct(ctx, id, product); err != nil {
		return nil, errors.Wrap(err, "updating product")
	}

	return product, nil
}

func (s *ProductService) DeleteProduct(ctx context.Context, id, userID string) error {
	if _, err := s.getOwnedProduct(ctx, id, userID); err != nil {
		return err
	}
	if err := s.repo.DeleteProduct(ctx, id); err != nil {
		return errors.Wrap(err, "soft deleting product")
	}
	return nil
}

func (s *ProductService) getOwnedProduct(ctx context.Context, id, userID string) (*models.Product, error) {
	product, err := s.GetProductByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := storeServices.RequireStoreOwner(ctx, s.storeRepo, product.StoreID.Hex(), userID); err != nil {
		return nil, err
	}
	return product, nil
}
