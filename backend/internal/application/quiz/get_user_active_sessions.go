package quiz

import (
	"github.com/barsukov/quiz-sprint/backend/internal/domain/quiz"
	"github.com/barsukov/quiz-sprint/backend/internal/domain/shared"
)

// GetUserActiveSessionsUseCase handles the business logic for getting user's active sessions
type GetUserActiveSessionsUseCase struct {
	quizRepo    quiz.QuizRepository
	sessionRepo quiz.SessionRepository
}

// NewGetUserActiveSessionsUseCase creates a new GetUserActiveSessionsUseCase
func NewGetUserActiveSessionsUseCase(
	quizRepo quiz.QuizRepository,
	sessionRepo quiz.SessionRepository,
) *GetUserActiveSessionsUseCase {
	return &GetUserActiveSessionsUseCase{
		quizRepo:    quizRepo,
		sessionRepo: sessionRepo,
	}
}

// Execute retrieves all active sessions for a user
func (uc *GetUserActiveSessionsUseCase) Execute(input GetUserActiveSessionsInput) (GetUserActiveSessionsOutput, error) {
	// 1. Validate and convert input
	userID, err := shared.NewUserID(input.UserID)
	if err != nil {
		return GetUserActiveSessionsOutput{}, err
	}

	// 2. Find all active sessions for this user
	sessions, err := uc.sessionRepo.FindAllActiveByUser(userID)
	if err != nil {
		return GetUserActiveSessionsOutput{}, err
	}

	// 3. Build session summaries with quiz information
	var sessionSummaries []SessionSummaryDTO

	for _, session := range sessions {
		// Load quiz to get title and questions count
		quizAggregate, err := uc.quizRepo.FindByID(session.QuizID())
		if err != nil {
			// Skip sessions where quiz cannot be loaded (might be deleted)
			continue
		}

		summary := SessionSummaryDTO{
			SessionID:       session.ID().String(),
			QuizID:          session.QuizID().String(),
			QuizTitle:       quizAggregate.Title().String(),
			CurrentQuestion: session.CurrentQuestion(),
			TotalQuestions:  quizAggregate.QuestionsCount(),
			Score:           session.Score().Value(),
			StartedAt:       session.StartedAt(),
		}

		sessionSummaries = append(sessionSummaries, summary)
	}

	// 4. Return DTOs
	return GetUserActiveSessionsOutput{
		Sessions: sessionSummaries,
	}, nil
}
