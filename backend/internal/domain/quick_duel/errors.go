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

	// Challenge errors
	ErrInvalidChallengeID   = errors.New("invalid challenge ID")
	ErrChallengeNotFound    = errors.New("challenge not found")
	ErrChallengeExpired     = errors.New("challenge has expired")
	ErrChallengeNotPending  = errors.New("challenge is not pending")
	ErrNotChallengedPlayer  = errors.New("player is not the challenged player")
	ErrCannotChallengeSelf  = errors.New("cannot challenge yourself")
	ErrFriendBusy           = errors.New("friend is already in a game")
	ErrAlreadyInQueue       = errors.New("already in matchmaking queue")
	ErrAlreadyInGame        = errors.New("already in an active game")
	ErrInsufficientTickets  = errors.New("insufficient tickets")

	// Referral errors
	ErrReferralNotFound     = errors.New("referral not found")
	ErrSelfReferral         = errors.New("cannot refer yourself")
	ErrAlreadyReferred      = errors.New("player already has a referrer")
	ErrReferralAlreadyExists = errors.New("referral already exists")
	ErrMilestoneNotReached  = errors.New("milestone not reached")
	ErrRewardAlreadyClaimed = errors.New("reward already claimed")
)
