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

	// Hints errors
	ErrNoHintsAvailable = errors.New("no hints available")
	ErrInvalidHintType  = errors.New("invalid hint type")
	ErrHintAlreadyUsed  = errors.New("hint already used for this question")

	// Question errors
	ErrInvalidQuestion = errors.New("invalid question")

	// Record errors
	ErrInvalidPersonalBestID = errors.New("invalid personal best ID")
	ErrPersonalBestNotFound  = errors.New("personal best not found")
)
