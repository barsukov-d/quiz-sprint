package classic_mode

import (
	"errors"

	"github.com/google/uuid"
)

// GameID is a value object representing a unique game identifier
type GameID struct {
	value uuid.UUID
}

// NewGameID generates a new GameID
func NewGameID() GameID {
	return GameID{value: uuid.New()}
}

// NewGameIDFromString creates a GameID from a string
func NewGameIDFromString(s string) (GameID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return GameID{}, ErrInvalidGameID
	}
	return GameID{value: id}, nil
}

// String returns the string representation
func (gid GameID) String() string {
	return gid.value.String()
}

// UUID returns the underlying UUID
func (gid GameID) UUID() uuid.UUID {
	return gid.value
}

// IsZero checks if the GameID is zero value
func (gid GameID) IsZero() bool {
	return gid.value == uuid.Nil
}

// Equals checks if two GameIDs are equal
func (gid GameID) Equals(other GameID) bool {
	return gid.value == other.value
}

// GameStatus represents the lifecycle of a classic game
type GameStatus int

const (
	GameStatusInProgress GameStatus = iota
	GameStatusFinished
)

func (s GameStatus) String() string {
	switch s {
	case GameStatusInProgress:
		return "InProgress"
	case GameStatusFinished:
		return "Finished"
	default:
		return "Unknown"
	}
}

// VisualState represents the visual intensity state for game juice feedback
type VisualState int

const (
	VisualStateNormal VisualState = iota // Streak < 3
	VisualStateHeat                      // Streak >= 3 and < 6 ("On Fire")
	VisualStateFire                      // Streak >= 6 ("Godlike")
)

func (vs VisualState) String() string {
	switch vs {
	case VisualStateNormal:
		return "Normal"
	case VisualStateHeat:
		return "Heat"
	case VisualStateFire:
		return "Fire"
	default:
		return "Unknown"
	}
}

// VisualStateFromStreak determines the visual state based on streak count
func VisualStateFromStreak(streak int) VisualState {
	if streak >= 6 {
		return VisualStateFire
	}
	if streak >= 3 {
		return VisualStateHeat
	}
	return VisualStateNormal
}

// Multiplier represents the score multiplier based on streak
type Multiplier float64

const (
	MultiplierNormal  Multiplier = 1.0 // Streak 0-2
	MultiplierOnFire  Multiplier = 1.5 // Streak 3-5
	MultiplierGodlike Multiplier = 2.0 // Streak 6+
)

func (m Multiplier) Float64() float64 {
	return float64(m)
}

func (m Multiplier) String() string {
	switch m {
	case MultiplierNormal:
		return "x1.0"
	case MultiplierOnFire:
		return "x1.5"
	case MultiplierGodlike:
		return "x2.0"
	default:
		return "x?.?"
	}
}

// MultiplierFromStreak calculates the multiplier based on streak count
// Business Rule: x1.0 (0-2), x1.5 (3-5), x2.0 (6+)
func MultiplierFromStreak(streak int) Multiplier {
	if streak >= 6 {
		return MultiplierGodlike
	}
	if streak >= 3 {
		return MultiplierOnFire
	}
	return MultiplierNormal
}

// IsStreakMilestone checks if the streak count is a milestone (3 or 6)
func IsStreakMilestone(streak int) bool {
	return streak == 3 || streak == 6
}

// Domain errors for classic_mode
var (
	ErrInvalidGameID       = errors.New("invalid game ID")
	ErrGameAlreadyFinished = errors.New("game is already finished")
	ErrGameNotStarted      = errors.New("game has not started")
	ErrNotEnoughQuestions  = errors.New("not enough questions to start game")
)
