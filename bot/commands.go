package bot

import (
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (h *BotHandler) handleCommand(msg *tgbotapi.Message) {
	switch msg.Command() {
	case "start":
		h.handleStart(msg)
	case "help":
		h.handleHelp(msg)
	case "test", "start_test":
		h.handleTestMe(msg)
	case "finish_test":
		h.handleFinishTest(msg)
	case "result":
		h.handleResult(msg)
	default:
		h.sendMessageWithMenu(msg.Chat.ID, "Unknown command. Use /help to see available commands.")
	}
}

func (h *BotHandler) handleMenuButton(msg *tgbotapi.Message) bool {
	text := msg.Text
	switch text {
	case "📚 Start Test", "Start Test":
		h.handleTestMe(msg)
		return true
	case "📊 My Results", "My Results":
		h.handleResult(msg)
		return true
	case "✅ Finish Test", "Finish Test":
		h.handleFinishTest(msg)
		return true
	case "ℹ️ Help", "Help":
		h.handleHelp(msg)
		return true
	}
	return false
}

func (h *BotHandler) handleStart(msg *tgbotapi.Message) {
	text := "Welcome to the English Level Test Bot! 🇬🇧\n\n" +
		"Use the menu buttons below or commands to interact with the bot.\n\n" +
		"Use 'Start Test' to begin the test.\n\n" +
		"The bot will ask you questions one by one. " +
		"Select your answer using the buttons. " +
		"Your answers are kept secret until the end of the test."

	h.sendMessageWithMenu(msg.Chat.ID, text)
}

func (h *BotHandler) handleHelp(msg *tgbotapi.Message) {
	text := "I'll bot to test you\n\n" +
		"Available commands:\n" +
		"📚 Start Test - Start a new test\n" +
		"✅ Finish Test - Finish current test session\n" +
		"📊 My Results - Show results of your last completed test\n" +
		"ℹ️ Help - Show this help message\n\n" +
		"You can use menu buttons or commands: /start_test, /finish_test, /result"

	h.sendMessageWithMenu(msg.Chat.ID, text)
}

func (h *BotHandler) handleTestMe(msg *tgbotapi.Message) {
	userID := msg.From.ID

	// Find or create user
	user, err := h.userRepo.FindOrCreate(
		int64(userID),
		msg.From.UserName,
		msg.From.FirstName,
		msg.From.LastName,
	)
	if err != nil {
		log.Printf("Error finding/creating user: %v", err)
		h.sendMessage(msg.Chat.ID, "Error starting test. Please try again later.")
		return
	}

	// Check if user already has an active session in memory
	if _, exists := h.activeSessions[userID]; exists {
		h.sendMessage(msg.Chat.ID, "You already have an active test session. Please complete it first.")
		return
	}

	// Check for existing active session in database
	dbSession, err := h.sessionRepo.GetActiveByUserID(user.ID)
	if err != nil {
		log.Printf("Error checking for existing session: %v", err)
		h.sendMessage(msg.Chat.ID, "Error starting test. Please try again later.")
		return
	}

	// If there's an active session in DB, resume it
	if dbSession != nil {
		h.activeSessions[userID] = &ActiveSession{
			SessionID:         dbSession.ID,
			UserID:            dbSession.UserID,
			QuestionIDs:       dbSession.QuestionIDs,
			CurrentIdx:        dbSession.CurrentIdx,
			Score:             dbSession.TotalScore,
			ConsecutiveErrors: 0, // Will be recalculated from answers when callback is processed
		}
		h.sendMessage(msg.Chat.ID, "Resuming your test...")
		h.sendNextQuestion(msg.Chat.ID, userID)
		return
	}

	// Get all questions
	questions, err := h.questionRepo.GetAll()
	if err != nil {
		log.Printf("Error getting questions: %v", err)
		h.sendMessage(msg.Chat.ID, "Error loading questions. Please try again later.")
		return
	}

	if len(questions) == 0 {
		h.sendMessage(msg.Chat.ID, "No questions available. Please contact administrator.")
		return
	}

	// Create question IDs list
	questionIDs := make([]primitive.ObjectID, len(questions))
	for i, q := range questions {
		questionIDs[i] = q.ID
	}

	// Create new session in database
	session, err := h.sessionRepo.Create(user.ID, questionIDs)
	if err != nil {
		log.Printf("Error creating session: %v", err)
		h.sendMessage(msg.Chat.ID, "Error starting test. Please try again later.")
		return
	}

	// Create active session in memory
	h.activeSessions[userID] = &ActiveSession{
		SessionID:         session.ID,
		UserID:            user.ID,
		QuestionIDs:       questionIDs,
		CurrentIdx:        0,
		Score:             0,
		ConsecutiveErrors: 0,
	}

	// Remove menu keyboard during test
	h.removeMenu(msg.Chat.ID)

	// Send first question
	h.sendNextQuestion(msg.Chat.ID, userID)
}

func (h *BotHandler) handleFinishTest(msg *tgbotapi.Message) {
	userID := msg.From.ID

	// Check if user has an active session in memory
	session, exists := h.activeSessions[userID]
	if !exists {
		// Try to load from database
		user, err := h.userRepo.FindOrCreate(
			int64(userID),
			msg.From.UserName,
			msg.From.FirstName,
			msg.From.LastName,
		)
		if err != nil {
			log.Printf("Error finding/creating user: %v", err)
			h.sendMessage(msg.Chat.ID, "Error processing request. Please try again later.")
			return
		}

		dbSession, err := h.sessionRepo.GetActiveByUserID(user.ID)
		if err != nil || dbSession == nil {
			h.sendMessage(msg.Chat.ID, "You don't have an active test session.")
			return
		}

		// Load session into memory
		h.activeSessions[userID] = &ActiveSession{
			SessionID:   dbSession.ID,
			UserID:      dbSession.UserID,
			QuestionIDs: dbSession.QuestionIDs,
			CurrentIdx:  dbSession.CurrentIdx,
			Score:       dbSession.TotalScore,
		}
		session = h.activeSessions[userID]
	}

	// Finish the test manually (hide detailed results)
	h.finishTest(msg.Chat.ID, userID, session, false)
}

func (h *BotHandler) handleResult(msg *tgbotapi.Message) {
	userID := msg.From.ID

	// Find or create user
	user, err := h.userRepo.FindOrCreate(
		int64(userID),
		msg.From.UserName,
		msg.From.FirstName,
		msg.From.LastName,
	)
	if err != nil {
		log.Printf("Error finding/creating user: %v", err)
		h.sendMessage(msg.Chat.ID, "Error processing request. Please try again later.")
		return
	}

	// Get last completed session
	session, err := h.sessionRepo.GetLastCompletedByUserID(user.ID)
	if err != nil {
		log.Printf("Error getting last session: %v", err)
		h.sendMessage(msg.Chat.ID, "Error retrieving results. Please try again later.")
		return
	}

	if session == nil {
		h.sendMessage(msg.Chat.ID, "You haven't completed any test yet. Use /start_test to begin.")
		return
	}

	// Get all answers for this session
	answers, err := h.answerRepo.GetBySession(session.ID)
	if err != nil {
		log.Printf("Error getting answers: %v", err)
		h.sendMessage(msg.Chat.ID, "Error retrieving answers. Please try again later.")
		return
	}

	// Calculate statistics
	totalQuestions := session.TotalQuestions
	correctAnswers := 0
	incorrectAnswers := 0

	for _, answer := range answers {
		if answer.IsCorrect {
			correctAnswers++
		} else {
			incorrectAnswers++
		}
	}

	skippedQuestions := totalQuestions - len(answers)
	percentage := float64(correctAnswers) / float64(totalQuestions) * 100.0

	// Format result message
	resultText := fmt.Sprintf(
		"📊 Results of your last test:\n\n"+
			"Total Questions: %d\n"+
			"✅ Correct Answers: %d\n"+
			"❌ Incorrect Answers: %d\n"+
			"⏭️  Skipped Questions: %d\n"+
			"📈 Score: %d/%d (%.1f%%)",
		totalQuestions,
		correctAnswers,
		incorrectAnswers,
		skippedQuestions,
		session.TotalScore,
		totalQuestions,
		percentage,
	)

	h.sendMessageWithMenu(msg.Chat.ID, resultText)
}
