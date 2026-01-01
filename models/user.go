package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User represents a user in the system
type User struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	TelegramID   int64              `bson:"telegram_id" json:"telegram_id"`
	Username     string             `bson:"username,omitempty" json:"username,omitempty"`
	FirstName    string             `bson:"first_name,omitempty" json:"first_name,omitempty"`
	LastName     string             `bson:"last_name,omitempty" json:"last_name,omitempty"`
	TotalScore   int                `bson:"total_score" json:"total_score"`
	TestsTaken   int                `bson:"tests_taken" json:"tests_taken"`
	CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time          `bson:"updated_at" json:"updated_at"`
}

