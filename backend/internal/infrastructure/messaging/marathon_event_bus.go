package messaging

import (
	"log"

	"github.com/barsukov/quiz-sprint/backend/internal/domain/solo_marathon"
)

// MarathonEventBus is an in-memory implementation of marathon.EventBus
type MarathonEventBus struct {
	enableLogging bool
}

// NewMarathonEventBus creates a new marathon event bus
func NewMarathonEventBus(enableLogging bool) *MarathonEventBus {
	return &MarathonEventBus{
		enableLogging: enableLogging,
	}
}

// Publish publishes a single marathon event
func (eb *MarathonEventBus) Publish(event solo_marathon.Event) {
	go eb.dispatch(event)
}

func (eb *MarathonEventBus) dispatch(event solo_marathon.Event) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[MARATHON EVENT BUS] Panic during event dispatch: %v", r)
		}
	}()

	if eb.enableLogging {
		eb.logEvent(event)
	}
}

func (eb *MarathonEventBus) logEvent(event solo_marathon.Event) {
	switch e := event.(type) {
	case solo_marathon.MarathonGameStartedEvent:
		log.Printf("[MARATHON EVENT] Game Started: gameId=%s, playerId=%s, category=%s",
			e.GameID().String(), e.PlayerID().String(), e.Category().Name())

	case solo_marathon.MarathonQuestionAnsweredEvent:
		correctStr := "incorrect"
		if e.IsCorrect() {
			correctStr = "correct"
		}
		shieldStr := ""
		if e.ShieldConsumed() {
			shieldStr = " (shield consumed)"
		}
		log.Printf("[MARATHON EVENT] Question Answered: gameId=%s, playerId=%s, %s%s, score=%d, lives=%d",
			e.GameID().String(), e.PlayerID().String(), correctStr, shieldStr, e.CurrentScore(), e.LivesRemaining())

	case solo_marathon.BonusUsedEvent:
		log.Printf("[MARATHON EVENT] Bonus Used: gameId=%s, playerId=%s, bonusType=%s, remaining=%d",
			e.GameID().String(), e.PlayerID().String(), string(e.BonusType()), e.RemainingCount())

	case solo_marathon.LifeLostEvent:
		log.Printf("[MARATHON EVENT] Life Lost: gameId=%s, playerId=%s, remainingLives=%d",
			e.GameID().String(), e.PlayerID().String(), e.RemainingLives())

	case solo_marathon.MarathonGameOverEvent:
		log.Printf("[MARATHON EVENT] Game Over: gameId=%s, playerId=%s, finalScore=%d, totalQuestions=%d, isNewRecord=%t, continueCount=%d",
			e.GameID().String(), e.PlayerID().String(), e.FinalScore(), e.TotalQuestions(), e.IsNewRecord(), e.ContinueCount())

	case solo_marathon.ContinueUsedEvent:
		log.Printf("[MARATHON EVENT] Continue Used: gameId=%s, playerId=%s, continueCount=%d, payment=%s, cost=%d",
			e.GameID().String(), e.PlayerID().String(), e.ContinueCount(), string(e.PaymentMethod()), e.CostCoins())

	case solo_marathon.DifficultyIncreasedEvent:
		log.Printf("[MARATHON EVENT] Difficulty Increased: gameId=%s, playerId=%s, newLevel=%s, questionIndex=%d",
			e.GameID().String(), e.PlayerID().String(), string(e.NewLevel()), e.QuestionIndex())

	default:
		log.Printf("[MARATHON EVENT] Unknown event type: %s at %d", event.EventType(), event.OccurredAt())
	}
}

// NoOpMarathonEventBus is a no-op implementation for testing
type NoOpMarathonEventBus struct{}

// NewNoOpMarathonEventBus creates a no-op event bus
func NewNoOpMarathonEventBus() *NoOpMarathonEventBus {
	return &NoOpMarathonEventBus{}
}

// Publish does nothing
func (eb *NoOpMarathonEventBus) Publish(event solo_marathon.Event) {
	// No-op
}
