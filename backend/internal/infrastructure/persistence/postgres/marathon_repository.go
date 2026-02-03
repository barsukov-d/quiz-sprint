package postgres

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/barsukov/quiz-sprint/backend/internal/domain/quiz"
	"github.com/barsukov/quiz-sprint/backend/internal/domain/shared"
	"github.com/barsukov/quiz-sprint/backend/internal/domain/solo_marathon"
)

// MarathonRepository is a PostgreSQL implementation of solo_marathon.Repository
type MarathonRepository struct {
	db           *sql.DB
	questionRepo quiz.QuestionRepository
}

// NewMarathonRepository creates a new PostgreSQL marathon repository
func NewMarathonRepository(db *sql.DB, questionRepo quiz.QuestionRepository) *MarathonRepository {
	return &MarathonRepository{
		db:           db,
		questionRepo: questionRepo,
	}
}

// Save persists a marathon game
func (r *MarathonRepository) Save(game *solo_marathon.MarathonGameV2) error {
	// Marshal JSONB fields
	answeredIDsJSON, err := r.marshalQuestionIDs(game.AnsweredQuestionIDs())
	if err != nil {
		return fmt.Errorf("failed to marshal answered_question_ids: %w", err)
	}

	recentIDsJSON, err := r.marshalQuestionIDs(game.RecentQuestionIDs())
	if err != nil {
		return fmt.Errorf("failed to marshal recent_question_ids: %w", err)
	}

	// Get current question ID (nullable)
	var currentQuestionID *string
	if game.CurrentQuestion() != nil {
		qid := game.CurrentQuestion().ID().String()
		currentQuestionID = &qid
	}

	// Upsert query
	query := `
		INSERT INTO marathon_games (
			id, player_id, category_id, status, started_at, finished_at,
			current_question_id, answered_question_ids, recent_question_ids,
			score, total_questions,
			current_lives, lives_last_update,
			bonus_shield, bonus_fifty_fifty, bonus_skip, bonus_freeze,
			shield_active, continue_count,
			difficulty_level, personal_best_score
		) VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8, $9,
			$10, $11,
			$12, $13,
			$14, $15, $16, $17,
			$18, $19,
			$20, $21
		)
		ON CONFLICT (id) DO UPDATE SET
			status = EXCLUDED.status,
			finished_at = EXCLUDED.finished_at,
			current_question_id = EXCLUDED.current_question_id,
			answered_question_ids = EXCLUDED.answered_question_ids,
			recent_question_ids = EXCLUDED.recent_question_ids,
			score = EXCLUDED.score,
			total_questions = EXCLUDED.total_questions,
			current_lives = EXCLUDED.current_lives,
			lives_last_update = EXCLUDED.lives_last_update,
			bonus_shield = EXCLUDED.bonus_shield,
			bonus_fifty_fifty = EXCLUDED.bonus_fifty_fifty,
			bonus_skip = EXCLUDED.bonus_skip,
			bonus_freeze = EXCLUDED.bonus_freeze,
			shield_active = EXCLUDED.shield_active,
			continue_count = EXCLUDED.continue_count,
			difficulty_level = EXCLUDED.difficulty_level
	`

	// Get category ID (nullable for "all categories")
	var categoryID *string
	if !game.Category().IsAllCategories() {
		cid := game.Category().CategoryID().String()
		categoryID = &cid
	}

	_, err = r.db.Exec(query,
		game.ID().String(),
		game.PlayerID().String(),
		categoryID,
		string(game.Status()),
		game.StartedAt(),
		r.nullInt64(game.FinishedAt()),
		currentQuestionID,
		answeredIDsJSON,
		recentIDsJSON,
		game.Score(),
		game.TotalQuestions(),
		game.Lives().CurrentLives(),
		game.Lives().LastUpdate(),
		game.BonusInventory().Shield(),
		game.BonusInventory().FiftyFifty(),
		game.BonusInventory().Skip(),
		game.BonusInventory().Freeze(),
		game.ShieldActive(),
		game.ContinueCount(),
		string(game.Difficulty().Level()),
		r.nullInt(game.PersonalBestScore()),
	)

	if err != nil {
		return fmt.Errorf("failed to save marathon game: %w", err)
	}

	return nil
}

// FindByID retrieves a marathon game by ID
func (r *MarathonRepository) FindByID(id solo_marathon.GameID) (*solo_marathon.MarathonGameV2, error) {
	query := `
		SELECT
			id, player_id, category_id, status, started_at, finished_at,
			current_question_id, answered_question_ids, recent_question_ids,
			score, total_questions,
			current_lives, lives_last_update,
			bonus_shield, bonus_fifty_fifty, bonus_skip, bonus_freeze,
			shield_active, continue_count,
			difficulty_level, personal_best_score
		FROM marathon_games
		WHERE id = $1
	`

	return r.scanGame(r.db.QueryRow(query, id.String()))
}

// FindActiveByPlayer retrieves the active marathon game for a player
func (r *MarathonRepository) FindActiveByPlayer(playerID solo_marathon.UserID) (*solo_marathon.MarathonGameV2, error) {
	query := `
		SELECT
			id, player_id, category_id, status, started_at, finished_at,
			current_question_id, answered_question_ids, recent_question_ids,
			score, total_questions,
			current_lives, lives_last_update,
			bonus_shield, bonus_fifty_fifty, bonus_skip, bonus_freeze,
			shield_active, continue_count,
			difficulty_level, personal_best_score
		FROM marathon_games
		WHERE player_id = $1 AND status IN ('in_progress', 'game_over')
		ORDER BY started_at DESC
		LIMIT 1
	`

	return r.scanGame(r.db.QueryRow(query, playerID.String()))
}

// Delete removes a marathon game
func (r *MarathonRepository) Delete(id solo_marathon.GameID) error {
	query := `DELETE FROM marathon_games WHERE id = $1`

	result, err := r.db.Exec(query, id.String())
	if err != nil {
		return fmt.Errorf("failed to delete marathon game: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return solo_marathon.ErrGameNotFound
	}

	return nil
}

// ========================================
// Helper Methods
// ========================================

// scanGame scans a database row into a MarathonGameV2
func (r *MarathonRepository) scanGame(row *sql.Row) (*solo_marathon.MarathonGameV2, error) {
	var (
		gameID              string
		playerID            string
		categoryID          sql.NullString
		status              string
		startedAt           int64
		finishedAt          sql.NullInt64
		currentQuestionID   sql.NullString
		answeredQuestionIDs []byte
		recentQuestionIDs   []byte
		score               int
		totalQuestions      int
		currentLives        int
		livesLastUpdate     int64
		bonusShield         int
		bonusFiftyFifty     int
		bonusSkip           int
		bonusFreeze         int
		shieldActive        bool
		continueCount       int
		difficultyLevel     string
		personalBestScore   sql.NullInt32
	)

	err := row.Scan(
		&gameID, &playerID, &categoryID, &status, &startedAt, &finishedAt,
		&currentQuestionID, &answeredQuestionIDs, &recentQuestionIDs,
		&score, &totalQuestions,
		&currentLives, &livesLastUpdate,
		&bonusShield, &bonusFiftyFifty, &bonusSkip, &bonusFreeze,
		&shieldActive, &continueCount,
		&difficultyLevel, &personalBestScore,
	)

	if err == sql.ErrNoRows {
		return nil, solo_marathon.ErrGameNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query marathon game: %w", err)
	}

	return r.reconstructGame(
		gameID, playerID, categoryID, status, startedAt, finishedAt,
		currentQuestionID, answeredQuestionIDs, recentQuestionIDs,
		score, totalQuestions,
		currentLives, livesLastUpdate,
		bonusShield, bonusFiftyFifty, bonusSkip, bonusFreeze,
		shieldActive, continueCount,
		difficultyLevel, personalBestScore,
	)
}

// reconstructGame reconstructs a MarathonGameV2 from database values
func (r *MarathonRepository) reconstructGame(
	gameID string,
	playerID string,
	categoryID sql.NullString,
	status string,
	startedAt int64,
	finishedAt sql.NullInt64,
	currentQuestionID sql.NullString,
	answeredQuestionIDsJSON []byte,
	recentQuestionIDsJSON []byte,
	score int,
	totalQuestions int,
	currentLives int,
	livesLastUpdate int64,
	bonusShield int,
	bonusFiftyFifty int,
	bonusSkip int,
	bonusFreeze int,
	shieldActive bool,
	continueCount int,
	difficultyLevel string,
	personalBestScore sql.NullInt32,
) (*solo_marathon.MarathonGameV2, error) {
	// Parse IDs
	id := solo_marathon.NewGameIDFromString(gameID)
	userID, err := shared.NewUserID(playerID)
	if err != nil {
		return nil, fmt.Errorf("invalid player_id: %w", err)
	}

	// Reconstruct category
	var category solo_marathon.MarathonCategory
	if categoryID.Valid {
		catID, err := quiz.NewCategoryIDFromString(categoryID.String)
		if err != nil {
			return nil, fmt.Errorf("invalid category_id: %w", err)
		}
		category = solo_marathon.NewMarathonCategory(catID, "")
	} else {
		category = solo_marathon.NewMarathonCategoryAll()
	}

	// Load current question if exists
	var currentQuestion *quiz.Question
	if currentQuestionID.Valid {
		qid, err := quiz.NewQuestionIDFromString(currentQuestionID.String)
		if err != nil {
			return nil, fmt.Errorf("invalid current_question_id: %w", err)
		}
		currentQuestion, err = r.questionRepo.FindByID(qid)
		if err != nil {
			currentQuestion = nil
		}
	}

	// Unmarshal question IDs
	answeredIDs, err := r.unmarshalQuestionIDs(answeredQuestionIDsJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal answered_question_ids: %w", err)
	}

	recentIDs, err := r.unmarshalQuestionIDs(recentQuestionIDsJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal recent_question_ids: %w", err)
	}

	// Reconstruct value objects
	lives := solo_marathon.ReconstructLivesSystem(currentLives, livesLastUpdate)
	bonuses := solo_marathon.ReconstructBonusInventory(bonusShield, bonusFiftyFifty, bonusSkip, bonusFreeze)

	// Reconstruct difficulty from question index (score = correct answers = approximate question index)
	difficulty := solo_marathon.NewDifficultyProgression().UpdateFromQuestionIndex(score)

	// Extract personal best score
	var pbScore *int
	if personalBestScore.Valid {
		val := int(personalBestScore.Int32)
		pbScore = &val
	}

	// Reconstruct game
	game := solo_marathon.ReconstructMarathonGameV2(
		id,
		userID,
		category,
		solo_marathon.GameStatus(status),
		startedAt,
		r.int64Value(finishedAt),
		currentQuestion,
		answeredIDs,
		recentIDs,
		score,
		totalQuestions,
		lives,
		bonuses,
		difficulty,
		shieldActive,
		continueCount,
		pbScore,
		make(map[solo_marathon.QuestionID][]solo_marathon.BonusType),
	)

	return game, nil
}

// marshalQuestionIDs marshals question IDs to JSONB
func (r *MarathonRepository) marshalQuestionIDs(ids []solo_marathon.QuestionID) ([]byte, error) {
	if len(ids) == 0 {
		return []byte("[]"), nil
	}

	stringIDs := make([]string, len(ids))
	for i, id := range ids {
		stringIDs[i] = id.String()
	}

	return json.Marshal(stringIDs)
}

// unmarshalQuestionIDs unmarshals question IDs from JSONB
func (r *MarathonRepository) unmarshalQuestionIDs(data []byte) ([]solo_marathon.QuestionID, error) {
	if len(data) == 0 || string(data) == "null" {
		return []solo_marathon.QuestionID{}, nil
	}

	var stringIDs []string
	if err := json.Unmarshal(data, &stringIDs); err != nil {
		return nil, err
	}

	ids := make([]solo_marathon.QuestionID, len(stringIDs))
	for i, str := range stringIDs {
		id, err := quiz.NewQuestionIDFromString(str)
		if err != nil {
			return nil, fmt.Errorf("invalid question_id in array: %w", err)
		}
		ids[i] = id
	}

	return ids, nil
}

// nullInt64 converts int64 to sql.NullInt64
func (r *MarathonRepository) nullInt64(val int64) sql.NullInt64 {
	if val == 0 {
		return sql.NullInt64{Valid: false}
	}
	return sql.NullInt64{Int64: val, Valid: true}
}

// nullInt converts *int to sql.NullInt32
func (r *MarathonRepository) nullInt(val *int) sql.NullInt32 {
	if val == nil {
		return sql.NullInt32{Valid: false}
	}
	return sql.NullInt32{Int32: int32(*val), Valid: true}
}

// int64Value extracts value from sql.NullInt64
func (r *MarathonRepository) int64Value(val sql.NullInt64) int64 {
	if !val.Valid {
		return 0
	}
	return val.Int64
}
