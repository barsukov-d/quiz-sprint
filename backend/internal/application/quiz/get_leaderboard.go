package quiz

import (
	"context"

	"github.com/barsukov/quiz-sprint/backend/internal/domain/quiz"
	"github.com/google/uuid"
)

// GetLeaderboardQuery contains the parameters for getting leaderboard
type GetLeaderboardQuery struct {
	QuizID uuid.UUID
	Limit  int
}

// GetLeaderboardUseCase handles the business logic for getting leaderboard
type GetLeaderboardUseCase struct {
	repo quiz.QuizRepository
}

// NewGetLeaderboardUseCase creates a new GetLeaderboardUseCase
func NewGetLeaderboardUseCase(repo quiz.QuizRepository) *GetLeaderboardUseCase {
	return &GetLeaderboardUseCase{repo: repo}
}

// Execute retrieves the leaderboard
func (uc *GetLeaderboardUseCase) Execute(ctx context.Context, query GetLeaderboardQuery) ([]quiz.LeaderboardEntry, error) {
	if query.Limit <= 0 {
		query.Limit = 10
	}

	if query.Limit > 100 {
		query.Limit = 100
	}

	return uc.repo.GetLeaderboard(ctx, query.QuizID, query.Limit)
}
