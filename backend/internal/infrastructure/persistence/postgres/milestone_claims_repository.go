package postgres

import (
	"database/sql"
	"errors"
	"time"
)

// MilestoneClaimsRepository implements marathon.MilestoneClaimsRepository.
type MilestoneClaimsRepository struct {
	db *sql.DB
}

func NewMilestoneClaimsRepository(db *sql.DB) *MilestoneClaimsRepository {
	return &MilestoneClaimsRepository{db: db}
}

// HasClaimed returns true if the player has already claimed the given milestone.
func (r *MilestoneClaimsRepository) HasClaimed(playerID string, milestone int) (bool, error) {
	var dummy int
	err := r.db.QueryRow(
		`SELECT milestone FROM marathon_milestone_claims WHERE player_id = $1 AND milestone = $2`,
		playerID, milestone,
	).Scan(&dummy)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// MarkClaimed records that the player has claimed the given milestone.
func (r *MilestoneClaimsRepository) MarkClaimed(playerID string, milestone int) error {
	_, err := r.db.Exec(
		`INSERT INTO marathon_milestone_claims (player_id, milestone, claimed_at)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (player_id, milestone) DO NOTHING`,
		playerID, milestone, time.Now().UTC().Unix(),
	)
	return err
}
