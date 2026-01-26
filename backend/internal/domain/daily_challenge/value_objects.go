package daily_challenge

import (
	"github.com/barsukov/quiz-sprint/backend/internal/domain/quiz"
	"github.com/barsukov/quiz-sprint/backend/internal/domain/shared"
	"github.com/google/uuid"
	"time"
)

// Type aliases from other domains
type UserID = shared.UserID
type QuizID = quiz.QuizID
type QuestionID = quiz.QuestionID
type AnswerID = quiz.AnswerID
type Points = quiz.Points

// DailyQuizID uniquely identifies a daily quiz (one per day)
type DailyQuizID struct {
	value string
}

func NewDailyQuizID() DailyQuizID {
	return DailyQuizID{value: uuid.New().String()}
}

func NewDailyQuizIDFromString(value string) DailyQuizID {
	return DailyQuizID{value: value}
}

func (id DailyQuizID) String() string {
	return id.value
}

func (id DailyQuizID) IsZero() bool {
	return id.value == ""
}

func (id DailyQuizID) Equals(other DailyQuizID) bool {
	return id.value == other.value
}

// GameID uniquely identifies a daily game (player's attempt)
type GameID struct {
	value string
}

func NewGameID() GameID {
	return GameID{value: uuid.New().String()}
}

func NewGameIDFromString(value string) GameID {
	return GameID{value: value}
}

func (id GameID) String() string {
	return id.value
}

func (id GameID) IsZero() bool {
	return id.value == ""
}

func (id GameID) Equals(other GameID) bool {
	return id.value == other.value
}

// Date represents a calendar date (YYYY-MM-DD)
type Date struct {
	value string // Format: "2026-01-25"
}

func NewDate(year int, month time.Month, day int) Date {
	t := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
	return Date{value: t.Format("2006-01-02")}
}

func NewDateFromString(value string) Date {
	return Date{value: value}
}

func NewDateFromTime(t time.Time) Date {
	return Date{value: t.UTC().Format("2006-01-02")}
}

func TodayUTC() Date {
	return NewDateFromTime(time.Now().UTC())
}

func (d Date) String() string {
	return d.value
}

func (d Date) IsZero() bool {
	return d.value == ""
}

func (d Date) Equals(other Date) bool {
	return d.value == other.value
}

// Next returns the next day
func (d Date) Next() Date {
	t, _ := time.Parse("2006-01-02", d.value)
	next := t.AddDate(0, 0, 1)
	return NewDateFromTime(next)
}

// Previous returns the previous day
func (d Date) Previous() Date {
	t, _ := time.Parse("2006-01-02", d.value)
	prev := t.AddDate(0, 0, -1)
	return NewDateFromTime(prev)
}

// StreakSystem manages daily streak tracking
type StreakSystem struct {
	currentStreak  int    // Days in a row played
	bestStreak     int    // All-time best streak
	lastPlayedDate Date   // Last date player completed daily challenge
}

func NewStreakSystem() StreakSystem {
	return StreakSystem{
		currentStreak:  0,
		bestStreak:     0,
		lastPlayedDate: Date{},
	}
}

// ReconstructStreakSystem reconstructs from persistence
func ReconstructStreakSystem(currentStreak int, bestStreak int, lastPlayedDate Date) StreakSystem {
	return StreakSystem{
		currentStreak:  currentStreak,
		bestStreak:     bestStreak,
		lastPlayedDate: lastPlayedDate,
	}
}

// UpdateForDate updates streak based on the date played
// Returns updated StreakSystem (immutable)
func (ss StreakSystem) UpdateForDate(playedDate Date) StreakSystem {
	// First time playing
	if ss.lastPlayedDate.IsZero() {
		return StreakSystem{
			currentStreak:  1,
			bestStreak:     1,
			lastPlayedDate: playedDate,
		}
	}

	// Same day (shouldn't happen with one attempt per day, but handle it)
	if ss.lastPlayedDate.Equals(playedDate) {
		return ss
	}

	// Consecutive day
	expectedDate := ss.lastPlayedDate.Next()
	if expectedDate.Equals(playedDate) {
		newCurrent := ss.currentStreak + 1
		newBest := ss.bestStreak
		if newCurrent > newBest {
			newBest = newCurrent
		}

		return StreakSystem{
			currentStreak:  newCurrent,
			bestStreak:     newBest,
			lastPlayedDate: playedDate,
		}
	}

	// Streak broken (missed one or more days)
	return StreakSystem{
		currentStreak:  1, // Start new streak
		bestStreak:     ss.bestStreak,
		lastPlayedDate: playedDate,
	}
}

// GetBonus returns score multiplier based on current streak
func (ss StreakSystem) GetBonus() float64 {
	switch {
	case ss.currentStreak >= 100:
		return 2.0 // +100%
	case ss.currentStreak >= 30:
		return 1.6 // +60%
	case ss.currentStreak >= 14:
		return 1.4 // +40%
	case ss.currentStreak >= 7:
		return 1.25 // +25%
	case ss.currentStreak >= 3:
		return 1.1 // +10%
	default:
		return 1.0 // No bonus
	}
}

// IsActive checks if streak is still active for today
func (ss StreakSystem) IsActive(today Date) bool {
	if ss.lastPlayedDate.IsZero() {
		return false
	}

	// Streak is active if last played yesterday or today
	return ss.lastPlayedDate.Equals(today) || ss.lastPlayedDate.Equals(today.Previous())
}

// Getters
func (ss StreakSystem) CurrentStreak() int { return ss.currentStreak }
func (ss StreakSystem) BestStreak() int    { return ss.bestStreak }
func (ss StreakSystem) LastPlayedDate() Date { return ss.lastPlayedDate }

// GameStatus represents the current status of a daily game
type GameStatus string

const (
	GameStatusInProgress GameStatus = "in_progress"
	GameStatusCompleted  GameStatus = "completed"
)

// State transition diagram for Daily Game:
//
//   in_progress ──(complete)──> completed (terminal)
//
// Allowed transitions map
var dailyGameTransitions = map[GameStatus][]GameStatus{
	GameStatusInProgress: {GameStatusCompleted},
	GameStatusCompleted:  {}, // Terminal state - no transitions allowed
}

// CanTransitionTo checks if transition to target status is valid
func (gs GameStatus) CanTransitionTo(target GameStatus) bool {
	allowedTargets, exists := dailyGameTransitions[gs]
	if !exists {
		return false
	}

	for _, allowedTarget := range allowedTargets {
		if allowedTarget == target {
			return true
		}
	}

	return false
}

// IsTerminal checks if status is a terminal state (no further transitions)
func (gs GameStatus) IsTerminal() bool {
	transitions, exists := dailyGameTransitions[gs]
	return exists && len(transitions) == 0
}
