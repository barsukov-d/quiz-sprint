package postgres

import (
	"database/sql"
	"fmt"

	"github.com/barsukov/quiz-sprint/backend/internal/domain/quiz"
	"github.com/barsukov/quiz-sprint/backend/internal/domain/shared"
)

// SessionRepository is a PostgreSQL implementation of quiz.SessionRepository
type SessionRepository struct {
	db *sql.DB
}

// NewSessionRepository creates a new PostgreSQL session repository
func NewSessionRepository(db *sql.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

// FindByID retrieves a quiz session by ID with all user answers
func (r *SessionRepository) FindByID(id quiz.SessionID) (*quiz.QuizSession, error) {
	// Load session data
	query := `
		SELECT id, quiz_id, user_id, current_question, score, status, started_at, completed_at, correct_answer_streak
		FROM quiz_sessions
		WHERE id = $1
	`

	var (
		sessionID           string
		quizID              string
		userID              string
		currentQuestion     int
		score               int
		status              string
		startedAt           int64
		completedAtNullable sql.NullInt64
		correctAnswerStreak int
	)

	err := r.db.QueryRow(query, id.String()).Scan(
		&sessionID,
		&quizID,
		&userID,
		&currentQuestion,
		&score,
		&status,
		&startedAt,
		&completedAtNullable,
		&correctAnswerStreak,
	)

	if err == sql.ErrNoRows {
		return nil, quiz.ErrSessionNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("failed to query session: %w", err)
	}

	// Parse IDs
	sessionIDVO, err := quiz.NewSessionIDFromString(sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse session ID: %w", err)
	}

	quizIDVO, err := quiz.NewQuizIDFromString(quizID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse quiz ID: %w", err)
	}

	userIDVO, err := shared.NewUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse user ID: %w", err)
	}

	scoreVO, err := quiz.NewPoints(score)
	if err != nil {
		return nil, fmt.Errorf("failed to parse score: %w", err)
	}

	// Parse status
	var sessionStatus quiz.SessionStatus
	switch status {
	case "active":
		sessionStatus = quiz.SessionStatusActive
	case "completed":
		sessionStatus = quiz.SessionStatusCompleted
	case "abandoned":
		sessionStatus = quiz.SessionStatusAbandoned
	default:
		sessionStatus = quiz.SessionStatusActive
	}

	// Parse completedAt (nullable)
	var completedAt int64
	if completedAtNullable.Valid {
		completedAt = completedAtNullable.Int64
	} else {
		completedAt = 0
	}

	// Load user answers
	answers, err := r.loadUserAnswers(sessionIDVO)
	if err != nil {
		return nil, fmt.Errorf("failed to load user answers: %w", err)
	}

	// Reconstruct session aggregate
	session := quiz.ReconstructQuizSession(
		sessionIDVO,
		quizIDVO,
		userIDVO,
		currentQuestion,
		scoreVO,
		answers,
		startedAt,
		completedAt,
		sessionStatus,
		correctAnswerStreak,
	)

	return session, nil
}

// loadUserAnswers retrieves all user answers for a session
func (r *SessionRepository) loadUserAnswers(sessionID quiz.SessionID) ([]quiz.UserAnswer, error) {
	query := `
		SELECT question_id, answer_id, is_correct, base_points, time_bonus, streak_bonus, time_spent, answered_at
		FROM user_answers
		WHERE session_id = $1
		ORDER BY answered_at ASC
	`

	rows, err := r.db.Query(query, sessionID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to query user answers: %w", err)
	}
	defer rows.Close()

	var answers []quiz.UserAnswer

	for rows.Next() {
		var (
			questionID  string
			answerID    string
			isCorrect   bool
			basePoints  int
			timeBonus   int
			streakBonus int
			timeSpent   int64
			answeredAt  int64
		)

		err := rows.Scan(
			&questionID,
			&answerID,
			&isCorrect,
			&basePoints,
			&timeBonus,
			&streakBonus,
			&timeSpent,
			&answeredAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan user answer: %w", err)
		}

		// Parse IDs
		questionIDVO, err := quiz.NewQuestionIDFromString(questionID)
		if err != nil {
			return nil, fmt.Errorf("failed to parse question ID: %w", err)
		}

		answerIDVO, err := quiz.NewAnswerIDFromString(answerID)
		if err != nil {
			return nil, fmt.Errorf("failed to parse answer ID: %w", err)
		}

		// Parse points
		basePointsVO, _ := quiz.NewPoints(basePoints)
		timeBonusVO, _ := quiz.NewPoints(timeBonus)
		streakBonusVO, _ := quiz.NewPoints(streakBonus)

		// Create UserAnswer with breakdown
		answer := quiz.NewUserAnswerWithBreakdown(
			questionIDVO,
			answerIDVO,
			isCorrect,
			basePointsVO,
			timeBonusVO,
			streakBonusVO,
			timeSpent,
			answeredAt,
		)

		answers = append(answers, answer)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating user answers: %w", err)
	}

	return answers, nil
}

// FindActiveByUserAndQuiz finds an active session for a user and quiz
func (r *SessionRepository) FindActiveByUserAndQuiz(userID shared.UserID, quizID quiz.QuizID) (*quiz.QuizSession, error) {
	query := `
		SELECT id
		FROM quiz_sessions
		WHERE user_id = $1 AND quiz_id = $2 AND status = 'active'
		LIMIT 1
	`

	var sessionID string

	err := r.db.QueryRow(query, userID.String(), quizID.String()).Scan(&sessionID)

	if err == sql.ErrNoRows {
		return nil, quiz.ErrSessionNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("failed to query active session: %w", err)
	}

	// Parse and load full session
	sessionIDVO, err := quiz.NewSessionIDFromString(sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse session ID: %w", err)
	}

	return r.FindByID(sessionIDVO)
}

// FindAllActiveByUser retrieves all active sessions for a user
func (r *SessionRepository) FindAllActiveByUser(userID shared.UserID) ([]*quiz.QuizSession, error) {
	query := `
		SELECT id
		FROM quiz_sessions
		WHERE user_id = $1 AND status = 'active'
		ORDER BY started_at DESC
	`

	rows, err := r.db.Query(query, userID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to query active sessions: %w", err)
	}
	defer rows.Close()

	var sessions []*quiz.QuizSession

	for rows.Next() {
		var sessionID string
		if err := rows.Scan(&sessionID); err != nil {
			return nil, fmt.Errorf("failed to scan session ID: %w", err)
		}

		// Parse and load full session
		sessionIDVO, err := quiz.NewSessionIDFromString(sessionID)
		if err != nil {
			return nil, fmt.Errorf("failed to parse session ID: %w", err)
		}

		session, err := r.FindByID(sessionIDVO)
		if err != nil {
			// Skip sessions that fail to load
			continue
		}

		sessions = append(sessions, session)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating sessions: %w", err)
	}

	return sessions, nil
}

// FindCompletedByUserQuizAndDate finds a completed session for a user, quiz, and date range
func (r *SessionRepository) FindCompletedByUserQuizAndDate(userID shared.UserID, quizID quiz.QuizID, startTime, endTime int64) (*quiz.QuizSession, error) {
	query := `
		SELECT id
		FROM quiz_sessions
		WHERE user_id = $1
		  AND quiz_id = $2
		  AND status = 'completed'
		  AND completed_at >= $3
		  AND completed_at < $4
		ORDER BY completed_at DESC
		LIMIT 1
	`

	var sessionID string

	err := r.db.QueryRow(query, userID.String(), quizID.String(), startTime, endTime).Scan(&sessionID)

	if err == sql.ErrNoRows {
		return nil, quiz.ErrSessionNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("failed to query completed session: %w", err)
	}

	// Parse and load full session
	sessionIDVO, err := quiz.NewSessionIDFromString(sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse session ID: %w", err)
	}

	return r.FindByID(sessionIDVO)
}

// Save persists a quiz session (create or update) with all user answers
func (r *SessionRepository) Save(session *quiz.QuizSession) error {
	// Start transaction
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback() // Rollback if not committed

	// UPSERT quiz_sessions
	sessionQuery := `
		INSERT INTO quiz_sessions (id, quiz_id, user_id, current_question, score, status, started_at, completed_at, correct_answer_streak)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (id) DO UPDATE SET
			current_question = EXCLUDED.current_question,
			score = EXCLUDED.score,
			status = EXCLUDED.status,
			completed_at = EXCLUDED.completed_at,
			correct_answer_streak = EXCLUDED.correct_answer_streak
	`

	var completedAtNullable sql.NullInt64
	if session.CompletedAt() > 0 {
		completedAtNullable = sql.NullInt64{Int64: session.CompletedAt(), Valid: true}
	} else {
		completedAtNullable = sql.NullInt64{Valid: false}
	}

	_, err = tx.Exec(
		sessionQuery,
		session.ID().String(),
		session.QuizID().String(),
		session.UserID().String(),
		session.CurrentQuestion(),
		session.Score().Value(),
		session.Status().String(),
		session.StartedAt(),
		completedAtNullable,
		session.CurrentStreak(),
	)

	if err != nil {
		return fmt.Errorf("failed to upsert session: %w", err)
	}

	// Insert new user answers (only inserts, no updates - answers are immutable)
	// First, get existing answer count to know which answers are new
	var existingCount int
	err = tx.QueryRow("SELECT COUNT(*) FROM user_answers WHERE session_id = $1", session.ID().String()).Scan(&existingCount)
	if err != nil {
		return fmt.Errorf("failed to count existing answers: %w", err)
	}

	// Insert only new answers (answers beyond existingCount)
	answers := session.Answers()
	for i := existingCount; i < len(answers); i++ {
		answer := answers[i]

		answerQuery := `
			INSERT INTO user_answers (session_id, question_id, answer_id, is_correct, base_points, time_bonus, streak_bonus, time_spent, answered_at, points)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			ON CONFLICT (session_id, question_id) DO NOTHING
		`

		_, err = tx.Exec(
			answerQuery,
			session.ID().String(),
			answer.QuestionID().String(),
			answer.AnswerID().String(),
			answer.IsCorrect(),
			answer.BasePoints().Value(),
			answer.TimeBonus().Value(),
			answer.StreakBonus().Value(),
			answer.TimeSpent(),
			answer.AnsweredAt(),
			answer.TotalPoints().Value(), // Legacy points column
		)

		if err != nil {
			return fmt.Errorf("failed to insert user answer: %w", err)
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Delete removes a session by ID (CASCADE deletes user_answers)
func (r *SessionRepository) Delete(id quiz.SessionID) error {
	query := `DELETE FROM quiz_sessions WHERE id = $1`

	result, err := r.db.Exec(query, id.String())
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return quiz.ErrSessionNotFound
	}

	return nil
}
