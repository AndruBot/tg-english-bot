package database

import (
	"context"
	"fmt"
	"time"

	"github.com/andru_bot/tg-bot/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var DB *mongo.Database
var Client *mongo.Client

// Connect initializes MongoDB connection
func Connect() error {
	mongoURI := config.GetMongoDBURI()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Test connection
	err = client.Ping(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	Client = client

	// Get database name from config
	dbName := config.GetMongoDatabase()
	DB = client.Database(dbName)

	fmt.Println("Connected to MongoDB successfully")
	return nil
}

// Disconnect closes MongoDB connection
func Disconnect() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return Client.Disconnect(ctx)
}
