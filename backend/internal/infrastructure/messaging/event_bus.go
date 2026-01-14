package messaging

import (
	"log"
	"sync"

	"github.com/barsukov/quiz-sprint/backend/internal/domain/quiz"
)

// InMemoryEventBus is an in-memory implementation of quiz.EventBus
type InMemoryEventBus struct {
	handlers map[string][]quiz.EventHandler
	mu       sync.RWMutex
}

// NewInMemoryEventBus creates a new in-memory event bus
func NewInMemoryEventBus() *InMemoryEventBus {
	return &InMemoryEventBus{
		handlers: make(map[string][]quiz.EventHandler),
	}
}

// Publish publishes events asynchronously
func (eb *InMemoryEventBus) Publish(events ...quiz.Event) {
	for _, event := range events {
		go eb.dispatch(event)
	}
}

func (eb *InMemoryEventBus) dispatch(event quiz.Event) {
	eb.mu.RLock()
	handlers, exists := eb.handlers[event.EventName()]
	eb.mu.RUnlock()

	if !exists {
		return
	}

	for _, handler := range handlers {
		// Run each handler in a separate goroutine for true async
		go func(h quiz.EventHandler) {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("Event handler panicked: %v", r)
				}
			}()
			h(event)
		}(handler)
	}
}

// Subscribe registers a handler for a specific event type
func (eb *InMemoryEventBus) Subscribe(eventName string, handler quiz.EventHandler) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	eb.handlers[eventName] = append(eb.handlers[eventName], handler)
}

// LoggingEventBus wraps an EventBus and logs all events
type LoggingEventBus struct {
	inner quiz.EventBus
}

// NewLoggingEventBus creates a new logging event bus
func NewLoggingEventBus(inner quiz.EventBus) *LoggingEventBus {
	return &LoggingEventBus{inner: inner}
}

// Publish publishes events with logging
func (eb *LoggingEventBus) Publish(events ...quiz.Event) {
	for _, event := range events {
		log.Printf("[EVENT] %s at %d", event.EventName(), event.OccurredAt())
	}
	eb.inner.Publish(events...)
}

// Subscribe registers a handler with logging
func (eb *LoggingEventBus) Subscribe(eventName string, handler quiz.EventHandler) {
	log.Printf("[EVENT] Subscribed to: %s", eventName)
	eb.inner.Subscribe(eventName, handler)
}
