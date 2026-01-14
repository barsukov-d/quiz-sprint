package quiz

import (
	"time"

	"github.com/google/uuid"
)

// Quiz represents a quiz aggregate root
type Quiz struct {
	ID           uuid.UUID  `json:"id"`
	Title        string     `json:"title"`
	Description  string     `json:"description"`
	Questions    []Question `json:"questions"`
	TimeLimit    int        `json:"timeLimit"` // seconds per question
	PassingScore int        `json:"passingScore"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
}

// Question represents a quiz question
type Question struct {
	ID       uuid.UUID `json:"id"`
	QuizID   uuid.UUID `json:"quizId"`
	Text     string    `json:"text"`
	Answers  []Answer  `json:"answers"`
	Points   int       `json:"points"`
	Position int       `json:"position"`
}

// Answer represents a possible answer to a question
type Answer struct {
	ID         uuid.UUID `json:"id"`
	QuestionID uuid.UUID `json:"questionId"`
	Text       string    `json:"text"`
	IsCorrect  bool      `json:"isCorrect"`
	Position   int       `json:"position"`
}

// QuizSession represents an active quiz session
type QuizSession struct {
	ID              uuid.UUID `json:"id"`
	QuizID          uuid.UUID `json:"quizId"`
	UserID          string    `json:"userId"` // Telegram user ID
	CurrentQuestion int       `json:"currentQuestion"`
	Score           int       `json:"score"`
	Answers         []UserAnswer `json:"answers"`
	StartedAt       time.Time `json:"startedAt"`
	CompletedAt     *time.Time `json:"completedAt,omitempty"`
	Status          SessionStatus `json:"status"`
}

// UserAnswer represents a user's answer to a question
type UserAnswer struct {
	QuestionID uuid.UUID `json:"questionId"`
	AnswerID   uuid.UUID `json:"answerId"`
	IsCorrect  bool      `json:"isCorrect"`
	Points     int       `json:"points"`
	AnsweredAt time.Time `json:"answeredAt"`
}

// SessionStatus represents the status of a quiz session
type SessionStatus string

const (
	SessionStatusActive    SessionStatus = "active"
	SessionStatusCompleted SessionStatus = "completed"
	SessionStatusAbandoned SessionStatus = "abandoned"
)

// Business rules

// CanStart checks if the quiz can be started
func (q *Quiz) CanStart() bool {
	return len(q.Questions) > 0 && q.TimeLimit > 0
}

// HasMinimumQuestions checks if the quiz has minimum required questions
func (q *Quiz) HasMinimumQuestions() bool {
	return len(q.Questions) >= 5
}

// GetTotalPoints calculates total points available in the quiz
func (q *Quiz) GetTotalPoints() int {
	total := 0
	for _, question := range q.Questions {
		total += question.Points
	}
	return total
}

// HasPassed checks if the user passed the quiz
func (qs *QuizSession) HasPassed(quiz *Quiz) bool {
	return qs.Score >= quiz.PassingScore
}

// IsCompleted checks if the session is completed
func (qs *QuizSession) IsCompleted() bool {
	return qs.Status == SessionStatusCompleted
}

// GetNextQuestion returns the next question to be answered
func (qs *QuizSession) GetNextQuestion(quiz *Quiz) *Question {
	if qs.CurrentQuestion >= len(quiz.Questions) {
		return nil
	}
	return &quiz.Questions[qs.CurrentQuestion]
}
