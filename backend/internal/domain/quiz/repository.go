package quiz

import (
	"github.com/barsukov/quiz-sprint/backend/internal/domain/shared"
)

// QuizRepository defines the interface for quiz persistence
// NOTE: No context.Context - domain layer is pure
// Infrastructure implementations add context internally
type QuizRepository interface {
	// FindByID retrieves a quiz by its ID
	FindByID(id QuizID) (*Quiz, error)

	// FindAll retrieves all quizzes
	FindAll() ([]Quiz, error)

	// FindAllSummaries retrieves a list of quiz summaries (read model for list views)
	FindAllSummaries() ([]*QuizSummary, error)

	// FindSummariesByCategory retrieves a list of quiz summaries filtered by category
	FindSummariesByCategory(categoryID CategoryID) ([]*QuizSummary, error)

	// Save persists a quiz (create or update)
	Save(quiz *Quiz) error

	// Delete removes a quiz by ID
	Delete(id QuizID) error
}

// SessionRepository defines the interface for quiz session persistence
type SessionRepository interface {
	// FindByID retrieves a session by its ID
	FindByID(id SessionID) (*QuizSession, error)

	// FindActiveByUserAndQuiz finds an active session for a user and quiz
	FindActiveByUserAndQuiz(userID shared.UserID, quizID QuizID) (*QuizSession, error)

	// Save persists a session (create or update)
	Save(session *QuizSession) error

	// Delete removes a session by ID
	Delete(id SessionID) error
}

// LeaderboardEntry represents a leaderboard entry
// This is a read model (CQRS pattern)
type LeaderboardEntry struct {
	userID      shared.UserID
	username    string
	score       Points
	rank        int
	quizID      QuizID
	completedAt int64
}

// NewLeaderboardEntry creates a new leaderboard entry
func NewLeaderboardEntry(userID shared.UserID, username string, score Points, rank int, quizID QuizID, completedAt int64) LeaderboardEntry {
	return LeaderboardEntry{
		userID:      userID,
		username:    username,
		score:       score,
		rank:        rank,
		quizID:      quizID,
		completedAt: completedAt,
	}
}

// Getters
func (le LeaderboardEntry) UserID() shared.UserID { return le.userID }
func (le LeaderboardEntry) Username() string      { return le.username }
func (le LeaderboardEntry) Score() Points         { return le.score }
func (le LeaderboardEntry) Rank() int             { return le.rank }
func (le LeaderboardEntry) QuizID() QuizID        { return le.quizID }
func (le LeaderboardEntry) CompletedAt() int64    { return le.completedAt }

// LeaderboardRepository defines the interface for leaderboard queries
// Separate from SessionRepository (CQRS: read model)
type LeaderboardRepository interface {
	// GetLeaderboard retrieves top scores for a quiz
	GetLeaderboard(quizID QuizID, limit int) ([]LeaderboardEntry, error)

	// GetUserRank retrieves a user's rank in a quiz leaderboard
	GetUserRank(quizID QuizID, userID shared.UserID) (int, error)
}

// CategoryRepository defines the interface for category persistence
type CategoryRepository interface {
	FindByID(id CategoryID) (*Category, error)
	FindAll() ([]*Category, error)
	Save(category *Category) error
	Delete(id CategoryID) error
}
