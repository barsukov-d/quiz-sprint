package kernel

import (
	"errors"

	"github.com/google/uuid"
)

// SessionID is a value object representing a unique gameplay session identifier
type SessionID struct {
	value uuid.UUID
}

// NewSessionID generates a new SessionID
func NewSessionID() SessionID {
	return SessionID{value: uuid.New()}
}

// NewSessionIDFromString creates a SessionID from a string
func NewSessionIDFromString(s string) (SessionID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return SessionID{}, ErrInvalidSessionID
	}
	return SessionID{value: id}, nil
}

// String returns the string representation
func (sid SessionID) String() string {
	return sid.value.String()
}

// UUID returns the underlying UUID
func (sid SessionID) UUID() uuid.UUID {
	return sid.value
}

// IsZero checks if the SessionID is zero value
func (sid SessionID) IsZero() bool {
	return sid.value == uuid.Nil
}

// Equals checks if two SessionIDs are equal
func (sid SessionID) Equals(other SessionID) bool {
	return sid.value == other.value
}

// Domain errors for kernel
var (
	ErrInvalidSessionID = errors.New("invalid session ID")
	ErrSessionFinished  = errors.New("session is already finished")
	ErrInvalidQuiz      = errors.New("invalid quiz: quiz cannot be nil")
)
