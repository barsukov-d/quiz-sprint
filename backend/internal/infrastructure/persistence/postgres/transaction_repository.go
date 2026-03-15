package postgres

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/barsukov/quiz-sprint/backend/internal/domain/shared"
	"github.com/barsukov/quiz-sprint/backend/internal/domain/user"
)

// TransactionRepository is a PostgreSQL implementation of user.TransactionRepository
type TransactionRepository struct {
	db *sql.DB
}

// NewTransactionRepository creates a new PostgreSQL transaction repository
func NewTransactionRepository(db *sql.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

// Save persists a transaction log
func (r *TransactionRepository) Save(tx *user.TransactionLog) error {
	detailsJSON, err := json.Marshal(tx.Details())
	if err != nil {
		return fmt.Errorf("failed to marshal transaction details: %w", err)
	}

	query := `
		INSERT INTO user_transactions (id, player_id, tx_type, source, details, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err = r.db.Exec(query,
		tx.ID(),
		tx.PlayerID().String(),
		string(tx.Type()),
		detailsJSON,
		tx.CreatedAt(),
	)

	if err != nil {
		return fmt.Errorf("failed to save transaction: %w", err)
	}

	return nil
}

// FindByPlayer retrieves transactions for a player ordered by created_at DESC
func (r *TransactionRepository) FindByPlayer(playerID user.UserID, limit int) ([]user.TransactionLog, error) {
	query := `
		SELECT id, player_id, tx_type, source, details, created_at
		FROM user_transactions
		WHERE player_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := r.db.Query(query, playerID.String(), limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query transactions: %w", err)
	}
	defer rows.Close()

	var transactions []user.TransactionLog

	for rows.Next() {
		var (
			id          string
			dbPlayerID  string
			txType      string
			source      string
			detailsJSON []byte
			createdAt   int64
		)

		err := rows.Scan(&id, &dbPlayerID, &txType, &source, &detailsJSON, &createdAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan transaction row: %w", err)
		}

		var details map[string]int
		if err := json.Unmarshal(detailsJSON, &details); err != nil {
			return nil, fmt.Errorf("failed to unmarshal transaction details: %w", err)
		}

		uid, err := shared.NewUserID(dbPlayerID)
		if err != nil {
			return nil, fmt.Errorf("invalid player_id in transaction: %w", err)
		}

		tx := user.ReconstructTransactionLog(id, uid, user.TransactionType(txType), source, details, createdAt)
		transactions = append(transactions, *tx)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating transaction rows: %w", err)
	}

	return transactions, nil
}
