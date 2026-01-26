package quick_duel

import "errors"

// Domain errors for quick duel
var (
	// Game errors
	ErrInvalidGameID       = errors.New("invalid game ID")
	ErrGameNotFound        = errors.New("duel game not found")
	ErrGameAlreadyFinished = errors.New("duel game already finished")
	ErrGameNotActive       = errors.New("duel game is not active")
	ErrGameNotStarted      = errors.New("duel game not started yet")
	ErrInvalidGameStatus   = errors.New("invalid game status transition")

	// Player errors
	ErrPlayerNotInGame     = errors.New("player not in this game")
	ErrPlayerAlreadyAnswered = errors.New("player already answered this question")
	ErrBothPlayersDisconnected = errors.New("both players disconnected")

	// Question errors
	ErrAllQuestionsAnswered = errors.New("all questions already answered")
	ErrQuestionNotInGame    = errors.New("question not in this game")
	ErrInvalidRound         = errors.New("invalid round number")

	// Answer errors
	ErrInvalidAnswerTime = errors.New("invalid answer time (anti-cheat)")
)
