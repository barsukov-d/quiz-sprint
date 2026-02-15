package postgres

import (
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/barsukov/quiz-sprint/backend/internal/domain/quick_duel"
	"github.com/barsukov/quiz-sprint/backend/internal/domain/quiz"
	"github.com/barsukov/quiz-sprint/backend/internal/domain/shared"
)

type DuelGameRepository struct {
	db *sql.DB
}

func NewDuelGameRepository(db *sql.DB) *DuelGameRepository {
	return &DuelGameRepository{db: db}
}

func (r *DuelGameRepository) Save(game *quick_duel.DuelGame) error {
	questionIDsJSON, err := json.Marshal(questionIDsToStrings(game.QuestionIDs()))
	if err != nil {
		return err
	}

	roundAnswersJSON, err := json.Marshal(map[string]interface{}{})
	if err != nil {
		return err
	}

	query := `
		INSERT INTO duel_matches (
			id, status, player1_id, player2_id, winner_id,
			player1_score, player2_score, player1_total_time, player2_total_time,
			player1_mmr_before, player2_mmr_before, player1_mmr_after, player2_mmr_after,
			win_reason, is_friend_match, current_round,
			question_ids, round_answers, started_at, finished_at, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21)
		ON CONFLICT (id) DO UPDATE SET
			status = EXCLUDED.status,
			winner_id = EXCLUDED.winner_id,
			player1_score = EXCLUDED.player1_score,
			player2_score = EXCLUDED.player2_score,
			player1_total_time = EXCLUDED.player1_total_time,
			player2_total_time = EXCLUDED.player2_total_time,
			player1_mmr_after = EXCLUDED.player1_mmr_after,
			player2_mmr_after = EXCLUDED.player2_mmr_after,
			win_reason = EXCLUDED.win_reason,
			current_round = EXCLUDED.current_round,
			round_answers = EXCLUDED.round_answers,
			started_at = EXCLUDED.started_at,
			finished_at = EXCLUDED.finished_at
	`

	var winnerID *string
	if game.Player1().Score() > game.Player2().Score() {
		id := game.Player1().UserID().String()
		winnerID = &id
	} else if game.Player2().Score() > game.Player1().Score() {
		id := game.Player2().UserID().String()
		winnerID = &id
	}

	var player1MMRAfter, player2MMRAfter *int
	if game.Status() == quick_duel.GameStatusFinished {
		p1mmr := game.Player1().Elo().Rating()
		p2mmr := game.Player2().Elo().Rating()
		player1MMRAfter = &p1mmr
		player2MMRAfter = &p2mmr
	}

	var startedAt, finishedAt *int64
	if game.StartedAt() > 0 {
		sa := game.StartedAt()
		startedAt = &sa
	}
	if game.FinishedAt() > 0 {
		fa := game.FinishedAt()
		finishedAt = &fa
	}

	_, err = r.db.Exec(query,
		game.ID().String(),
		string(game.Status()),
		game.Player1().UserID().String(),
		game.Player2().UserID().String(),
		winnerID,
		game.Player1().Score(),
		game.Player2().Score(),
		0, // player1_total_time (TODO)
		0, // player2_total_time (TODO)
		game.Player1().Elo().Rating(),
		game.Player2().Elo().Rating(),
		player1MMRAfter,
		player2MMRAfter,
		nil, // win_reason (TODO)
		false, // is_friend_match (TODO)
		game.CurrentRound(),
		questionIDsJSON,
		roundAnswersJSON,
		startedAt,
		finishedAt,
		game.StartedAt(),
	)

	return err
}

func (r *DuelGameRepository) FindByID(id quick_duel.GameID) (*quick_duel.DuelGame, error) {
	query := `
		SELECT id, status, player1_id, player2_id,
			player1_score, player2_score,
			player1_mmr_before, player2_mmr_before,
			current_round, question_ids, round_answers,
			started_at, finished_at
		FROM duel_matches
		WHERE id = $1
	`

	return r.scanGame(r.db.QueryRow(query, id.String()))
}

func (r *DuelGameRepository) FindActiveByPlayer(playerID quick_duel.UserID) (*quick_duel.DuelGame, error) {
	query := `
		SELECT id, status, player1_id, player2_id,
			player1_score, player2_score,
			player1_mmr_before, player2_mmr_before,
			current_round, question_ids, round_answers,
			started_at, finished_at
		FROM duel_matches
		WHERE (player1_id = $1 OR player2_id = $1)
		AND status IN ('waiting_start', 'in_progress')
		ORDER BY created_at DESC
		LIMIT 1
	`

	return r.scanGame(r.db.QueryRow(query, playerID.String()))
}

func (r *DuelGameRepository) FindByPlayerPaginated(playerID quick_duel.UserID, limit int, offset int, filter string) ([]*quick_duel.DuelGame, int, error) {
	baseQuery := `
		FROM duel_matches
		WHERE (player1_id = $1 OR player2_id = $1)
		AND status = 'finished'
	`

	switch filter {
	case "wins":
		baseQuery += " AND winner_id = $1"
	case "losses":
		baseQuery += " AND winner_id IS NOT NULL AND winner_id != $1"
	case "friends":
		baseQuery += " AND is_friend_match = TRUE"
	}

	// Get total count
	countQuery := "SELECT COUNT(*) " + baseQuery
	var total int
	err := r.db.QueryRow(countQuery, playerID.String()).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get paginated results
	query := `
		SELECT id, status, player1_id, player2_id,
			player1_score, player2_score,
			player1_mmr_before, player2_mmr_before,
			current_round, question_ids, round_answers,
			started_at, finished_at
	` + baseQuery + `
		ORDER BY finished_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(query, playerID.String(), limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	games, err := r.scanGames(rows)
	if err != nil {
		return nil, 0, err
	}

	return games, total, nil
}

func (r *DuelGameRepository) Delete(id quick_duel.GameID) error {
	query := `DELETE FROM duel_matches WHERE id = $1`
	_, err := r.db.Exec(query, id.String())
	return err
}

func (r *DuelGameRepository) scanGame(row *sql.Row) (*quick_duel.DuelGame, error) {
	var (
		id             string
		status         string
		player1ID      string
		player2ID      string
		player1Score   int
		player2Score   int
		player1MMR     int
		player2MMR     int
		currentRound   int
		questionIDsJSON []byte
		roundAnswersJSON []byte
		startedAt      sql.NullInt64
		finishedAt     sql.NullInt64
	)

	err := row.Scan(
		&id, &status, &player1ID, &player2ID,
		&player1Score, &player2Score,
		&player1MMR, &player2MMR,
		&currentRound, &questionIDsJSON, &roundAnswersJSON,
		&startedAt, &finishedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, quick_duel.ErrGameNotFound
	}
	if err != nil {
		return nil, err
	}

	return r.reconstructGame(
		id, status, player1ID, player2ID,
		player1Score, player2Score,
		player1MMR, player2MMR,
		currentRound, questionIDsJSON, roundAnswersJSON,
		startedAt, finishedAt,
	)
}

func (r *DuelGameRepository) scanGames(rows *sql.Rows) ([]*quick_duel.DuelGame, error) {
	var games []*quick_duel.DuelGame

	for rows.Next() {
		var (
			id             string
			status         string
			player1ID      string
			player2ID      string
			player1Score   int
			player2Score   int
			player1MMR     int
			player2MMR     int
			currentRound   int
			questionIDsJSON []byte
			roundAnswersJSON []byte
			startedAt      sql.NullInt64
			finishedAt     sql.NullInt64
		)

		err := rows.Scan(
			&id, &status, &player1ID, &player2ID,
			&player1Score, &player2Score,
			&player1MMR, &player2MMR,
			&currentRound, &questionIDsJSON, &roundAnswersJSON,
			&startedAt, &finishedAt,
		)
		if err != nil {
			return nil, err
		}

		game, err := r.reconstructGame(
			id, status, player1ID, player2ID,
			player1Score, player2Score,
			player1MMR, player2MMR,
			currentRound, questionIDsJSON, roundAnswersJSON,
			startedAt, finishedAt,
		)
		if err != nil {
			return nil, err
		}

		games = append(games, game)
	}

	return games, nil
}

func (r *DuelGameRepository) reconstructGame(
	id, status, player1ID, player2ID string,
	player1Score, player2Score int,
	player1MMR, player2MMR int,
	currentRound int,
	questionIDsJSON, roundAnswersJSON []byte,
	startedAt, finishedAt sql.NullInt64,
) (*quick_duel.DuelGame, error) {
	// Parse question IDs
	var questionIDStrs []string
	if err := json.Unmarshal(questionIDsJSON, &questionIDStrs); err != nil {
		return nil, err
	}

	questionIDs := make([]quick_duel.QuestionID, 0, len(questionIDStrs))
	for _, idStr := range questionIDStrs {
		qid, _ := quiz.NewQuestionIDFromString(idStr)
		questionIDs = append(questionIDs, qid)
	}

	// Create players
	p1id, _ := shared.NewUserID(player1ID)
	p2id, _ := shared.NewUserID(player2ID)

	player1 := quick_duel.NewDuelPlayer(
		p1id,
		"Player1", // TODO: get username
		quick_duel.ReconstructEloRating(player1MMR, 0),
	)
	// Simulate adding score
	for i := 0; i < player1Score/100; i++ {
		player1 = player1.AddScore(100)
	}

	player2 := quick_duel.NewDuelPlayer(
		p2id,
		"Player2",
		quick_duel.ReconstructEloRating(player2MMR, 0),
	)
	for i := 0; i < player2Score/100; i++ {
		player2 = player2.AddScore(100)
	}

	var sa, fa int64
	if startedAt.Valid {
		sa = startedAt.Int64
	}
	if finishedAt.Valid {
		fa = finishedAt.Int64
	}

	return quick_duel.ReconstructDuelGame(
		quick_duel.NewGameIDFromString(id),
		player1,
		player2,
		questionIDs,
		currentRound,
		quick_duel.GameStatus(status),
		make(map[int][]quick_duel.RoundAnswer), // TODO: parse round answers
		sa,
		fa,
	), nil
}

// Helper function to convert QuestionIDs to strings
func questionIDsToStrings(ids []quick_duel.QuestionID) []string {
	strs := make([]string, 0, len(ids))
	for _, id := range ids {
		strs = append(strs, id.String())
	}
	return strs
}
