package bot

import (
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *BotHandler) sendMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "HTML"
	_, err := h.bot.Send(msg)
	if err != nil {
		log.Printf("Error sending message: %v", err)
	}
}

func (h *BotHandler) getMenuKeyboard() tgbotapi.ReplyKeyboardMarkup {
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("📚 Start Test"),
			tgbotapi.NewKeyboardButton("📊 My Results"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("✅ Finish Test"),
			tgbotapi.NewKeyboardButton("ℹ️ Help"),
		),
	)
	keyboard.ResizeKeyboard = true
	return keyboard
}

func (h *BotHandler) sendMessageWithMenu(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "HTML"
	msg.ReplyMarkup = h.getMenuKeyboard()
	_, err := h.bot.Send(msg)
	if err != nil {
		log.Printf("Error sending message: %v", err)
	}
}

func (h *BotHandler) removeMenu(chatID int64) {
	removeKeyboard := tgbotapi.NewRemoveKeyboard(true)
	msg := tgbotapi.NewMessage(chatID, "")
	msg.ReplyMarkup = removeKeyboard
	_, err := h.bot.Send(msg)
	if err != nil {
		log.Printf("Error removing keyboard: %v", err)
	}
}

func (h *BotHandler) answerCallback(callbackID string, text string) {
	callback := tgbotapi.NewCallback(callbackID, text)
	_, err := h.bot.Request(callback)
	if err != nil {
		log.Printf("Error answering callback: %v", err)
	}
}

func (h *BotHandler) deleteResultsCSV() {
	err := os.Remove(h.resultsCSVPath)
	if err != nil {
		// File might not exist, which is okay
		if !os.IsNotExist(err) {
			log.Printf("Error deleting results.csv: %v", err)
		}
	} else {
		log.Printf("Deleted results.csv to save space")
	}
}
