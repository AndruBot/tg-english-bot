package json

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/andru_bot/tg-bot/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// QuestionData represents the structure of the JSON file
type QuestionData struct {
	Questions []QuestionJSON `json:"questions"`
}

// QuestionJSON represents a question in the JSON file
type QuestionJSON struct {
	Text            string `json:"text"`
	TextHTML        string `json:"text_html"`
	Answer1         string `json:"answer_1"`
	Answer1HTML     string `json:"answer_1_html"`
	Answer2         string `json:"answer_2"`
	Answer2HTML     string `json:"answer_2_html"`
	Answer3         string `json:"answer_3"`
	Answer3HTML     string `json:"answer_3_html"`
	Answer4         string `json:"answer_4"`
	Answer4HTML     string `json:"answer_4_html"`
	CorrectAnswerID int    `json:"correct_answer_id"`
	Score           int    `json:"score"`
}

// LoadQuestions loads questions from JSON file
func LoadQuestions(filename string) ([]models.Question, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open JSON file: %w", err)
	}
	defer file.Close()

	var data QuestionData
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode JSON file: %w", err)
	}

	if len(data.Questions) == 0 {
		return nil, fmt.Errorf("JSON file must contain at least one question")
	}

	var questions []models.Question
	for i, qJSON := range data.Questions {
		// Validate correct answer ID is within range
		maxAnswerID := 3
		if qJSON.Answer4 != "" {
			maxAnswerID = 4
		}
		if qJSON.CorrectAnswerID < 1 || qJSON.CorrectAnswerID > maxAnswerID {
			return nil, fmt.Errorf("invalid correct_answer_id at question %d: must be between 1 and %d (question has %d answers)", i+1, maxAnswerID, maxAnswerID)
		}

		question := models.Question{
			ID:              primitive.NewObjectID(),
			Text:            qJSON.Text,
			TextHTML:        qJSON.TextHTML,
			Answer1:         qJSON.Answer1,
			Answer2:         qJSON.Answer2,
			Answer3:         qJSON.Answer3,
			Answer4:         qJSON.Answer4,
			CorrectAnswerID: qJSON.CorrectAnswerID,
			Score:           qJSON.Score,
		}

		questions = append(questions, question)
	}

	return questions, nil
}
