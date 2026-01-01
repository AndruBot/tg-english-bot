package bot

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/andru_bot/tg-bot/models"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (h *BotHandler) handleCallbackQuery(query *tgbotapi.CallbackQuery) {
	userID := query.From.ID
	session, exists := h.activeSessions[userID]

	// If session not in memory, try to load from database
	if !exists {
		// Find or create user
		user, err := h.userRepo.FindOrCreate(
			int64(userID),
			query.From.UserName,
			query.From.FirstName,
			query.From.LastName,
		)
		if err != nil {
			log.Printf("Error finding/creating user: %v", err)
			h.answerCallback(query.ID, "Error processing answer. Please try again.")
			return
		}

		// Try to load active session from database
		dbSession, err := h.sessionRepo.GetActiveByUserID(user.ID)
		if err != nil || dbSession == nil {
			h.answerCallback(query.ID, "Your test session has expired. Please start a new test with /test.")
			return
		}

		// Load session into memory
		h.activeSessions[userID] = &ActiveSession{
			SessionID:         dbSession.ID,
			UserID:            dbSession.UserID,
			QuestionIDs:       dbSession.QuestionIDs,
			CurrentIdx:        dbSession.CurrentIdx,
			Score:             dbSession.TotalScore,
			ConsecutiveErrors: 0, // Reset when loading from DB (we'll recalculate if needed)
		}
		session = h.activeSessions[userID]
	}

	// Parse selected answer
	selectedAnswerID, err := strconv.Atoi(query.Data)
	if err != nil {
		h.answerCallback(query.ID, "Invalid answer. Please try again.")
		return
	}

	// Get current question
	if session.CurrentIdx >= len(session.QuestionIDs) {
		h.answerCallback(query.ID, "Test already completed.")
		return
	}

	questionID := session.QuestionIDs[session.CurrentIdx]
	question, err := h.questionRepo.GetByID(questionID)
	if err != nil {
		log.Printf("Error getting question: %v", err)
		h.answerCallback(query.ID, "Error processing answer. Please try again.")
		return
	}

	// Validate selected answer ID is within range
	maxAnswerID := question.GetAnswerCount()
	if selectedAnswerID < 1 || selectedAnswerID > maxAnswerID {
		h.answerCallback(query.ID, "Invalid answer. Please try again.")
		return
	}

	// Check if answer is correct
	isCorrect := selectedAnswerID == question.CorrectAnswerID
	score := 0
	if isCorrect {
		score = question.Score
		session.Score += score
		// Reset consecutive errors on correct answer
		session.ConsecutiveErrors = 0
	} else {
		// Increment consecutive errors on incorrect answer
		session.ConsecutiveErrors++
	}

	// Save answer
	answer := &models.Answer{
		ID:               primitive.NewObjectID(),
		SessionID:        session.SessionID,
		UserID:           session.UserID,
		QuestionID:       questionID,
		SelectedAnswerID: selectedAnswerID,
		IsCorrect:        isCorrect,
		Score:            score,
		AnsweredAt:       time.Now(),
	}

	err = h.answerRepo.Create(answer)
	if err != nil {
		log.Printf("Error saving answer: %v", err)
	}

	// Acknowledge callback (don't reveal if answer is correct)
	h.answerCallback(query.ID, "")

	// Check if max consecutive errors occurred
	if session.ConsecutiveErrors >= h.maxConsecutiveErrors {
		// Test failed due to consecutive errors
		h.finishTestWithFailure(query.Message.Chat.ID, userID, session)
		return
	}

	// Move to next question
	session.CurrentIdx++

	// Update session progress in database
	err = h.sessionRepo.UpdateProgress(session.SessionID, session.CurrentIdx, session.Score)
	if err != nil {
		log.Printf("Error updating session progress: %v", err)
	}

	// Send next question or finish test
	if session.CurrentIdx >= len(session.QuestionIDs) {
		// Test completed naturally (all questions answered)
		h.finishTest(query.Message.Chat.ID, userID, session, true)
	} else {
		// Send next question immediately
		h.sendNextQuestion(query.Message.Chat.ID, userID)
	}
}

func (h *BotHandler) sendNextQuestion(chatID int64, userID int64) {
	session := h.activeSessions[userID]
	if session == nil {
		return
	}

	if session.CurrentIdx >= len(session.QuestionIDs) {
		return
	}

	questionID := session.QuestionIDs[session.CurrentIdx]
	question, err := h.questionRepo.GetByID(questionID)
	if err != nil {
		log.Printf("Error getting question: %v", err)
		h.sendMessage(chatID, "Error loading question. Please try again.")
		return
	}

	// Create inline keyboard with answer options (3 or 4 answers)
	var keyboard [][]tgbotapi.InlineKeyboardButton
	answers := []string{question.Answer1, question.Answer2, question.Answer3}
	if question.Answer4 != "" {
		answers = append(answers, question.Answer4)
	}

	for i, answer := range answers {
		row := []tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("%d. %s", i+1, answer),
				strconv.Itoa(i+1),
			),
		}
		keyboard = append(keyboard, row)
	}

	questionNum := session.CurrentIdx + 1
	totalQuestions := len(session.QuestionIDs)

	// Format question text with HTML - text already contains newlines from JSON
	text := fmt.Sprintf("<b>Question %d/%d</b>\n\n%s", questionNum, totalQuestions, question.Text)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "HTML"
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(keyboard...)

	_, err = h.bot.Send(msg)
	if err != nil {
		log.Printf("Error sending message: %v", err)
	}
}

func (h *BotHandler) finishTestWithFailure(chatID int64, userID int64, session *ActiveSession) {
	// Finish session in database
	err := h.sessionRepo.Finish(session.SessionID, session.Score, len(session.QuestionIDs))
	if err != nil {
		log.Printf("Error finishing session: %v", err)
	}

	// Update user score
	err = h.userRepo.UpdateScore(session.UserID, session.Score)
	if err != nil {
		log.Printf("Error updating user score: %v", err)
	}

	// Get all answers for this session
	answers, err := h.answerRepo.GetBySession(session.SessionID)
	if err != nil {
		log.Printf("Error getting answers: %v", err)
	}

	// Get questions for this session only
	var questions []models.Question
	for _, questionID := range session.QuestionIDs {
		question, err := h.questionRepo.GetByID(questionID)
		if err != nil {
			log.Printf("Error getting question %s: %v", questionID.Hex(), err)
			continue
		}
		questions = append(questions, *question)
	}

	// Calculate statistics
	correctAnswers := 0
	incorrectAnswers := 0
	for _, answer := range answers {
		if answer.IsCorrect {
			correctAnswers++
		} else {
			incorrectAnswers++
		}
	}
	totalQuestions := len(session.QuestionIDs)

	// Send failure message to user
	failureMessage := fmt.Sprintf("You made %d errors in row, test has been failed", h.maxConsecutiveErrors)
	h.sendMessage(chatID, failureMessage)

	// Show menu again after test failure
	h.sendMessageWithMenu(chatID, "Test failed. Use menu to start a new test.")

	// Send notification to admin with skipped questions marked
	// currentIdx is the question that was just answered (the last error)
	// Mark questions from currentIdx+1 onwards as skip
	h.sendAdminNotificationWithSkipped(userID, session.SessionID, correctAnswers, incorrectAnswers, totalQuestions, answers, questions, session.CurrentIdx+1, h.maxConsecutiveErrors)

	// Delete results.csv file to save space
	h.deleteResultsCSV()

	// Log result to console
	log.Printf("Test failed (%d consecutive errors) - UserID: %d, SessionID: %s, Score: %d/%d",
		h.maxConsecutiveErrors, userID, session.SessionID.Hex(), session.Score, totalQuestions)

	// Remove active session
	delete(h.activeSessions, userID)
}

func (h *BotHandler) finishTest(chatID int64, userID int64, session *ActiveSession, showDetailedResults bool) {
	// Finish session in database
	err := h.sessionRepo.Finish(session.SessionID, session.Score, len(session.QuestionIDs))
	if err != nil {
		log.Printf("Error finishing session: %v", err)
	}

	// Update user score
	err = h.userRepo.UpdateScore(session.UserID, session.Score)
	if err != nil {
		log.Printf("Error updating user score: %v", err)
	}

	// Get all answers for this session
	answers, err := h.answerRepo.GetBySession(session.SessionID)
	if err != nil {
		log.Printf("Error getting answers: %v", err)
	}

	// Get questions for this session only
	var questions []models.Question
	for _, questionID := range session.QuestionIDs {
		question, err := h.questionRepo.GetByID(questionID)
		if err != nil {
			log.Printf("Error getting question %s: %v", questionID.Hex(), err)
			continue
		}
		questions = append(questions, *question)
	}

	// Calculate statistics
	correctAnswers := 0
	incorrectAnswers := 0
	for _, answer := range answers {
		if answer.IsCorrect {
			correctAnswers++
		} else {
			incorrectAnswers++
		}
	}
	totalQuestions := len(session.QuestionIDs)

	// Calculate percentage
	percentage := float64(session.Score) / float64(totalQuestions) * 100.0

	// Send result to user (hide detailed info if manually finished)
	var resultText string
	if showDetailedResults {
		// Show detailed results (when test completes naturally)
		resultText = fmt.Sprintf(
			"🎉 Test Completed!\n\n"+
				"Your Score: %d/%d (%.1f%%)\n\n"+
				"Thank you for taking the test!",
			session.Score,
			totalQuestions,
			percentage,
		)
	} else {
		// Hide detailed results (when manually finished)
		resultText = "✅ Test session finished.\n\n" +
			"Thank you for taking the test!"
	}

	h.sendMessage(chatID, resultText)

	// Show menu again after test completion
	h.sendMessageWithMenu(chatID, "Test completed! Use menu to start a new test or view results.")

	// Send notification to admin
	h.sendAdminNotification(userID, session.SessionID, correctAnswers, incorrectAnswers, totalQuestions, answers, questions)

	// Delete results.csv file to save space
	h.deleteResultsCSV()

	// Log result to console
	log.Printf("Test completed - UserID: %d, SessionID: %s, Score: %d/%d (%.1f%%)",
		userID, session.SessionID.Hex(), session.Score, totalQuestions, percentage)

	// Remove active session
	delete(h.activeSessions, userID)
}
