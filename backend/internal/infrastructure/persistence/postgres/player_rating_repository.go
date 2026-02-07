package postgres

import (
	"database/sql"
	"errors"

	"github.com/barsukov/quiz-sprint/backend/internal/domain/quick_duel"
	"github.com/barsukov/quiz-sprint/backend/internal/domain/shared"
)

type PlayerRatingRepository struct {
	db *sql.DB
}

func NewPlayerRatingRepository(db *sql.DB) *PlayerRatingRepository {
	return &PlayerRatingRepository{db: db}
}

func (r *PlayerRatingRepository) Save(rating *quick_duel.PlayerRating) error {
	query := `
		INSERT INTO player_ratings (
			player_id, mmr, league, division,
			peak_mmr, peak_league, peak_division,
			games_at_rank, season_id, season_wins, season_losses, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		ON CONFLICT (player_id) DO UPDATE SET
			mmr = EXCLUDED.mmr,
			league = EXCLUDED.league,
			division = EXCLUDED.division,
			peak_mmr = EXCLUDED.peak_mmr,
			peak_league = EXCLUDED.peak_league,
			peak_division = EXCLUDED.peak_division,
			games_at_rank = EXCLUDED.games_at_rank,
			season_id = EXCLUDED.season_id,
			season_wins = EXCLUDED.season_wins,
			season_losses = EXCLUDED.season_losses,
			updated_at = EXCLUDED.updated_at
	`

	_, err := r.db.Exec(query,
		rating.PlayerID().String(),
		rating.MMR(),
		rating.League().String(),
		rating.Division().Value(),
		rating.PeakMMR(),
		rating.PeakLeague().String(),
		rating.PeakDivision().Value(),
		rating.GamesAtRank(),
		rating.SeasonID(),
		rating.SeasonWins(),
		rating.SeasonLosses(),
		rating.UpdatedAt(),
	)

	return err
}

func (r *PlayerRatingRepository) FindByPlayerID(playerID quick_duel.UserID) (*quick_duel.PlayerRating, error) {
	query := `
		SELECT player_id, mmr, league, division,
			peak_mmr, peak_league, peak_division,
			games_at_rank, season_id, season_wins, season_losses, updated_at
		FROM player_ratings
		WHERE player_id = $1
	`

	row := r.db.QueryRow(query, playerID.String())

	var (
		playerIDStr    string
		mmr            int
		leagueStr      string
		division       int
		peakMMR        int
		peakLeagueStr  string
		peakDivision   int
		gamesAtRank    int
		seasonID       string
		seasonWins     int
		seasonLosses   int
		updatedAt      int64
	)

	err := row.Scan(
		&playerIDStr, &mmr, &leagueStr, &division,
		&peakMMR, &peakLeagueStr, &peakDivision,
		&gamesAtRank, &seasonID, &seasonWins, &seasonLosses, &updatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, quick_duel.ErrGameNotFound
	}
	if err != nil {
		return nil, err
	}

	pid, _ := shared.NewUserID(playerIDStr)
	league := stringToLeague(leagueStr)
	peakLeague := stringToLeague(peakLeagueStr)

	return quick_duel.ReconstructPlayerRating(
		pid,
		mmr,
		league,
		quick_duel.Division(division),
		peakMMR,
		peakLeague,
		quick_duel.Division(peakDivision),
		gamesAtRank,
		seasonID,
		seasonWins,
		seasonLosses,
		updatedAt,
	), nil
}

func (r *PlayerRatingRepository) FindOrCreate(playerID quick_duel.UserID, seasonID string, createdAt int64) (*quick_duel.PlayerRating, error) {
	rating, err := r.FindByPlayerID(playerID)
	if err == nil {
		return rating, nil
	}

	if !errors.Is(err, quick_duel.ErrGameNotFound) {
		return nil, err
	}

	// Create new rating
	rating = quick_duel.NewPlayerRating(playerID, seasonID, createdAt)
	if err := r.Save(rating); err != nil {
		return nil, err
	}

	return rating, nil
}

func (r *PlayerRatingRepository) GetLeaderboard(seasonID string, limit int, offset int) ([]*quick_duel.PlayerRating, error) {
	query := `
		SELECT player_id, mmr, league, division,
			peak_mmr, peak_league, peak_division,
			games_at_rank, season_id, season_wins, season_losses, updated_at
		FROM player_ratings
		WHERE season_id = $1
		ORDER BY mmr DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(query, seasonID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ratings []*quick_duel.PlayerRating
	for rows.Next() {
		var (
			playerIDStr    string
			mmr            int
			leagueStr      string
			division       int
			peakMMR        int
			peakLeagueStr  string
			peakDivision   int
			gamesAtRank    int
			sid            string
			seasonWins     int
			seasonLosses   int
			updatedAt      int64
		)

		err := rows.Scan(
			&playerIDStr, &mmr, &leagueStr, &division,
			&peakMMR, &peakLeagueStr, &peakDivision,
			&gamesAtRank, &sid, &seasonWins, &seasonLosses, &updatedAt,
		)
		if err != nil {
			return nil, err
		}

		pid, _ := shared.NewUserID(playerIDStr)
		league := stringToLeague(leagueStr)
		peakLeague := stringToLeague(peakLeagueStr)

		ratings = append(ratings, quick_duel.ReconstructPlayerRating(
			pid, mmr, league, quick_duel.Division(division),
			peakMMR, peakLeague, quick_duel.Division(peakDivision),
			gamesAtRank, sid, seasonWins, seasonLosses, updatedAt,
		))
	}

	return ratings, nil
}

func (r *PlayerRatingRepository) GetFriendsLeaderboard(playerID quick_duel.UserID, friendIDs []quick_duel.UserID, limit int) ([]*quick_duel.PlayerRating, error) {
	// TODO: Implement friends leaderboard
	return nil, nil
}

func (r *PlayerRatingRepository) GetPlayerRank(playerID quick_duel.UserID, seasonID string) (int, error) {
	query := `
		SELECT rank FROM (
			SELECT player_id, RANK() OVER (ORDER BY mmr DESC) as rank
			FROM player_ratings
			WHERE season_id = $1
		) ranked
		WHERE player_id = $2
	`

	var rank int
	err := r.db.QueryRow(query, seasonID, playerID.String()).Scan(&rank)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}
	return rank, err
}

func (r *PlayerRatingRepository) GetTotalPlayers(seasonID string) (int, error) {
	query := `SELECT COUNT(*) FROM player_ratings WHERE season_id = $1`

	var count int
	err := r.db.QueryRow(query, seasonID).Scan(&count)
	return count, err
}

// Helper function to convert string to League
func stringToLeague(s string) quick_duel.League {
	switch s {
	case "bronze":
		return quick_duel.LeagueBronze
	case "silver":
		return quick_duel.LeagueSilver
	case "gold":
		return quick_duel.LeagueGold
	case "platinum":
		return quick_duel.LeaguePlatinum
	case "diamond":
		return quick_duel.LeagueDiamond
	case "legend":
		return quick_duel.LeagueLegend
	default:
		return quick_duel.LeagueBronze
	}
}
