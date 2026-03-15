package marathon

// InventoryService defines the interface for crediting/debiting player resources.
// Implementation is in application/user layer.
type InventoryService interface {
	Credit(playerID string, source string, details map[string]int) error
	Debit(playerID string, source string, details map[string]int) error
}

// MilestoneClaimsRepository tracks which milestone rewards each player has already claimed,
// preventing double-crediting when CompleteMarathon is called.
type MilestoneClaimsRepository interface {
	HasClaimed(playerID string, milestone int) (bool, error)
	MarkClaimed(playerID string, milestone int) error
}
