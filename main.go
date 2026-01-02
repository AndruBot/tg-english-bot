package main

import (
	"log"

	"github.com/andru_bot/tg-bot/bot"
	"github.com/andru_bot/tg-bot/config"
	"github.com/andru_bot/tg-bot/database"
	"github.com/andru_bot/tg-bot/json"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from the appropriate file based on ENV variable
	envFile := config.GetEnvFile()
	log.Printf("Loading environment from %s", envFile)

	err := godotenv.Load(envFile)
	if err != nil {
		log.Printf("Warning: Error loading %s file: %v (using system environment variables)", envFile, err)
	}

	// Get bot token from config
	botToken, err := config.GetTelegramBotToken()
	if err != nil {
		log.Fatal(err)
	}

	// Connect to MongoDB
	err = database.Connect()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Disconnect()

	// Load questions from JSON
	questions, err := json.LoadQuestions("questions.json")
	if err != nil {
		log.Fatalf("Failed to load questions: %v", err)
	}

	// Store questions in database
	questionRepo := database.NewQuestionRepository()
	existingQuestions, err := questionRepo.GetAll()
	if err != nil {
		log.Printf("Error checking existing questions: %v", err)
	}

	// Only insert questions if database is empty
	if len(existingQuestions) == 0 {
		log.Println("Loading questions into database...")
		for _, q := range questions {
			err := questionRepo.Create(&q)
			if err != nil {
				log.Printf("Error inserting question: %v", err)
			}
		}
		log.Printf("Loaded %d questions into database", len(questions))
	} else {
		log.Printf("Database already contains %d questions, skipping import", len(existingQuestions))
	}

	// Initialize Telegram bot
	telegramBot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Authorized on account %s", telegramBot.Self.UserName)

	// Set bot commands menu
	commands := []tgbotapi.BotCommand{
		{Command: "help", Description: "Show help message"},
		{Command: "start_test", Description: "Start a new test"},
		{Command: "finish_test", Description: "Finish current test"},
		{Command: "result", Description: "Show last test results"},
	}
	cmdConfig := tgbotapi.NewSetMyCommands(commands...)
	_, err = telegramBot.Request(cmdConfig)
	if err != nil {
		log.Printf("Warning: Failed to set bot commands: %v", err)
	} else {
		log.Println("Bot commands menu set successfully")
	}

	// Create bot handler
	resultsCSVPath := "results.csv"
	botHandler := bot.NewBotHandler(telegramBot, resultsCSVPath)
	botHandler.LoadQuestions(questions)

	// Set up update config
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := telegramBot.GetUpdatesChan(u)

	log.Println("Bot is running. Press Ctrl+C to stop.")

	// Handle updates
	for update := range updates {
		botHandler.HandleUpdate(update)
	}
}
