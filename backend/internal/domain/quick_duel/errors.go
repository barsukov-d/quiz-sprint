package quick_duel

import "errors"

// ErrorCode is a machine-readable string identifier for a domain error.
// Frontend uses these codes to handle errors programmatically without parsing messages.
type ErrorCode string

const (
	// Game error codes
	CodeInvalidGameID       ErrorCode = "INVALID_GAME_ID"
	CodeGameNotFound        ErrorCode = "GAME_NOT_FOUND"
	CodeGameAlreadyFinished ErrorCode = "GAME_ALREADY_FINISHED"
	CodeGameNotActive       ErrorCode = "GAME_NOT_ACTIVE"
	CodeGameNotStarted      ErrorCode = "GAME_NOT_STARTED"
	CodeInvalidGameStatus   ErrorCode = "INVALID_GAME_STATUS"

	// Player error codes
	CodePlayerNotInGame        ErrorCode = "PLAYER_NOT_IN_GAME"
	CodePlayerAlreadyAnswered  ErrorCode = "PLAYER_ALREADY_ANSWERED"
	CodeBothPlayersDisconnected ErrorCode = "BOTH_PLAYERS_DISCONNECTED"

	// Question error codes
	CodeAllQuestionsAnswered ErrorCode = "ALL_QUESTIONS_ANSWERED"
	CodeQuestionNotInGame    ErrorCode = "QUESTION_NOT_IN_GAME"
	CodeInvalidRound         ErrorCode = "INVALID_ROUND"

	// Answer error codes
	CodeInvalidAnswerTime   ErrorCode = "INVALID_ANSWER_TIME"
	CodeTooEarlyToSurrender ErrorCode = "TOO_EARLY_TO_SURRENDER"

	// Challenge error codes
	CodeInvalidChallengeID   ErrorCode = "INVALID_CHALLENGE_ID"
	CodeChallengeNotFound    ErrorCode = "CHALLENGE_NOT_FOUND"
	CodeChallengeExpired     ErrorCode = "CHALLENGE_EXPIRED"
	CodeChallengeNotPending  ErrorCode = "CHALLENGE_NOT_PENDING"
	CodeNotChallengedPlayer  ErrorCode = "NOT_CHALLENGED_PLAYER"
	CodeCannotChallengeSelf  ErrorCode = "CANNOT_CHALLENGE_SELF"
	CodeFriendBusy           ErrorCode = "FRIEND_BUSY"
	CodeChallengeAlreadySent ErrorCode = "CHALLENGE_ALREADY_SENT"
	CodeAlreadyInQueue       ErrorCode = "ALREADY_IN_QUEUE"
	CodeAlreadyInGame        ErrorCode = "ALREADY_IN_GAME"
	CodeInsufficientTickets  ErrorCode = "INSUFFICIENT_TICKETS"

	// Referral error codes
	CodeReferralNotFound     ErrorCode = "REFERRAL_NOT_FOUND"
	CodeSelfReferral         ErrorCode = "SELF_REFERRAL"
	CodeAlreadyReferred      ErrorCode = "ALREADY_REFERRED"
	CodeReferralAlreadyExists ErrorCode = "REFERRAL_ALREADY_EXISTS"
	CodeMilestoneNotReached  ErrorCode = "MILESTONE_NOT_REACHED"
	CodeRewardAlreadyClaimed ErrorCode = "REWARD_ALREADY_CLAIMED"
)

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
	ErrInvalidAnswerTime   = errors.New("invalid answer time (anti-cheat)")
	ErrTooEarlyToSurrender = errors.New("cannot surrender before answering 3 questions")

	// Challenge errors
	ErrInvalidChallengeID   = errors.New("invalid challenge ID")
	ErrChallengeNotFound    = errors.New("challenge not found")
	ErrChallengeExpired     = errors.New("challenge has expired")
	ErrChallengeNotPending  = errors.New("challenge is not pending")
	ErrNotChallengedPlayer  = errors.New("player is not the challenged player")
	ErrCannotChallengeSelf  = errors.New("cannot challenge yourself")
	ErrFriendBusy           = errors.New("friend is already in a game")
	ErrChallengeAlreadySent = errors.New("challenge already sent to this player")
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
