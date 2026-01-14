package quiz

import (
	"context"
	"time"

	"github.com/barsukov/quiz-sprint/backend/internal/domain/quiz"
	"github.com/google/uuid"
)

// StartQuizCommand contains the data needed to start a quiz
type StartQuizCommand struct {
	QuizID uuid.UUID
	UserID string
}

// StartQuizResult contains the result of starting a quiz
type StartQuizResult struct {
	Session *quiz.QuizSession
	Quiz    *quiz.Quiz
}

// StartQuizUseCase handles the business logic for starting a quiz
type StartQuizUseCase struct {
	repo quiz.QuizRepository
}

// NewStartQuizUseCase creates a new StartQuizUseCase
func NewStartQuizUseCase(repo quiz.QuizRepository) *StartQuizUseCase {
	return &StartQuizUseCase{repo: repo}
}

// Execute starts a quiz session
func (uc *StartQuizUseCase) Execute(ctx context.Context, cmd StartQuizCommand) (*StartQuizResult, error) {
	// Validate user ID
	if cmd.UserID == "" {
		return nil, quiz.ErrInvalidUserID
	}

	// Get the quiz
	quizData, err := uc.repo.FindByID(ctx, cmd.QuizID)
	if err != nil {
		return nil, err
	}

	// Validate quiz can be started
	if !quizData.CanStart() {
		return nil, quiz.ErrQuizCannotStart
	}

	// Check for existing active session
	existingSession, err := uc.repo.FindActiveSessionByUserAndQuiz(ctx, cmd.UserID, cmd.QuizID)
	if err == nil && existingSession != nil {
		return nil, quiz.ErrSessionAlreadyExists
	}

	// Create new session
	session := &quiz.QuizSession{
		ID:              uuid.New(),
		QuizID:          cmd.QuizID,
		UserID:          cmd.UserID,
		CurrentQuestion: 0,
		Score:           0,
		Answers:         make([]quiz.UserAnswer, 0),
		StartedAt:       time.Now(),
		Status:          quiz.SessionStatusActive,
	}

	// Save session
	if err := uc.repo.SaveSession(ctx, session); err != nil {
		return nil, err
	}

	return &StartQuizResult{
		Session: session,
		Quiz:    quizData,
	}, nil
}
