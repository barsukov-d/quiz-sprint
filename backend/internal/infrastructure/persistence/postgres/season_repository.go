package postgres

import (
	"database/sql"
	"errors"
)

type SeasonRepository struct {
	db *sql.DB
}

func NewSeasonRepository(db *sql.DB) *SeasonRepository {
	return &SeasonRepository{db: db}
}

func (r *SeasonRepository) GetCurrentSeason() (string, error) {
	query := `
		SELECT id FROM seasons
		WHERE status = 'active'
		ORDER BY starts_at DESC
		LIMIT 1
	`

	var seasonID string
	err := r.db.QueryRow(query).Scan(&seasonID)
	if errors.Is(err, sql.ErrNoRows) {
		// Return current month as default
		return "2026-02", nil
	}
	if err != nil {
		return "", err
	}

	return seasonID, nil
}

func (r *SeasonRepository) CreateSeason(seasonID string, startsAt int64, endsAt int64) error {
	query := `
		INSERT INTO seasons (id, starts_at, ends_at, status)
		VALUES ($1, $2, $3, 'active')
		ON CONFLICT (id) DO NOTHING
	`

	_, err := r.db.Exec(query, seasonID, startsAt, endsAt)
	return err
}

func (r *SeasonRepository) GetSeasonInfo(seasonID string) (int64, int64, error) {
	query := `SELECT starts_at, ends_at FROM seasons WHERE id = $1`

	var startsAt, endsAt int64
	err := r.db.QueryRow(query, seasonID).Scan(&startsAt, &endsAt)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, 0, nil
	}
	if err != nil {
		return 0, 0, err
	}

	return startsAt, endsAt, nil
}
