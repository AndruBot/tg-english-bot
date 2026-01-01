# MongoDB Collection Schemas

This document describes the structure of all MongoDB collections used in the English Test Bot.

## User Collection

**Collection Name:** `users`

Stores main information about users.

```json
{
  "_id": ObjectId("..."),
  "telegram_id": 123456789,
  "username": "john_doe",
  "first_name": "John",
  "last_name": "Doe",
  "total_score": 150,
  "tests_taken": 3,
  "created_at": ISODate("2024-01-01T00:00:00Z"),
  "updated_at": ISODate("2024-01-15T10:30:00Z")
}
```

**Fields:**
- `_id`: ObjectID - Unique identifier (auto-generated)
- `telegram_id`: int64 - Telegram user ID (unique, indexed)
- `username`: string (optional) - Telegram username
- `first_name`: string (optional) - User's first name
- `last_name`: string (optional) - User's last name
- `total_score`: int - Cumulative score across all tests
- `tests_taken`: int - Number of completed tests
- `created_at`: timestamp - Account creation time
- `updated_at`: timestamp - Last update time

## Session Collection

**Collection Name:** `sessions`

Stores information about test sessions (when test started and finished).

```json
{
  "_id": ObjectId("..."),
  "user_id": ObjectId("..."),
  "started_at": ISODate("2024-01-15T10:00:00Z"),
  "finished_at": ISODate("2024-01-15T10:30:00Z"),
  "total_score": 50,
  "total_questions": 100,
  "status": "completed"
}
```

**Fields:**
- `_id`: ObjectID - Unique identifier (auto-generated)
- `user_id`: ObjectID - Reference to User collection
- `started_at`: timestamp - When the test session started
- `finished_at`: timestamp (optional) - When the test session finished (null if in progress)
- `total_score`: int - Total score achieved in this session
- `total_questions`: int - Number of questions in this session
- `status`: string - Session status: "in_progress" or "completed"

## Question Collection

**Collection Name:** `questions`

Stores questions loaded from CSV file.

```json
{
  "_id": ObjectId("..."),
  "text": "What is the capital of France?",
  "answer_1": "Paris",
  "answer_2": "London",
  "answer_3": "Berlin",
  "answer_4": "Madrid",
  "correct_answer_id": 1,
  "score": 1
}
```

**Fields:**
- `_id`: ObjectID - Unique identifier (auto-generated)
- `text`: string - Question text
- `answer_1`: string - First answer option
- `answer_2`: string - Second answer option
- `answer_3`: string - Third answer option
- `answer_4`: string - Fourth answer option
- `correct_answer_id`: int - Correct answer (1-4)
- `score`: int - Points awarded for correct answer

## Answer Collection

**Collection Name:** `answers`

Stores user answers to questions.

```json
{
  "_id": ObjectId("..."),
  "session_id": ObjectId("..."),
  "user_id": ObjectId("..."),
  "question_id": ObjectId("..."),
  "selected_answer_id": 1,
  "is_correct": true,
  "score": 1,
  "answered_at": ISODate("2024-01-15T10:05:00Z")
}
```

**Fields:**
- `_id`: ObjectID - Unique identifier (auto-generated)
- `session_id`: ObjectID - Reference to Session collection
- `user_id`: ObjectID - Reference to User collection
- `question_id`: ObjectID - Reference to Question collection
- `selected_answer_id`: int - User's selected answer (1-4)
- `is_correct`: bool - Whether the answer is correct
- `score`: int - Points earned for this answer (0 if incorrect)
- `answered_at`: timestamp - When the answer was submitted

## Relationships

- **User** → **Session**: One-to-Many (a user can have multiple test sessions)
- **Session** → **Answer**: One-to-Many (a session contains multiple answers)
- **Question** → **Answer**: One-to-Many (a question can be answered multiple times by different users)
- **User** → **Answer**: One-to-Many (a user can have multiple answers across sessions)

## Indexes Recommendations

For better performance, consider creating the following indexes:

```javascript
// Users collection
db.users.createIndex({ "telegram_id": 1 }, { unique: true })

// Sessions collection
db.sessions.createIndex({ "user_id": 1 })
db.sessions.createIndex({ "status": 1 })

// Answers collection
db.answers.createIndex({ "session_id": 1 })
db.answers.createIndex({ "user_id": 1 })
db.answers.createIndex({ "question_id": 1 })
```

