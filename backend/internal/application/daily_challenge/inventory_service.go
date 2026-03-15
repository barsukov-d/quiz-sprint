package daily_challenge

// InventoryService defines the interface for crediting/debiting player resources
// Implementation is in application/user layer
type InventoryService interface {
	GetCoins(playerID string) (int, error)
	Credit(playerID string, source string, details map[string]int) error
	Debit(playerID string, source string, details map[string]int) error
}
