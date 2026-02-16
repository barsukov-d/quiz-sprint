package postgres

import (
	"database/sql"
	"fmt"

	"github.com/barsukov/quiz-sprint/backend/internal/domain/shared"
	"github.com/barsukov/quiz-sprint/backend/internal/domain/solo_marathon"
)

// BonusWalletRepository is a PostgreSQL implementation of solo_marathon.BonusWalletRepository
type BonusWalletRepository struct {
	db *sql.DB
}

// NewBonusWalletRepository creates a new PostgreSQL bonus wallet repository
func NewBonusWalletRepository(db *sql.DB) *BonusWalletRepository {
	return &BonusWalletRepository{db: db}
}

// FindByPlayer retrieves the bonus wallet for a player
func (r *BonusWalletRepository) FindByPlayer(playerID solo_marathon.UserID) (*solo_marathon.BonusWallet, error) {
	query := `
		SELECT player_id, bonus_shield, bonus_fifty_fifty, bonus_skip, bonus_freeze
		FROM player_bonus_wallet
		WHERE player_id = $1
	`

	var (
		dbPlayerID string
		shield     int
		fiftyFifty int
		skip       int
		freeze     int
	)

	err := r.db.QueryRow(query, playerID.String()).Scan(
		&dbPlayerID, &shield, &fiftyFifty, &skip, &freeze,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query bonus wallet: %w", err)
	}

	uid, err := shared.NewUserID(dbPlayerID)
	if err != nil {
		return nil, fmt.Errorf("invalid player_id in bonus wallet: %w", err)
	}

	return solo_marathon.ReconstructBonusWallet(uid, shield, fiftyFifty, skip, freeze), nil
}

// Save persists a bonus wallet (upsert)
func (r *BonusWalletRepository) Save(wallet *solo_marathon.BonusWallet) error {
	query := `
		INSERT INTO player_bonus_wallet (player_id, bonus_shield, bonus_fifty_fifty, bonus_skip, bonus_freeze, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		ON CONFLICT (player_id) DO UPDATE SET
			bonus_shield = EXCLUDED.bonus_shield,
			bonus_fifty_fifty = EXCLUDED.bonus_fifty_fifty,
			bonus_skip = EXCLUDED.bonus_skip,
			bonus_freeze = EXCLUDED.bonus_freeze,
			updated_at = NOW()
	`

	_, err := r.db.Exec(query,
		wallet.PlayerID().String(),
		wallet.Shield(),
		wallet.FiftyFifty(),
		wallet.Skip(),
		wallet.Freeze(),
	)

	if err != nil {
		return fmt.Errorf("failed to save bonus wallet: %w", err)
	}

	return nil
}
