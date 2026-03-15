package postgres

import (
	"database/sql"
	"errors"
	"time"
)

// WeeklyDistributionRepository is a PostgreSQL implementation of
// marathon.WeeklyRewardDistributionRepository.
type WeeklyDistributionRepository struct {
	db *sql.DB
}

func NewWeeklyDistributionRepository(db *sql.DB) *WeeklyDistributionRepository {
	return &WeeklyDistributionRepository{db: db}
}

// HasDistributed returns true if rewards for the given weekID have already been distributed.
func (r *WeeklyDistributionRepository) HasDistributed(weekID string) (bool, error) {
	var dummy string
	err := r.db.QueryRow(
		`SELECT week_id FROM marathon_weekly_distribution WHERE week_id = $1`,
		weekID,
	).Scan(&dummy)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// MarkDistributed records that rewards for the given weekID have been distributed.
func (r *WeeklyDistributionRepository) MarkDistributed(weekID string) error {
	_, err := r.db.Exec(
		`INSERT INTO marathon_weekly_distribution (week_id, distributed_at)
		 VALUES ($1, $2)
		 ON CONFLICT (week_id) DO NOTHING`,
		weekID,
		time.Now().UTC().Unix(),
	)
	return err
}
