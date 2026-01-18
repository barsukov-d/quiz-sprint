package quiz

import "errors"

var (
	// Value Object errors
	ErrInvalidQuizID         = errors.New("invalid quiz ID")
	ErrInvalidQuestionID     = errors.New("invalid question ID")
	ErrInvalidAnswerID       = errors.New("invalid answer ID")
	ErrInvalidSessionID      = errors.New("invalid session ID")
	ErrInvalidCategoryID     = errors.New("invalid category ID")
	ErrInvalidTitle          = errors.New("invalid title")
	ErrTitleTooLong          = errors.New("title too long")
	ErrInvalidQuestionText   = errors.New("invalid question text")
	ErrQuestionTextTooLong   = errors.New("question text too long")
	ErrInvalidAnswerText     = errors.New("invalid answer text")
	ErrAnswerTextTooLong     = errors.New("answer text too long")
	ErrNegativePoints        = errors.New("points cannot be negative")
	ErrPointsTooHigh         = errors.New("points too high")
	ErrInvalidTimeLimit      = errors.New("invalid time limit")
	ErrTimeLimitTooHigh      = errors.New("time limit too high")
	ErrInvalidPassingScore   = errors.New("invalid passing score")

	// Quiz errors
	ErrQuizNotFound          = errors.New("quiz not found")
	ErrQuizCannotStart       = errors.New("quiz cannot be started")
	ErrQuizInvalidData       = errors.New("invalid quiz data")
	ErrNoQuestions           = errors.New("quiz has no questions")
	ErrTooManyQuestions      = errors.New("too many questions")
	ErrTooManyAnswers        = errors.New("too many answers per question")

	// Session errors
	ErrSessionNotFound       = errors.New("quiz session not found")
	ErrSessionAlreadyExists  = errors.New("active session already exists")
	ErrSessionCompleted      = errors.New("quiz session already completed")
	ErrSessionExpired        = errors.New("quiz session expired")

	// Answer errors
	ErrInvalidAnswer         = errors.New("invalid answer")
	ErrQuestionNotFound      = errors.New("question not found")
	ErrAnswerNotFound        = errors.New("answer not found")
	ErrAlreadyAnswered       = errors.New("question already answered")

	// User errors
	ErrUnauthorized          = errors.New("unauthorized")
)
