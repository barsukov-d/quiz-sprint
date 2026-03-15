package user

import (
	"errors"
	"testing"

	"github.com/barsukov/quiz-sprint/backend/internal/domain/shared"
)

func mustUserID(s string) UserID {
	id, err := shared.NewUserID(s)
	if err != nil {
		panic(err)
	}
	return id
}

const ts = int64(1700000000)

// ---------------------------------------------------------------------------
// Inventory
// ---------------------------------------------------------------------------

func TestNewInventory_WelcomeBonus(t *testing.T) {
	inv, err := NewInventory(mustUserID("player1"), ts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if inv.PvpTickets() != 3 {
		t.Errorf("PvpTickets = %d, want 3", inv.PvpTickets())
	}
	if inv.Coins() != 0 {
		t.Errorf("Coins = %d, want 0", inv.Coins())
	}
	if inv.Shield() != 0 {
		t.Errorf("Shield = %d, want 0", inv.Shield())
	}
	if inv.FiftyFifty() != 0 {
		t.Errorf("FiftyFifty = %d, want 0", inv.FiftyFifty())
	}
	if inv.Skip() != 0 {
		t.Errorf("Skip = %d, want 0", inv.Skip())
	}
	if inv.Freeze() != 0 {
		t.Errorf("Freeze = %d, want 0", inv.Freeze())
	}
}

func TestInventory_Credit_Success(t *testing.T) {
	inv, _ := NewInventory(mustUserID("player1"), ts)

	err := inv.Credit(ResourceCoins, 100, ts+1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inv.Coins() != 100 {
		t.Errorf("Coins = %d, want 100", inv.Coins())
	}
	if inv.UpdatedAt() != ts+1 {
		t.Errorf("UpdatedAt = %d, want %d", inv.UpdatedAt(), ts+1)
	}
}

func TestInventory_Credit_InvalidResource(t *testing.T) {
	inv, _ := NewInventory(mustUserID("player1"), ts)

	err := inv.Credit("unknown_resource", 10, ts+1)
	if !errors.Is(err, ErrInvalidResource) {
		t.Errorf("expected ErrInvalidResource, got %v", err)
	}
}

func TestInventory_Credit_InvalidAmount(t *testing.T) {
	tests := []struct {
		name   string
		amount int
	}{
		{"zero", 0},
		{"negative", -5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inv, _ := NewInventory(mustUserID("player1"), ts)
			err := inv.Credit(ResourceCoins, tt.amount, ts+1)
			if !errors.Is(err, ErrInvalidAmount) {
				t.Errorf("expected ErrInvalidAmount, got %v", err)
			}
		})
	}
}

func TestInventory_Debit_Success(t *testing.T) {
	inv := ReconstructInventory(mustUserID("player1"), 200, 3, 0, 0, 0, 0, ts)

	err := inv.Debit(ResourceCoins, 50, ts+1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inv.Coins() != 150 {
		t.Errorf("Coins = %d, want 150", inv.Coins())
	}
	if inv.UpdatedAt() != ts+1 {
		t.Errorf("UpdatedAt = %d, want %d", inv.UpdatedAt(), ts+1)
	}
}

func TestInventory_Debit_InsufficientBalance(t *testing.T) {
	inv := ReconstructInventory(mustUserID("player1"), 30, 3, 0, 0, 0, 0, ts)

	err := inv.Debit(ResourceCoins, 50, ts+1)
	if !errors.Is(err, ErrInsufficientBalance) {
		t.Errorf("expected ErrInsufficientBalance, got %v", err)
	}
	// balance must be unchanged
	if inv.Coins() != 30 {
		t.Errorf("Coins = %d, want 30 (must not change on error)", inv.Coins())
	}
}

func TestInventory_CreditMultiple(t *testing.T) {
	inv, _ := NewInventory(mustUserID("player1"), ts)

	credits := map[string]int{
		ResourceCoins:  500,
		ResourceShield: 2,
		ResourceSkip:   1,
	}
	err := inv.CreditMultiple(credits, ts+1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inv.Coins() != 500 {
		t.Errorf("Coins = %d, want 500", inv.Coins())
	}
	if inv.Shield() != 2 {
		t.Errorf("Shield = %d, want 2", inv.Shield())
	}
	if inv.Skip() != 1 {
		t.Errorf("Skip = %d, want 1", inv.Skip())
	}
}

func TestInventory_DebitMultiple_Success(t *testing.T) {
	inv := ReconstructInventory(mustUserID("player1"), 1000, 3, 2, 1, 2, 3, ts)

	debits := map[string]int{
		ResourceCoins:      200,
		ResourcePvpTickets: 1,
		ResourceShield:     1,
	}
	err := inv.DebitMultiple(debits, ts+1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inv.Coins() != 800 {
		t.Errorf("Coins = %d, want 800", inv.Coins())
	}
	if inv.PvpTickets() != 2 {
		t.Errorf("PvpTickets = %d, want 2", inv.PvpTickets())
	}
	if inv.Shield() != 1 {
		t.Errorf("Shield = %d, want 1", inv.Shield())
	}
}

func TestInventory_DebitMultiple_InsufficientBalance(t *testing.T) {
	// coins=100, pvpTickets=1 — debit will exceed pvpTickets
	inv := ReconstructInventory(mustUserID("player1"), 100, 1, 0, 0, 0, 0, ts)

	debits := map[string]int{
		ResourceCoins:      50,  // would succeed individually
		ResourcePvpTickets: 5,   // exceeds balance → triggers error
	}
	err := inv.DebitMultiple(debits, ts+1)
	if !errors.Is(err, ErrInsufficientBalance) {
		t.Errorf("expected ErrInsufficientBalance, got %v", err)
	}
	// atomic: no resource must have been deducted
	if inv.Coins() != 100 {
		t.Errorf("Coins = %d, want 100 (must not change on error)", inv.Coins())
	}
	if inv.PvpTickets() != 1 {
		t.Errorf("PvpTickets = %d, want 1 (must not change on error)", inv.PvpTickets())
	}
}

// ---------------------------------------------------------------------------
// TransactionLog
// ---------------------------------------------------------------------------

func TestNewTransactionLog_Success(t *testing.T) {
	playerID := mustUserID("player1")
	details := map[string]int{ResourceCoins: 100}

	tx, err := NewTransactionLog("tx-001", playerID, TransactionCredit, "quiz_reward", details, ts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tx.ID() != "tx-001" {
		t.Errorf("ID = %s, want tx-001", tx.ID())
	}
	if tx.PlayerID() != playerID {
		t.Errorf("PlayerID mismatch")
	}
	if tx.Type() != TransactionCredit {
		t.Errorf("Type = %s, want %s", tx.Type(), TransactionCredit)
	}
	if tx.Source() != "quiz_reward" {
		t.Errorf("Source = %s, want quiz_reward", tx.Source())
	}
	if tx.CreatedAt() != ts {
		t.Errorf("CreatedAt = %d, want %d", tx.CreatedAt(), ts)
	}
	got := tx.Details()
	if got[ResourceCoins] != 100 {
		t.Errorf("Details[coins] = %d, want 100", got[ResourceCoins])
	}
}

func TestNewTransactionLog_InvalidType(t *testing.T) {
	playerID := mustUserID("player1")
	details := map[string]int{ResourceCoins: 50}

	tests := []struct {
		name    string
		txType  TransactionType
		wantErr error
	}{
		{"empty type", "", ErrInvalidTransactionType},
		{"unknown type", "refund", ErrInvalidTransactionType},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewTransactionLog("tx-001", playerID, tt.txType, "source", details, ts)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("expected %v, got %v", tt.wantErr, err)
			}
		})
	}
}
