package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/devbenho/luka-platform/internal/store/models"
	"github.com/devbenho/luka-platform/pkg/database"
	apperrors "github.com/devbenho/luka-platform/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// Store has no bson tags, so the driver stores each field under its
// lowercased Go name (OwnerId -> "ownerid", DeletedAt -> "deletedat").
// Queries must use those names to match existing documents.
const (
	fieldName      = "name"
	fieldSlug      = "slug"
	fieldUpdatedAt = "updatedat"
	fieldDeletedAt = "deletedat"
)

type IStoreRepository interface {
	CreateStore(ctx context.Context, store *models.Store) (*models.Store, error)
	// GetStoreByID returns (nil, nil) when no store has the given ID.
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

func parseStoreID(id string) (primitive.ObjectID, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return primitive.NilObjectID, apperrors.NewBadRequestError("invalid store ID")
	}
	return objID, nil
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
	objID, err := parseStoreID(id)
	if err != nil {
		return nil, err
	}
	var store models.Store
	filter := bson.M{"_id": objID}
	if err := r.db.FindOne(ctx, "stores", filter, &store); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}

	return &store, nil
}

func (r *StoreRepository) UpdateStore(ctx context.Context, id string, store *models.Store) error {
	objID, err := parseStoreID(id)
	if err != nil {
		return err
	}
	store.UpdatedAt = time.Now()
	filter := bson.M{"_id": objID, fieldDeletedAt: nil}
	update := bson.M{
		"$set": bson.M{
			fieldName:      store.Name,
			fieldSlug:      store.Slug,
			fieldUpdatedAt: store.UpdatedAt,
		},
	}
	return r.db.Update(ctx, "stores", filter, update)
}

func (r *StoreRepository) DeleteStore(ctx context.Context, id string) error {
	objID, err := parseStoreID(id)
	if err != nil {
		return err
	}
	filter := bson.M{"_id": objID, fieldDeletedAt: nil}
	return r.db.SoftDelete(ctx, "stores", fieldDeletedAt, filter)
}
