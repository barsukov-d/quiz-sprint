package postgres

import (
	"context"
	"database/sql"
	"fmt"

	appUser "github.com/barsukov/quiz-sprint/backend/internal/application/user"
	"github.com/barsukov/quiz-sprint/backend/internal/domain/quick_duel"
)

// ProfileStatsRepository implements appUser.StatsQueryRepository using PostgreSQL.
type ProfileStatsRepository struct {
	db *sql.DB
}

// NewProfileStatsRepository creates a new ProfileStatsRepository.
func NewProfileStatsRepository(db *sql.DB) *ProfileStatsRepository {
	return &ProfileStatsRepository{db: db}
}

// GetProfileStats executes a single aggregation query across all game-mode tables
// and returns the aggregated profile stats for the given user.
func (r *ProfileStatsRepository) GetProfileStats(ctx context.Context, userID string) (*appUser.ProfileStatsOutput, error) {
	query := `
		SELECT
			-- General: quiz sessions completed
			COALESCE(us.total_quizzes_completed, 0)     AS quiz_completed,
			COALESCE(qs.quiz_points, 0)                  AS quiz_points,

			-- Daily challenge counts and points
			COALESCE(dc.daily_count, 0)                  AS daily_count,
			COALESCE(dc.daily_points, 0)                 AS daily_points,

			-- Marathon counts and points
			COALESCE(mr.marathon_count, 0)               AS marathon_count,
			COALESCE(mr.marathon_points, 0)              AS marathon_points,

			-- Duel game count (where user is either player)
			COALESCE(du.duel_count, 0)                   AS duel_count,

			-- Streak from user_stats
			COALESCE(us.current_streak, 0)               AS current_streak,
			COALESCE(us.longest_streak, 0)               AS longest_streak,

			-- Duel rating
			COALESCE(pr.mmr, 0)                          AS duel_mmr,
			COALESCE(pr.league, '')                      AS duel_league,
			COALESCE(pr.division, 0)                     AS duel_division,
			COALESCE(pr.season_wins, 0)                  AS duel_wins,
			COALESCE(pr.season_losses, 0)                AS duel_losses,

			-- Marathon best (top single score row)
			COALESCE(mpb.best_score, 0)                  AS marathon_best_score,
			COALESCE(mpb.best_streak, 0)                 AS marathon_best_streak,
			COALESCE(mpb.category_id, '')                AS marathon_best_category_id

		FROM users u

		LEFT JOIN user_stats us ON us.user_id = u.id

		LEFT JOIN (
			SELECT user_id,
			       SUM(score) AS quiz_points
			FROM quiz_sessions
			WHERE status = 'completed'
			GROUP BY user_id
		) qs ON qs.user_id = u.id

		LEFT JOIN (
			SELECT player_id,
			       COUNT(*) AS daily_count,
			       COALESCE(SUM((session_state->>'base_score')::int), 0) AS daily_points
			FROM daily_games
			WHERE status = 'completed'
			GROUP BY player_id
		) dc ON dc.player_id = u.id

		LEFT JOIN (
			SELECT player_id,
			       COUNT(*) AS marathon_count,
			       COALESCE(SUM(best_score), 0) AS marathon_points
			FROM marathon_personal_bests
			GROUP BY player_id
		) mr ON mr.player_id = u.id

		LEFT JOIN (
			SELECT $1 AS player_id, COUNT(*) AS duel_count
			FROM duel_matches
			WHERE status = 'finished'
			  AND (player1_id = $1 OR player2_id = $1)
		) du ON true

		LEFT JOIN player_ratings pr ON pr.player_id = u.id

		LEFT JOIN LATERAL (
			SELECT best_score, best_streak, COALESCE(category_id::text, '') AS category_id
			FROM marathon_personal_bests
			WHERE player_id = u.id
			ORDER BY best_score DESC
			LIMIT 1
		) mpb ON true

		WHERE u.id = $1
	`

	var (
		quizCompleted          int
		quizPoints             int
		dailyCount             int
		dailyPoints            int
		marathonCount          int
		marathonPoints         int
		duelCount              int
		currentStreak          int
		longestStreak          int
		duelMMR                int
		duelLeagueStr          string
		duelDivision           int
		duelWins               int
		duelLosses             int
		marathonBestScore      int
		marathonBestStreak     int
		marathonBestCategoryID string
	)

	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&quizCompleted,
		&quizPoints,
		&dailyCount,
		&dailyPoints,
		&marathonCount,
		&marathonPoints,
		&duelCount,
		&currentStreak,
		&longestStreak,
		&duelMMR,
		&duelLeagueStr,
		&duelDivision,
		&duelWins,
		&duelLosses,
		&marathonBestScore,
		&marathonBestStreak,
		&marathonBestCategoryID,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found: %s", userID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query profile stats: %w", err)
	}

	// Aggregate totals
	totalGames := quizCompleted + dailyCount + marathonCount + duelCount
	totalPoints := quizPoints + dailyPoints + marathonPoints

	// Build league label and icon from MMR using domain logic
	duelLeagueLabel := ""
	duelLeagueIcon := ""
	if duelMMR > 0 || duelLeagueStr != "" {
		league := stringToLeague(duelLeagueStr)
		leagueInfo := quick_duel.GetLeagueFromMMR(duelMMR)
		duelLeagueLabel = leagueInfo.Label()
		duelLeagueIcon = league.Icon()
	}

	return &appUser.ProfileStatsOutput{
		TotalGamesCompleted: totalGames,
		TotalPoints:         totalPoints,
		CurrentStreak:       currentStreak,
		LongestStreak:       longestStreak,
		DuelMMR:             duelMMR,
		DuelLeague:          duelLeagueStr,
		DuelDivision:        duelDivision,
		DuelLeagueLabel:     duelLeagueLabel,
		DuelLeagueIcon:      duelLeagueIcon,
		DuelWins:            duelWins,
		DuelLosses:          duelLosses,
		MarathonBestScore:   marathonBestScore,
		MarathonBestStreak:  marathonBestStreak,
		MarathonCategory:    marathonBestCategoryID,
	}, nil
}
