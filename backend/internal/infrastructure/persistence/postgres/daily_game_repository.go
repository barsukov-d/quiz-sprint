package postgres

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/barsukov/quiz-sprint/backend/internal/domain/daily_challenge"
	"github.com/barsukov/quiz-sprint/backend/internal/domain/kernel"
	"github.com/barsukov/quiz-sprint/backend/internal/domain/quiz"
	"github.com/barsukov/quiz-sprint/backend/internal/domain/shared"
)

// DailyGameRepository is a PostgreSQL implementation of daily_challenge.DailyGameRepository
type DailyGameRepository struct {
	db            *sql.DB
	quizRepo      quiz.QuizRepository
	questionRepo  quiz.QuestionRepository
	dailyQuizRepo daily_challenge.DailyQuizRepository
}

// NewDailyGameRepository creates a new PostgreSQL daily game repository
func NewDailyGameRepository(
	db *sql.DB,
	quizRepo quiz.QuizRepository,
	questionRepo quiz.QuestionRepository,
	dailyQuizRepo daily_challenge.DailyQuizRepository,
) *DailyGameRepository {
	return &DailyGameRepository{
		db:            db,
		quizRepo:      quizRepo,
		questionRepo:  questionRepo,
		dailyQuizRepo: dailyQuizRepo,
	}
}

// Save persists a daily game
func (r *DailyGameRepository) Save(game *daily_challenge.DailyGame) error {
	session := game.Session()

	// Serialize session state
	sessionState, err := serializeGameplaySession(session)
	if err != nil {
		return fmt.Errorf("failed to serialize session: %w", err)
	}

	rank := sql.NullInt32{}
	if game.Rank() != nil {
		rank.Int32 = int32(*game.Rank())
		rank.Valid = true
	}

	lastPlayedDate := sql.NullString{}
	if !game.Streak().LastPlayedDate().IsZero() {
		lastPlayedDate.String = game.Streak().LastPlayedDate().String()
		lastPlayedDate.Valid = true
	}

	query := `
		INSERT INTO daily_games (
			id, player_id, daily_quiz_id, date, status,
			session_state, current_streak, best_streak, last_played_date, rank
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (id) DO UPDATE SET
			status = EXCLUDED.status,
			session_state = EXCLUDED.session_state,
			current_streak = EXCLUDED.current_streak,
			best_streak = EXCLUDED.best_streak,
			last_played_date = EXCLUDED.last_played_date,
			rank = EXCLUDED.rank
	`

	_, err = r.db.Exec(query,
		game.ID().String(),
		game.PlayerID().String(),
		game.DailyQuizID().String(),
		game.Date().String(),
		string(game.Status()),
		sessionState,
		game.Streak().CurrentStreak(),
		game.Streak().BestStreak(),
		lastPlayedDate,
		rank,
	)

	return err
}

// FindByID retrieves a daily game by ID
func (r *DailyGameRepository) FindByID(id daily_challenge.GameID) (*daily_challenge.DailyGame, error) {
	query := `
		SELECT id, player_id, daily_quiz_id, date, status,
			session_state, current_streak, best_streak, last_played_date, rank
		FROM daily_games WHERE id = $1
	`

	return r.scanGame(r.db.QueryRow(query, id.String()))
}

// FindByPlayerAndDate retrieves a player's game for a specific date
func (r *DailyGameRepository) FindByPlayerAndDate(playerID daily_challenge.UserID, date daily_challenge.Date) (*daily_challenge.DailyGame, error) {
	query := `
		SELECT id, player_id, daily_quiz_id, date, status,
			session_state, current_streak, best_streak, last_played_date, rank
		FROM daily_games WHERE player_id = $1 AND date = $2
	`

	return r.scanGame(r.db.QueryRow(query, playerID.String(), date.String()))
}

// FindTopByDate retrieves top N players for a specific date
func (r *DailyGameRepository) FindTopByDate(date daily_challenge.Date, limit int) ([]*daily_challenge.DailyGame, error) {
	query := `
		SELECT id, player_id, daily_quiz_id, date, status,
			session_state, current_streak, best_streak, last_played_date, rank
		FROM daily_games
		WHERE date = $1 AND status = 'completed'
		ORDER BY (session_state->>'base_score')::int * (CASE
			WHEN current_streak >= 100 THEN 2.0
			WHEN current_streak >= 30 THEN 1.6
			WHEN current_streak >= 14 THEN 1.4
			WHEN current_streak >= 7 THEN 1.25
			WHEN current_streak >= 3 THEN 1.1
			ELSE 1.0
		END) DESC,
		(session_state->>'completed_at')::bigint ASC
		LIMIT $2
	`

	rows, err := r.db.Query(query, date.String(), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	games := make([]*daily_challenge.DailyGame, 0)
	for rows.Next() {
		game, err := r.scanGameFromRows(rows)
		if err != nil {
			return nil, err
		}
		games = append(games, game)
	}

	return games, rows.Err()
}

// GetPlayerRankByDate calculates player's rank for a specific date
func (r *DailyGameRepository) GetPlayerRankByDate(playerID daily_challenge.UserID, date daily_challenge.Date) (int, error) {
	query := `
		WITH player_score AS (
			SELECT (session_state->>'base_score')::int * (CASE
				WHEN current_streak >= 100 THEN 2.0
				WHEN current_streak >= 30 THEN 1.6
				WHEN current_streak >= 14 THEN 1.4
				WHEN current_streak >= 7 THEN 1.25
				WHEN current_streak >= 3 THEN 1.1
				ELSE 1.0
			END) as final_score
			FROM daily_games
			WHERE player_id = $1 AND date = $2 AND status = 'completed'
		)
		SELECT COUNT(*) + 1 FROM daily_games
		WHERE date = $2 AND status = 'completed'
		AND (session_state->>'base_score')::int * (CASE
			WHEN current_streak >= 100 THEN 2.0
			WHEN current_streak >= 30 THEN 1.6
			WHEN current_streak >= 14 THEN 1.4
			WHEN current_streak >= 7 THEN 1.25
			WHEN current_streak >= 3 THEN 1.1
			ELSE 1.0
		END) > (SELECT final_score FROM player_score)
	`

	var rank int
	err := r.db.QueryRow(query, playerID.String(), date.String()).Scan(&rank)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return rank, err
}

// GetTotalPlayersByDate returns total number of players who played on date
func (r *DailyGameRepository) GetTotalPlayersByDate(date daily_challenge.Date) (int, error) {
	query := `SELECT COUNT(*) FROM daily_games WHERE date = $1 AND status = 'completed'`

	var count int
	err := r.db.QueryRow(query, date.String()).Scan(&count)
	return count, err
}

// Delete removes a daily game
func (r *DailyGameRepository) Delete(id daily_challenge.GameID) error {
	_, err := r.db.Exec(`DELETE FROM daily_games WHERE id = $1`, id.String())
	return err
}

// ========================================
// Helper Methods
// ========================================

func (r *DailyGameRepository) scanGame(row *sql.Row) (*daily_challenge.DailyGame, error) {
	var (
		id, playerID, dailyQuizID, date, status string
		sessionState                             []byte
		currentStreak, bestStreak                int
		lastPlayedDate                          sql.NullString
		rank                                     sql.NullInt32
	)

	err := row.Scan(&id, &playerID, &dailyQuizID, &date, &status, &sessionState, &currentStreak, &bestStreak, &lastPlayedDate, &rank)
	if err == sql.ErrNoRows {
		return nil, daily_challenge.ErrGameNotFound
	}
	if err != nil {
		return nil, err
	}

	return r.reconstructGame(id, playerID, dailyQuizID, date, status, sessionState, currentStreak, bestStreak, lastPlayedDate, rank)
}

func (r *DailyGameRepository) scanGameFromRows(rows *sql.Rows) (*daily_challenge.DailyGame, error) {
	var (
		id, playerID, dailyQuizID, date, status string
		sessionState                             []byte
		currentStreak, bestStreak                int
		lastPlayedDate                          sql.NullString
		rank                                     sql.NullInt32
	)

	err := rows.Scan(&id, &playerID, &dailyQuizID, &date, &status, &sessionState, &currentStreak, &bestStreak, &lastPlayedDate, &rank)
	if err != nil {
		return nil, err
	}

	return r.reconstructGame(id, playerID, dailyQuizID, date, status, sessionState, currentStreak, bestStreak, lastPlayedDate, rank)
}

func (r *DailyGameRepository) reconstructGame(
	id, playerID, dailyQuizID, date, status string,
	sessionState []byte,
	currentStreak, bestStreak int,
	lastPlayedDate sql.NullString,
	rank sql.NullInt32,
) (*daily_challenge.DailyGame, error) {
	// Deserialize session - need daily quiz ID to reconstruct the quiz
	session, err := r.deserializeDailyChallengeSession(sessionState, dailyQuizID)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize session: %w", err)
	}

	// Convert sql.NullString to Date
	var lastPlayedDateValue daily_challenge.Date
	if lastPlayedDate.Valid && lastPlayedDate.String != "" {
		lastPlayedDateValue = daily_challenge.NewDateFromString(lastPlayedDate.String)
	}
	// else: zero value Date (empty)

	// Reconstruct streak
	streak := daily_challenge.ReconstructStreakSystem(
		currentStreak,
		bestStreak,
		lastPlayedDateValue,
	)

	// Reconstruct rank
	var rankPtr *int
	if rank.Valid {
		r := int(rank.Int32)
		rankPtr = &r
	}

	// Parse user ID
	userID, err := daily_challenge.UserID{}, nil
	if playerID != "" {
		userID, err = shared.NewUserID(playerID)
		if err != nil {
			return nil, fmt.Errorf("invalid player_id: %w", err)
		}
	}

	return daily_challenge.ReconstructDailyGame(
		daily_challenge.NewGameIDFromString(id),
		userID,
		daily_challenge.NewDailyQuizIDFromString(dailyQuizID),
		daily_challenge.NewDateFromString(date),
		daily_challenge.GameStatus(status),
		session,
		streak,
		rankPtr,
	), nil
}

// Reuse session serialization from quiz repository
func serializeGameplaySession(session *kernel.QuizGameplaySession) ([]byte, error) {
	return serializeSession(session) // From quiz_repository.go helper
}

func deserializeGameplaySession(data []byte, quizRepo quiz.QuizRepository, questionRepo quiz.QuestionRepository) (*kernel.QuizGameplaySession, error) {
	// NOTE: This function is kept for compatibility but should not be used for Daily Challenge
	// Use deserializeDailyChallengeSession instead
	return deserializeSession(data, quizRepo, questionRepo) // From session_serialization.go helper
}

// deserializeDailyChallengeSession reconstructs a quiz gameplay session for Daily Challenge
// Unlike regular quizzes, Daily Challenge quizzes are ephemeral (not stored in quizzes table)
// We reconstruct the quiz from the daily_quiz question list
func (r *DailyGameRepository) deserializeDailyChallengeSession(
	sessionData []byte,
	dailyQuizIDStr string,
) (*kernel.QuizGameplaySession, error) {
	// 1. Load daily quiz to get question IDs
	dailyQuizID := daily_challenge.NewDailyQuizIDFromString(dailyQuizIDStr)
	dailyQuiz, err := r.dailyQuizRepo.FindByID(dailyQuizID)
	if err != nil {
		return nil, fmt.Errorf("failed to load daily quiz: %w", err)
	}

	// 2. Load questions
	questions, err := r.questionRepo.FindByIDs(dailyQuiz.QuestionIDs())
	if err != nil {
		return nil, fmt.Errorf("failed to load questions: %w", err)
	}

	// 3. Reconstruct quiz aggregate
	now := time.Now().Unix()
	quizID := quiz.NewQuizID() // Ephemeral ID
	quizTitle, _ := quiz.NewQuizTitle("Daily Challenge")
	quizTimeLimit, _ := quiz.NewTimeLimit(150) // 15 seconds per question * 10
	quizPassingScore, _ := quiz.NewPassingScore(0)

	quizAggregate, err := quiz.NewQuiz(
		quizID,
		quizTitle,
		"",
		quiz.CategoryID{},
		quizTimeLimit,
		quizPassingScore,
		now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create quiz: %w", err)
	}

	// Add questions in order
	for _, question := range questions {
		if err := quizAggregate.AddQuestion(*question); err != nil {
			return nil, fmt.Errorf("failed to add question: %w", err)
		}
	}

	// 4. Deserialize session data (manually, without quiz repo lookup)
	var serialized SerializedSession
	if err := json.Unmarshal(sessionData, &serialized); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}

	// 5. Reconstruct userAnswers
	userAnswers := make(map[kernel.QuestionID]kernel.AnswerData)
	for questionIDStr, serializedAnswer := range serialized.UserAnswers {
		questionID, err := quiz.NewQuestionIDFromString(questionIDStr)
		if err != nil {
			return nil, fmt.Errorf("invalid question_id: %w", err)
		}

		answerID, err := quiz.NewAnswerIDFromString(serializedAnswer.AnswerID)
		if err != nil {
			return nil, fmt.Errorf("invalid answer_id: %w", err)
		}

		userAnswers[questionID] = kernel.NewAnswerData(
			answerID,
			serializedAnswer.IsCorrect,
			serializedAnswer.TimeTaken,
			serializedAnswer.AnsweredAt,
		)
	}

	// 6. Reconstruct session
	sessionID, err := kernel.NewSessionIDFromString(serialized.SessionID)
	if err != nil {
		return nil, fmt.Errorf("invalid session_id: %w", err)
	}

	baseScore, err := quiz.NewPoints(serialized.BaseScore)
	if err != nil {
		return nil, fmt.Errorf("invalid base_score: %w", err)
	}

	return kernel.ReconstructQuizGameplaySession(
		sessionID,
		quizAggregate,
		userAnswers,
		serialized.CurrentQuestionIndex,
		baseScore,
		serialized.StartedAt,
		serialized.FinishedAt,
	), nil
}
