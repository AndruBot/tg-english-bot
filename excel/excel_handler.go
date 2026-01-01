package excel

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/andru_bot/tg-bot/models"
	"github.com/xuri/excelize/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CreateResultsExcel creates an Excel file with test results
func CreateResultsExcel(sessionID primitive.ObjectID, answers []models.Answer, questions []models.Question) (string, error) {
	// Create a map of question IDs to questions for quick lookup
	questionMap := make(map[primitive.ObjectID]models.Question)
	for _, q := range questions {
		questionMap[q.ID] = q
	}

	// Create a map of question IDs to answers for quick lookup
	answerMap := make(map[primitive.ObjectID]models.Answer)
	for _, a := range answers {
		answerMap[a.QuestionID] = a
	}

	// Create new Excel file
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Printf("Error closing Excel file: %v\n", err)
		}
	}()

	// Set sheet name
	sheetName := "Results"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return "", fmt.Errorf("failed to create sheet: %w", err)
	}
	f.SetActiveSheet(index)

	// Delete default sheet
	f.DeleteSheet("Sheet1")

	// Set headers
	headers := []string{"Question", "Answer 1", "Answer 2", "Answer 3", "Answer 4", "Correct Answer", "User Answer", "Result"}
	for i, header := range headers {
		cell := fmt.Sprintf("%c1", 'A'+i)
		f.SetCellValue(sheetName, cell, header)
	}

	// Style headers
	headerStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#E0E0E0"}, Pattern: 1},
	})
	if err == nil {
		f.SetCellStyle(sheetName, "A1", fmt.Sprintf("%c1", 'A'+len(headers)-1), headerStyle)
	}

	// Write data rows
	row := 2
	for _, question := range questions {
		answer, answered := answerMap[question.ID]

		// Question text
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), question.Text)

		// Answer options
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), question.Answer1)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), question.Answer2)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), question.Answer3)
		// Answer4 can be empty for 3-answer questions
		if question.Answer4 != "" {
			f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), question.Answer4)
		} else {
			f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), "")
		}

		// Correct answer
		correctAnswerText := question.GetAnswer(question.CorrectAnswerID)
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), correctAnswerText)

		// User answer
		if answered {
			userAnswerText := question.GetAnswer(answer.SelectedAnswerID)
			f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), userAnswerText)

			// Result (+ or -)
			if answer.IsCorrect {
				f.SetCellValue(sheetName, fmt.Sprintf("H%d", row), "+")
			} else {
				f.SetCellValue(sheetName, fmt.Sprintf("H%d", row), "-")
			}
		} else {
			f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), "Not answered")
			f.SetCellValue(sheetName, fmt.Sprintf("H%d", row), "-")
		}

		row++
	}

	// Auto-fit columns
	for i := 0; i < len(headers); i++ {
		col := string(rune('A' + i))
		f.SetColWidth(sheetName, col, col, 20)
	}

	// Save file
	filename := fmt.Sprintf("results_%s.xlsx", sessionID.Hex())
	filepath := filepath.Join(os.TempDir(), filename)
	if err := f.SaveAs(filepath); err != nil {
		return "", fmt.Errorf("failed to save Excel file: %w", err)
	}

	return filepath, nil
}

// CreateResultsExcelWithSkipped creates an Excel file with test results, marking unanswered questions as "skip"
func CreateResultsExcelWithSkipped(sessionID primitive.ObjectID, answers []models.Answer, questions []models.Question, currentIdx int) (string, error) {
	// Create a map of question IDs to answers for quick lookup
	answerMap := make(map[primitive.ObjectID]models.Answer)
	for _, a := range answers {
		answerMap[a.QuestionID] = a
	}

	// Create new Excel file
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Printf("Error closing Excel file: %v\n", err)
		}
	}()

	// Set sheet name
	sheetName := "Results"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return "", fmt.Errorf("failed to create sheet: %w", err)
	}
	f.SetActiveSheet(index)

	// Delete default sheet
	f.DeleteSheet("Sheet1")

	// Set headers
	headers := []string{"Question", "Answer 1", "Answer 2", "Answer 3", "Answer 4", "Correct Answer", "User Answer", "Result"}
	for i, header := range headers {
		cell := fmt.Sprintf("%c1", 'A'+i)
		f.SetCellValue(sheetName, cell, header)
	}

	// Style headers
	headerStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#E0E0E0"}, Pattern: 1},
	})
	if err == nil {
		f.SetCellStyle(sheetName, "A1", fmt.Sprintf("%c1", 'A'+len(headers)-1), headerStyle)
	}

	// Write data rows
	row := 2
	for i, question := range questions {
		answer, answered := answerMap[question.ID]

		// Question text
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), question.Text)

		// Answer options
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), question.Answer1)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), question.Answer2)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), question.Answer3)
		// Answer4 can be empty for 3-answer questions
		if question.Answer4 != "" {
			f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), question.Answer4)
		} else {
			f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), "")
		}

		// Correct answer
		correctAnswerText := question.GetAnswer(question.CorrectAnswerID)
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), correctAnswerText)

		// User answer and result
		if answered {
			userAnswerText := question.GetAnswer(answer.SelectedAnswerID)
			f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), userAnswerText)

			// Result (+ or -)
			if answer.IsCorrect {
				f.SetCellValue(sheetName, fmt.Sprintf("H%d", row), "+")
			} else {
				f.SetCellValue(sheetName, fmt.Sprintf("H%d", row), "-")
			}
		} else {
			// Check if this question was skipped (not answered due to early test failure)
			// currentIdx is the first unanswered question index (after the last answered question)
			if i >= currentIdx {
				f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), "Not answered")
				f.SetCellValue(sheetName, fmt.Sprintf("H%d", row), "skip")
			} else {
				f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), "Not answered")
				f.SetCellValue(sheetName, fmt.Sprintf("H%d", row), "-")
			}
		}

		row++
	}

	// Auto-fit columns
	for i := 0; i < len(headers); i++ {
		col := string(rune('A' + i))
		f.SetColWidth(sheetName, col, col, 20)
	}

	// Save file
	filename := fmt.Sprintf("results_%s.xlsx", sessionID.Hex())
	filepath := filepath.Join(os.TempDir(), filename)
	if err := f.SaveAs(filepath); err != nil {
		return "", fmt.Errorf("failed to save Excel file: %w", err)
	}

	return filepath, nil
}
