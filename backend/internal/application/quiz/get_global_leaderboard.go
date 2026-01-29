package quiz

import (
	"github.com/barsukov/quiz-sprint/backend/internal/domain/quiz"
)

// GetGlobalLeaderboardUseCase handles the business logic for getting global leaderboard
type GetGlobalLeaderboardUseCase struct {
	leaderboardRepo quiz.GlobalLeaderboardRepository
}

// NewGetGlobalLeaderboardUseCase creates a new GetGlobalLeaderboardUseCase
func NewGetGlobalLeaderboardUseCase(leaderboardRepo quiz.GlobalLeaderboardRepository) *GetGlobalLeaderboardUseCase {
	return &GetGlobalLeaderboardUseCase{
		leaderboardRepo: leaderboardRepo,
	}
}

// Execute retrieves the global leaderboard across all quizzes
func (uc *GetGlobalLeaderboardUseCase) Execute(input GetGlobalLeaderboardInput) (GetGlobalLeaderboardOutput, error) {
	// 1. Normalize limit
	limit := input.Limit
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	// 2. Get global leaderboard from repository
	entries, err := uc.leaderboardRepo.GetGlobalLeaderboard(limit)
	if err != nil {
		return GetGlobalLeaderboardOutput{}, err
	}

	// 3. Return DTOs
	return GetGlobalLeaderboardOutput{
		Entries: ToGlobalLeaderboardEntriesDTO(entries),
	}, nil
}
