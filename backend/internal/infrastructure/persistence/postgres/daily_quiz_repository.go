package postgres

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/barsukov/quiz-sprint/backend/internal/domain/daily_challenge"
	"github.com/barsukov/quiz-sprint/backend/internal/domain/quiz"
)

// DailyQuizRepository is a PostgreSQL implementation of daily_challenge.DailyQuizRepository
type DailyQuizRepository struct {
	db *sql.DB
}

// NewDailyQuizRepository creates a new PostgreSQL daily quiz repository
func NewDailyQuizRepository(db *sql.DB) *DailyQuizRepository {
	return &DailyQuizRepository{db: db}
}

// Save persists a daily quiz
func (r *DailyQuizRepository) Save(dailyQuiz *daily_challenge.DailyQuiz) error {
	// Marshal question IDs to JSONB
	questionIDsJSON, err := r.marshalQuestionIDs(dailyQuiz.QuestionIDs())
	if err != nil {
		return fmt.Errorf("failed to marshal question_ids: %w", err)
	}

	query := `
		INSERT INTO daily_quizzes (
			id, date, question_ids, expires_at, created_at
		) VALUES (
			$1, $2, $3, $4, $5
		)
		ON CONFLICT (date) DO UPDATE SET
			question_ids = EXCLUDED.question_ids,
			expires_at = EXCLUDED.expires_at
	`

	_, err = r.db.Exec(query,
		dailyQuiz.ID().String(),
		dailyQuiz.Date().String(),
		questionIDsJSON,
		dailyQuiz.ExpiresAt(),
		dailyQuiz.CreatedAt(),
	)

	if err != nil {
		return fmt.Errorf("failed to save daily quiz: %w", err)
	}

	return nil
}

// FindByID retrieves a daily quiz by ID
func (r *DailyQuizRepository) FindByID(id daily_challenge.DailyQuizID) (*daily_challenge.DailyQuiz, error) {
	query := `
		SELECT id, date, question_ids, expires_at, created_at
		FROM daily_quizzes
		WHERE id = $1
	`

	var (
		quizID       string
		date         string
		questionIDs  []byte
		expiresAt    int64
		createdAt    int64
	)

	err := r.db.QueryRow(query, id.String()).Scan(&quizID, &date, &questionIDs, &expiresAt, &createdAt)
	if err == sql.ErrNoRows {
		return nil, daily_challenge.ErrDailyQuizNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query daily quiz: %w", err)
	}

	return r.reconstructDailyQuiz(quizID, date, questionIDs, expiresAt, createdAt)
}

// FindByDate retrieves the daily quiz for a specific date
func (r *DailyQuizRepository) FindByDate(date daily_challenge.Date) (*daily_challenge.DailyQuiz, error) {
	query := `
		SELECT id, date, question_ids, expires_at, created_at
		FROM daily_quizzes
		WHERE date = $1
	`

	var (
		quizID       string
		dbDate       string
		questionIDs  []byte
		expiresAt    int64
		createdAt    int64
	)

	err := r.db.QueryRow(query, date.String()).Scan(&quizID, &dbDate, &questionIDs, &expiresAt, &createdAt)
	if err == sql.ErrNoRows {
		return nil, daily_challenge.ErrDailyQuizNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query daily quiz by date: %w", err)
	}

	return r.reconstructDailyQuiz(quizID, dbDate, questionIDs, expiresAt, createdAt)
}

// Delete removes a daily quiz
func (r *DailyQuizRepository) Delete(id daily_challenge.DailyQuizID) error {
	query := `DELETE FROM daily_quizzes WHERE id = $1`

	result, err := r.db.Exec(query, id.String())
	if err != nil {
		return fmt.Errorf("failed to delete daily quiz: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return daily_challenge.ErrDailyQuizNotFound
	}

	return nil
}

// ========================================
// Helper Methods
// ========================================

func (r *DailyQuizRepository) marshalQuestionIDs(ids []daily_challenge.QuestionID) ([]byte, error) {
	stringIDs := make([]string, len(ids))
	for i, id := range ids {
		stringIDs[i] = id.String()
	}
	return json.Marshal(stringIDs)
}

func (r *DailyQuizRepository) unmarshalQuestionIDs(data []byte) ([]daily_challenge.QuestionID, error) {
	var stringIDs []string
	if err := json.Unmarshal(data, &stringIDs); err != nil {
		return nil, err
	}

	ids := make([]daily_challenge.QuestionID, len(stringIDs))
	for i, str := range stringIDs {
		id, err := quiz.NewQuestionIDFromString(str)
		if err != nil {
			return nil, fmt.Errorf("invalid question_id in array: %w", err)
		}
		ids[i] = id
	}

	return ids, nil
}

func (r *DailyQuizRepository) reconstructDailyQuiz(
	id string,
	date string,
	questionIDsJSON []byte,
	expiresAt int64,
	createdAt int64,
) (*daily_challenge.DailyQuiz, error) {
	quizID := daily_challenge.NewDailyQuizIDFromString(id)
	dateVO := daily_challenge.NewDateFromString(date)

	questionIDs, err := r.unmarshalQuestionIDs(questionIDsJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal question_ids: %w", err)
	}

	return daily_challenge.ReconstructDailyQuiz(
		quizID,
		dateVO,
		questionIDs,
		expiresAt,
		createdAt,
	), nil
}
