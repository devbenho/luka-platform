package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/devbenho/luka-platform/internal/product/models"
	"github.com/devbenho/luka-platform/pkg/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type IProductRepository interface {
	CreateProduct(ctx context.Context, product *models.Product) (*models.Product, error)
	GetProductByID(ctx context.Context, id string) (*models.Product, error)
	UpdateProduct(ctx context.Context, id string, product *models.Product) error
	DeleteProduct(ctx context.Context, id string) error
}

type ProductRepository struct {
	db database.IDatabase
}

func NewProductRepository(db database.IDatabase) IProductRepository {
	return &ProductRepository{
		db: db,
	}
}

func (r *ProductRepository) CreateProduct(ctx context.Context, product *models.Product) (*models.Product, error) {
	product.ID = primitive.NewObjectID()
	product.CreatedAt = time.Now()
	product.UpdatedAt = time.Now()
	if err := product.Validate(); err != nil {
		return nil, err
	}
	if err := r.db.Create(ctx, "products", product); err != nil {
		return nil, err
	}
	return product, nil
}

func (r *ProductRepository) GetProductByID(ctx context.Context, id string) (*models.Product, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid product ID")
	}
	var product models.Product
	filter := bson.M{"_id": objID}
	if err := r.db.FindOne(ctx, "products", filter, &product); err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *ProductRepository) UpdateProduct(ctx context.Context, id string, product *models.Product) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid product ID")
	}
	product.UpdatedAt = time.Now()
	filter := bson.M{"_id": objID}
	update := bson.M{
		"$set": bson.M{
			"name":        product.Name,
			"description": product.Description,
			"price":       product.Price,
			"updated_at":  product.UpdatedAt,
		},
	}
	return r.db.Update(ctx, "products", filter, update)
}

func (r *ProductRepository) DeleteProduct(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid product ID")
	}
	filter := bson.M{"_id": objID}
	return r.db.SoftDelete(ctx, "products", filter)
}
