package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/devbenho/luka-platform/internal/user/models"
	"github.com/devbenho/luka-platform/pkg/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// IUserRepository defines the methods that any repository implementation must have.
type IUserRepository interface {
	CreateUser(ctx context.Context, user *models.User) error
	GetUserByID(ctx context.Context, id string) (*models.User, error)
	IsUserExists(ctx context.Context, login string) (bool, error)
	UpdateUser(ctx context.Context, id string, user *models.User) error
	DeleteUser(ctx context.Context, id string) error
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)
}

type userRepository struct {
	db database.IDatabase
}

func NewUserRepository(db database.IDatabase) IUserRepository {
	return &userRepository{
		db: db,
	}
}

// CreateUser inserts a new user into the database
func (r *userRepository) CreateUser(ctx context.Context, user *models.User) error {
	user.ID = primitive.NewObjectID()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	return r.db.Create(ctx, "users", user)
}

// GetUserByID fetches a user by their ID from the database
func (r *userRepository) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}

	var user models.User
	filter := bson.M{"_id": objID}
	if err := r.db.FindOne(ctx, "users", filter, &user); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// UpdateUser updates an existing user in the database
func (r *userRepository) UpdateUser(ctx context.Context, id string, user *models.User) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid user ID")
	}

	user.UpdatedAt = time.Now()
	filter := bson.M{"_id": objID}
	update := bson.M{
		"$set": bson.M{
			"email":      user.Email,
			"updated_at": user.UpdatedAt,
		},
	}

	r.db.Update(ctx, "users", filter, update)
	return nil
}

// DeleteUser deletes a user by their ID from the database
func (r *userRepository) DeleteUser(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid user ID")
	}

	filter := bson.M{"_id": objID}

	return r.db.Delete(ctx, "users", filter)
}

// IsUserExists checks if a user with the given login already exists in the database
func (r *userRepository) IsUserExists(ctx context.Context, login string) (bool, error) {
	filter := bson.M{"$or": []bson.M{
		{"email": login},
		{"username": login},
	}}

	count, err := r.db.Count(ctx, "users", filter)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetUserByEmail fetches a user by their email from the database
func (r *userRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	user := &models.User{}

	filter := bson.M{"email": email}
	err := r.db.FindOne(ctx, "users", filter, user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// GetUserByUsername fetches a user by their username from the database
func (r *userRepository) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	user := &models.User{}

	filter := bson.M{"username": username}

	err := r.db.FindOne(ctx, "users", filter, user)
	if err != nil {
		return nil, err
	}
	return user, nil
}
