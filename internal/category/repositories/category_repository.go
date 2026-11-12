package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/devbenho/luka-platform/internal/category/models"
	"github.com/devbenho/luka-platform/pkg/database"
	apperrors "github.com/devbenho/luka-platform/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// fieldDeletedAt is the bson name of Category.DeletedAt ("delete_at" for
// historical reasons; kept so existing documents still match).
const fieldDeletedAt = "delete_at"

type ICategoryRepository interface {
	CreateCategory(ctx context.Context, category *models.Category) (*models.Category, error)
	// GetCategoryByID returns (nil, nil) when no category has the given ID.
	GetCategoryByID(ctx context.Context, id string) (*models.Category, error)
	UpdateCategory(ctx context.Context, id string, category *models.Category) error
	DeleteCategory(ctx context.Context, id string) error
}

type CategoryRepository struct {
	db database.IDatabase
}

func NewCategoryRepository(db database.IDatabase) ICategoryRepository {
	return &CategoryRepository{
		db: db,
	}
}

func parseCategoryID(id string) (primitive.ObjectID, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return primitive.NilObjectID, apperrors.NewBadRequestError("invalid category ID")
	}
	return objID, nil
}

func (r *CategoryRepository) CreateCategory(ctx context.Context, category *models.Category) (*models.Category, error) {
	category.ID = primitive.NewObjectID()
	category.CreatedAt = time.Now()
	category.UpdatedAt = time.Now()
	if err := category.Validate(); err != nil {
		return nil, err
	}
	if err := r.db.Create(ctx, "categories", category); err != nil {
		return nil, err
	}
	return category, nil
}

func (r *CategoryRepository) GetCategoryByID(ctx context.Context, id string) (*models.Category, error) {
	objID, err := parseCategoryID(id)
	if err != nil {
		return nil, err
	}
	var category models.Category
	filter := bson.M{"_id": objID}
	if err := r.db.FindOne(ctx, "categories", filter, &category); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &category, nil
}

func (r *CategoryRepository) UpdateCategory(ctx context.Context, id string, category *models.Category) error {
	objID, err := parseCategoryID(id)
	if err != nil {
		return err
	}
	category.UpdatedAt = time.Now()
	filter := bson.M{"_id": objID, fieldDeletedAt: nil}
	update := bson.M{
		"$set": bson.M{
			"name":        category.Name,
			"slug":        category.Slug,
			"description": category.Description,
			"parent_id":   category.ParentID,
			"updated_at":  category.UpdatedAt,
		},
	}
	return r.db.Update(ctx, "categories", filter, update)
}

func (r *CategoryRepository) DeleteCategory(ctx context.Context, id string) error {
	objID, err := parseCategoryID(id)
	if err != nil {
		return err
	}
	filter := bson.M{"_id": objID, fieldDeletedAt: nil}
	return r.db.SoftDelete(ctx, "categories", fieldDeletedAt, filter)
}
