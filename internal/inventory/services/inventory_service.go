package services

import (
	"context"
	stderrors "errors"
	"fmt"

	"github.com/devbenho/luka-platform/internal/inventory/dtos"
	"github.com/devbenho/luka-platform/internal/inventory/models"
	"github.com/devbenho/luka-platform/internal/inventory/repositories"
	"github.com/devbenho/luka-platform/pkg/errors"
	"github.com/devbenho/luka-platform/pkg/validation"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type IInventoryService interface {
	CreateInventory(ctx context.Context, dto dtos.CreateInventoryRequest) (*models.Inventory, error)
	UpdateInventory(ctx context.Context, id string, dto dtos.UpdateInventoryRequest) (*models.Inventory, error)
	DeleteInventory(ctx context.Context, id string) error
	GetInventoryByID(ctx context.Context, id string) (*models.Inventory, error)
	// ReserveStock takes quantity units of the product from one inventory
	// record and returns that record's ID (needed to release the stock).
	ReserveStock(ctx context.Context, productID primitive.ObjectID, quantity int) (primitive.ObjectID, error)
	// ReleaseStock returns previously reserved units to an inventory record.
	ReleaseStock(ctx context.Context, inventoryID primitive.ObjectID, quantity int) error
}

type InventoryService struct {
	repo      repositories.IInventoryRepository
	validator *validation.Validator
}

func NewInventoryService(repo repositories.IInventoryRepository, validator *validation.Validator) *InventoryService {
	return &InventoryService{
		repo:      repo,
		validator: validator,
	}
}

func (s *InventoryService) CreateInventory(ctx context.Context, dto dtos.CreateInventoryRequest) (*models.Inventory, error) {
	if err := s.validator.ValidateStruct(dto); err != nil {
		return nil, err
	}
	inventory, err := s.repo.CreateInventory(ctx, dto.ToInventory())
	if err != nil {
		return nil, errors.Wrap(err, "creating inventory")
	}
	return inventory, nil
}

func (s *InventoryService) UpdateInventory(ctx context.Context, id string, updateBody dtos.UpdateInventoryRequest) (*models.Inventory, error) {
	if err := s.validator.ValidateStruct(updateBody); err != nil {
		return nil, err
	}

	inventory, err := s.GetInventoryByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := updateBody.ApplyTo(inventory); err != nil {
		return nil, err
	}

	if err := s.repo.UpdateInventory(ctx, id, inventory); err != nil {
		return nil, errors.Wrap(err, "updating inventory")
	}

	return inventory, nil
}

func (s *InventoryService) DeleteInventory(ctx context.Context, id string) error {
	err := s.repo.DeleteInventory(ctx, id)
	if stderrors.Is(err, mongo.ErrNoDocuments) {
		return errors.NewNotFoundError("inventory", id)
	}
	if err != nil {
		return errors.Wrap(err, "deleting inventory")
	}
	return nil
}

func (s *InventoryService) GetInventoryByID(ctx context.Context, id string) (*models.Inventory, error) {
	inventory, err := s.repo.GetInventoryByID(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, "fetching inventory")
	}
	if inventory == nil || inventory.DeletedAt != nil {
		return nil, errors.NewNotFoundError("inventory", id)
	}
	return inventory, nil
}

func (s *InventoryService) ReserveStock(ctx context.Context, productID primitive.ObjectID, quantity int) (primitive.ObjectID, error) {
	if quantity <= 0 {
		return primitive.NilObjectID, errors.NewBadRequestError("quantity must be positive")
	}
	inventoryID, err := s.repo.ReserveStock(ctx, productID, quantity)
	if stderrors.Is(err, repositories.ErrInsufficientStock) {
		return primitive.NilObjectID, errors.NewConflictError(fmt.Sprintf(
			"insufficient stock for product %s: requested %d", productID.Hex(), quantity,
		))
	}
	if err != nil {
		return primitive.NilObjectID, errors.Wrap(err, "reserving stock")
	}
	return inventoryID, nil
}

func (s *InventoryService) ReleaseStock(ctx context.Context, inventoryID primitive.ObjectID, quantity int) error {
	if err := s.repo.ReleaseStock(ctx, inventoryID, quantity); err != nil {
		return errors.Wrap(err, fmt.Sprintf("releasing %d units to inventory %s", quantity, inventoryID.Hex()))
	}
	return nil
}
