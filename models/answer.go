package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Answer represents a user's answer to a question
type Answer struct {
	ID               primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	SessionID        primitive.ObjectID `bson:"session_id" json:"session_id"`
	UserID           primitive.ObjectID `bson:"user_id" json:"user_id"`
	QuestionID       primitive.ObjectID `bson:"question_id" json:"question_id"`
	SelectedAnswerID int                `bson:"selected_answer_id" json:"selected_answer_id"`
	IsCorrect        bool               `bson:"is_correct" json:"is_correct"`
	Score            int                `bson:"score" json:"score"`
	AnsweredAt       time.Time          `bson:"answered_at" json:"answered_at"`
}
