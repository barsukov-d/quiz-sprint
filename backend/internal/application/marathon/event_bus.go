package marathon

import "github.com/barsukov/quiz-sprint/backend/internal/domain/solo_marathon"

// EventBus defines the interface for publishing domain events
// Implementation is in infrastructure layer
type EventBus interface {
	Publish(event solo_marathon.Event)
}
