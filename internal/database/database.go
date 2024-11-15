package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	Client   *mongo.Client
	Database *mongo.Database
)

// Connect initializes the MongoDB client and connects to the database.
func Connect() error {
	uri := os.Getenv("MONGO_URI")
	dbName := os.Getenv("DB_NAME")

	if uri == "" || dbName == "" {
		return &DBConnectionError{
			Operation: "initialization",
			Err:       fmt.Errorf("MONGO_URI and DB_NAME environment variables must be set"),
		}
	}

	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.NewClient(clientOptions)
	if err != nil {
		return &DBConnectionError{
			Operation: "creating client",
			Err:       fmt.Errorf("failed to create MongoDB client: %w", err),
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err = client.Connect(ctx); err != nil {
		return &DBConnectionError{
			Operation: "connecting to MongoDB",
			Err:       fmt.Errorf("failed to connect to MongoDB: %w", err),
		}
	}

	log.Println("Connected to MongoDB!")
	Client = client
	Database = client.Database(dbName)

	// Ensure unique indexes on username and email
	if err := ensureIndexes(Database); err != nil {
		return &DBConnectionError{
			Operation: "creating indexes",
			Err:       fmt.Errorf("failed to create indexes: %w", err),
		}
	}

	return nil
}

// ensureIndexes creates unique indexes on the username and email fields.
func ensureIndexes(db *mongo.Database) error {
	userCollection := db.Collection("users")

	indexModels := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "username", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys:    bson.D{{Key: "email", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	}

	_, err := userCollection.Indexes().CreateMany(context.Background(), indexModels)
	return err
}

// Disconnect closes the MongoDB connection.
func Disconnect() error {
	if err := Client.Disconnect(context.TODO()); err != nil {
		return &DBConnectionError{
			Operation: "disconnecting from MongoDB",
			Err:       fmt.Errorf("failed to disconnect from MongoDB: %w", err),
		}
	}
	log.Println("Disconnected from MongoDB.")
	return nil
}
