package csv

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"

	"github.com/andru_bot/tg-bot/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// LoadQuestions loads questions from CSV file
func LoadQuestions(filename string) ([]models.Question, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV file: %w", err)
	}

	if len(records) < 2 {
		return nil, fmt.Errorf("CSV file must have at least a header and one question")
	}

	var questions []models.Question
	for i, record := range records[1:] { // Skip header
		// New CSV format: text, text_html, answer_1, answer_1_html, answer_2, answer_2_html, answer_3, answer_3_html, answer_4, answer_4_html, correct_answer_id, score
		// We only need: text, answer_1, answer_2, answer_3, answer_4, correct_answer_id, score
		if len(record) < 12 {
			return nil, fmt.Errorf("invalid CSV format at line %d: expected at least 12 columns (text, text_html, answer_1, answer_1_html, answer_2, answer_2_html, answer_3, answer_3_html, answer_4, answer_4_html, correct_answer_id, score)", i+2)
		}

		// Extract fields from new format
		text := record[0]                // text
		textHTML := record[1]            // text_html
		answer1 := record[2]             // answer_1
		answer2 := record[4]             // answer_2 (skip answer_1_html at index 3)
		answer3 := record[6]             // answer_3 (skip answer_2_html at index 5)
		answer4 := record[8]             // answer_4 (skip answer_3_html at index 7)
		correctAnswerIDStr := record[10] // correct_answer_id (skip answer_4_html at index 9)
		scoreStr := record[11]           // score

		// Answer4 is optional - can be empty string for 3-answer questions
		if answer4 == "" {
			answer4 = ""
		}

		// Parse correct answer ID
		correctAnswerID, err := strconv.Atoi(correctAnswerIDStr)
		if err != nil {
			return nil, fmt.Errorf("invalid correct_answer_id at line %d: %w", i+2, err)
		}

		// Validate correct answer ID is within range
		maxAnswerID := 3
		if answer4 != "" {
			maxAnswerID = 4
		}
		if correctAnswerID < 1 || correctAnswerID > maxAnswerID {
			return nil, fmt.Errorf("invalid correct_answer_id at line %d: must be between 1 and %d (question has %d answers)", i+2, maxAnswerID, maxAnswerID)
		}

		// Parse score
		score, err := strconv.Atoi(scoreStr)
		if err != nil {
			return nil, fmt.Errorf("invalid score at line %d: %w", i+2, err)
		}

		question := models.Question{
			ID:              primitive.NewObjectID(),
			Text:            text,
			TextHTML:        textHTML,
			Answer1:         answer1,
			Answer2:         answer2,
			Answer3:         answer3,
			Answer4:         answer4,
			CorrectAnswerID: correctAnswerID,
			Score:           score,
		}

		questions = append(questions, question)
	}

	return questions, nil
}

// SaveResults saves test results to CSV file
func SaveResults(filename string, sessionID primitive.ObjectID, answers []models.Answer, questions []models.Question) error {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open results file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Create a map of question IDs to questions for quick lookup
	questionMap := make(map[primitive.ObjectID]models.Question)
	for _, q := range questions {
		questionMap[q.ID] = q
	}

	// Write header if file is empty
	stat, _ := file.Stat()
	if stat.Size() == 0 {
		header := []string{"session_id", "question_text", "answer_1", "answer_2", "answer_3", "answer_4", "correct_answer_id", "user_answer_id", "is_correct", "score"}
		if err := writer.Write(header); err != nil {
			return fmt.Errorf("failed to write header: %w", err)
		}
	}

	// Write results
	for _, answer := range answers {
		question, exists := questionMap[answer.QuestionID]
		if !exists {
			continue
		}

		record := []string{
			sessionID.Hex(),
			question.Text,
			question.Answer1,
			question.Answer2,
			question.Answer3,
			question.Answer4,
			strconv.Itoa(question.CorrectAnswerID),
			strconv.Itoa(answer.SelectedAnswerID),
			strconv.FormatBool(answer.IsCorrect),
			strconv.Itoa(answer.Score),
		}

		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write record: %w", err)
		}
	}

	return nil
}
