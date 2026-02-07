package quick_duel

import "github.com/barsukov/quiz-sprint/backend/internal/domain/quick_duel"

// EventBus defines the interface for publishing domain events
type EventBus interface {
	Publish(event quick_duel.Event)
}

// NoOpEventBus is a no-operation event bus for testing
type NoOpEventBus struct{}

func (n *NoOpEventBus) Publish(event quick_duel.Event) {
	// No-op
}

// NewNoOpEventBus creates a new no-operation event bus
func NewNoOpEventBus() *NoOpEventBus {
	return &NoOpEventBus{}
}
