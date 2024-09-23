package database

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"os"
	"time"
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
		return fmt.Errorf("MONGO_URI and DB_NAME environment variables must be set")
	}

	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.NewClient(clientOptions)
	if err != nil {
		return fmt.Errorf("failed to create MongoDB client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err = client.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	log.Println("Connected to MongoDB!")
	Client = client
	Database = client.Database(dbName)
	return nil
}

// Disconnect closes the MongoDB connection.
func Disconnect() error {
	if err := Client.Disconnect(context.TODO()); err != nil {
		return fmt.Errorf("failed to disconnect from MongoDB: %w", err)
	}
	log.Println("Disconnected from MongoDB.")
	return nil
}
