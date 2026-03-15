package daily_challenge

import (
	"context"

	"github.com/barsukov/quiz-sprint/backend/internal/domain/daily_challenge"
)

// CleanupAbandonedGamesUseCase marks stale in_progress games as abandoned.
// Should be run periodically (e.g. daily cron) to clean up games that were
// never finished by the player.
type CleanupAbandonedGamesUseCase struct {
	gameRepo daily_challenge.DailyGameRepository
}

func NewCleanupAbandonedGamesUseCase(gameRepo daily_challenge.DailyGameRepository) *CleanupAbandonedGamesUseCase {
	return &CleanupAbandonedGamesUseCase{gameRepo: gameRepo}
}

// Execute marks games as abandoned where status='in_progress' AND date < today - 1.
// Returns the number of games that were marked abandoned.
func (uc *CleanupAbandonedGamesUseCase) Execute(ctx context.Context) (int, error) {
	return uc.gameRepo.MarkAbandonedGames()
}
