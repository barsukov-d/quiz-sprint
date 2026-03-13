package quick_duel_test

import (
	"testing"

	quick_duel "github.com/barsukov/quiz-sprint/backend/internal/domain/quick_duel"
)

func TestChallengeCreatedEvent_WithChallengerUsername(t *testing.T) {
	challengerID, _ := quick_duel.NewUserID("111")
	friendID, _ := quick_duel.NewUserID("222")
	challengeID := quick_duel.NewChallengeID()

	evt := quick_duel.NewChallengeCreatedEvent(
		challengeID,
		challengerID,
		&friendID,
		quick_duel.DirectChallenge,
		0,
		0,
	)

	enriched := evt.WithChallengerUsername("Pavel")

	if enriched.ChallengerUsername() != "Pavel" {
		t.Errorf("expected 'Pavel', got %q", enriched.ChallengerUsername())
	}
	// Original must not be modified (value receiver copy semantics)
	if evt.ChallengerUsername() != "" {
		t.Errorf("original event should not be modified, got %q", evt.ChallengerUsername())
	}
}
