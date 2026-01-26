package solo_marathon

import (
	"testing"
)

// TestLivesSystem_NewLivesSystem tests creation of lives system
func TestLivesSystem_NewLivesSystem(t *testing.T) {
	now := int64(1000000)

	lives := NewLivesSystem(now)

	if lives.CurrentLives() != MaxLives {
		t.Errorf("Expected %d lives, got %d", MaxLives, lives.CurrentLives())
	}
	if lives.MaxLives() != MaxLives {
		t.Errorf("Expected max lives %d, got %d", MaxLives, lives.MaxLives())
	}
	if lives.LastUpdate() != now {
		t.Errorf("Expected last update %d, got %d", now, lives.LastUpdate())
	}
}

// TestLivesSystem_LoseLife tests losing a life
func TestLivesSystem_LoseLife(t *testing.T) {
	now := int64(1000000)
	lives := NewLivesSystem(now)

	// Lose first life
	newLives := lives.LoseLife(now + 100)

	if newLives.CurrentLives() != 2 {
		t.Errorf("Expected 2 lives after losing one, got %d", newLives.CurrentLives())
	}
	if newLives.LastUpdate() != now+100 {
		t.Errorf("Expected last update %d, got %d", now+100, newLives.LastUpdate())
	}

	// Original should be unchanged (immutable)
	if lives.CurrentLives() != 3 {
		t.Errorf("Original lives should be unchanged, got %d", lives.CurrentLives())
	}
}

// TestLivesSystem_LoseLife_CannotGoNegative tests lives cannot go below zero
func TestLivesSystem_LoseLife_CannotGoNegative(t *testing.T) {
	now := int64(1000000)
	lives := ReconstructLivesSystem(0, now)

	newLives := lives.LoseLife(now + 100)

	if newLives.CurrentLives() != 0 {
		t.Errorf("Lives should not go negative, got %d", newLives.CurrentLives())
	}
}

// TestLivesSystem_HasLives tests checking if player has lives
func TestLivesSystem_HasLives(t *testing.T) {
	tests := []struct {
		name          string
		currentLives  int
		expectedHas   bool
	}{
		{"3 lives", 3, true},
		{"1 life", 1, true},
		{"0 lives", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lives := ReconstructLivesSystem(tt.currentLives, 1000000)

			if lives.HasLives() != tt.expectedHas {
				t.Errorf("HasLives() = %v, want %v", lives.HasLives(), tt.expectedHas)
			}
		})
	}
}

// TestLivesSystem_RegenerateLives tests life regeneration over time
func TestLivesSystem_RegenerateLives(t *testing.T) {
	now := int64(1000000)
	lives := ReconstructLivesSystem(1, now)

	tests := []struct {
		name          string
		timeElapsed   int64
		expectedLives int
	}{
		{"No time passed", 0, 1},
		{"2 hours passed", 2 * 60 * 60, 1}, // Not enough for 1 life
		{"4 hours passed", LifeRegenInterval, 2}, // Exactly 1 life
		{"8 hours passed", 2 * LifeRegenInterval, 3}, // 2 lives (capped at max)
		{"12 hours passed", 3 * LifeRegenInterval, 3}, // 3 lives but capped at max
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newLives := lives.RegenerateLives(now + tt.timeElapsed)

			if newLives.CurrentLives() != tt.expectedLives {
				t.Errorf("Expected %d lives after %d seconds, got %d",
					tt.expectedLives, tt.timeElapsed, newLives.CurrentLives())
			}
		})
	}
}

// TestLivesSystem_RegenerateLives_AlreadyFull tests regeneration when already at max
func TestLivesSystem_RegenerateLives_AlreadyFull(t *testing.T) {
	now := int64(1000000)
	lives := NewLivesSystem(now)

	newLives := lives.RegenerateLives(now + LifeRegenInterval)

	if newLives.CurrentLives() != MaxLives {
		t.Errorf("Expected max lives %d, got %d", MaxLives, newLives.CurrentLives())
	}
}

// TestLivesSystem_AddLives tests adding lives
func TestLivesSystem_AddLives(t *testing.T) {
	now := int64(1000000)
	lives := ReconstructLivesSystem(1, now)

	newLives := lives.AddLives(2, now+100)

	if newLives.CurrentLives() != 3 {
		t.Errorf("Expected 3 lives after adding 2, got %d", newLives.CurrentLives())
	}

	// Test capping at max
	newLives2 := newLives.AddLives(5, now+200)
	if newLives2.CurrentLives() != MaxLives {
		t.Errorf("Lives should be capped at %d, got %d", MaxLives, newLives2.CurrentLives())
	}
}

// TestLivesSystem_TimeToNextLife tests time calculation
func TestLivesSystem_TimeToNextLife(t *testing.T) {
	now := int64(1000000)

	tests := []struct {
		name           string
		currentLives   int
		lastUpdate     int64
		currentTime    int64
		expectedTime   int64
	}{
		{
			name:         "Already at max",
			currentLives: 3,
			lastUpdate:   now,
			currentTime:  now,
			expectedTime: 0,
		},
		{
			name:         "Just lost a life",
			currentLives: 2,
			lastUpdate:   now,
			currentTime:  now,
			expectedTime: LifeRegenInterval,
		},
		{
			name:         "2 hours passed",
			currentLives: 2,
			lastUpdate:   now,
			currentTime:  now + (2 * 60 * 60),
			expectedTime: LifeRegenInterval - (2 * 60 * 60),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lives := ReconstructLivesSystem(tt.currentLives, tt.lastUpdate)
			timeToNext := lives.TimeToNextLife(tt.currentTime)

			if timeToNext != tt.expectedTime {
				t.Errorf("Expected time to next life %d, got %d", tt.expectedTime, timeToNext)
			}
		})
	}
}

// TestHintsSystem_NewHintsSystem tests creation of hints system
func TestHintsSystem_NewHintsSystem(t *testing.T) {
	hints := NewHintsSystem()

	if hints.FiftyFifty() != DefaultFiftyFifty {
		t.Errorf("Expected %d 50/50 hints, got %d", DefaultFiftyFifty, hints.FiftyFifty())
	}
	if hints.ExtraTime() != DefaultExtraTime {
		t.Errorf("Expected %d extra time hints, got %d", DefaultExtraTime, hints.ExtraTime())
	}
	if hints.Skip() != DefaultSkip {
		t.Errorf("Expected %d skip hints, got %d", DefaultSkip, hints.Skip())
	}
}

// TestHintsSystem_UseHint tests using hints
func TestHintsSystem_UseHint(t *testing.T) {
	tests := []struct {
		name        string
		hintType    HintType
		initialFF   int
		initialET   int
		initialSkip int
		expectError bool
		expectedFF  int
		expectedET  int
		expectedSkip int
	}{
		{
			name:        "Use 50/50",
			hintType:    HintFiftyFifty,
			initialFF:   3,
			initialET:   2,
			initialSkip: 1,
			expectError: false,
			expectedFF:  2,
			expectedET:  2,
			expectedSkip: 1,
		},
		{
			name:        "Use extra time",
			hintType:    HintExtraTime,
			initialFF:   3,
			initialET:   2,
			initialSkip: 1,
			expectError: false,
			expectedFF:  3,
			expectedET:  1,
			expectedSkip: 1,
		},
		{
			name:        "Use skip",
			hintType:    HintSkip,
			initialFF:   3,
			initialET:   2,
			initialSkip: 1,
			expectError: false,
			expectedFF:  3,
			expectedET:  2,
			expectedSkip: 0,
		},
		{
			name:        "No 50/50 available",
			hintType:    HintFiftyFifty,
			initialFF:   0,
			initialET:   2,
			initialSkip: 1,
			expectError: true,
			expectedFF:  0,
			expectedET:  2,
			expectedSkip: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hints := ReconstructHintsSystem(tt.initialFF, tt.initialET, tt.initialSkip)

			newHints, err := hints.UseHint(tt.hintType)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error, got nil")
				}
				if err != ErrNoHintsAvailable {
					t.Errorf("Expected ErrNoHintsAvailable, got %v", err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if newHints.FiftyFifty() != tt.expectedFF {
					t.Errorf("Expected %d 50/50 hints, got %d", tt.expectedFF, newHints.FiftyFifty())
				}
				if newHints.ExtraTime() != tt.expectedET {
					t.Errorf("Expected %d extra time hints, got %d", tt.expectedET, newHints.ExtraTime())
				}
				if newHints.Skip() != tt.expectedSkip {
					t.Errorf("Expected %d skip hints, got %d", tt.expectedSkip, newHints.Skip())
				}

				// Test immutability
				if hints.FiftyFifty() != tt.initialFF {
					t.Errorf("Original hints should be unchanged")
				}
			}
		})
	}
}

// TestHintsSystem_HasHint tests checking hint availability
func TestHintsSystem_HasHint(t *testing.T) {
	hints := ReconstructHintsSystem(1, 0, 1)

	if !hints.HasHint(HintFiftyFifty) {
		t.Errorf("Should have 50/50 hint")
	}
	if hints.HasHint(HintExtraTime) {
		t.Errorf("Should not have extra time hint")
	}
	if !hints.HasHint(HintSkip) {
		t.Errorf("Should have skip hint")
	}
}

// TestDifficultyProgression_UpdateFromStreak tests difficulty calculation
func TestDifficultyProgression_UpdateFromStreak(t *testing.T) {
	tests := []struct {
		name              string
		streak            int
		expectedLevel     DifficultyLevel
	}{
		{"Beginner (1)", 1, DifficultyBeginner},
		{"Beginner (5)", 5, DifficultyBeginner},
		{"Medium (6)", 6, DifficultyMedium},
		{"Medium (15)", 15, DifficultyMedium},
		{"Hard (16)", 16, DifficultyHard},
		{"Hard (30)", 30, DifficultyHard},
		{"Expert (31)", 31, DifficultyExpert},
		{"Expert (50)", 50, DifficultyExpert},
		{"Master (51)", 51, DifficultyMaster},
		{"Master (100)", 100, DifficultyMaster},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dp := NewDifficultyProgression()
			newDP := dp.UpdateFromStreak(tt.streak)

			if newDP.Level() != tt.expectedLevel {
				t.Errorf("Expected level %s for streak %d, got %s",
					tt.expectedLevel, tt.streak, newDP.Level())
			}
		})
	}
}

// TestDifficultyProgression_GetDistribution tests difficulty distribution
func TestDifficultyProgression_GetDistribution(t *testing.T) {
	tests := []struct {
		name     string
		level    DifficultyLevel
		expected map[string]float64
	}{
		{
			name:  "Beginner",
			level: DifficultyBeginner,
			expected: map[string]float64{"easy": 0.8, "medium": 0.2, "hard": 0.0},
		},
		{
			name:  "Master",
			level: DifficultyMaster,
			expected: map[string]float64{"easy": 0.0, "medium": 0.3, "hard": 0.7},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dp := DifficultyProgression{level: tt.level}
			dist := dp.GetDistribution()

			for key, expectedVal := range tt.expected {
				if dist[key] != expectedVal {
					t.Errorf("Distribution[%s] = %f, want %f", key, dist[key], expectedVal)
				}
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
		{"in_progress -> finished", GameStatusInProgress, GameStatusFinished, true},
		{"in_progress -> abandoned", GameStatusInProgress, GameStatusAbandoned, true},
		{"finished -> in_progress", GameStatusFinished, GameStatusInProgress, false},
		{"finished -> abandoned", GameStatusFinished, GameStatusAbandoned, false},
		{"abandoned -> in_progress", GameStatusAbandoned, GameStatusInProgress, false},
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
		{GameStatusFinished, true},
		{GameStatusAbandoned, true},
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
