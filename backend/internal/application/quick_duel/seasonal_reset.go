package quick_duel

import (
	"fmt"
	"time"

	"github.com/barsukov/quiz-sprint/backend/internal/domain/quick_duel"
)

// SeasonalResetInput holds parameters for a seasonal MMR reset.
type SeasonalResetInput struct {
	// NewSeasonID is the ID of the new season. If empty, auto-generated as "YYYY-MM".
	NewSeasonID string
}

// SeasonalResetOutput describes the result of a seasonal reset operation.
type SeasonalResetOutput struct {
	PlayersReset int    `json:"playersReset"`
	OldSeasonID  string `json:"oldSeasonId"`
	NewSeasonID  string `json:"newSeasonId"`
}

// SeasonalResetUseCase resets all player MMRs using the soft-reset formula and advances the season.
type SeasonalResetUseCase struct {
	playerRatingRepo quick_duel.PlayerRatingRepository
	seasonRepo       quick_duel.SeasonRepository
}

func NewSeasonalResetUseCase(
	playerRatingRepo quick_duel.PlayerRatingRepository,
	seasonRepo quick_duel.SeasonRepository,
) *SeasonalResetUseCase {
	return &SeasonalResetUseCase{
		playerRatingRepo: playerRatingRepo,
		seasonRepo:       seasonRepo,
	}
}

// Execute performs the soft MMR reset for all players and registers the new season.
// Formula: newMMR = 1000 + (currentMMR - 1000) * 0.5, minimum 500.
func (uc *SeasonalResetUseCase) Execute(input SeasonalResetInput) (SeasonalResetOutput, error) {
	oldSeasonID, err := uc.seasonRepo.GetCurrentSeason()
	if err != nil {
		return SeasonalResetOutput{}, fmt.Errorf("seasonal reset: get current season: %w", err)
	}

	newSeasonID := input.NewSeasonID
	if newSeasonID == "" {
		newSeasonID = time.Now().UTC().Format("2006-01")
	}

	// Count players before reset for reporting
	total, err := uc.playerRatingRepo.GetTotalPlayers(oldSeasonID)
	if err != nil {
		return SeasonalResetOutput{}, fmt.Errorf("seasonal reset: count players: %w", err)
	}

	resetFn := func(currentMMR int) int {
		newMMR := 1000 + int(float64(currentMMR-1000)*0.5)
		if newMMR < 500 {
			newMMR = 500
		}
		return newMMR
	}

	if err := uc.playerRatingRepo.ResetAllRatingsForSeason(newSeasonID, resetFn); err != nil {
		return SeasonalResetOutput{}, fmt.Errorf("seasonal reset: reset ratings: %w", err)
	}

	// Register the new season record
	now := time.Now().UTC()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	monthEnd := monthStart.AddDate(0, 1, 0).Add(-time.Second)
	_ = uc.seasonRepo.CreateSeason(newSeasonID, monthStart.Unix(), monthEnd.Unix())

	return SeasonalResetOutput{
		PlayersReset: total,
		OldSeasonID:  oldSeasonID,
		NewSeasonID:  newSeasonID,
	}, nil
}
