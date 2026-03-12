package quick_duel

// DuelRoundCache stores per-round player answers for an ongoing duel game.
// Implementations must be safe for concurrent use and survive process restarts.
type DuelRoundCache interface {
	// AddAnswer records a player's answer for a specific round.
	AddAnswer(gameID string, round int, answer PlayerAnswer) error

	// GetAnswers returns all answers stored for a specific round.
	GetAnswers(gameID string, round int) ([]PlayerAnswer, error)

	// DeleteGame removes all cached answers for a finished game.
	DeleteGame(gameID string) error
}
