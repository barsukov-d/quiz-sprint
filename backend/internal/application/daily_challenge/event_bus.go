package daily_challenge

import "github.com/barsukov/quiz-sprint/backend/internal/domain/daily_challenge"

// EventBus defines the interface for publishing domain events
// Implementation is in infrastructure layer
type EventBus interface {
	Publish(event daily_challenge.Event)
}
