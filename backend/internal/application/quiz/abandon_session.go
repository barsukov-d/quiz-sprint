package quiz

import (
	"github.com/barsukov/quiz-sprint/backend/internal/domain/quiz"
	"github.com/barsukov/quiz-sprint/backend/internal/domain/shared"
)

// AbandonSessionUseCase handles abandoning an active quiz session
type AbandonSessionUseCase struct {
	sessionRepo quiz.SessionRepository
}

// NewAbandonSessionUseCase creates a new AbandonSessionUseCase
func NewAbandonSessionUseCase(
	sessionRepo quiz.SessionRepository,
) *AbandonSessionUseCase {
	return &AbandonSessionUseCase{
		sessionRepo: sessionRepo,
	}
}

// AbandonSessionInput is the input DTO
type AbandonSessionInput struct {
	SessionID string
	UserID    string // For authorization
}

// Execute abandons (deletes) a quiz session
func (uc *AbandonSessionUseCase) Execute(input AbandonSessionInput) error {
	// 1. Validate and convert input
	sessionID, err := quiz.NewSessionIDFromString(input.SessionID)
	if err != nil {
		return err
	}

	userID, err := shared.NewUserID(input.UserID)
	if err != nil {
		return err
	}

	// 2. Find session (to verify it exists and belongs to user)
	session, err := uc.sessionRepo.FindByID(sessionID)
	if err != nil {
		return err
	}

	// 3. Authorization check - ensure user owns this session
	if !session.UserID().Equals(userID) {
		return quiz.ErrUnauthorized
	}

	// 4. Delete the session
	return uc.sessionRepo.Delete(sessionID)
}
