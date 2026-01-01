package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Session represents a test session
type Session struct {
	ID             primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	UserID         primitive.ObjectID   `bson:"user_id" json:"user_id"`
	StartedAt      time.Time            `bson:"started_at" json:"started_at"`
	FinishedAt     *time.Time           `bson:"finished_at,omitempty" json:"finished_at,omitempty"`
	TotalScore     int                  `bson:"total_score" json:"total_score"`
	TotalQuestions int                  `bson:"total_questions" json:"total_questions"`
	Status         string               `bson:"status" json:"status"`             // "in_progress", "completed"
	CurrentIdx     int                  `bson:"current_idx" json:"current_idx"`   // Current question index
	QuestionIDs    []primitive.ObjectID `bson:"question_ids" json:"question_ids"` // List of question IDs in order
}
