package handlers

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
)

// AdminHandler handles admin/debug endpoints for testing.
// Uses direct SQL queries (no DDD layers) — this is intentional for dev tooling.
type AdminHandler struct {
	db *sql.DB
}

func NewAdminHandler(db *sql.DB) *AdminHandler {
	return &AdminHandler{db: db}
}

// AdminKeyMiddleware validates X-Admin-Key header against ADMIN_API_KEY env var
func AdminKeyMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		expected := os.Getenv("ADMIN_API_KEY")
		if expected == "" {
			// No key configured — admin disabled
			return fiber.NewError(fiber.StatusForbidden, "Admin API not configured")
		}
		if c.Get("X-Admin-Key") != expected {
			return fiber.NewError(fiber.StatusUnauthorized, "Invalid admin key")
		}
		return c.Next()
	}
}

// UpdateStreak handles PATCH /api/v1/admin/daily-challenge/streak
// @Summary Update player streak
// @Description Set streak values for testing (updates the latest game for given date)
// @Tags admin
// @Accept json
// @Produce json
// @Param X-Admin-Key header string true "Admin API key"
// @Param request body AdminUpdateStreakRequest true "Streak update"
// @Success 200 {object} AdminUpdateStreakResponse "Updated"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 404 {object} ErrorResponse "No game found"
// @Router /admin/daily-challenge/streak [patch]
func (h *AdminHandler) UpdateStreak(c fiber.Ctx) error {
	var req AdminUpdateStreakRequest
	if err := c.Bind().Body(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}
	if req.PlayerID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "playerId is required")
	}

	// Build dynamic SET clause
	setClauses := ""
	args := []interface{}{}
	argIdx := 1

	if req.CurrentStreak != nil {
		setClauses += fmt.Sprintf("current_streak = $%d, ", argIdx)
		args = append(args, *req.CurrentStreak)
		argIdx++
	}
	if req.BestStreak != nil {
		setClauses += fmt.Sprintf("best_streak = $%d, ", argIdx)
		args = append(args, *req.BestStreak)
		argIdx++
	}
	if req.LastPlayedDate != nil {
		setClauses += fmt.Sprintf("last_played_date = $%d, ", argIdx)
		args = append(args, *req.LastPlayedDate)
		argIdx++
	}

	if setClauses == "" {
		return fiber.NewError(fiber.StatusBadRequest, "No fields to update")
	}
	// Remove trailing ", "
	setClauses = setClauses[:len(setClauses)-2]

	// Update ALL games for this player (so streak is consistent across attempts)
	query := fmt.Sprintf(`UPDATE daily_games SET %s WHERE player_id = $%d`, setClauses, argIdx)
	args = append(args, req.PlayerID)

	result, err := h.db.Exec(query, args...)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to update: "+err.Error())
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fiber.NewError(fiber.StatusNotFound, "No games found for player")
	}

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"updated":  rows,
			"playerId": req.PlayerID,
		},
	})
}

// DeleteGames handles DELETE /api/v1/admin/daily-challenge/games
// @Summary Delete player games
// @Description Delete games for a player on a specific date (or all dates)
// @Tags admin
// @Accept json
// @Produce json
// @Param X-Admin-Key header string true "Admin API key"
// @Param playerId query string true "Player ID"
// @Param date query string false "Date (YYYY-MM-DD). If empty, deletes ALL games"
// @Success 200 {object} AdminDeleteGamesResponse "Deleted"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Router /admin/daily-challenge/games [delete]
func (h *AdminHandler) DeleteGames(c fiber.Ctx) error {
	playerID := c.Query("playerId")
	if playerID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "playerId is required")
	}

	date := c.Query("date")
	var result sql.Result
	var err error

	if date != "" {
		result, err = h.db.Exec(
			`DELETE FROM daily_games WHERE player_id = $1 AND date = $2`,
			playerID, date,
		)
	} else {
		result, err = h.db.Exec(
			`DELETE FROM daily_games WHERE player_id = $1`,
			playerID,
		)
	}

	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to delete: "+err.Error())
	}

	rows, _ := result.RowsAffected()
	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"deleted":  rows,
			"playerId": playerID,
			"date":     date,
		},
	})
}

// ListGames handles GET /api/v1/admin/daily-challenge/games
// @Summary List player games
// @Description List all daily games for a player (debug view)
// @Tags admin
// @Produce json
// @Param X-Admin-Key header string true "Admin API key"
// @Param playerId query string true "Player ID"
// @Param limit query int false "Limit (default 20)"
// @Success 200 {object} AdminListGamesResponse "Games list"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Router /admin/daily-challenge/games [get]
func (h *AdminHandler) ListGames(c fiber.Ctx) error {
	playerID := c.Query("playerId")
	if playerID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "playerId is required")
	}

	limit := fiber.Query[int](c, "limit", 20)
	if limit < 1 || limit > 100 {
		limit = 20
	}

	rows, err := h.db.Query(`
		SELECT id, date, status, attempt_number,
			   current_streak, best_streak, last_played_date,
			   (session_state->>'base_score')::int as base_score,
			   chest_type, chest_coins, rank
		FROM daily_games
		WHERE player_id = $1
		ORDER BY date DESC, attempt_number ASC
		LIMIT $2
	`, playerID, limit)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Query failed: "+err.Error())
	}
	defer rows.Close()

	var games []fiber.Map
	for rows.Next() {
		var (
			id, date, status   string
			attempt, streak    int
			bestStreak, score  int
			lastPlayed         sql.NullString
			chestType          sql.NullString
			chestCoins, rank   sql.NullInt64
		)
		if err := rows.Scan(&id, &date, &status, &attempt,
			&streak, &bestStreak, &lastPlayed,
			&score, &chestType, &chestCoins, &rank); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Scan failed: "+err.Error())
		}

		game := fiber.Map{
			"id":             id,
			"date":           date,
			"status":         status,
			"attemptNumber":  attempt,
			"currentStreak":  streak,
			"bestStreak":     bestStreak,
			"lastPlayedDate": lastPlayed.String,
			"baseScore":      score,
		}
		if chestType.Valid {
			game["chestType"] = chestType.String
		}
		if chestCoins.Valid {
			game["chestCoins"] = chestCoins.Int64
		}
		if rank.Valid {
			game["rank"] = rank.Int64
		}
		games = append(games, game)
	}

	if games == nil {
		games = []fiber.Map{}
	}

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"playerId": playerID,
			"games":    games,
			"count":    len(games),
		},
	})
}

// ResetPlayer handles DELETE /api/v1/admin/player/reset
// @Summary Full player reset
// @Description Delete ALL player data: daily games, marathon games, personal bests, quiz sessions, user stats. User profile is preserved.
// @Tags admin
// @Produce json
// @Param X-Admin-Key header string true "Admin API key"
// @Param playerId query string true "Player ID"
// @Success 200 {object} AdminResetPlayerResponse "Reset complete"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Router /admin/player/reset [delete]
func (h *AdminHandler) ResetPlayer(c fiber.Ctx) error {
	playerID := c.Query("playerId")
	if playerID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "playerId is required")
	}

	type deleteResult struct {
		table string
		count int64
	}

	results := []deleteResult{}

	// Order matters: user_answers cascade from quiz_sessions, so sessions first is fine
	queries := []struct {
		table string
		sql   string
	}{
		{"quiz_sessions", `DELETE FROM quiz_sessions WHERE user_id = $1`},
		{"daily_games", `DELETE FROM daily_games WHERE player_id = $1`},
		{"marathon_games", `DELETE FROM marathon_games WHERE player_id = $1`},
		{"marathon_personal_bests", `DELETE FROM marathon_personal_bests WHERE player_id = $1`},
		{"user_stats", `DELETE FROM user_stats WHERE user_id = $1`},
	}

	for _, q := range queries {
		res, err := h.db.Exec(q.sql, playerID)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to delete from "+q.table+": "+err.Error())
		}
		rows, _ := res.RowsAffected()
		results = append(results, deleteResult{table: q.table, count: rows})
	}

	deleted := fiber.Map{}
	totalDeleted := int64(0)
	for _, r := range results {
		deleted[r.table] = r.count
		totalDeleted += r.count
	}

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"playerId":     playerID,
			"totalDeleted": totalDeleted,
			"deleted":      deleted,
		},
	})
}

// SimulateStreak handles POST /api/v1/admin/daily-challenge/simulate-streak
// @Summary Simulate streak
// @Description Create fake completed games for N consecutive days to build up a streak
// @Tags admin
// @Accept json
// @Produce json
// @Param X-Admin-Key header string true "Admin API key"
// @Param request body AdminSimulateStreakRequest true "Simulate request"
// @Success 201 {object} AdminSimulateStreakResponse "Simulated"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Router /admin/daily-challenge/simulate-streak [post]
func (h *AdminHandler) SimulateStreak(c fiber.Ctx) error {
	var req AdminSimulateStreakRequest
	if err := c.Bind().Body(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}
	if req.PlayerID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "playerId is required")
	}
	if req.Days < 1 || req.Days > 365 {
		return fiber.NewError(fiber.StatusBadRequest, "days must be 1-365")
	}
	if req.BaseScore < 0 {
		req.BaseScore = 40
	}

	today := time.Now().UTC().Truncate(24 * time.Hour)
	created := 0

	for i := req.Days; i >= 1; i-- {
		date := today.AddDate(0, 0, -i)
		dateStr := date.Format("2006-01-02")
		streak := req.Days - i + 1

		// Ensure daily_quiz exists for this date
		var dailyQuizID string
		err := h.db.QueryRow(
			`SELECT id FROM daily_quizzes WHERE date = $1`, dateStr,
		).Scan(&dailyQuizID)

		if err == sql.ErrNoRows {
			// Create a minimal daily quiz with random questions
			dailyQuizID = fmt.Sprintf("00000000-0000-0000-0000-%012d", i)
			expiresAt := date.Add(24 * time.Hour).Unix()

			// Get 10 random question IDs
			qRows, err := h.db.Query(`SELECT id FROM questions ORDER BY RANDOM() LIMIT 10`)
			if err != nil {
				return fiber.NewError(fiber.StatusInternalServerError, "Failed to get questions: "+err.Error())
			}
			var qIDs []string
			for qRows.Next() {
				var qid string
				qRows.Scan(&qid)
				qIDs = append(qIDs, qid)
			}
			qRows.Close()

			if len(qIDs) < 10 {
				return fiber.NewError(fiber.StatusBadRequest, "Not enough questions in DB (need 10, have "+strconv.Itoa(len(qIDs))+")")
			}

			// Build JSONB array
			questionIDsJSON := "["
			for j, qid := range qIDs {
				if j > 0 {
					questionIDsJSON += ","
				}
				questionIDsJSON += `"` + qid + `"`
			}
			questionIDsJSON += "]"

			_, err = h.db.Exec(
				`INSERT INTO daily_quizzes (id, date, question_ids, expires_at, created_at)
				 VALUES ($1, $2, $3::jsonb, $4, $5)
				 ON CONFLICT (date) DO UPDATE SET id = daily_quizzes.id
				 RETURNING id`,
				dailyQuizID, dateStr, questionIDsJSON, expiresAt, date.Unix(),
			)
			if err != nil {
				return fiber.NewError(fiber.StatusInternalServerError, "Failed to create daily quiz: "+err.Error())
			}
			// Re-read actual ID in case of conflict
			h.db.QueryRow(`SELECT id FROM daily_quizzes WHERE date = $1`, dateStr).Scan(&dailyQuizID)
		} else if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to check daily quiz: "+err.Error())
		}

		// Build minimal session_state JSON
		sessionState := fmt.Sprintf(`{"base_score": %d, "user_answers": {}, "current_question_index": 10, "started_at": %d, "finished_at": %d}`,
			req.BaseScore, date.Unix(), date.Add(5*time.Minute).Unix())

		gameID := fmt.Sprintf("aaaaaaaa-bbbb-cccc-dddd-%012d", i)
		questionStartedAt := date.Unix()

		_, err = h.db.Exec(`
			INSERT INTO daily_games (id, player_id, daily_quiz_id, date, status, session_state,
				current_streak, best_streak, last_played_date, attempt_number, question_started_at)
			VALUES ($1, $2, $3, $4, 'completed', $5::jsonb, $6, $7, $8, 1, $9)
			ON CONFLICT (player_id, date, attempt_number) DO UPDATE SET
				current_streak = EXCLUDED.current_streak,
				best_streak = EXCLUDED.best_streak,
				last_played_date = EXCLUDED.last_played_date
		`, gameID, req.PlayerID, dailyQuizID, dateStr, sessionState,
			streak, streak, dateStr, questionStartedAt)

		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to insert game for "+dateStr+": "+err.Error())
		}
		created++
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"data": fiber.Map{
			"playerId":     req.PlayerID,
			"daysCreated":  created,
			"streakBuilt":  req.Days,
			"dateRange": fiber.Map{
				"from": today.AddDate(0, 0, -req.Days).Format("2006-01-02"),
				"to":   today.AddDate(0, 0, -1).Format("2006-01-02"),
			},
		},
	})
}
