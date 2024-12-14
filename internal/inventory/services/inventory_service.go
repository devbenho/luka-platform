package services

import (
	"context"
	"fmt"
	"log"

	"github.com/devbenho/luka-platform/internal/inventory/dtos"
	"github.com/devbenho/luka-platform/internal/inventory/models"
	"github.com/devbenho/luka-platform/internal/inventory/repositories"
	"github.com/devbenho/luka-platform/internal/utils"
	"github.com/devbenho/luka-platform/pkg/errors"
	"github.com/devbenho/luka-platform/pkg/validation"
)

type IInventoryService interface {
	CreateInventory(ctx context.Context, dto dtos.CreateInventoryRequest) (*models.Inventory, error)
	UpdateInventory(ctx context.Context, id string, dto dtos.UpdateInventoryRequest) (*models.Inventory, error)
	DeleteInventory(ctx context.Context, id string) error
	GetInventoryByID(ctx context.Context, id string) (*models.Inventory, error)
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
		return nil, fmt.Errorf("invalid create request: %w", err)
	}
	inventory := dto.ToInventory()
	return s.repo.CreateInventory(ctx, inventory)
}

func (s *InventoryService) UpdateInventory(ctx context.Context, id string, updateBody dtos.UpdateInventoryRequest) (*models.Inventory, error) {
	if err := s.validator.ValidateStruct(updateBody); err != nil {
		log.Println(`lol: `, err)
		return nil, fmt.Errorf("invalid update request: %w", err)
	}

	existingInventory, err := s.repo.GetInventoryByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch inventory: %w", err)
	}
	if existingInventory == nil {
		return nil, errors.NewNotFoundError(
			"inventory",
			id,
		)
	}

	utils.Copy(existingInventory, updateBody)

	if err := s.repo.UpdateInventory(ctx, id, existingInventory); err != nil {
		return nil, fmt.Errorf("failed to update inventory: %w", err)
	}

	return existingInventory, nil
}

func (s *InventoryService) DeleteInventory(ctx context.Context, id string) error {
	return s.repo.DeleteInventory(ctx, id)
}

func (s *InventoryService) GetInventoryByID(ctx context.Context, id string) (*models.Inventory, error) {
	existInventory, err := s.repo.GetInventoryByID(ctx, id)
	if err != nil {
		return nil, errors.NewNotFoundError("inventory", id)
	}

	return existInventory, nil
}
