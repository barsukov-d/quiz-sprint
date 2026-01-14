package quiz

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// QuizRepository defines the interface for quiz persistence
type QuizRepository interface {
	// Quiz operations
	FindByID(ctx context.Context, id uuid.UUID) (*Quiz, error)
	FindAll(ctx context.Context) ([]Quiz, error)
	Save(ctx context.Context, quiz *Quiz) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Session operations
	FindSessionByID(ctx context.Context, id uuid.UUID) (*QuizSession, error)
	FindActiveSessionByUserAndQuiz(ctx context.Context, userID string, quizID uuid.UUID) (*QuizSession, error)
	SaveSession(ctx context.Context, session *QuizSession) error
	UpdateSession(ctx context.Context, session *QuizSession) error

	// Leaderboard operations
	GetLeaderboard(ctx context.Context, quizID uuid.UUID, limit int) ([]LeaderboardEntry, error)
}

// LeaderboardEntry represents a leaderboard entry
type LeaderboardEntry struct {
	UserID    string    `json:"userId"`
	Username  string    `json:"username"`
	Score     int       `json:"score"`
	Rank      int       `json:"rank"`
	QuizID    uuid.UUID `json:"quizId"`
	CompletedAt time.Time `json:"completedAt"`
}
