package database

import (
	"context"
	"time"

	"github.com/andru_bot/tg-bot/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// UserRepository handles user operations
type UserRepository struct {
	collection *mongo.Collection
}

func NewUserRepository() *UserRepository {
	return &UserRepository{
		collection: DB.Collection("users"),
	}
}

func (r *UserRepository) FindOrCreate(telegramID int64, username, firstName, lastName string) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var user models.User
	err := r.collection.FindOne(ctx, bson.M{"telegram_id": telegramID}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		// Create new user
		user = models.User{
			ID:         primitive.NewObjectID(),
			TelegramID: telegramID,
			Username:   username,
			FirstName:  firstName,
			LastName:   lastName,
			TotalScore: 0,
			TestsTaken: 0,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
		_, err = r.collection.InsertOne(ctx, user)
		if err != nil {
			return nil, err
		}
		return &user, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetByID(userID primitive.ObjectID) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var user models.User
	err := r.collection.FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetByTelegramID(telegramID int64) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var user models.User
	err := r.collection.FindOne(ctx, bson.M{"telegram_id": telegramID}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) UpdateScore(userID primitive.ObjectID, score int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": userID},
		bson.M{
			"$inc": bson.M{
				"total_score": score,
				"tests_taken": 1,
			},
			"$set": bson.M{"updated_at": time.Now()},
		},
	)
	return err
}

// SessionRepository handles session operations
type SessionRepository struct {
	collection *mongo.Collection
}

func NewSessionRepository() *SessionRepository {
	return &SessionRepository{
		collection: DB.Collection("sessions"),
	}
}

func (r *SessionRepository) Create(userID primitive.ObjectID, questionIDs []primitive.ObjectID) (*models.Session, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	session := models.Session{
		ID:             primitive.NewObjectID(),
		UserID:         userID,
		StartedAt:      time.Now(),
		TotalScore:     0,
		TotalQuestions: len(questionIDs),
		Status:         "in_progress",
		CurrentIdx:     0,
		QuestionIDs:    questionIDs,
	}

	_, err := r.collection.InsertOne(ctx, session)
	if err != nil {
		return nil, err
	}

	return &session, nil
}

func (r *SessionRepository) Finish(sessionID primitive.ObjectID, totalScore, totalQuestions int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	now := time.Now()
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": sessionID},
		bson.M{
			"$set": bson.M{
				"finished_at":     now,
				"total_score":     totalScore,
				"total_questions": totalQuestions,
				"status":          "completed",
			},
		},
	)
	return err
}

func (r *SessionRepository) GetByID(sessionID primitive.ObjectID) (*models.Session, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var session models.Session
	err := r.collection.FindOne(ctx, bson.M{"_id": sessionID}).Decode(&session)
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *SessionRepository) GetActiveByUserID(userID primitive.ObjectID) (*models.Session, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var session models.Session
	err := r.collection.FindOne(
		ctx,
		bson.M{
			"user_id": userID,
			"status":  "in_progress",
		},
	).Decode(&session)
	if err == mongo.ErrNoDocuments {
		return nil, nil // No active session found
	}
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *SessionRepository) UpdateProgress(sessionID primitive.ObjectID, currentIdx int, score int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": sessionID},
		bson.M{
			"$set": bson.M{
				"current_idx": currentIdx,
				"total_score": score,
			},
		},
	)
	return err
}

func (r *SessionRepository) GetAllActive() ([]models.Session, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := r.collection.Find(ctx, bson.M{"status": "in_progress"})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var sessions []models.Session
	if err = cursor.All(ctx, &sessions); err != nil {
		return nil, err
	}

	return sessions, nil
}

func (r *SessionRepository) GetLastCompletedByUserID(userID primitive.ObjectID) (*models.Session, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var session models.Session
	err := r.collection.FindOne(
		ctx,
		bson.M{
			"user_id": userID,
			"status":  "completed",
		},
		options.FindOne().SetSort(bson.M{"finished_at": -1}),
	).Decode(&session)
	if err == mongo.ErrNoDocuments {
		return nil, nil // No completed session found
	}
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// QuestionRepository handles question operations
type QuestionRepository struct {
	collection *mongo.Collection
}

func NewQuestionRepository() *QuestionRepository {
	return &QuestionRepository{
		collection: DB.Collection("questions"),
	}
}

func (r *QuestionRepository) Create(question *models.Question) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := r.collection.InsertOne(ctx, question)
	return err
}

func (r *QuestionRepository) GetAll() ([]models.Question, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var questions []models.Question
	if err = cursor.All(ctx, &questions); err != nil {
		return nil, err
	}

	return questions, nil
}

func (r *QuestionRepository) GetByID(questionID primitive.ObjectID) (*models.Question, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var question models.Question
	err := r.collection.FindOne(ctx, bson.M{"_id": questionID}).Decode(&question)
	if err != nil {
		return nil, err
	}
	return &question, nil
}

// AnswerRepository handles answer operations
type AnswerRepository struct {
	collection *mongo.Collection
}

func NewAnswerRepository() *AnswerRepository {
	return &AnswerRepository{
		collection: DB.Collection("answers"),
	}
}

func (r *AnswerRepository) Create(answer *models.Answer) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := r.collection.InsertOne(ctx, answer)
	return err
}

func (r *AnswerRepository) GetBySession(sessionID primitive.ObjectID) ([]models.Answer, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := r.collection.Find(ctx, bson.M{"session_id": sessionID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var answers []models.Answer
	if err = cursor.All(ctx, &answers); err != nil {
		return nil, err
	}

	return answers, nil
}
