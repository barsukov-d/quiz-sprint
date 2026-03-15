package daily_challenge

// PremiumService checks whether a player has an active premium subscription.
// This is a port (interface) — the implementation lives in the user application layer.
type PremiumService interface {
	IsPremium(playerID string) (bool, error)
}
