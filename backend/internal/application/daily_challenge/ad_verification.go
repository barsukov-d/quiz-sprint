package daily_challenge

// AdVerificationService defines the interface for verifying ad views before granting free retries.
type AdVerificationService interface {
	VerifyAdWatched(playerID string, adType string) (bool, error)
}

// NoopAdVerificationService is a stub until real ad network integration is wired.
// It always reports the ad as watched.
type NoopAdVerificationService struct{}

func (s *NoopAdVerificationService) VerifyAdWatched(playerID string, adType string) (bool, error) {
	return true, nil
}
