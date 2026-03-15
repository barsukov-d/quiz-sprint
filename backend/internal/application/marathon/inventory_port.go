package marathon

// InventoryService defines the interface for crediting/debiting player resources.
// Implementation is in application/user layer.
type InventoryService interface {
	Credit(playerID string, source string, details map[string]int) error
	Debit(playerID string, source string, details map[string]int) error
}
