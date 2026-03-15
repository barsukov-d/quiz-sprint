package quick_duel

// InventoryService defines the interface for crediting/debiting player resources.
// Implementation is in application/user layer.
type InventoryService interface {
	GetPvpTickets(playerID string) (int, error)
	Credit(playerID string, source string, details map[string]int) error
	Debit(playerID string, source string, details map[string]int) error
}
