package daily_challenge

import "errors"

// Domain errors for daily challenge
var (
	// DailyQuiz errors
	ErrInvalidDailyQuizID = errors.New("invalid daily quiz ID")
	ErrDailyQuizNotFound  = errors.New("daily quiz not found")
	ErrDailyQuizExpired   = errors.New("daily quiz expired")
	ErrInvalidDate        = errors.New("invalid date")

	// DailyGame errors
	ErrInvalidGameID        = errors.New("invalid game ID")
	ErrGameNotFound         = errors.New("daily game not found")
	ErrGameAlreadyCompleted = errors.New("daily game already completed")
	ErrGameNotActive        = errors.New("daily game is not active")
	ErrAlreadyPlayedToday   = errors.New("already played today")
	ErrInvalidGameStatus    = errors.New("invalid game status transition")

	// Question errors
	ErrAllQuestionsAnswered = errors.New("all questions already answered")
	ErrQuestionNotInQuiz    = errors.New("question not in daily quiz")
)
