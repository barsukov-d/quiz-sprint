package solo_marathon

// BonusWallet accumulates marathon bonuses earned from other game modes (e.g., Daily Challenge chests).
// Bonuses are consumed when starting a new marathon game and merged with defaults.
type BonusWallet struct {
	playerID   UserID
	shield     int
	fiftyFifty int
	skip       int
	freeze     int
}

// NewBonusWallet creates an empty bonus wallet for a player
func NewBonusWallet(playerID UserID) *BonusWallet {
	return &BonusWallet{
		playerID: playerID,
	}
}

// ReconstructBonusWallet reconstructs a BonusWallet from persistence
func ReconstructBonusWallet(playerID UserID, shield, fiftyFifty, skip, freeze int) *BonusWallet {
	return &BonusWallet{
		playerID:   playerID,
		shield:     shield,
		fiftyFifty: fiftyFifty,
		skip:       skip,
		freeze:     freeze,
	}
}

// AddBonus increments the count of a specific bonus type
func (w *BonusWallet) AddBonus(bonusType BonusType) {
	switch bonusType {
	case BonusShield:
		w.shield++
	case BonusFiftyFifty:
		w.fiftyFifty++
	case BonusSkip:
		w.skip++
	case BonusFreeze:
		w.freeze++
	}
}

// ToBonusInventory converts current wallet contents to a BonusInventory (non-destructive)
func (w *BonusWallet) ToBonusInventory() BonusInventory {
	return ReconstructBonusInventory(w.shield, w.fiftyFifty, w.skip, w.freeze)
}

// ConsumeAll returns all accumulated bonuses as a BonusInventory and zeroes the wallet
func (w *BonusWallet) ConsumeAll() BonusInventory {
	inv := w.ToBonusInventory()
	w.shield = 0
	w.fiftyFifty = 0
	w.skip = 0
	w.freeze = 0
	return inv
}

// IsEmpty returns true if all bonus counts are zero
func (w *BonusWallet) IsEmpty() bool {
	return w.shield == 0 && w.fiftyFifty == 0 && w.skip == 0 && w.freeze == 0
}

// Getters
func (w *BonusWallet) PlayerID() UserID { return w.playerID }
func (w *BonusWallet) Shield() int      { return w.shield }
func (w *BonusWallet) FiftyFifty() int  { return w.fiftyFifty }
func (w *BonusWallet) Skip() int        { return w.skip }
func (w *BonusWallet) Freeze() int      { return w.freeze }
