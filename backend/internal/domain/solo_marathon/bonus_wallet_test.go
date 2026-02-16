package solo_marathon

import (
	"testing"

	"github.com/barsukov/quiz-sprint/backend/internal/domain/shared"
)

func mustUserID(s string) shared.UserID {
	id, err := shared.NewUserID(s)
	if err != nil {
		panic(err)
	}
	return id
}

func TestNewBonusWallet(t *testing.T) {
	wallet := NewBonusWallet(mustUserID("player1"))

	if !wallet.IsEmpty() {
		t.Error("new wallet should be empty")
	}
	if wallet.Shield() != 0 || wallet.FiftyFifty() != 0 || wallet.Skip() != 0 || wallet.Freeze() != 0 {
		t.Error("new wallet should have all zeros")
	}
	if wallet.PlayerID().String() != "player1" {
		t.Errorf("PlayerID = %s, want player1", wallet.PlayerID().String())
	}
}

func TestBonusWallet_AddBonus(t *testing.T) {
	wallet := NewBonusWallet(mustUserID("player1"))

	wallet.AddBonus(BonusShield)
	wallet.AddBonus(BonusShield)
	wallet.AddBonus(BonusFiftyFifty)
	wallet.AddBonus(BonusSkip)
	wallet.AddBonus(BonusFreeze)
	wallet.AddBonus(BonusFreeze)
	wallet.AddBonus(BonusFreeze)

	if wallet.Shield() != 2 {
		t.Errorf("Shield = %d, want 2", wallet.Shield())
	}
	if wallet.FiftyFifty() != 1 {
		t.Errorf("FiftyFifty = %d, want 1", wallet.FiftyFifty())
	}
	if wallet.Skip() != 1 {
		t.Errorf("Skip = %d, want 1", wallet.Skip())
	}
	if wallet.Freeze() != 3 {
		t.Errorf("Freeze = %d, want 3", wallet.Freeze())
	}
	if wallet.IsEmpty() {
		t.Error("wallet should not be empty after adding bonuses")
	}
}

func TestBonusWallet_ToBonusInventory(t *testing.T) {
	wallet := ReconstructBonusWallet(mustUserID("player1"), 3, 2, 1, 4)

	inv := wallet.ToBonusInventory()

	if inv.Shield() != 3 {
		t.Errorf("Shield = %d, want 3", inv.Shield())
	}
	if inv.FiftyFifty() != 2 {
		t.Errorf("FiftyFifty = %d, want 2", inv.FiftyFifty())
	}
	if inv.Skip() != 1 {
		t.Errorf("Skip = %d, want 1", inv.Skip())
	}
	if inv.Freeze() != 4 {
		t.Errorf("Freeze = %d, want 4", inv.Freeze())
	}

	// Wallet should NOT be modified
	if wallet.Shield() != 3 {
		t.Error("ToBonusInventory should not modify wallet")
	}
}

func TestBonusWallet_ConsumeAll(t *testing.T) {
	wallet := ReconstructBonusWallet(mustUserID("player1"), 2, 1, 0, 3)

	inv := wallet.ConsumeAll()

	// Returned inventory should have the original values
	if inv.Shield() != 2 || inv.FiftyFifty() != 1 || inv.Skip() != 0 || inv.Freeze() != 3 {
		t.Errorf("ConsumeAll returned wrong inventory: shield=%d, fiftyFifty=%d, skip=%d, freeze=%d",
			inv.Shield(), inv.FiftyFifty(), inv.Skip(), inv.Freeze())
	}

	// Wallet should be zeroed
	if !wallet.IsEmpty() {
		t.Error("wallet should be empty after ConsumeAll")
	}
	if wallet.Shield() != 0 || wallet.FiftyFifty() != 0 || wallet.Skip() != 0 || wallet.Freeze() != 0 {
		t.Error("wallet should have all zeros after ConsumeAll")
	}
}

func TestBonusWallet_IsEmpty(t *testing.T) {
	tests := []struct {
		name     string
		wallet   *BonusWallet
		expected bool
	}{
		{"empty", NewBonusWallet(mustUserID("p")), true},
		{"has shield", ReconstructBonusWallet(mustUserID("p"), 1, 0, 0, 0), false},
		{"has fifty_fifty", ReconstructBonusWallet(mustUserID("p"), 0, 1, 0, 0), false},
		{"has skip", ReconstructBonusWallet(mustUserID("p"), 0, 0, 1, 0), false},
		{"has freeze", ReconstructBonusWallet(mustUserID("p"), 0, 0, 0, 1), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.wallet.IsEmpty(); got != tt.expected {
				t.Errorf("IsEmpty() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestBonusInventory_Add(t *testing.T) {
	defaults := NewBonusInventory() // shield=2, fiftyFifty=1, skip=0, freeze=3
	wallet := ReconstructBonusInventory(1, 2, 3, 1)

	combined := defaults.Add(wallet)

	if combined.Shield() != 3 {
		t.Errorf("Shield = %d, want 3 (2+1)", combined.Shield())
	}
	if combined.FiftyFifty() != 3 {
		t.Errorf("FiftyFifty = %d, want 3 (1+2)", combined.FiftyFifty())
	}
	if combined.Skip() != 3 {
		t.Errorf("Skip = %d, want 3 (0+3)", combined.Skip())
	}
	if combined.Freeze() != 4 {
		t.Errorf("Freeze = %d, want 4 (3+1)", combined.Freeze())
	}

	// Original inventories should not be modified (they're value types)
	if defaults.Shield() != 2 {
		t.Error("Add should not modify original")
	}
}

func TestBonusInventory_Add_ZeroValues(t *testing.T) {
	defaults := NewBonusInventory()
	empty := BonusInventory{} // zero value

	combined := defaults.Add(empty)

	if combined.Shield() != defaults.Shield() ||
		combined.FiftyFifty() != defaults.FiftyFifty() ||
		combined.Skip() != defaults.Skip() ||
		combined.Freeze() != defaults.Freeze() {
		t.Error("Adding zero inventory should return same values as original")
	}
}
