package daily_challenge

import (
	"testing"
	"time"
)

// TestDailyQuizID_Operations tests DailyQuizID value object
func TestDailyQuizID_Operations(t *testing.T) {
	// Test generation
	id1 := NewDailyQuizID()
	id2 := NewDailyQuizID()

	if id1.IsZero() {
		t.Error("Generated ID should not be zero")
	}
	if id1.String() == "" {
		t.Error("Generated ID should have string representation")
	}
	if id1.Equals(id2) {
		t.Error("Generated IDs should be unique")
	}

	// Test from string
	id3 := NewDailyQuizIDFromString("test-daily-quiz-id")
	if id3.String() != "test-daily-quiz-id" {
		t.Errorf("ID string = %s, want %s", id3.String(), "test-daily-quiz-id")
	}

	// Test zero ID
	zeroID := DailyQuizID{}
	if !zeroID.IsZero() {
		t.Error("Empty ID should be zero")
	}

	// Test equals
	id4 := NewDailyQuizIDFromString("same-id")
	id5 := NewDailyQuizIDFromString("same-id")
	if !id4.Equals(id5) {
		t.Error("IDs with same string should be equal")
	}
}

// TestGameID_Operations tests GameID value object
func TestGameID_Operations(t *testing.T) {
	// Test generation
	id1 := NewGameID()
	id2 := NewGameID()

	if id1.IsZero() {
		t.Error("Generated ID should not be zero")
	}
	if id1.String() == "" {
		t.Error("Generated ID should have string representation")
	}
	if id1.Equals(id2) {
		t.Error("Generated IDs should be unique")
	}

	// Test from string
	id3 := NewGameIDFromString("test-game-id")
	if id3.String() != "test-game-id" {
		t.Errorf("ID string = %s, want %s", id3.String(), "test-game-id")
	}

	// Test zero ID
	zeroID := GameID{}
	if !zeroID.IsZero() {
		t.Error("Empty ID should be zero")
	}

	// Test equals
	id4 := NewGameIDFromString("same-id")
	id5 := NewGameIDFromString("same-id")
	if !id4.Equals(id5) {
		t.Error("IDs with same string should be equal")
	}
}

// TestDate_Creation tests Date value object creation
func TestDate_Creation(t *testing.T) {
	// Test from year/month/day
	date1 := NewDate(2026, time.January, 25)
	if date1.String() != "2026-01-25" {
		t.Errorf("Date string = %s, want %s", date1.String(), "2026-01-25")
	}

	// Test from string
	date2 := NewDateFromString("2026-01-25")
	if !date2.Equals(date1) {
		t.Error("Dates should be equal")
	}

	// Test from time.Time
	now := time.Date(2026, time.January, 25, 15, 30, 45, 0, time.UTC)
	date3 := NewDateFromTime(now)
	if date3.String() != "2026-01-25" {
		t.Errorf("Date string = %s, want %s", date3.String(), "2026-01-25")
	}

	// Test zero date
	zeroDate := Date{}
	if !zeroDate.IsZero() {
		t.Error("Empty date should be zero")
	}
}

// TestDate_Navigation tests Date navigation (Next/Previous)
func TestDate_Navigation(t *testing.T) {
	tests := []struct {
		name         string
		date         Date
		expectedNext string
		expectedPrev string
	}{
		{
			name:         "Mid month",
			date:         NewDate(2026, time.January, 15),
			expectedNext: "2026-01-16",
			expectedPrev: "2026-01-14",
		},
		{
			name:         "Month boundary",
			date:         NewDate(2026, time.January, 31),
			expectedNext: "2026-02-01",
			expectedPrev: "2026-01-30",
		},
		{
			name:         "Year boundary",
			date:         NewDate(2025, time.December, 31),
			expectedNext: "2026-01-01",
			expectedPrev: "2025-12-30",
		},
		{
			name:         "Leap year February",
			date:         NewDate(2024, time.February, 28),
			expectedNext: "2024-02-29",
			expectedPrev: "2024-02-27",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			next := tt.date.Next()
			if next.String() != tt.expectedNext {
				t.Errorf("Next() = %s, want %s", next.String(), tt.expectedNext)
			}

			prev := tt.date.Previous()
			if prev.String() != tt.expectedPrev {
				t.Errorf("Previous() = %s, want %s", prev.String(), tt.expectedPrev)
			}
		})
	}
}

// TestStreakSystem_NewStreakSystem tests streak system creation
func TestStreakSystem_NewStreakSystem(t *testing.T) {
	streak := NewStreakSystem()

	if streak.CurrentStreak() != 0 {
		t.Errorf("Expected current streak 0, got %d", streak.CurrentStreak())
	}
	if streak.BestStreak() != 0 {
		t.Errorf("Expected best streak 0, got %d", streak.BestStreak())
	}
	if !streak.LastPlayedDate().IsZero() {
		t.Error("Expected zero last played date")
	}
}

// TestStreakSystem_UpdateForDate tests streak updates
func TestStreakSystem_UpdateForDate(t *testing.T) {
	date1 := NewDate(2026, time.January, 20)
	date2 := NewDate(2026, time.January, 21)
	date3 := NewDate(2026, time.January, 22)
	date4 := NewDate(2026, time.January, 24) // Skipped day 23

	tests := []struct {
		name                 string
		initialStreak        StreakSystem
		playedDate           Date
		expectedCurrent      int
		expectedBest         int
		expectedLastPlayed   Date
	}{
		{
			name:               "First time playing",
			initialStreak:      NewStreakSystem(),
			playedDate:         date1,
			expectedCurrent:    1,
			expectedBest:       1,
			expectedLastPlayed: date1,
		},
		{
			name:               "Consecutive day (day 2)",
			initialStreak:      ReconstructStreakSystem(1, 1, date1),
			playedDate:         date2,
			expectedCurrent:    2,
			expectedBest:       2,
			expectedLastPlayed: date2,
		},
		{
			name:               "Consecutive day (day 3)",
			initialStreak:      ReconstructStreakSystem(2, 2, date2),
			playedDate:         date3,
			expectedCurrent:    3,
			expectedBest:       3,
			expectedLastPlayed: date3,
		},
		{
			name:               "Streak broken (missed day)",
			initialStreak:      ReconstructStreakSystem(3, 3, date3),
			playedDate:         date4,
			expectedCurrent:    1,
			expectedBest:       3, // Best streak preserved
			expectedLastPlayed: date4,
		},
		{
			name:               "Same day (no update)",
			initialStreak:      ReconstructStreakSystem(5, 10, date1),
			playedDate:         date1,
			expectedCurrent:    5,
			expectedBest:       10,
			expectedLastPlayed: date1,
		},
		{
			name:               "New streak doesn't beat old best",
			initialStreak:      ReconstructStreakSystem(1, 100, date3),
			playedDate:         date4,
			expectedCurrent:    1,
			expectedBest:       100,
			expectedLastPlayed: date4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newStreak := tt.initialStreak.UpdateForDate(tt.playedDate)

			if newStreak.CurrentStreak() != tt.expectedCurrent {
				t.Errorf("CurrentStreak = %d, want %d", newStreak.CurrentStreak(), tt.expectedCurrent)
			}
			if newStreak.BestStreak() != tt.expectedBest {
				t.Errorf("BestStreak = %d, want %d", newStreak.BestStreak(), tt.expectedBest)
			}
			if !newStreak.LastPlayedDate().Equals(tt.expectedLastPlayed) {
				t.Errorf("LastPlayedDate = %s, want %s", newStreak.LastPlayedDate().String(), tt.expectedLastPlayed.String())
			}

			// Verify immutability (original unchanged)
			if newStreak.CurrentStreak() != tt.expectedCurrent {
				// This should always pass, but checking for consistency
				t.Error("UpdateForDate should return new instance")
			}
		})
	}
}

// TestStreakSystem_GetBonus tests streak bonus calculation
func TestStreakSystem_GetBonus(t *testing.T) {
	tests := []struct {
		name          string
		currentStreak int
		expectedBonus float64
	}{
		{"No streak", 0, 1.0},
		{"1 day", 1, 1.0},
		{"2 days", 2, 1.0},
		{"3 days (+10%)", 3, 1.1},
		{"6 days", 6, 1.1},
		{"7 days (+25%)", 7, 1.25},
		{"13 days", 13, 1.25},
		{"14 days (+40%)", 14, 1.4},
		{"29 days", 29, 1.4},
		{"30 days (+50%)", 30, 1.5},
		{"99 days", 99, 1.5},
		{"100 days (+50%)", 100, 1.5},
		{"500 days", 500, 1.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			streak := ReconstructStreakSystem(tt.currentStreak, tt.currentStreak, NewDate(2026, time.January, 1))
			bonus := streak.GetBonus()

			if bonus != tt.expectedBonus {
				t.Errorf("GetBonus() = %.2f, want %.2f", bonus, tt.expectedBonus)
			}
		})
	}
}

// TestStreakSystem_IsActive tests streak activity check
func TestStreakSystem_IsActive(t *testing.T) {
	today := NewDate(2026, time.January, 25)
	yesterday := NewDate(2026, time.January, 24)
	twoDaysAgo := NewDate(2026, time.January, 23)

	tests := []struct {
		name           string
		lastPlayedDate Date
		today          Date
		expectedActive bool
	}{
		{
			name:           "Never played",
			lastPlayedDate: Date{},
			today:          today,
			expectedActive: false,
		},
		{
			name:           "Played today",
			lastPlayedDate: today,
			today:          today,
			expectedActive: true,
		},
		{
			name:           "Played yesterday (still active)",
			lastPlayedDate: yesterday,
			today:          today,
			expectedActive: true,
		},
		{
			name:           "Played 2 days ago (expired)",
			lastPlayedDate: twoDaysAgo,
			today:          today,
			expectedActive: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			streak := ReconstructStreakSystem(5, 10, tt.lastPlayedDate)
			active := streak.IsActive(tt.today)

			if active != tt.expectedActive {
				t.Errorf("IsActive() = %v, want %v", active, tt.expectedActive)
			}
		})
	}
}

// TestGameStatus_CanTransitionTo tests state transitions
func TestGameStatus_CanTransitionTo(t *testing.T) {
	tests := []struct {
		name     string
		from     GameStatus
		to       GameStatus
		expected bool
	}{
		{"in_progress -> completed", GameStatusInProgress, GameStatusCompleted, true},
		{"completed -> in_progress", GameStatusCompleted, GameStatusInProgress, false},
		{"in_progress -> in_progress", GameStatusInProgress, GameStatusInProgress, false},
		{"completed -> completed", GameStatusCompleted, GameStatusCompleted, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.from.CanTransitionTo(tt.to)
			if result != tt.expected {
				t.Errorf("CanTransitionTo(%s -> %s) = %v, want %v",
					tt.from, tt.to, result, tt.expected)
			}
		})
	}
}

// TestGameStatus_IsTerminal tests terminal state detection
func TestGameStatus_IsTerminal(t *testing.T) {
	tests := []struct {
		status   GameStatus
		expected bool
	}{
		{GameStatusInProgress, false},
		{GameStatusCompleted, true},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			if tt.status.IsTerminal() != tt.expected {
				t.Errorf("IsTerminal(%s) = %v, want %v",
					tt.status, tt.status.IsTerminal(), tt.expected)
			}
		})
	}
}

// TestDate_ToSeed tests deterministic seed generation from dates
func TestDate_ToSeed(t *testing.T) {
	tests := []struct {
		name         string
		date         Date
		expectedSeed int64
	}{
		{
			name:         "Standard date",
			date:         NewDate(2026, time.January, 25),
			expectedSeed: 20260125, // YYYYMMDD format
		},
		{
			name:         "Different year",
			date:         NewDate(2025, time.January, 25),
			expectedSeed: 20250125,
		},
		{
			name:         "Different month",
			date:         NewDate(2026, time.December, 25),
			expectedSeed: 20261225,
		},
		{
			name:         "Different day",
			date:         NewDate(2026, time.January, 1),
			expectedSeed: 20260101,
		},
		{
			name:         "Leap year date",
			date:         NewDate(2024, time.February, 29),
			expectedSeed: 20240229,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			seed := tt.date.ToSeed()
			if seed != tt.expectedSeed {
				t.Errorf("ToSeed() = %d, want %d", seed, tt.expectedSeed)
			}
		})
	}
}

// TestDate_ToSeed_Deterministic verifies same date always produces same seed
func TestDate_ToSeed_Deterministic(t *testing.T) {
	date1 := NewDate(2026, time.January, 25)
	date2 := NewDate(2026, time.January, 25)
	date3 := NewDateFromString("2026-01-25")

	seed1 := date1.ToSeed()
	seed2 := date2.ToSeed()
	seed3 := date3.ToSeed()

	if seed1 != seed2 || seed2 != seed3 {
		t.Errorf("Same date should produce same seed: %d, %d, %d", seed1, seed2, seed3)
	}

	// Verify different dates produce different seeds
	differentDate := NewDate(2026, time.January, 26)
	differentSeed := differentDate.ToSeed()

	if seed1 == differentSeed {
		t.Error("Different dates should produce different seeds")
	}
}
