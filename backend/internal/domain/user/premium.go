package user

// PremiumStatus is a value object representing a user's premium subscription state.
type PremiumStatus struct {
	active    bool
	expiresAt int64 // Unix timestamp, 0 if never had premium
}

// NewPremiumStatus creates a PremiumStatus with the given values.
func NewPremiumStatus(active bool, expiresAt int64) PremiumStatus {
	return PremiumStatus{active: active, expiresAt: expiresAt}
}

// NoPremium returns the zero PremiumStatus (free user, never had premium).
func NoPremium() PremiumStatus {
	return PremiumStatus{}
}

// IsActive reports whether the subscription is currently active.
func (ps PremiumStatus) IsActive() bool {
	return ps.active
}

// ExpiresAt returns the Unix timestamp when the subscription expires.
// Returns 0 if the user has never had premium.
func (ps PremiumStatus) ExpiresAt() int64 {
	return ps.expiresAt
}
