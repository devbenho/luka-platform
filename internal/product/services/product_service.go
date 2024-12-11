package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/devbenho/luka-platform/internal/product/dtos"
	"github.com/devbenho/luka-platform/internal/product/models"
	"github.com/devbenho/luka-platform/internal/product/repositories"
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
	validator *validation.Validator
}

func NewProductService(repository repositories.IProductRepository, validator *validation.Validator) IProductService {
	return &ProductService{
		repo:      repository,
		validator: validator,
	}
}

func (s *ProductService) CreateProduct(ctx context.Context, product *dtos.CreateProductRequest) (*dtos.CreateProductResponse, error) {
	log.Println(`Context is `, ctx)
	if err := product.Validate(); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			validationErrorsResult := convertValidationErrors(validationErrors)
			return nil, validationErrorsResult
		}
		return nil, err
	}
	log.Println(`The product entity is `, product.ToProduct())
	productResult, err := s.repo.CreateProduct(ctx, product.ToProduct())
	if err != nil {
		return nil, err
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
		return nil, &errors.NotFoundError{Entity: "product", Field: "id", Value: id}
	}
	return product, nil
}

func (s *ProductService) UpdateProduct(ctx context.Context, id string, product *dtos.UpdateProductRequest) (*models.Product, error) {
	if err := product.Validate(); err != nil {
		log.Println(`yalahwy `, err)
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			validationErrorsResult := convertValidationErrors(validationErrors)
			return nil, validationErrorsResult
		}
		return nil, err
	}
	log.Println(`Hey`)

	existingProduct, err := s.repo.GetProductByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existingProduct.DeletedAt != nil {
		return nil, &errors.NotFoundError{Entity: "product", Field: "id", Value: id}
	}
	utils.Copy(existingProduct, product)
	err = s.repo.UpdateProduct(ctx, id, existingProduct)
	if err != nil {
		return nil, err
	}

	return existingProduct, nil
}

func (s *ProductService) DeleteProduct(ctx context.Context, id string) error {
	product, err := s.repo.GetProductByID(ctx, id)
	if err != nil {
		return err
	}
	if product.DeletedAt != nil {
		return &errors.NotFoundError{Entity: "product", Field: "id", Value: id}
	}

	now := time.Now()
	product.DeletedAt = &now
	err = s.repo.UpdateProduct(ctx, id, product)
	if err != nil {
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
