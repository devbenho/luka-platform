package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/devbenho/luka-platform/internal/store/models"
	"github.com/devbenho/luka-platform/pkg/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type IStoreRepository interface {
	CreateStore(ctx context.Context, store *models.Store) (*models.Store, error)
	GetStoreByID(ctx context.Context, id string) (*models.Store, error)
	UpdateStore(ctx context.Context, id string, store *models.Store) error
	DeleteStore(ctx context.Context, id string) error
}

type StoreRepository struct {
	db database.IDatabase
}

func NewStoreRepository(db database.IDatabase) IStoreRepository {
	return &StoreRepository{
		db: db,
	}
}

func (r *StoreRepository) CreateStore(ctx context.Context, store *models.Store) (*models.Store, error) {
	store.ID = primitive.NewObjectID()
	store.CreatedAt = time.Now()
	store.UpdatedAt = time.Now()

	if err := store.Validate(); err != nil {
		return nil, err
	}

	if err := r.db.Create(ctx, "stores", store); err != nil {
		return nil, err
	}

	return store, nil
}

func (r *StoreRepository) GetStoreByID(ctx context.Context, id string) (*models.Store, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid store ID")
	}
	var store models.Store
	filter := bson.M{"_id": objID}
	if err := r.db.FindOne(ctx, "stores", filter, &store); err != nil {
		return nil, err
	}

	return &store, nil
}

func (r *StoreRepository) UpdateStore(ctx context.Context, id string, store *models.Store) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid store ID")
	}
	store.UpdatedAt = time.Now()
	filter := bson.M{"_id": objID}
	update := bson.M{
		"$set": bson.M{
			"name":       store.Name,
			"owner":      store.OwnerId,
			"slug":       store.Slug,
			"updated_at": store.UpdatedAt,
		},
	}

	r.db.Update(ctx, "stores", filter, update)
	return nil
}

func (r *StoreRepository) DeleteStore(ctx context.Context, id string) error {
	return nil
}
