package marathon

import (
	"testing"
	"time"

	"github.com/barsukov/quiz-sprint/backend/internal/domain/solo_marathon"
)

func TestToLivesDTO_TimeToNextLifeIsAlwaysZero(t *testing.T) {
	now := time.Now().Unix()

	// Player lost 2 lives — normally would show a timer
	lives := solo_marathon.ReconstructLivesSystem(3, now-7200)

	dto := ToLivesDTO(lives, now)

	if dto.TimeToNextLife != 0 {
		t.Errorf("Expected TimeToNextLife=0 (no time-gate between runs), got %d", dto.TimeToNextLife)
	}
}
