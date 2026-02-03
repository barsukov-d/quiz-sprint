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

// TestLivesSystem_ResetForContinue tests resetting lives for continue
func TestLivesSystem_ResetForContinue(t *testing.T) {
	now := int64(1000000)
	lives := ReconstructLivesSystem(0, now)

	newLives := lives.ResetForContinue(now + 100)

	if newLives.CurrentLives() != 1 {
		t.Errorf("Expected 1 life after continue, got %d", newLives.CurrentLives())
	}
	if newLives.LastUpdate() != now+100 {
		t.Errorf("Expected last update %d, got %d", now+100, newLives.LastUpdate())
	}

	// Original should be unchanged (immutable)
	if lives.CurrentLives() != 0 {
		t.Errorf("Original lives should be unchanged, got %d", lives.CurrentLives())
	}
}

// TestLivesSystem_Label tests visual representation of lives
func TestLivesSystem_Label(t *testing.T) {
	tests := []struct {
		name     string
		lives    int
		expected string
	}{
		{"3 lives", 3, "❤️❤️❤️"},
		{"2 lives", 2, "❤️❤️🖤"},
		{"1 life", 1, "❤️🖤🖤"},
		{"0 lives", 0, "🖤🖤🖤"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lives := ReconstructLivesSystem(tt.lives, 1000000)
			if lives.Label() != tt.expected {
				t.Errorf("Label() = %q, want %q", lives.Label(), tt.expected)
			}
		})
	}
}

// TestBonusInventory_NewBonusInventory tests creation of bonus inventory
func TestBonusInventory_NewBonusInventory(t *testing.T) {
	bonuses := NewBonusInventory()

	if bonuses.Shield() != DefaultShield {
		t.Errorf("Expected %d shields, got %d", DefaultShield, bonuses.Shield())
	}
	if bonuses.FiftyFifty() != DefaultFiftyFifty {
		t.Errorf("Expected %d 50/50 bonuses, got %d", DefaultFiftyFifty, bonuses.FiftyFifty())
	}
	if bonuses.Skip() != DefaultSkip {
		t.Errorf("Expected %d skip bonuses, got %d", DefaultSkip, bonuses.Skip())
	}
	if bonuses.Freeze() != DefaultFreeze {
		t.Errorf("Expected %d freeze bonuses, got %d", DefaultFreeze, bonuses.Freeze())
	}
}

// TestBonusInventory_UseBonus tests using bonuses
func TestBonusInventory_UseBonus(t *testing.T) {
	tests := []struct {
		name            string
		bonusType       BonusType
		initialShield   int
		initialFF       int
		initialSkip     int
		initialFreeze   int
		expectError     bool
		expectedShield  int
		expectedFF      int
		expectedSkip    int
		expectedFreeze  int
	}{
		{
			name:           "Use shield",
			bonusType:      BonusShield,
			initialShield:  2,
			initialFF:      1,
			initialSkip:    1,
			initialFreeze:  3,
			expectError:    false,
			expectedShield: 1,
			expectedFF:     1,
			expectedSkip:   1,
			expectedFreeze: 3,
		},
		{
			name:           "Use 50/50",
			bonusType:      BonusFiftyFifty,
			initialShield:  2,
			initialFF:      1,
			initialSkip:    1,
			initialFreeze:  3,
			expectError:    false,
			expectedShield: 2,
			expectedFF:     0,
			expectedSkip:   1,
			expectedFreeze: 3,
		},
		{
			name:           "Use freeze",
			bonusType:      BonusFreeze,
			initialShield:  2,
			initialFF:      1,
			initialSkip:    1,
			initialFreeze:  3,
			expectError:    false,
			expectedShield: 2,
			expectedFF:     1,
			expectedSkip:   1,
			expectedFreeze: 2,
		},
		{
			name:           "Use skip",
			bonusType:      BonusSkip,
			initialShield:  2,
			initialFF:      1,
			initialSkip:    1,
			initialFreeze:  3,
			expectError:    false,
			expectedShield: 2,
			expectedFF:     1,
			expectedSkip:   0,
			expectedFreeze: 3,
		},
		{
			name:           "No 50/50 available",
			bonusType:      BonusFiftyFifty,
			initialShield:  2,
			initialFF:      0,
			initialSkip:    1,
			initialFreeze:  3,
			expectError:    true,
			expectedShield: 2,
			expectedFF:     0,
			expectedSkip:   1,
			expectedFreeze: 3,
		},
		{
			name:           "No shield available",
			bonusType:      BonusShield,
			initialShield:  0,
			initialFF:      1,
			initialSkip:    1,
			initialFreeze:  3,
			expectError:    true,
			expectedShield: 0,
			expectedFF:     1,
			expectedSkip:   1,
			expectedFreeze: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bonuses := ReconstructBonusInventory(tt.initialShield, tt.initialFF, tt.initialSkip, tt.initialFreeze)

			newBonuses, err := bonuses.UseBonus(tt.bonusType)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error, got nil")
				}
				if err != ErrNoBonusesAvailable {
					t.Errorf("Expected ErrNoBonusesAvailable, got %v", err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if newBonuses.Shield() != tt.expectedShield {
					t.Errorf("Expected %d shields, got %d", tt.expectedShield, newBonuses.Shield())
				}
				if newBonuses.FiftyFifty() != tt.expectedFF {
					t.Errorf("Expected %d 50/50, got %d", tt.expectedFF, newBonuses.FiftyFifty())
				}
				if newBonuses.Skip() != tt.expectedSkip {
					t.Errorf("Expected %d skip, got %d", tt.expectedSkip, newBonuses.Skip())
				}
				if newBonuses.Freeze() != tt.expectedFreeze {
					t.Errorf("Expected %d freeze, got %d", tt.expectedFreeze, newBonuses.Freeze())
				}

				// Test immutability
				if bonuses.Shield() != tt.initialShield {
					t.Errorf("Original bonuses should be unchanged")
				}
			}
		})
	}
}

// TestBonusInventory_HasBonus tests checking bonus availability
func TestBonusInventory_HasBonus(t *testing.T) {
	bonuses := ReconstructBonusInventory(1, 0, 1, 2)

	if !bonuses.HasBonus(BonusShield) {
		t.Errorf("Should have shield bonus")
	}
	if bonuses.HasBonus(BonusFiftyFifty) {
		t.Errorf("Should not have 50/50 bonus")
	}
	if !bonuses.HasBonus(BonusSkip) {
		t.Errorf("Should have skip bonus")
	}
	if !bonuses.HasBonus(BonusFreeze) {
		t.Errorf("Should have freeze bonus")
	}
}

// TestBonusInventory_Count tests counting specific bonus type
func TestBonusInventory_Count(t *testing.T) {
	bonuses := ReconstructBonusInventory(2, 1, 0, 3)

	if bonuses.Count(BonusShield) != 2 {
		t.Errorf("Expected 2 shields, got %d", bonuses.Count(BonusShield))
	}
	if bonuses.Count(BonusFiftyFifty) != 1 {
		t.Errorf("Expected 1 fifty-fifty, got %d", bonuses.Count(BonusFiftyFifty))
	}
	if bonuses.Count(BonusSkip) != 0 {
		t.Errorf("Expected 0 skip, got %d", bonuses.Count(BonusSkip))
	}
	if bonuses.Count(BonusFreeze) != 3 {
		t.Errorf("Expected 3 freeze, got %d", bonuses.Count(BonusFreeze))
	}
}

// TestBonusInventory_InvalidType tests using invalid bonus type
func TestBonusInventory_InvalidType(t *testing.T) {
	bonuses := NewBonusInventory()

	_, err := bonuses.UseBonus(BonusType("invalid"))
	if err != ErrInvalidBonusType {
		t.Errorf("Expected ErrInvalidBonusType, got %v", err)
	}
}

// TestDifficultyProgression_UpdateFromQuestionIndex tests difficulty calculation
func TestDifficultyProgression_UpdateFromQuestionIndex(t *testing.T) {
	tests := []struct {
		name              string
		questionIndex     int
		expectedLevel     DifficultyLevel
	}{
		{"Beginner (1)", 1, DifficultyBeginner},
		{"Beginner (5)", 5, DifficultyBeginner},
		{"Beginner (10)", 10, DifficultyBeginner},
		{"Medium (11)", 11, DifficultyMedium},
		{"Medium (30)", 30, DifficultyMedium},
		{"Hard (31)", 31, DifficultyHard},
		{"Hard (50)", 50, DifficultyHard},
		{"Master (51)", 51, DifficultyMaster},
		{"Master (100)", 100, DifficultyMaster},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dp := NewDifficultyProgression()
			newDP := dp.UpdateFromQuestionIndex(tt.questionIndex)

			if newDP.Level() != tt.expectedLevel {
				t.Errorf("Expected level %s for question index %d, got %s",
					tt.expectedLevel, tt.questionIndex, newDP.Level())
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
			name:  "Medium",
			level: DifficultyMedium,
			expected: map[string]float64{"easy": 0.0, "medium": 1.0, "hard": 0.0},
		},
		{
			name:  "Hard",
			level: DifficultyHard,
			expected: map[string]float64{"easy": 0.0, "medium": 0.7, "hard": 0.3},
		},
		{
			name:  "Master",
			level: DifficultyMaster,
			expected: map[string]float64{"easy": 0.0, "medium": 0.0, "hard": 1.0},
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

// TestDifficultyProgression_GetTimeLimit tests time limit by question index
func TestDifficultyProgression_GetTimeLimit(t *testing.T) {
	dp := NewDifficultyProgression()

	tests := []struct {
		name          string
		questionIndex int
		expectedTime  int
	}{
		{"Question 1", 1, 15},
		{"Question 10", 10, 15},
		{"Question 11", 11, 12},
		{"Question 25", 25, 12},
		{"Question 26", 26, 10},
		{"Question 50", 50, 10},
		{"Question 51", 51, 8},
		{"Question 100", 100, 8},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			timeLimit := dp.GetTimeLimit(tt.questionIndex)
			if timeLimit != tt.expectedTime {
				t.Errorf("GetTimeLimit(%d) = %d, want %d", tt.questionIndex, timeLimit, tt.expectedTime)
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
		{"in_progress -> game_over", GameStatusInProgress, GameStatusGameOver, true},
		{"in_progress -> abandoned", GameStatusInProgress, GameStatusAbandoned, true},
		{"in_progress -> completed", GameStatusInProgress, GameStatusCompleted, false},
		{"game_over -> in_progress (continue)", GameStatusGameOver, GameStatusInProgress, true},
		{"game_over -> completed", GameStatusGameOver, GameStatusCompleted, true},
		{"game_over -> abandoned", GameStatusGameOver, GameStatusAbandoned, false},
		{"completed -> in_progress", GameStatusCompleted, GameStatusInProgress, false},
		{"completed -> abandoned", GameStatusCompleted, GameStatusAbandoned, false},
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
		{GameStatusGameOver, false},  // Intermediate, not terminal
		{GameStatusCompleted, true},
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

// TestContinueCostCalculator tests continue cost calculation
func TestContinueCostCalculator_GetCost(t *testing.T) {
	calc := ContinueCostCalculator{}

	tests := []struct {
		continueCount int
		expectedCost  int
	}{
		{0, 200},
		{1, 400},
		{2, 600},
		{3, 800},
		{4, 1000},
	}

	for _, tt := range tests {
		cost := calc.GetCost(tt.continueCount)
		if cost != tt.expectedCost {
			t.Errorf("GetCost(%d) = %d, want %d", tt.continueCount, cost, tt.expectedCost)
		}
	}
}

// TestContinueCostCalculator_HasAdOption tests ad availability
func TestContinueCostCalculator_HasAdOption(t *testing.T) {
	calc := ContinueCostCalculator{}

	tests := []struct {
		continueCount int
		expectedHasAd bool
	}{
		{0, true},
		{1, true},
		{2, true},
		{3, false},
		{4, false},
	}

	for _, tt := range tests {
		hasAd := calc.HasAdOption(tt.continueCount)
		if hasAd != tt.expectedHasAd {
			t.Errorf("HasAdOption(%d) = %v, want %v", tt.continueCount, hasAd, tt.expectedHasAd)
		}
	}
}

// TestGetNextMilestone tests milestone calculation
func TestGetNextMilestone(t *testing.T) {
	tests := []struct {
		score             int
		expectedNext      int
		expectedRemaining int
	}{
		{0, 25, 25},
		{10, 25, 15},
		{25, 50, 25},
		{49, 50, 1},
		{50, 100, 50},
		{100, 200, 100},
		{200, 500, 300},
		{500, 0, 0},  // Past all milestones
		{999, 0, 0},
	}

	for _, tt := range tests {
		next, remaining := GetNextMilestone(tt.score)
		if next != tt.expectedNext || remaining != tt.expectedRemaining {
			t.Errorf("GetNextMilestone(%d) = (%d, %d), want (%d, %d)",
				tt.score, next, remaining, tt.expectedNext, tt.expectedRemaining)
		}
	}
}
