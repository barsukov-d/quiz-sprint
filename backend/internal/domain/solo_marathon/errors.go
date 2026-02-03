package solo_marathon

import "errors"

// Domain errors for solo marathon
var (
	// Game errors
	ErrInvalidGameID       = errors.New("invalid game ID")
	ErrGameNotFound        = errors.New("marathon game not found")
	ErrGameAlreadyFinished = errors.New("marathon game already finished")
	ErrGameNotActive       = errors.New("marathon game is not active")
	ErrInvalidGameStatus   = errors.New("invalid game status transition")
	ErrActiveGameExists    = errors.New("player already has an active marathon game")

	// Lives errors
	ErrNoLivesRemaining = errors.New("no lives remaining")

	// Bonus errors
	ErrNoBonusesAvailable = errors.New("no bonuses available")
	ErrInvalidBonusType   = errors.New("invalid bonus type")
	ErrBonusAlreadyUsed   = errors.New("bonus already used for this question")
	ErrShieldAlreadyActive = errors.New("shield is already active")

	// Continue errors
	ErrContinueNotAvailable = errors.New("continue not available in current game state")
	ErrInsufficientCoins    = errors.New("insufficient coins for continue")

	// Question errors
	ErrInvalidQuestion = errors.New("invalid question")

	// Record errors
	ErrInvalidPersonalBestID = errors.New("invalid personal best ID")
	ErrPersonalBestNotFound  = errors.New("personal best not found")
)
