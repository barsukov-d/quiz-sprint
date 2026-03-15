package postgres

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/barsukov/quiz-sprint/backend/internal/domain/shared"
	"github.com/barsukov/quiz-sprint/backend/internal/domain/user"
)

// InventoryRepository is a PostgreSQL implementation of user.InventoryRepository
type InventoryRepository struct {
	db *sql.DB
}

// NewInventoryRepository creates a new PostgreSQL inventory repository
func NewInventoryRepository(db *sql.DB) *InventoryRepository {
	return &InventoryRepository{db: db}
}

// FindByPlayerID retrieves the inventory for a player.
// If not found, creates a default inventory and returns it.
func (r *InventoryRepository) FindByPlayerID(playerID user.UserID) (*user.Inventory, error) {
	// Try INSERT with defaults first (ON CONFLICT DO NOTHING handles existing rows)
	insertQuery := `
		INSERT INTO user_inventory (player_id, coins, pvp_tickets, shield, fifty_fifty, "skip", "freeze", updated_at)
		VALUES ($1, 0, 3, 0, 0, 0, 0, $2)
		ON CONFLICT (player_id) DO NOTHING
	`

	now := time.Now().Unix()
	_, err := r.db.Exec(insertQuery, playerID.String(), now)
	if err != nil {
		return nil, fmt.Errorf("failed to ensure inventory exists: %w", err)
	}

	// Now SELECT the row (guaranteed to exist)
	selectQuery := `
		SELECT player_id, coins, pvp_tickets, shield, fifty_fifty, "skip", "freeze", updated_at
		FROM user_inventory
		WHERE player_id = $1
	`

	var (
		dbPlayerID string
		coins      int
		pvpTickets int
		shield     int
		fiftyFifty int
		skip       int
		freeze     int
		updatedAt  int64
	)

	err = r.db.QueryRow(selectQuery, playerID.String()).Scan(
		&dbPlayerID, &coins, &pvpTickets, &shield, &fiftyFifty, &skip, &freeze, &updatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query inventory: %w", err)
	}

	uid, err := shared.NewUserID(dbPlayerID)
	if err != nil {
		return nil, fmt.Errorf("invalid player_id in inventory: %w", err)
	}

	return user.ReconstructInventory(uid, coins, pvpTickets, shield, fiftyFifty, skip, freeze, updatedAt), nil
}

// Save persists an inventory (upsert)
func (r *InventoryRepository) Save(inventory *user.Inventory) error {
	query := `
		INSERT INTO user_inventory (player_id, coins, pvp_tickets, shield, fifty_fifty, "skip", "freeze", updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (player_id) DO UPDATE SET
			coins = EXCLUDED.coins,
			pvp_tickets = EXCLUDED.pvp_tickets,
			shield = EXCLUDED.shield,
			fifty_fifty = EXCLUDED.fifty_fifty,
			"skip" = EXCLUDED."skip",
			"freeze" = EXCLUDED."freeze",
			updated_at = EXCLUDED.updated_at
	`

	_, err := r.db.Exec(query,
		inventory.PlayerID().String(),
		inventory.Coins(),
		inventory.PvpTickets(),
		inventory.Shield(),
		inventory.FiftyFifty(),
		inventory.Skip(),
		inventory.Freeze(),
		inventory.UpdatedAt(),
	)

	if err != nil {
		return fmt.Errorf("failed to save inventory: %w", err)
	}

	return nil
}
