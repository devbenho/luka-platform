package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/devbenho/luka-platform/internal/category/models"
	"github.com/devbenho/luka-platform/pkg/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ICategoryRepository interface {
	CreateCategory(ctx context.Context, category *models.Category) (*models.Category, error)
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
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid category ID")
	}
	var category models.Category
	filter := bson.M{"_id": objID}
	if err := r.db.FindOne(ctx, "categories", filter, &category); err != nil {
		return nil, err
	}
	return &category, nil
}

func (r *CategoryRepository) UpdateCategory(ctx context.Context, id string, category *models.Category) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid category ID")
	}
	category.UpdatedAt = time.Now()
	filter := bson.M{"_id": objID}
	update := bson.M{
		"$set": bson.M{
			"name":       category.Name,
			"updated_at": category.UpdatedAt,
		},
	}
	return r.db.Update(ctx, "categories", filter, update)
}

func (r *CategoryRepository) DeleteCategory(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid category ID")
	}
	filter := bson.M{"_id": objID}
	return r.db.SoftDelete(ctx, "categories", filter)
}
