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

	// FindAllActiveByUser retrieves all active sessions for a user
	FindAllActiveByUser(userID shared.UserID) ([]*QuizSession, error)

	// FindCompletedByUserQuizAndDate finds a completed session for a user, quiz, and date range
	// startTime and endTime are Unix timestamps
	FindCompletedByUserQuizAndDate(userID shared.UserID, quizID QuizID, startTime, endTime int64) (*QuizSession, error)

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

// GlobalLeaderboardEntry represents a global leaderboard entry (read model)
// Shows aggregated score across all quizzes for a user
type GlobalLeaderboardEntry struct {
	userID           shared.UserID
	username         string
	totalScore       Points
	quizzesCompleted int
	rank             int
	lastActivityAt   int64
}

// NewGlobalLeaderboardEntry creates a new global leaderboard entry
func NewGlobalLeaderboardEntry(
	userID shared.UserID,
	username string,
	totalScore Points,
	quizzesCompleted int,
	rank int,
	lastActivityAt int64,
) GlobalLeaderboardEntry {
	return GlobalLeaderboardEntry{
		userID:           userID,
		username:         username,
		totalScore:       totalScore,
		quizzesCompleted: quizzesCompleted,
		rank:             rank,
		lastActivityAt:   lastActivityAt,
	}
}

// Getters
func (gle GlobalLeaderboardEntry) UserID() shared.UserID  { return gle.userID }
func (gle GlobalLeaderboardEntry) Username() string       { return gle.username }
func (gle GlobalLeaderboardEntry) TotalScore() Points     { return gle.totalScore }
func (gle GlobalLeaderboardEntry) QuizzesCompleted() int  { return gle.quizzesCompleted }
func (gle GlobalLeaderboardEntry) Rank() int              { return gle.rank }
func (gle GlobalLeaderboardEntry) LastActivityAt() int64  { return gle.lastActivityAt }

// GlobalLeaderboardRepository defines the interface for global leaderboard queries
// Aggregates scores across all quizzes (sum of best scores per quiz)
type GlobalLeaderboardRepository interface {
	// GetGlobalLeaderboard retrieves top scores across all quizzes
	GetGlobalLeaderboard(limit int) ([]GlobalLeaderboardEntry, error)

	// GetUserGlobalRank retrieves a user's rank in the global leaderboard
	GetUserGlobalRank(userID shared.UserID) (int, error)
}

// CategoryRepository defines the interface for category persistence
type CategoryRepository interface {
	FindByID(id CategoryID) (*Category, error)
	FindAll() ([]*Category, error)
	Save(category *Category) error
	Delete(id CategoryID) error
}

// TagRepository defines the interface for tag persistence
type TagRepository interface {
	// Save persists a tag (create or update)
	Save(tag *Tag) error

	// SaveAll persists multiple tags at once
	SaveAll(tags []*Tag) error

	// FindByName retrieves a tag by its name
	FindByName(name string) (*Tag, error)

	// FindByNames retrieves multiple tags by their names
	FindByNames(names []string) ([]*Tag, error)

	// FindAll retrieves all tags
	FindAll() ([]*Tag, error)

	// FindByQuizID retrieves all tags assigned to a quiz
	FindByQuizID(quizID QuizID) ([]*Tag, error)

	// AssignTagsToQuiz creates relationships between quiz and tags
	AssignTagsToQuiz(quizID QuizID, tags []*Tag) error

	// RemoveTagsFromQuiz removes relationships between quiz and tags
	RemoveTagsFromQuiz(quizID QuizID) error

	// FindQuizzesByTag retrieves quizzes that have a specific tag
	FindQuizzesByTag(tagName string, limit, offset int) ([]*QuizSummary, error)
}
