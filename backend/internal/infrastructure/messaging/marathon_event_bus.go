package messaging

import (
	"log"

	"github.com/barsukov/quiz-sprint/backend/internal/domain/solo_marathon"
)

// MarathonEventBus is an in-memory implementation of marathon.EventBus
// Events are published asynchronously for fire-and-forget semantics
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
// Events are logged if logging is enabled
// In production, this could be extended to:
// - Send to Redis Pub/Sub
// - Send to Kafka
// - Send to RabbitMQ
// - Store in event store for event sourcing
func (eb *MarathonEventBus) Publish(event solo_marathon.Event) {
	// Async dispatch
	go eb.dispatch(event)
}

func (eb *MarathonEventBus) dispatch(event solo_marathon.Event) {
	// Panic recovery to prevent event publishing from crashing the app
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[MARATHON EVENT BUS] Panic during event dispatch: %v", r)
		}
	}()

	if eb.enableLogging {
		eb.logEvent(event)
	}

	// TODO: Add event handlers here
	// For now, we just log. In the future:
	// - Update leaderboards in real-time (via WebSocket)
	// - Send notifications
	// - Update analytics/metrics
	// - Trigger achievement unlocks
}

func (eb *MarathonEventBus) logEvent(event solo_marathon.Event) {
	switch e := event.(type) {
	case *solo_marathon.MarathonGameStartedEvent:
		log.Printf("[MARATHON EVENT] Game Started: gameId=%s, playerId=%s, category=%s",
			e.GameID().String(), e.PlayerID().String(), e.Category().Name())

	case *solo_marathon.MarathonQuestionAnsweredEvent:
		correctStr := "incorrect"
		if e.IsCorrect() {
			correctStr = "correct"
		}
		log.Printf("[MARATHON EVENT] Question Answered: gameId=%s, playerId=%s, %s, streak=%d",
			e.GameID().String(), e.PlayerID().String(), correctStr, e.CurrentStreak())

	case *solo_marathon.HintUsedEvent:
		log.Printf("[MARATHON EVENT] Hint Used: gameId=%s, playerId=%s, hintType=%s, remaining=%d",
			e.GameID().String(), e.PlayerID().String(), string(e.HintType()), e.RemainingHints())

	case *solo_marathon.LifeLostEvent:
		log.Printf("[MARATHON EVENT] Life Lost: gameId=%s, playerId=%s, remainingLives=%d",
			e.GameID().String(), e.PlayerID().String(), e.RemainingLives())

	case *solo_marathon.MarathonGameOverEvent:
		log.Printf("[MARATHON EVENT] Game Over: gameId=%s, playerId=%s, streak=%d, isNewRecord=%t",
			e.GameID().String(), e.PlayerID().String(), e.FinalStreak(), e.IsNewRecord())

	case *solo_marathon.DifficultyIncreasedEvent:
		log.Printf("[MARATHON EVENT] Difficulty Increased: gameId=%s, playerId=%s, newLevel=%s, streak=%d",
			e.GameID().String(), e.PlayerID().String(), string(e.NewLevel()), e.StreakReached())

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
