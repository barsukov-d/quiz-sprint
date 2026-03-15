package user

import (
	"context"

	domainUser "github.com/barsukov/quiz-sprint/backend/internal/domain/user"
)

// ========================================
// StatsQueryRepository — read-only cross-aggregate query interface
// Defined in application layer (not domain) because it spans multiple aggregates.
// ========================================

// StatsQueryRepository fetches aggregated profile stats for a user.
type StatsQueryRepository interface {
	GetProfileStats(ctx context.Context, userID string) (*ProfileStatsOutput, error)
}

// ========================================
// ProfileStatsOutput — output DTO
// ========================================

// ProfileStatsOutput holds aggregated stats for a player's profile.
type ProfileStatsOutput struct {
	TotalGamesCompleted int
	TotalPoints         int
	AverageScore        int
	CurrentStreak       int
	LongestStreak       int
	DuelMMR             int
	DuelLeague          string
	DuelDivision        int
	DuelLeagueLabel     string
	DuelLeagueIcon      string
	DuelWins            int
	DuelLosses          int
	MarathonBestScore   int
	MarathonBestStreak  int
	MarathonCategory    string
}

// ========================================
// GetUserProfileStatsUseCase
// ========================================

// GetUserProfileStatsUseCase retrieves aggregated statistics for a player's profile.
type GetUserProfileStatsUseCase struct {
	statsRepo StatsQueryRepository
}

// NewGetUserProfileStatsUseCase creates a new GetUserProfileStatsUseCase.
func NewGetUserProfileStatsUseCase(statsRepo StatsQueryRepository) *GetUserProfileStatsUseCase {
	return &GetUserProfileStatsUseCase{statsRepo: statsRepo}
}

// Execute retrieves aggregated profile stats for the given user ID.
func (uc *GetUserProfileStatsUseCase) Execute(ctx context.Context, userID string) (*ProfileStatsOutput, error) {
	if userID == "" {
		return nil, domainUser.ErrInvalidUserID
	}

	output, err := uc.statsRepo.GetProfileStats(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Calculate average score server-side
	if output.TotalGamesCompleted > 0 {
		output.AverageScore = output.TotalPoints / output.TotalGamesCompleted
	}

	return output, nil
}
