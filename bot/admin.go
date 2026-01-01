package bot

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/andru_bot/tg-bot/excel"
	"github.com/andru_bot/tg-bot/models"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (h *BotHandler) getAdminTelegramIDs() []int64 {
	adminIDStr := os.Getenv("ADMIN_TELEGRAM_ID")
	if adminIDStr == "" {
		return []int64{} // No default value
	}

	// Split by comma and parse each ID
	parts := strings.Split(adminIDStr, ",")
	var adminIDs []int64

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		adminID, err := strconv.ParseInt(part, 10, 64)
		if err != nil {
			log.Printf("Error parsing admin ID '%s': %v, skipping", part, err)
			continue
		}
		adminIDs = append(adminIDs, adminID)
	}

	return adminIDs
}

func (h *BotHandler) sendAdminNotification(userTelegramID int64, sessionID primitive.ObjectID, correctAnswers, incorrectAnswers, totalQuestions int, answers []models.Answer, questions []models.Question) {
	adminIDs := h.getAdminTelegramIDs()
	if len(adminIDs) == 0 {
		log.Printf("No admin IDs configured, skipping admin notification")
		return
	}

	// Get user info by telegram ID
	user, err := h.userRepo.GetByTelegramID(userTelegramID)
	var userLink string
	if err != nil || user == nil {
		// Fallback: create user link with just ID
		userLink = fmt.Sprintf("[User %d](tg://user?id=%d)", userTelegramID, userTelegramID)
		log.Printf("Error getting user for admin notification: %v", err)
	} else {
		// Create user link
		if user.Username != "" {
			userLink = fmt.Sprintf("[@%s](https://t.me/%s)", user.Username, user.Username)
		} else {
			userLink = fmt.Sprintf("[User %d](tg://user?id=%d)", userTelegramID, userTelegramID)
		}
	}

	// Create admin message
	adminMessage := fmt.Sprintf(
		"📊 Test Completed\n\n"+
			"👤 User: %s\n"+
			"✅ Correct Answers: %d\n"+
			"❌ Incorrect Answers: %d\n"+
			"📝 Total Questions: %d",
		userLink,
		correctAnswers,
		incorrectAnswers,
		totalQuestions,
	)

	// Create Excel file
	excelPath, err := excel.CreateResultsExcel(sessionID, answers, questions)
	if err != nil {
		log.Printf("Error creating Excel file: %v", err)
		return
	}
	defer os.Remove(excelPath) // Clean up temp file

	// Send message and Excel file to all admins
	for _, adminID := range adminIDs {
		// Send message to admin
		msg := tgbotapi.NewMessage(adminID, adminMessage)
		msg.ParseMode = "Markdown"
		_, err = h.bot.Send(msg)
		if err != nil {
			log.Printf("Error sending admin notification to %d: %v", adminID, err)
		}

		// Send Excel file to admin
		doc := tgbotapi.NewDocument(adminID, tgbotapi.FilePath(excelPath))
		doc.Caption = fmt.Sprintf("Test results for user %s", userLink)
		doc.ParseMode = "Markdown"
		_, err = h.bot.Send(doc)
		if err != nil {
			log.Printf("Error sending Excel file to admin %d: %v", adminID, err)
		}
	}
}

func (h *BotHandler) sendAdminNotificationWithSkipped(userTelegramID int64, sessionID primitive.ObjectID, correctAnswers, incorrectAnswers, totalQuestions int, answers []models.Answer, questions []models.Question, currentIdx int, maxConsecutiveErrors int) {
	adminIDs := h.getAdminTelegramIDs()
	if len(adminIDs) == 0 {
		log.Printf("No admin IDs configured, skipping admin notification")
		return
	}

	// Get user info by telegram ID
	user, err := h.userRepo.GetByTelegramID(userTelegramID)
	var userLink string
	if err != nil || user == nil {
		// Fallback: create user link with just ID
		userLink = fmt.Sprintf("[User %d](tg://user?id=%d)", userTelegramID, userTelegramID)
		log.Printf("Error getting user for admin notification: %v", err)
	} else {
		// Create user link
		if user.Username != "" {
			userLink = fmt.Sprintf("[@%s](https://t.me/%s)", user.Username, user.Username)
		} else {
			userLink = fmt.Sprintf("[User %d](tg://user?id=%d)", userTelegramID, userTelegramID)
		}
	}

	// Create admin message
	adminMessage := fmt.Sprintf(
		"📊 Test Failed (%d Consecutive Errors)\n\n"+
			"👤 User: %s\n"+
			"✅ Correct Answers: %d\n"+
			"❌ Incorrect Answers: %d\n"+
			"📝 Total Questions: %d",
		maxConsecutiveErrors,
		userLink,
		correctAnswers,
		incorrectAnswers,
		totalQuestions,
	)

	// Create Excel file with skipped questions
	excelPath, err := excel.CreateResultsExcelWithSkipped(sessionID, answers, questions, currentIdx)
	if err != nil {
		log.Printf("Error creating Excel file: %v", err)
		return
	}
	defer os.Remove(excelPath) // Clean up temp file

	// Send message and Excel file to all admins
	for _, adminID := range adminIDs {
		// Send message to admin
		msg := tgbotapi.NewMessage(adminID, adminMessage)
		msg.ParseMode = "Markdown"
		_, err = h.bot.Send(msg)
		if err != nil {
			log.Printf("Error sending admin notification to %d: %v", adminID, err)
		}

		// Send Excel file to admin
		doc := tgbotapi.NewDocument(adminID, tgbotapi.FilePath(excelPath))
		doc.Caption = fmt.Sprintf("Test results for user %s (failed due to %d consecutive errors)", userLink, maxConsecutiveErrors)
		doc.ParseMode = "Markdown"
		_, err = h.bot.Send(doc)
		if err != nil {
			log.Printf("Error sending Excel file to admin %d: %v", adminID, err)
		}
	}
}
