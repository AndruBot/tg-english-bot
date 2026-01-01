package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Question represents a question from CSV
type Question struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Text            string             `bson:"text" json:"text"`
	TextHTML        string             `bson:"text_html,omitempty" json:"text_html,omitempty"` // HTML formatted text for Telegram
	Answer1         string             `bson:"answer_1" json:"answer_1"`
	Answer2         string             `bson:"answer_2" json:"answer_2"`
	Answer3         string             `bson:"answer_3" json:"answer_3"`
	Answer4         string             `bson:"answer_4" json:"answer_4"` // Can be empty for 3-answer questions
	CorrectAnswerID int                `bson:"correct_answer_id" json:"correct_answer_id"`
	Score           int                `bson:"score" json:"score"`
}

// GetAnswerCount returns the number of available answers (3 or 4)
func (q *Question) GetAnswerCount() int {
	if q.Answer4 == "" {
		return 3
	}
	return 4
}

// GetAnswer returns the answer text for the given answer ID (1-4)
func (q *Question) GetAnswer(answerID int) string {
	switch answerID {
	case 1:
		return q.Answer1
	case 2:
		return q.Answer2
	case 3:
		return q.Answer3
	case 4:
		return q.Answer4
	default:
		return ""
	}
}
