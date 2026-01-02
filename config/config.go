package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

// GetMaxConsecutiveErrors returns the maximum allowed consecutive errors before test failure
// Defaults to 5 if MAX_CONSECUTIVE_ERRORS environment variable is not set or invalid
func GetMaxConsecutiveErrors() int {
	maxErrorsStr := os.Getenv("MAX_CONSECUTIVE_ERRORS")
	if maxErrorsStr == "" {
		return 5 // Default value
	}

	maxErrors, err := strconv.Atoi(maxErrorsStr)
	if err != nil {
		log.Printf("Error parsing MAX_CONSECUTIVE_ERRORS: %v, using default value 5", err)
		return 5
	}

	if maxErrors < 1 {
		log.Printf("MAX_CONSECUTIVE_ERRORS must be at least 1, using default value 5")
		return 5
	}

	return maxErrors
}

// GetTelegramBotToken returns the Telegram bot token from environment
// Returns error if TELEGRAM_BOT_TOKEN is not set (required)
func GetTelegramBotToken() (string, error) {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		return "", fmt.Errorf("TELEGRAM_BOT_TOKEN environment variable is required")
	}
	return token, nil
}

// GetMongoDBURI returns the MongoDB connection URI
// If MONGODB_URI is set, returns it directly
// Otherwise, builds URI from individual environment variables with defaults
func GetMongoDBURI() string {
	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI != "" {
		return mongoURI
	}

	// Build connection string from individual environment variables
	username := GetMongoRootUsername()
	password := GetMongoRootPassword()
	host := GetMongoHost()
	port := GetMongoPort()
	database := GetMongoDatabase()

	if username != "" && password != "" {
		return fmt.Sprintf("mongodb://%s:%s@%s:%s/%s?authSource=admin", username, password, host, port, database)
	}
	return fmt.Sprintf("mongodb://%s:%s/%s", host, port, database)
}

// GetMongoHost returns MongoDB host, defaults to "localhost"
func GetMongoHost() string {
	host := os.Getenv("MONGO_HOST")
	if host == "" {
		return "localhost"
	}
	return host
}

// GetMongoPort returns MongoDB port, defaults to "27017"
// Checks MONGODB_PORT first, then MONGO_PORT for backward compatibility
func GetMongoPort() string {
	port := os.Getenv("MONGODB_PORT")
	if port == "" {
		port = os.Getenv("MONGO_PORT")
	}
	if port == "" {
		return "27017"
	}
	return port
}

// GetMongoRootUsername returns MongoDB root username, empty string if not set
func GetMongoRootUsername() string {
	return os.Getenv("MONGO_ROOT_USERNAME")
}

// GetMongoRootPassword returns MongoDB root password, empty string if not set
func GetMongoRootPassword() string {
	return os.Getenv("MONGO_ROOT_PASSWORD")
}

// GetMongoDatabase returns MongoDB database name, defaults to "english_test_bot"
func GetMongoDatabase() string {
	database := os.Getenv("MONGO_DATABASE")
	if database == "" {
		return "english_test_bot"
	}
	return database
}

// GetEnvFile returns the environment file path based on the ENV variable
// Returns ".env.prod" for production, ".env.dev" for development, or ".env" as default
func GetEnvFile() string {
	env := strings.ToLower(os.Getenv("ENV"))

	switch env {
	case "production", "prod":
		return ".env.prod"
	case "development", "dev", "develop":
		return ".env.dev"
	default:
		// Default to .env for backward compatibility
		return ".env"
	}
}
