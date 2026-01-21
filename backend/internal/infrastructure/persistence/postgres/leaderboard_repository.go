package postgres

import (
	"database/sql"
	"fmt"

	"github.com/barsukov/quiz-sprint/backend/internal/domain/quiz"
	"github.com/barsukov/quiz-sprint/backend/internal/domain/shared"
)

// LeaderboardRepository is a PostgreSQL implementation of quiz.LeaderboardRepository and quiz.GlobalLeaderboardRepository
type LeaderboardRepository struct {
	db *sql.DB
}

// NewLeaderboardRepository creates a new PostgreSQL leaderboard repository
func NewLeaderboardRepository(db *sql.DB) *LeaderboardRepository {
	return &LeaderboardRepository{db: db}
}

// ==========================================
// Per-Quiz Leaderboard (LeaderboardRepository interface)
// ==========================================

// GetLeaderboard retrieves top scores for a specific quiz
// Returns leaderboard entries ranked by score (per-quiz)
// Shows only the BEST score per user (if user completed quiz multiple times)
func (r *LeaderboardRepository) GetLeaderboard(quizID quiz.QuizID, limit int) ([]quiz.LeaderboardEntry, error) {
	query := `
		WITH user_best_scores AS (
			SELECT
				user_id,
				MAX(score) as best_score,
				MAX(completed_at) as last_completed_at
			FROM quiz_sessions
			WHERE quiz_id = $1 AND status = 'completed'
			GROUP BY user_id
		)
		SELECT
			ubs.user_id,
			COALESCE(u.username, 'User ' || SUBSTRING(ubs.user_id, 1, 8)) as username,
			ubs.best_score as score,
			ubs.last_completed_at as completed_at,
			ROW_NUMBER() OVER (ORDER BY ubs.best_score DESC, ubs.last_completed_at ASC) as rank
		FROM user_best_scores ubs
		LEFT JOIN users u ON ubs.user_id = u.id
		ORDER BY rank
		LIMIT $2
	`

	rows, err := r.db.Query(query, quizID.String(), limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query leaderboard: %w", err)
	}
	defer rows.Close()

	var entries []quiz.LeaderboardEntry

	for rows.Next() {
		var (
			userID      string
			username    string
			score       int
			completedAt int64
			rank        int
		)

		err := rows.Scan(&userID, &username, &score, &completedAt, &rank)
		if err != nil {
			return nil, fmt.Errorf("failed to scan leaderboard entry: %w", err)
		}

		// Parse user ID
		userIDVO, err := shared.NewUserID(userID)
		if err != nil {
			return nil, fmt.Errorf("failed to parse user ID: %w", err)
		}

		// Parse score
		scoreVO, err := quiz.NewPoints(score)
		if err != nil {
			return nil, fmt.Errorf("failed to parse score: %w", err)
		}

		// Create leaderboard entry
		entry := quiz.NewLeaderboardEntry(
			userIDVO,
			username,
			scoreVO,
			rank,
			quizID,
			completedAt,
		)

		entries = append(entries, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating leaderboard entries: %w", err)
	}

	return entries, nil
}

// GetUserRank retrieves a user's rank in a quiz leaderboard
// Returns 0 if user has not completed the quiz
// Ranks are based on the BEST score per user
func (r *LeaderboardRepository) GetUserRank(quizID quiz.QuizID, userID shared.UserID) (int, error) {
	query := `
		WITH user_best_scores AS (
			SELECT
				user_id,
				MAX(score) as best_score,
				MAX(completed_at) as last_completed_at
			FROM quiz_sessions
			WHERE quiz_id = $1 AND status = 'completed'
			GROUP BY user_id
		),
		ranked_users AS (
			SELECT
				user_id,
				ROW_NUMBER() OVER (ORDER BY best_score DESC, last_completed_at ASC) as rank
			FROM user_best_scores
		)
		SELECT rank
		FROM ranked_users
		WHERE user_id = $2
	`

	var rank int

	err := r.db.QueryRow(query, quizID.String(), userID.String()).Scan(&rank)

	if err == sql.ErrNoRows {
		return 0, nil // User not found in leaderboard
	}

	if err != nil {
		return 0, fmt.Errorf("failed to query user rank: %w", err)
	}

	return rank, nil
}

// ==========================================
// Global Leaderboard (GlobalLeaderboardRepository interface)
// ==========================================

// GetGlobalLeaderboard retrieves top scores across all quizzes
// Returns aggregated leaderboard entries (sum of best scores per quiz)
func (r *LeaderboardRepository) GetGlobalLeaderboard(limit int) ([]quiz.GlobalLeaderboardEntry, error) {
	query := `
		SELECT
			gl.user_id,
			COALESCE(u.username, 'User ' || SUBSTRING(gl.user_id, 1, 8)) as username,
			gl.total_score,
			gl.quizzes_completed,
			gl.rank,
			gl.last_activity_at
		FROM global_leaderboard gl
		LEFT JOIN users u ON gl.user_id = u.id
		ORDER BY gl.rank
		LIMIT $1
	`

	rows, err := r.db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query global leaderboard: %w", err)
	}
	defer rows.Close()

	var entries []quiz.GlobalLeaderboardEntry

	for rows.Next() {
		var (
			userID           string
			username         string
			totalScore       int
			quizzesCompleted int
			rank             int
			lastActivityAt   int64
		)

		err := rows.Scan(&userID, &username, &totalScore, &quizzesCompleted, &rank, &lastActivityAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan global leaderboard entry: %w", err)
		}

		// Parse user ID
		userIDVO, err := shared.NewUserID(userID)
		if err != nil {
			return nil, fmt.Errorf("failed to parse user ID: %w", err)
		}

		// Parse total score
		totalScoreVO, err := quiz.NewPoints(totalScore)
		if err != nil {
			return nil, fmt.Errorf("failed to parse total score: %w", err)
		}

		// Create global leaderboard entry
		entry := quiz.NewGlobalLeaderboardEntry(
			userIDVO,
			username,
			totalScoreVO,
			quizzesCompleted,
			rank,
			lastActivityAt,
		)

		entries = append(entries, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating global leaderboard entries: %w", err)
	}

	return entries, nil
}

// GetUserGlobalRank retrieves a user's rank in the global leaderboard
// Returns 0 if user has not completed any quizzes
func (r *LeaderboardRepository) GetUserGlobalRank(userID shared.UserID) (int, error) {
	query := `
		SELECT rank
		FROM global_leaderboard
		WHERE user_id = $1
	`

	var rank int

	err := r.db.QueryRow(query, userID.String()).Scan(&rank)

	if err == sql.ErrNoRows {
		return 0, nil // User not found in global leaderboard
	}

	if err != nil {
		return 0, fmt.Errorf("failed to query user global rank: %w", err)
	}

	return rank, nil
}
