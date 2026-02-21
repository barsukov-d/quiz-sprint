package solo_marathon

import (
	"testing"

	"github.com/barsukov/quiz-sprint/backend/internal/domain/quiz"
	"github.com/barsukov/quiz-sprint/backend/internal/domain/shared"
)

// buildV2GameWithQuestion creates a MarathonGameV2 with a loaded question and optional streak
func buildV2GameWithQuestion(t *testing.T, streakCount int, currentLives int) (*MarathonGameV2, *quiz.Question) {
	t.Helper()

	playerID, _ := shared.NewUserID("player-v2-test")
	category := NewMarathonCategoryAll()
	now := int64(1000000)

	game, err := NewMarathonGameV2(playerID, category, nil, NewBonusInventory(), now)
	if err != nil {
		t.Fatalf("Failed to create game: %v", err)
	}

	// Build a question with a known correct answer
	q := createTestQuestion(t, "Test question", "Correct", true, 1)

	// Reconstruct with streak and lives already set
	reconstructed := ReconstructMarathonGameV2(
		game.id,
		game.playerID,
		game.category,
		GameStatusInProgress,
		now,
		0,
		&q,               // currentQuestion set
		[]QuestionID{},
		[]QuestionID{},
		0,                // score
		0,                // totalQuestions
		ReconstructLivesSystem(currentLives, now),
		game.bonusInventory,
		NewDifficultyProgression(),
		false,            // shieldActive
		0,                // continueCount
		nil,              // personalBest
		map[QuestionID][]BonusType{},
		streakCount,      // streakCount
		streakCount,      // bestStreak (same for simplicity)
		0,                // livesRestored
	)

	return reconstructed, &q
}

// findCorrectAnswerID finds the correct answer ID in a question
func findCorrectAnswerID(q *quiz.Question) quiz.AnswerID {
	for _, a := range q.Answers() {
		if a.IsCorrect() {
			return a.ID()
		}
	}
	panic("no correct answer found")
}

// findWrongAnswerID finds a wrong answer ID in a question
func findWrongAnswerID(q *quiz.Question) quiz.AnswerID {
	for _, a := range q.Answers() {
		if !a.IsCorrect() {
			return a.ID()
		}
	}
	panic("no wrong answer found")
}

// TestMarathonGameV2_StreakIncrementsOnCorrectAnswer verifies streak count grows
func TestMarathonGameV2_StreakIncrementsOnCorrectAnswer(t *testing.T) {
	game, q := buildV2GameWithQuestion(t, 0, 5)

	result, err := game.AnswerQuestion(q.ID(), findCorrectAnswerID(q), 1000, 1000001)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.StreakCount != 1 {
		t.Errorf("StreakCount = %d, want 1", result.StreakCount)
	}
	if result.LifeRestored {
		t.Error("LifeRestored should be false after only 1 correct answer")
	}
}

// TestMarathonGameV2_StreakResetsOnWrongAnswer verifies streak resets on mistake
func TestMarathonGameV2_StreakResetsOnWrongAnswer(t *testing.T) {
	// Start with a streak of 3
	game, q := buildV2GameWithQuestion(t, 3, 5)

	result, err := game.AnswerQuestion(q.ID(), findWrongAnswerID(q), 1000, 1000001)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.StreakCount != 0 {
		t.Errorf("StreakCount = %d after wrong answer, want 0", result.StreakCount)
	}
	if result.LifeRestored {
		t.Error("LifeRestored should be false on wrong answer")
	}
}

// TestMarathonGameV2_LifeRestoredAfterStreak5 verifies life regen at streak=5
func TestMarathonGameV2_LifeRestoredAfterStreak5(t *testing.T) {
	// Start with streak=4, lives=3 (not full) — answering correctly should trigger regen
	game, q := buildV2GameWithQuestion(t, 4, 3)

	result, err := game.AnswerQuestion(q.ID(), findCorrectAnswerID(q), 1000, 1000001)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.StreakCount != 5 {
		t.Errorf("StreakCount = %d, want 5", result.StreakCount)
	}
	if !result.LifeRestored {
		t.Error("LifeRestored should be true after reaching streak=5")
	}
	if result.RemainingLives != 4 {
		t.Errorf("RemainingLives = %d, want 4 (was 3, restored 1)", result.RemainingLives)
	}
}

// TestMarathonGameV2_NoLifeRestoredWhenFull verifies no over-regen at max lives
func TestMarathonGameV2_NoLifeRestoredWhenFull(t *testing.T) {
	// Start with streak=4, lives=5 (full max)
	game, q := buildV2GameWithQuestion(t, 4, MaxLives)

	result, err := game.AnswerQuestion(q.ID(), findCorrectAnswerID(q), 1000, 1000001)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.StreakCount != 5 {
		t.Errorf("StreakCount = %d, want 5", result.StreakCount)
	}
	if result.LifeRestored {
		t.Error("LifeRestored should be false when already at max lives")
	}
	if result.RemainingLives != MaxLives {
		t.Errorf("RemainingLives = %d, want %d (unchanged)", result.RemainingLives, MaxLives)
	}
}

// TestMarathonGameV2_ShieldResetsStreak verifies shield still resets streak
func TestMarathonGameV2_ShieldResetsStreak(t *testing.T) {
	// Start with streak=4, shield active
	game, q := buildV2GameWithQuestion(t, 4, 5)

	// Reconstruct with shield active
	gameWithShield := ReconstructMarathonGameV2(
		game.id, game.playerID, game.category, GameStatusInProgress,
		game.startedAt, 0, game.currentQuestion,
		game.answeredQuestionIDs, game.recentQuestionIDs,
		0, 0,
		game.lives, game.bonusInventory, game.difficulty,
		true,  // shieldActive = true
		0, nil, game.usedBonuses,
		4, 4, 0, // streak=4, bestStreak=4, livesRestored=0
	)

	result, err := gameWithShield.AnswerQuestion(q.ID(), findWrongAnswerID(q), 1000, 1000001)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Shield should protect life
	if result.LifeLost {
		t.Error("LifeLost should be false when shield was active")
	}
	if !result.ShieldConsumed {
		t.Error("ShieldConsumed should be true")
	}
	// But streak RESETS (wrong answer, even with shield)
	if result.StreakCount != 0 {
		t.Errorf("StreakCount = %d after shielded wrong answer, want 0 (shield resets streak)", result.StreakCount)
	}
}

// TestMarathonGameV2_StreakContinuesPast5 verifies streak continues counting after regen
func TestMarathonGameV2_StreakContinuesPast5(t *testing.T) {
	// streak=9, lives=3 — answering correctly makes streak=10, triggers 2nd regen
	game, q := buildV2GameWithQuestion(t, 9, 3)

	result, err := game.AnswerQuestion(q.ID(), findCorrectAnswerID(q), 1000, 1000001)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.StreakCount != 10 {
		t.Errorf("StreakCount = %d, want 10", result.StreakCount)
	}
	if !result.LifeRestored {
		t.Error("LifeRestored should be true at streak=10")
	}
}
