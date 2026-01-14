package quiz

import (
	"github.com/barsukov/quiz-sprint/backend/internal/domain/shared"
)

// Event is the interface for all domain events
type Event interface {
	EventName() string
	OccurredAt() int64
}

// QuizStartedEvent is published when a user starts a quiz
type QuizStartedEvent struct {
	quizID     QuizID
	sessionID  SessionID
	userID     shared.UserID
	occurredAt int64
}

func NewQuizStartedEvent(quizID QuizID, sessionID SessionID, userID shared.UserID, occurredAt int64) QuizStartedEvent {
	return QuizStartedEvent{
		quizID:     quizID,
		sessionID:  sessionID,
		userID:     userID,
		occurredAt: occurredAt,
	}
}

func (e QuizStartedEvent) EventName() string   { return "quiz.started" }
func (e QuizStartedEvent) OccurredAt() int64   { return e.occurredAt }
func (e QuizStartedEvent) QuizID() QuizID      { return e.quizID }
func (e QuizStartedEvent) SessionID() SessionID { return e.sessionID }
func (e QuizStartedEvent) UserID() shared.UserID { return e.userID }

// AnswerSubmittedEvent is published when a user submits an answer
type AnswerSubmittedEvent struct {
	sessionID  SessionID
	questionID QuestionID
	answerID   AnswerID
	isCorrect  bool
	points     Points
	occurredAt int64
}

func NewAnswerSubmittedEvent(sessionID SessionID, questionID QuestionID, answerID AnswerID, isCorrect bool, points Points, occurredAt int64) AnswerSubmittedEvent {
	return AnswerSubmittedEvent{
		sessionID:  sessionID,
		questionID: questionID,
		answerID:   answerID,
		isCorrect:  isCorrect,
		points:     points,
		occurredAt: occurredAt,
	}
}

func (e AnswerSubmittedEvent) EventName() string     { return "quiz.answer_submitted" }
func (e AnswerSubmittedEvent) OccurredAt() int64     { return e.occurredAt }
func (e AnswerSubmittedEvent) SessionID() SessionID  { return e.sessionID }
func (e AnswerSubmittedEvent) QuestionID() QuestionID { return e.questionID }
func (e AnswerSubmittedEvent) AnswerID() AnswerID    { return e.answerID }
func (e AnswerSubmittedEvent) IsCorrect() bool       { return e.isCorrect }
func (e AnswerSubmittedEvent) Points() Points        { return e.points }

// QuizCompletedEvent is published when a user completes a quiz
type QuizCompletedEvent struct {
	quizID      QuizID
	sessionID   SessionID
	userID      shared.UserID
	finalScore  Points
	occurredAt  int64
}

func NewQuizCompletedEvent(quizID QuizID, sessionID SessionID, userID shared.UserID, finalScore Points, occurredAt int64) QuizCompletedEvent {
	return QuizCompletedEvent{
		quizID:     quizID,
		sessionID:  sessionID,
		userID:     userID,
		finalScore: finalScore,
		occurredAt: occurredAt,
	}
}

func (e QuizCompletedEvent) EventName() string     { return "quiz.completed" }
func (e QuizCompletedEvent) OccurredAt() int64     { return e.occurredAt }
func (e QuizCompletedEvent) QuizID() QuizID        { return e.quizID }
func (e QuizCompletedEvent) SessionID() SessionID  { return e.sessionID }
func (e QuizCompletedEvent) UserID() shared.UserID { return e.userID }
func (e QuizCompletedEvent) FinalScore() Points    { return e.finalScore }

// EventBus is the interface for publishing domain events
// Defined in domain, implemented in infrastructure
type EventBus interface {
	// Publish publishes events asynchronously
	Publish(events ...Event)

	// Subscribe registers a handler for a specific event type
	Subscribe(eventName string, handler EventHandler)
}

// EventHandler is a function that handles a domain event
type EventHandler func(event Event)
