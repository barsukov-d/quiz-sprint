package postgres

import (
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/barsukov/quiz-sprint/backend/internal/domain/quick_duel"
	"github.com/barsukov/quiz-sprint/backend/internal/domain/shared"
)

type ReferralRepository struct {
	db *sql.DB
}

func NewReferralRepository(db *sql.DB) *ReferralRepository {
	return &ReferralRepository{db: db}
}

func (r *ReferralRepository) Save(referral *quick_duel.Referral) error {
	inviterClaimedJSON, err := json.Marshal(referral.InviterRewardsClaimed())
	if err != nil {
		return err
	}

	inviteeClaimedJSON, err := json.Marshal(referral.InviteeRewardsClaimed())
	if err != nil {
		return err
	}

	query := `
		INSERT INTO referrals (
			id, inviter_id, invitee_id,
			milestone_registered, milestone_played_5,
			milestone_reached_silver, milestone_reached_gold, milestone_reached_platinum,
			inviter_rewards_claimed, invitee_rewards_claimed, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (inviter_id, invitee_id) DO UPDATE SET
			milestone_played_5 = EXCLUDED.milestone_played_5,
			milestone_reached_silver = EXCLUDED.milestone_reached_silver,
			milestone_reached_gold = EXCLUDED.milestone_reached_gold,
			milestone_reached_platinum = EXCLUDED.milestone_reached_platinum,
			inviter_rewards_claimed = EXCLUDED.inviter_rewards_claimed,
			invitee_rewards_claimed = EXCLUDED.invitee_rewards_claimed
	`

	_, err = r.db.Exec(query,
		referral.ID().String(),
		referral.InviterID().String(),
		referral.InviteeID().String(),
		referral.MilestoneRegistered(),
		referral.MilestonePlayedFive(),
		referral.MilestoneReachedSilver(),
		referral.MilestoneReachedGold(),
		referral.MilestoneReachedPlatinum(),
		inviterClaimedJSON,
		inviteeClaimedJSON,
		referral.CreatedAt(),
	)

	return err
}

func (r *ReferralRepository) FindByID(id quick_duel.ReferralID) (*quick_duel.Referral, error) {
	query := `
		SELECT id, inviter_id, invitee_id,
			milestone_registered, milestone_played_5,
			milestone_reached_silver, milestone_reached_gold, milestone_reached_platinum,
			inviter_rewards_claimed, invitee_rewards_claimed, created_at
		FROM referrals
		WHERE id = $1
	`

	return r.scanReferral(r.db.QueryRow(query, id.String()))
}

func (r *ReferralRepository) FindByInviterAndInvitee(inviterID quick_duel.UserID, inviteeID quick_duel.UserID) (*quick_duel.Referral, error) {
	query := `
		SELECT id, inviter_id, invitee_id,
			milestone_registered, milestone_played_5,
			milestone_reached_silver, milestone_reached_gold, milestone_reached_platinum,
			inviter_rewards_claimed, invitee_rewards_claimed, created_at
		FROM referrals
		WHERE inviter_id = $1 AND invitee_id = $2
	`

	return r.scanReferral(r.db.QueryRow(query, inviterID.String(), inviteeID.String()))
}

func (r *ReferralRepository) FindByInvitee(inviteeID quick_duel.UserID) (*quick_duel.Referral, error) {
	query := `
		SELECT id, inviter_id, invitee_id,
			milestone_registered, milestone_played_5,
			milestone_reached_silver, milestone_reached_gold, milestone_reached_platinum,
			inviter_rewards_claimed, invitee_rewards_claimed, created_at
		FROM referrals
		WHERE invitee_id = $1
	`

	return r.scanReferral(r.db.QueryRow(query, inviteeID.String()))
}

func (r *ReferralRepository) FindByInviter(inviterID quick_duel.UserID) ([]*quick_duel.Referral, error) {
	query := `
		SELECT id, inviter_id, invitee_id,
			milestone_registered, milestone_played_5,
			milestone_reached_silver, milestone_reached_gold, milestone_reached_platinum,
			inviter_rewards_claimed, invitee_rewards_claimed, created_at
		FROM referrals
		WHERE inviter_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, inviterID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanReferrals(rows)
}

func (r *ReferralRepository) CountByInviter(inviterID quick_duel.UserID) (int, error) {
	query := `SELECT COUNT(*) FROM referrals WHERE inviter_id = $1`

	var count int
	err := r.db.QueryRow(query, inviterID.String()).Scan(&count)
	return count, err
}

func (r *ReferralRepository) CountActiveByInviter(inviterID quick_duel.UserID) (int, error) {
	query := `SELECT COUNT(*) FROM referrals WHERE inviter_id = $1 AND milestone_played_5 = TRUE`

	var count int
	err := r.db.QueryRow(query, inviterID.String()).Scan(&count)
	return count, err
}

func (r *ReferralRepository) GetReferralLeaderboard(limit int) ([]quick_duel.ReferralLeaderboardEntry, error) {
	query := `
		SELECT r.inviter_id, u.username, COUNT(*) as total, COUNT(*) FILTER (WHERE r.milestone_played_5 = TRUE) as active
		FROM referrals r
		LEFT JOIN users u ON u.id = r.inviter_id
		GROUP BY r.inviter_id, u.username
		ORDER BY total DESC
		LIMIT $1
	`

	rows, err := r.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []quick_duel.ReferralLeaderboardEntry
	for rows.Next() {
		var (
			inviterID       string
			username        sql.NullString
			totalReferrals  int
			activeReferrals int
		)

		err := rows.Scan(&inviterID, &username, &totalReferrals, &activeReferrals)
		if err != nil {
			return nil, err
		}

		uid, _ := shared.NewUserID(inviterID)
		un := "Player"
		if username.Valid {
			un = username.String
		}

		entries = append(entries, quick_duel.ReferralLeaderboardEntry{
			PlayerID:        uid,
			Username:        un,
			TotalReferrals:  totalReferrals,
			ActiveReferrals: activeReferrals,
		})
	}

	return entries, nil
}

func (r *ReferralRepository) GetPlayerReferralRank(inviterID quick_duel.UserID) (int, error) {
	query := `
		SELECT rank FROM (
			SELECT inviter_id, RANK() OVER (ORDER BY COUNT(*) DESC) as rank
			FROM referrals
			GROUP BY inviter_id
		) ranked
		WHERE inviter_id = $1
	`

	var rank int
	err := r.db.QueryRow(query, inviterID.String()).Scan(&rank)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}
	return rank, err
}

func (r *ReferralRepository) scanReferral(row *sql.Row) (*quick_duel.Referral, error) {
	var (
		id                       string
		inviterID                string
		inviteeID                string
		milestoneRegistered      bool
		milestonePlayedFive      bool
		milestoneReachedSilver   bool
		milestoneReachedGold     bool
		milestoneReachedPlatinum bool
		inviterClaimedJSON       []byte
		inviteeClaimedJSON       []byte
		createdAt                int64
	)

	err := row.Scan(
		&id, &inviterID, &inviteeID,
		&milestoneRegistered, &milestonePlayedFive,
		&milestoneReachedSilver, &milestoneReachedGold, &milestoneReachedPlatinum,
		&inviterClaimedJSON, &inviteeClaimedJSON, &createdAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, quick_duel.ErrReferralNotFound
	}
	if err != nil {
		return nil, err
	}

	return r.reconstructReferral(
		id, inviterID, inviteeID,
		milestoneRegistered, milestonePlayedFive,
		milestoneReachedSilver, milestoneReachedGold, milestoneReachedPlatinum,
		inviterClaimedJSON, inviteeClaimedJSON, createdAt,
	)
}

func (r *ReferralRepository) scanReferrals(rows *sql.Rows) ([]*quick_duel.Referral, error) {
	var referrals []*quick_duel.Referral

	for rows.Next() {
		var (
			id                       string
			inviterID                string
			inviteeID                string
			milestoneRegistered      bool
			milestonePlayedFive      bool
			milestoneReachedSilver   bool
			milestoneReachedGold     bool
			milestoneReachedPlatinum bool
			inviterClaimedJSON       []byte
			inviteeClaimedJSON       []byte
			createdAt                int64
		)

		err := rows.Scan(
			&id, &inviterID, &inviteeID,
			&milestoneRegistered, &milestonePlayedFive,
			&milestoneReachedSilver, &milestoneReachedGold, &milestoneReachedPlatinum,
			&inviterClaimedJSON, &inviteeClaimedJSON, &createdAt,
		)
		if err != nil {
			return nil, err
		}

		referral, err := r.reconstructReferral(
			id, inviterID, inviteeID,
			milestoneRegistered, milestonePlayedFive,
			milestoneReachedSilver, milestoneReachedGold, milestoneReachedPlatinum,
			inviterClaimedJSON, inviteeClaimedJSON, createdAt,
		)
		if err != nil {
			return nil, err
		}

		referrals = append(referrals, referral)
	}

	return referrals, nil
}

func (r *ReferralRepository) reconstructReferral(
	id, inviterID, inviteeID string,
	milestoneRegistered, milestonePlayedFive bool,
	milestoneReachedSilver, milestoneReachedGold, milestoneReachedPlatinum bool,
	inviterClaimedJSON, inviteeClaimedJSON []byte,
	createdAt int64,
) (*quick_duel.Referral, error) {
	inviterClaimed := make(map[string]bool)
	if err := json.Unmarshal(inviterClaimedJSON, &inviterClaimed); err != nil {
		return nil, err
	}

	inviteeClaimed := make(map[string]bool)
	if err := json.Unmarshal(inviteeClaimedJSON, &inviteeClaimed); err != nil {
		return nil, err
	}

	rid := quick_duel.NewReferralIDFromString(id)
	invUID, _ := shared.NewUserID(inviterID)
	invtUID, _ := shared.NewUserID(inviteeID)

	return quick_duel.ReconstructReferral(
		rid,
		invUID,
		invtUID,
		milestoneRegistered,
		milestonePlayedFive,
		milestoneReachedSilver,
		milestoneReachedGold,
		milestoneReachedPlatinum,
		inviterClaimed,
		inviteeClaimed,
		createdAt,
	), nil
}
