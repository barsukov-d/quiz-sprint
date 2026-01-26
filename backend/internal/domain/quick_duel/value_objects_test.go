package quick_duel

import (
	"github.com/barsukov/quiz-sprint/backend/internal/domain/shared"
	"testing"
)

// TestGameID_Operations tests GameID value object
func TestGameID_Operations(t *testing.T) {
	id1 := NewGameID()
	id2 := NewGameID()

	if id1.IsZero() {
		t.Error("Generated ID should not be zero")
	}
	if id1.Equals(id2) {
		t.Error("Generated IDs should be unique")
	}

	id3 := NewGameIDFromString("test-id")
	if id3.String() != "test-id" {
		t.Errorf("ID string = %s, want %s", id3.String(), "test-id")
	}
}

// TestEloRating_NewEloRating tests ELO creation
func TestEloRating_NewEloRating(t *testing.T) {
	elo := NewEloRating()

	if elo.Rating() != InitialEloRating {
		t.Errorf("Initial rating = %d, want %d", elo.Rating(), InitialEloRating)
	}
	if elo.GamesPlayed() != 0 {
		t.Errorf("Initial games played = %d, want 0", elo.GamesPlayed())
	}
	if !elo.IsNewPlayer() {
		t.Error("Should be new player")
	}
}

// TestEloRating_KFactor tests K-factor calculation
func TestEloRating_KFactor(t *testing.T) {
	tests := []struct {
		name        string
		gamesPlayed int
		expectedK   int
	}{
		{"New player (0 games)", 0, KFactorNew},
		{"New player (29 games)", 29, KFactorNew},
		{"Veteran (30 games)", 30, KFactorRegular},
		{"Veteran (100 games)", 100, KFactorRegular},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			elo := ReconstructEloRating(1000, tt.gamesPlayed)
			k := elo.KFactor()

			if k != tt.expectedK {
				t.Errorf("KFactor() = %d, want %d", k, tt.expectedK)
			}
		})
	}
}

// TestEloRating_CalculateNewRating tests ELO rating updates
func TestEloRating_CalculateNewRating(t *testing.T) {
	tests := []struct {
		name             string
		initialRating    int
		gamesPlayed      int
		won              bool
		opponentRating   int
		expectedIncrease bool
	}{
		{
			name:             "New player wins vs equal",
			initialRating:    1000,
			gamesPlayed:      5,
			won:              true,
			opponentRating:   1000,
			expectedIncrease: true,
		},
		{
			name:             "New player loses vs equal",
			initialRating:    1000,
			gamesPlayed:      5,
			won:              false,
			opponentRating:   1000,
			expectedIncrease: false,
		},
		{
			name:             "Veteran wins vs weaker",
			initialRating:    1200,
			gamesPlayed:      50,
			won:              true,
			opponentRating:   1000,
			expectedIncrease: true,
		},
		{
			name:             "Low rating enforces minimum",
			initialRating:    150,
			gamesPlayed:      10,
			won:              false,
			opponentRating:   1000,
			expectedIncrease: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			elo := ReconstructEloRating(tt.initialRating, tt.gamesPlayed)
			newElo := elo.CalculateNewRating(tt.won, tt.opponentRating)

			if tt.expectedIncrease {
				if newElo.Rating() <= tt.initialRating {
					t.Errorf("Rating should increase, got %d (was %d)", newElo.Rating(), tt.initialRating)
				}
			} else {
				if newElo.Rating() > tt.initialRating {
					t.Errorf("Rating should not increase, got %d (was %d)", newElo.Rating(), tt.initialRating)
				}
			}

			// Check games played incremented
			if newElo.GamesPlayed() != tt.gamesPlayed+1 {
				t.Errorf("Games played = %d, want %d", newElo.GamesPlayed(), tt.gamesPlayed+1)
			}

			// Check minimum rating enforced
			if newElo.Rating() < MinEloRating {
				t.Errorf("Rating %d below minimum %d", newElo.Rating(), MinEloRating)
			}

			// Verify immutability
			if elo.Rating() != tt.initialRating {
				t.Error("Original ELO should be unchanged")
			}
		})
	}
}

// TestEloRating_GetMatchmakingRange tests matchmaking ranges
func TestEloRating_GetMatchmakingRange(t *testing.T) {
	elo := ReconstructEloRating(1000, 10)

	tests := []struct {
		name          string
		searchSeconds int
		expectedMin   int
		expectedMax   int
	}{
		{"Instant (0s)", 0, 950, 1050},
		{"Quick (4s)", 4, 950, 1050},
		{"Medium (7s)", 7, 900, 1100},
		{"Long (12s)", 12, 800, 1200},
		{"Very long (20s)", 20, MinEloRating, 9999},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			min, max := elo.GetMatchmakingRange(tt.searchSeconds)

			if min != tt.expectedMin {
				t.Errorf("Min = %d, want %d", min, tt.expectedMin)
			}
			if max != tt.expectedMax {
				t.Errorf("Max = %d, want %d", max, tt.expectedMax)
			}
		})
	}
}

// TestDuelPlayer_Operations tests DuelPlayer value object
func TestDuelPlayer_Operations(t *testing.T) {
	userID, _ := shared.NewUserID("user123")
	elo := NewEloRating()
	player := NewDuelPlayer(userID, "TestPlayer", elo)

	if player.UserID() != userID {
		t.Errorf("UserID = %v, want %v", player.UserID(), userID)
	}
	if player.Username() != "TestPlayer" {
		t.Errorf("Username = %s, want TestPlayer", player.Username())
	}
	if player.Score() != 0 {
		t.Errorf("Initial score = %d, want 0", player.Score())
	}
	if !player.Connected() {
		t.Error("Should be connected initially")
	}
	if player.AnswersCount() != 0 {
		t.Errorf("Initial answers = %d, want 0", player.AnswersCount())
	}

	// Test AddScore (immutable)
	player2 := player.AddScore(100)
	if player2.Score() != 100 {
		t.Errorf("Score = %d, want 100", player2.Score())
	}
	if player2.AnswersCount() != 1 {
		t.Errorf("Answers count = %d, want 1", player2.AnswersCount())
	}
	if player.Score() != 0 {
		t.Error("Original player should be unchanged")
	}

	// Test SetConnected
	player3 := player.SetConnected(false)
	if player3.Connected() {
		t.Error("Should be disconnected")
	}
	if player.Connected() != true {
		t.Error("Original player should be unchanged")
	}
}

// TestWinStreak_Operations tests WinStreak value object
func TestWinStreak_Operations(t *testing.T) {
	streak := NewWinStreak()

	if streak.CurrentStreak() != 0 {
		t.Errorf("Initial current streak = %d, want 0", streak.CurrentStreak())
	}
	if streak.BestStreak() != 0 {
		t.Errorf("Initial best streak = %d, want 0", streak.BestStreak())
	}

	// Increment wins
	streak1 := streak.IncrementWin()
	if streak1.CurrentStreak() != 1 {
		t.Errorf("Current streak = %d, want 1", streak1.CurrentStreak())
	}
	if streak1.BestStreak() != 1 {
		t.Errorf("Best streak = %d, want 1", streak1.BestStreak())
	}

	streak2 := streak1.IncrementWin()
	if streak2.CurrentStreak() != 2 {
		t.Errorf("Current streak = %d, want 2", streak2.CurrentStreak())
	}

	streak3 := streak2.IncrementWin()
	if streak3.CurrentStreak() != 3 {
		t.Errorf("Current streak = %d, want 3", streak3.CurrentStreak())
	}
	if !streak3.IsMilestone() {
		t.Error("Streak 3 should be milestone")
	}

	// Reset on loss
	streakAfterLoss := streak3.ResetOnLoss()
	if streakAfterLoss.CurrentStreak() != 0 {
		t.Errorf("Current streak after loss = %d, want 0", streakAfterLoss.CurrentStreak())
	}
	if streakAfterLoss.BestStreak() != 3 {
		t.Errorf("Best streak should be preserved = %d, want 3", streakAfterLoss.BestStreak())
	}

	// Verify immutability
	if streak.CurrentStreak() != 0 {
		t.Error("Original streak should be unchanged")
	}
}

// TestWinStreak_GetBonusMultiplier tests bonus calculation
func TestWinStreak_GetBonusMultiplier(t *testing.T) {
	tests := []struct {
		streak         int
		expectedBonus  float64
	}{
		{0, 1.0},
		{2, 1.0},
		{3, 1.1},
		{4, 1.1},
		{5, 1.25},
		{9, 1.25},
		{10, 1.5},
		{50, 1.5},
	}

	for _, tt := range tests {
		t.Run("streak "+string(rune(tt.streak)), func(t *testing.T) {
			ws := ReconstructWinStreak(tt.streak, tt.streak)
			bonus := ws.GetBonusMultiplier()

			if bonus != tt.expectedBonus {
				t.Errorf("GetBonusMultiplier() = %.2f, want %.2f", bonus, tt.expectedBonus)
			}
		})
	}
}

// TestCalculateSpeedBonus tests speed bonus calculation
func TestCalculateSpeedBonus(t *testing.T) {
	tests := []struct {
		name          string
		timeTakenMs   int64
		expectedBonus int
	}{
		{"Super fast (1s)", 1000, 50},
		{"Fast (3s)", 3000, 50},
		{"Good (4s)", 4000, 25},
		{"OK (6s)", 6000, 10},
		{"Slow (8s)", 8000, 0},
		{"Very slow (10s)", 10000, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bonus := CalculateSpeedBonus(tt.timeTakenMs)

			if bonus != tt.expectedBonus {
				t.Errorf("CalculateSpeedBonus(%d) = %d, want %d", tt.timeTakenMs, bonus, tt.expectedBonus)
			}
		})
	}
}

// TestGameStatus_Transitions tests state transitions
func TestGameStatus_Transitions(t *testing.T) {
	tests := []struct {
		name     string
		from     GameStatus
		to       GameStatus
		expected bool
	}{
		{"waiting_start -> in_progress", GameStatusWaitingStart, GameStatusInProgress, true},
		{"waiting_start -> abandoned", GameStatusWaitingStart, GameStatusAbandoned, true},
		{"in_progress -> finished", GameStatusInProgress, GameStatusFinished, true},
		{"in_progress -> abandoned", GameStatusInProgress, GameStatusAbandoned, true},
		{"finished -> in_progress", GameStatusFinished, GameStatusInProgress, false},
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
		{GameStatusWaitingStart, false},
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
