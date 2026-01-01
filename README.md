# Telegram English Level Test Bot

A Telegram bot that tests users' English level through multiple-choice questions.

## Features

- Multiple-choice English level test
- Questions loaded from JSON file (115+ questions)
- Results stored in MongoDB
- Test results exported to Excel (XLSX) for admins
- Real-time test progress tracking
- Score calculation and reporting
- Persistent sessions (resume tests after bot restart)
- Automatic test failure after consecutive errors (configurable)
- Admin notifications with detailed results

## MongoDB Collections Structure

### User Collection
- `_id`: ObjectID (unique identifier)
- `telegram_id`: int64 (Telegram user ID)
- `username`: string (optional)
- `first_name`: string (optional)
- `last_name`: string (optional)
- `total_score`: int (cumulative score across all tests)
- `tests_taken`: int (number of tests completed)
- `created_at`: timestamp
- `updated_at`: timestamp

### Session Collection
- `_id`: ObjectID (unique identifier)
- `user_id`: ObjectID (reference to User)
- `started_at`: timestamp
- `finished_at`: timestamp (optional, set when test completes)
- `total_score`: int (score for this session)
- `total_questions`: int (number of questions in this session)
- `status`: string ("in_progress" or "completed")

### Question Collection
- `_id`: ObjectID (unique identifier)
- `text`: string (question text)
- `answer_1`: string (first answer option)
- `answer_2`: string (second answer option)
- `answer_3`: string (third answer option)
- `answer_4`: string (fourth answer option)
- `correct_answer_id`: int (1-4, indicating correct answer)
- `score`: int (points awarded for correct answer)

### Answer Collection
- `_id`: ObjectID (unique identifier)
- `session_id`: ObjectID (reference to Session)
- `user_id`: ObjectID (reference to User)
- `question_id`: ObjectID (reference to Question)
- `selected_answer_id`: int (1-4, user's selected answer)
- `is_correct`: bool (whether answer is correct)
- `score`: int (points earned for this answer)
- `answered_at`: timestamp

## Setup

### Prerequisites

- Go 1.25.0 or higher
- MongoDB (local or remote)
- Telegram Bot Token (get from [@BotFather](https://t.me/BotFather))

### Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd tg-english-bot
```

2. Install dependencies:
```bash
go mod download
```

3. Set up environment variables:
```bash
cp .env.example .env
# Edit .env and add your TELEGRAM_BOT_TOKEN
```

4. Start MongoDB using Docker Compose:
```bash
# Start MongoDB container
docker-compose -f docker-compose.mongodb.yml up -d

# Or if you prefer local MongoDB installation
mongod
```

5. Prepare your questions JSON file:
   - The file should be named `questions.json`
   - See "Questions" section below for format details
   - Or run `python3 generate_questions.py` to generate from `questions_text.txt`

6. Run the bot:
```bash
go run main.go
```

## Usage

1. Start a conversation with your bot on Telegram
2. Send `/start` to see welcome message
3. Send `/test me` to begin the test
4. Answer questions by selecting one of the four options
5. Complete all questions
6. Receive your score and percentage at the end

## Environment Variables

### Required
- `TELEGRAM_BOT_TOKEN`: Your Telegram bot token (required)

### MongoDB Configuration
- `MONGODB_URI`: MongoDB connection string (recommended). Example: `mongodb://admin:password@localhost:27017/english_test_bot?authSource=admin`

Alternatively, you can use individual MongoDB settings (all have defaults):
- `MONGO_HOST`: MongoDB host (default: `localhost`)
- `MONGODB_PORT` or `MONGO_PORT`: MongoDB port (default: `27017`)
- `MONGO_ROOT_USERNAME`: MongoDB root username (default: empty, no auth)
- `MONGO_ROOT_PASSWORD`: MongoDB root password (default: empty, no auth)
- `MONGO_DATABASE`: Database name (default: `english_test_bot`)

### Bot Configuration
- `ADMIN_TELEGRAM_ID`: Comma-separated list of admin Telegram IDs for notifications (default: empty, no notifications)
- `MAX_CONSECUTIVE_ERRORS`: Maximum consecutive errors before test failure (default: `5`)

### Docker Compose MongoDB

The `docker-compose.mongodb.yml` file sets up MongoDB with:
- Default credentials: `admin` / `password`
- Port: `27017` (configurable via `MONGODB_PORT` env var)
- Database: `english_test_bot` (configurable via `MONGO_DATABASE` env var)
- Persistent volumes for data storage

To start MongoDB:
```bash
docker-compose -f docker-compose.mongodb.yml up -d
```

To stop MongoDB:
```bash
docker-compose -f docker-compose.mongodb.yml down
```

## Questions

### Questions JSON (`questions.json`)
The bot loads questions from `questions.json` file. The file contains an array of question objects with the following structure:

```json
{
  "questions": [
    {
      "text": "Question text here",
      "text_html": "<b>Question 1</b>\n\nQuestion text here",
      "answer_1": "First answer option",
      "answer_1_html": "1. First answer option",
      "answer_2": "Second answer option",
      "answer_2_html": "2. Second answer option",
      "answer_3": "Third answer option",
      "answer_3_html": "3. Third answer option",
      "answer_4": "Fourth answer option (optional)",
      "answer_4_html": "4. Fourth answer option (optional)",
      "correct_answer_id": 1,
      "score": 1
    }
  ]
}
```

**Note:** Questions can have either 3 or 4 answer options. If a question has only 3 options, leave `answer_4` and `answer_4_html` as empty strings.

**Example file:** See `questions.json.example` for a complete example with sample questions.

### Updating Questions

To update the list of questions in `questions.json`:

1. **Using the provided script** (recommended):
   ```bash
   python3 generate_questions.py
   ```
   This script parses `questions_text.txt` and generates `questions.json` with all questions.

2. **Manually editing**:
   - Edit `questions.json` directly
   - Ensure the JSON structure is valid
   - Each question must have at least 3 answer options
   - `correct_answer_id` must be 1, 2, 3, or 4 (matching the number of available answers)
   - Use `\n` for line breaks in question text
   - Use `<b>` tags for bold text in HTML versions

3. **After updating**:
   - Restart the bot to load new questions
   - Questions are automatically imported into MongoDB on first run if the database is empty
   - To reload questions, clear the `questions` collection in MongoDB and restart the bot

## Project Structure

```
tg-english-bot/
├── main.go              # Application entry point
├── bot/
│   ├── handlers.go      # Bot handler structure
│   ├── commands.go      # Command handlers
│   ├── test_flow.go     # Test flow logic
│   ├── utils.go         # Utility functions
│   └── admin.go         # Admin notifications
├── database/
│   ├── db.go           # MongoDB connection
│   └── repository.go   # Database operations
├── models/
│   ├── user.go         # User model
│   ├── session.go      # Session model
│   ├── question.go     # Question model
│   └── answer.go       # Answer model
├── config/
│   └── config.go       # Configuration management
├── json/
│   └── json_handler.go  # JSON file operations
├── excel/
│   └── excel_handler.go # Excel file generation
├── questions.json       # Questions file (JSON format)
├── questions_text.txt   # Source questions text
├── generate_questions.py # Script to generate questions.json
├── go.mod
└── README.md
```

## Notes

- The bot doesn't reveal whether answers are correct during the test
- Results are logged to console when a test is completed
- All test data is stored in MongoDB for persistence
- Test sessions persist across bot restarts - users can resume their tests
- Admin notifications are sent via Telegram with Excel files containing detailed results
- Tests automatically fail if a user makes too many consecutive errors (configurable via `MAX_CONSECUTIVE_ERRORS`)
- Questions can have 3 or 4 answer options

