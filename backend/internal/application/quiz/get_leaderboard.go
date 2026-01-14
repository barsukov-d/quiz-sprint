package quiz

import (
	"github.com/barsukov/quiz-sprint/backend/internal/domain/quiz"
)

// GetLeaderboardUseCase handles the business logic for getting leaderboard
type GetLeaderboardUseCase struct {
	leaderboardRepo quiz.LeaderboardRepository
}

// NewGetLeaderboardUseCase creates a new GetLeaderboardUseCase
func NewGetLeaderboardUseCase(leaderboardRepo quiz.LeaderboardRepository) *GetLeaderboardUseCase {
	return &GetLeaderboardUseCase{
		leaderboardRepo: leaderboardRepo,
	}
}

// Execute retrieves the leaderboard for a quiz
func (uc *GetLeaderboardUseCase) Execute(input GetLeaderboardInput) (GetLeaderboardOutput, error) {
	// 1. Validate input
	quizID, err := quiz.NewQuizIDFromString(input.QuizID)
	if err != nil {
		return GetLeaderboardOutput{}, err
	}

	// 2. Normalize limit
	limit := input.Limit
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	// 3. Get leaderboard from repository
	entries, err := uc.leaderboardRepo.GetLeaderboard(quizID, limit)
	if err != nil {
		return GetLeaderboardOutput{}, err
	}

	// 4. Return DTOs
	return GetLeaderboardOutput{
		QuizID:  input.QuizID,
		Entries: ToLeaderboardEntriesDTO(entries),
	}, nil
}
