package services

import (
	"context"

	"github.com/devbenho/luka-platform/internal/store/repositories"
	"github.com/devbenho/luka-platform/pkg/errors"
)

// RequireStoreOwner returns nil if userID owns the (non-deleted) store,
// a not-found error if the store does not exist, and a forbidden error
// otherwise. Other modules use it to authorise writes to store-owned data
// (products, inventory, order fulfilment).
func RequireStoreOwner(ctx context.Context, repo repositories.IStoreRepository, storeID, userID string) error {
	store, err := repo.GetStoreByID(ctx, storeID)
	if err != nil {
		return errors.Wrap(err, "fetching store")
	}
	if store == nil || store.DeletedAt != nil {
		return errors.NewNotFoundError("store", storeID)
	}
	if store.OwnerId.Hex() != userID {
		return errors.NewForbiddenError("you do not own this store")
	}
	return nil
}
