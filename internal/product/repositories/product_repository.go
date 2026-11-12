package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/devbenho/luka-platform/internal/product/models"
	"github.com/devbenho/luka-platform/pkg/database"
	apperrors "github.com/devbenho/luka-platform/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type IProductRepository interface {
	CreateProduct(ctx context.Context, product *models.Product) (*models.Product, error)
	// GetProductByID returns (nil, nil) when no product has the given ID.
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

func parseProductID(id string) (primitive.ObjectID, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return primitive.NilObjectID, apperrors.NewBadRequestError("invalid product ID")
	}
	return objID, nil
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
	objID, err := parseProductID(id)
	if err != nil {
		return nil, err
	}
	var product models.Product
	filter := bson.M{"_id": objID}
	if err := r.db.FindOne(ctx, "products", filter, &product); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &product, nil
}

func (r *ProductRepository) UpdateProduct(ctx context.Context, id string, product *models.Product) error {
	objID, err := parseProductID(id)
	if err != nil {
		return err
	}
	product.UpdatedAt = time.Now()
	filter := bson.M{"_id": objID, "deleted_at": nil}
	update := bson.M{
		"$set": bson.M{
			"name":        product.Name,
			"description": product.Description,
			"price":       product.Price,
			"categories":  product.Categories,
			"images":      product.Images,
			"updated_at":  product.UpdatedAt,
		},
	}
	return r.db.Update(ctx, "products", filter, update)
}

func (r *ProductRepository) DeleteProduct(ctx context.Context, id string) error {
	objID, err := parseProductID(id)
	if err != nil {
		return err
	}
	filter := bson.M{"_id": objID, "deleted_at": nil}
	return r.db.SoftDelete(ctx, "products", "deleted_at", filter)
}
