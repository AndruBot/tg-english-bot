package bot

import (
	"github.com/andru_bot/tg-bot/config"
	"github.com/andru_bot/tg-bot/database"
	"github.com/andru_bot/tg-bot/models"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type BotHandler struct {
	bot                  *tgbotapi.BotAPI
	userRepo             *database.UserRepository
	sessionRepo          *database.SessionRepository
	questionRepo         *database.QuestionRepository
	answerRepo           *database.AnswerRepository
	activeSessions       map[int64]*ActiveSession
	questions            []models.Question
	resultsCSVPath       string
	maxConsecutiveErrors int
}

type ActiveSession struct {
	SessionID         primitive.ObjectID
	UserID            primitive.ObjectID
	QuestionIDs       []primitive.ObjectID
	CurrentIdx        int
	Score             int
	ConsecutiveErrors int // Track consecutive incorrect answers
}

func NewBotHandler(bot *tgbotapi.BotAPI, resultsCSVPath string) *BotHandler {
	return &BotHandler{
		bot:                  bot,
		userRepo:             database.NewUserRepository(),
		sessionRepo:          database.NewSessionRepository(),
		questionRepo:         database.NewQuestionRepository(),
		answerRepo:           database.NewAnswerRepository(),
		activeSessions:       make(map[int64]*ActiveSession),
		resultsCSVPath:       resultsCSVPath,
		maxConsecutiveErrors: config.GetMaxConsecutiveErrors(),
	}
}

func (h *BotHandler) LoadQuestions(questions []models.Question) {
	h.questions = questions
}

func (h *BotHandler) LoadActiveSessions() error {
	_, err := h.sessionRepo.GetAllActive()
	if err != nil {
		return err
	}
	// Sessions are loaded on-demand when user interacts
	return nil
}

func (h *BotHandler) HandleUpdate(update tgbotapi.Update) {
	// Handle callback queries first (button clicks)
	if update.CallbackQuery != nil {
		h.handleCallbackQuery(update.CallbackQuery)
		return
	}

	// Handle regular messages
	if update.Message == nil {
		return
	}

	msg := update.Message
	userID := msg.From.ID

	// Handle commands
	if msg.IsCommand() {
		h.handleCommand(msg)
		return
	}

	// Handle menu button clicks
	if h.handleMenuButton(msg) {
		return
	}

	// If user has active session, they might be trying to answer
	if _, exists := h.activeSessions[userID]; exists {
		h.sendMessage(msg.Chat.ID, "Please select an answer using the buttons below the question.")
		return
	}

	h.sendMessageWithMenu(msg.Chat.ID, "Use menu buttons or /start_test to start the English level test.")
}
