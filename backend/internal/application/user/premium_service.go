package user

// PremiumService checks whether a player has an active premium subscription.
type PremiumService interface {
	IsPremium(playerID string) (bool, error)
}

// NoopPremiumService always reports the player as a free user.
// Used until a real billing integration is wired in.
type NoopPremiumService struct{}

func (s *NoopPremiumService) IsPremium(_ string) (bool, error) {
	return false, nil
}
