package repositories

import (
	"context"
	"errors"
	"github.com/devbenho/bazar-user-service/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

// IUserRepository defines the methods that any repository implementation must have.
type IUserRepository interface {
	CreateUser(user *models.User) error
	GetUserByID(id string) (*models.User, error)
	IsUserExists(login string) (bool, error)
	UpdateUser(id string, user *models.User) error
	DeleteUser(id string) error
}

type userRepository struct {
	collection *mongo.Collection
}

func NewUserRepository(db *mongo.Database) IUserRepository {
	return &userRepository{
		collection: db.Collection("users"),
	}
}

// CreateUser inserts a new user into the database
func (r *userRepository) CreateUser(user *models.User) error {
	user.ID = primitive.NewObjectID() // Generate a new ObjectID for MongoDB
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	_, err := r.collection.InsertOne(context.TODO(), user)
	if err != nil {
		return err
	}
	return nil
}

// GetUserByID fetches a user by their ID from the database
func (r *userRepository) GetUserByID(id string) (*models.User, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}

	var user models.User
	filter := bson.M{"_id": objID}
	err = r.collection.FindOne(context.TODO(), filter).Decode(&user)
	if err == mongo.ErrNoDocuments {
		return nil, errors.New("user not found")
	} else if err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateUser updates an existing user in the database
func (r *userRepository) UpdateUser(id string, user *models.User) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid user ID")
	}

	user.UpdatedAt = time.Now()
	filter := bson.M{"_id": objID}
	update := bson.M{
		"$set": bson.M{
			//"first_name": user.FirstName,
			//"last_name":  user.LastName,
			"email":      user.Email,
			"updated_at": user.UpdatedAt,
		},
	}

	_, err = r.collection.UpdateOne(context.TODO(), filter, update, options.Update().SetUpsert(false))
	if err != nil {
		return err
	}
	return nil
}

// DeleteUser deletes a user by their ID from the database
func (r *userRepository) DeleteUser(id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid user ID")
	}

	filter := bson.M{"_id": objID}
	_, err = r.collection.DeleteOne(context.TODO(), filter)
	if err != nil {
		return err
	}
	return nil
}

// IsUserExists checks if a user with the given login already exists in the database
func (r *userRepository) IsUserExists(login string) (bool, error) {
	filter := bson.M{"username": login}
	count, err := r.collection.CountDocuments(context.Background(), filter)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
