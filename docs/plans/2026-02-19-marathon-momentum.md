# Marathon Momentum Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Change marathon starting lives from 3 to 5 and add streak-based life regeneration (+1 every 5 correct answers in a row).

**Architecture:** Domain-first TDD. Changes flow down: domain constants → domain aggregate → application DTO → swagger models → TypeScript regen → frontend composable → frontend view. The `LivesSystem` value object is immutable; streak state lives in `MarathonGameV2` aggregate.

**Tech Stack:** Go (domain/application/handlers), Fiber v3, Vue 3 + TypeScript, Vite, swaggo/swag, kubb (TypeScript gen)

---

## Critical Context

### Existing file map
```
backend/internal/domain/solo_marathon/
  value_objects.go            ← MaxLives const + LivesSystem
  marathon_game_aggregate_v2.go ← MarathonGameV2 aggregate
  value_objects_test.go       ← LivesSystem tests (SOME WILL BREAK)
  marathon_game_aggregate_test.go ← V1 tests (helper fns shared)

backend/internal/application/marathon/
  dto.go                      ← SubmitMarathonAnswerOutput
  submit_marathon_answer.go   ← SubmitMarathonAnswerUseCase.Execute

backend/internal/infrastructure/http/handlers/
  swagger_models.go           ← SubmitMarathonAnswerData (Swagger type)

tma/src/composables/useMarathon.ts   ← MarathonState + submitAnswer
tma/src/views/Marathon/MarathonPlayView.vue ← livesDisplay computed + template
```

### Key design decisions
- `MaxLives = 5` (was 3) — constant in `value_objects.go`
- `MarathonStreakForRegen = 5` — new constant (every 5 correct in a row → +1 life)
- Streak resets on **any** wrong answer, including Shield-protected ones
- Streak does NOT change on Skip bonus
- Life regen is silently skipped when already at max lives (streak counter continues)
- New fields on `AnswerQuestionResultV2`: `StreakCount int`, `LifeRestored bool`
- New fields on `MarathonGameV2` struct: `streakCount int`, `bestStreak int`, `livesRestored int`

---

## Task 1: Update MaxLives constant (3 → 5) + fix broken tests

**Files:**
- Modify: `backend/internal/domain/solo_marathon/value_objects.go:76-79`
- Modify: `backend/internal/domain/solo_marathon/value_objects_test.go`

### Step 1: Run existing tests (see what passes now)

```bash
cd /Users/barsukov/projects/quiz-sprint/backend
go test ./internal/domain/solo_marathon/... -v 2>&1 | tail -20
```
Expected: all PASS (baseline)

### Step 2: Write the new test for 5 starting lives

Add to end of `value_objects_test.go`:

```go
// TestLivesSystem_StartsWith5Lives verifies marathon starts with 5 lives
func TestLivesSystem_StartsWith5Lives(t *testing.T) {
	lives := NewLivesSystem(1000000)

	if lives.CurrentLives() != 5 {
		t.Errorf("Expected 5 starting lives, got %d", lives.CurrentLives())
	}
	if lives.MaxLives() != 5 {
		t.Errorf("Expected max 5 lives, got %d", lives.MaxLives())
	}
}
```

### Step 3: Run test to verify it fails

```bash
cd /Users/barsukov/projects/quiz-sprint/backend
go test ./internal/domain/solo_marathon/... -run TestLivesSystem_StartsWith5Lives -v
```
Expected: FAIL — "Expected 5 starting lives, got 3"

### Step 4: Change MaxLives constant

In `value_objects.go` lines 76-79, change:
```go
const (
	MaxLives            = 3
	LifeRegenInterval   = 4 * 60 * 60 // 4 hours in seconds
)
```
To:
```go
const (
	MaxLives            = 5
	LifeRegenInterval   = 4 * 60 * 60 // 4 hours in seconds
	MarathonStreakForRegen = 5 // correct answers in a row needed to restore 1 life
)
```

### Step 5: Fix broken tests in value_objects_test.go

These tests hardcode old MaxLives=3 values and will now fail. Fix them:

**Fix `TestLivesSystem_LoseLife` (around line 25):**
```go
// TestLivesSystem_LoseLife tests losing a life
func TestLivesSystem_LoseLife(t *testing.T) {
	now := int64(1000000)
	lives := NewLivesSystem(now)

	// Lose first life
	newLives := lives.LoseLife(now + 100)

	if newLives.CurrentLives() != MaxLives-1 {
		t.Errorf("Expected %d lives after losing one, got %d", MaxLives-1, newLives.CurrentLives())
	}
	if newLives.LastUpdate() != now+100 {
		t.Errorf("Expected last update %d, got %d", now+100, newLives.LastUpdate())
	}

	// Original should be unchanged (immutable)
	if lives.CurrentLives() != MaxLives {
		t.Errorf("Original lives should be unchanged at %d, got %d", MaxLives, lives.CurrentLives())
	}
}
```

**Fix `TestLivesSystem_RegenerateLives` (around line 81) — fix the "capped" test cases:**
```go
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
		{"2 hours passed", 2 * 60 * 60, 1},         // Not enough for 1 life
		{"4 hours passed", LifeRegenInterval, 2},    // Exactly 1 life
		{"8 hours passed", 2 * LifeRegenInterval, 3}, // 2 lives regened
		{"16 hours passed", 4 * LifeRegenInterval, 5}, // 4 lives regened, capped at MaxLives=5
		{"20 hours passed", 5 * LifeRegenInterval, 5}, // Capped at MaxLives=5
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
```

**Fix `TestLivesSystem_TimeToNextLife` (around line 139) — fix "Already at max" case:**
```go
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
			currentLives: MaxLives, // use constant, not hardcoded 3
			lastUpdate:   now,
			currentTime:  now,
			expectedTime: 0,
		},
		{
			name:         "Just lost a life",
			currentLives: MaxLives - 1,
			lastUpdate:   now,
			currentTime:  now,
			expectedTime: LifeRegenInterval,
		},
		{
			name:         "2 hours passed",
			currentLives: MaxLives - 1,
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
```

**Fix `TestLivesSystem_Label` (around line 206) — update labels to reflect MaxLives=5:**
```go
// TestLivesSystem_Label tests visual representation of lives
func TestLivesSystem_Label(t *testing.T) {
	tests := []struct {
		name     string
		lives    int
		expected string
	}{
		{"5 lives (full)", 5, "❤️❤️❤️❤️❤️"},
		{"3 lives", 3, "❤️❤️❤️🖤🖤"},
		{"1 life", 1, "❤️🖤🖤🖤🖤"},
		{"0 lives", 0, "🖤🖤🖤🖤🖤"},
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
```

### Step 6: Run all domain tests

```bash
cd /Users/barsukov/projects/quiz-sprint/backend
go test ./internal/domain/solo_marathon/... -v 2>&1 | grep -E "^(--- |PASS|FAIL|ok)"
```
Expected: all PASS

### Step 7: Commit

```bash
cd /Users/barsukov/projects/quiz-sprint
git add backend/internal/domain/solo_marathon/value_objects.go \
        backend/internal/domain/solo_marathon/value_objects_test.go
git commit -m "feat: increase MaxLives to 5, add MarathonStreakForRegen constant

Co-Authored-By: Claude Sonnet 4.6 <noreply@anthropic.com>"
```

---

## Task 2: Add streak fields to MarathonGameV2 + write failing tests

**Files:**
- Create: `backend/internal/domain/solo_marathon/marathon_game_aggregate_v2_test.go`
- Modify: `backend/internal/domain/solo_marathon/marathon_game_aggregate_v2.go`

### Step 1: Create test file with failing tests

Create `backend/internal/domain/solo_marathon/marathon_game_aggregate_v2_test.go`:

```go
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
		streakCount,      // NEW: streakCount
		streakCount,      // NEW: bestStreak (same for simplicity)
		0,                // NEW: livesRestored
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

	// Activate shield first
	gameWithShield := ReconstructMarathonGameV2(
		game.id, game.playerID, game.category, GameStatusInProgress,
		game.startedAt, 0, game.currentQuestion,
		game.answeredQuestionIDs, game.recentQuestionIDs,
		0, 0,
		game.lives, game.bonusInventory, game.difficulty,
		true,  // shieldActive = true
		0, nil, game.usedBonuses,
		4, 4, 0, // streak=4
	)

	result, err := gameWithShield.AnswerQuestion(q.ID(), findWrongAnswerID(q), 1000, 1000001)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Shield should protect life
	if result.LifeLost {
		t.Error("LifeLost should be false when shield was active")
	}
	if result.ShieldConsumed != true {
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
```

### Step 2: Run tests to verify they fail (compile errors first)

```bash
cd /Users/barsukov/projects/quiz-sprint/backend
go test ./internal/domain/solo_marathon/... -run TestMarathonGameV2_ -v 2>&1 | head -30
```
Expected: compile errors — `ReconstructMarathonGameV2 too many arguments` and `result.StreakCount undefined`

### Step 3: Add streak fields to MarathonGameV2 struct

In `marathon_game_aggregate_v2.go`, update the struct definition (around line 10-46):

Add 3 fields after `continueCount`:
```go
type MarathonGameV2 struct {
	id               GameID
	playerID         UserID
	category         MarathonCategory
	status           GameStatus
	startedAt        int64
	finishedAt       int64
	currentQuestion *quiz.Question
	answeredQuestionIDs []QuestionID
	recentQuestionIDs   []QuestionID
	score          int
	totalQuestions int
	lives          LivesSystem
	bonusInventory BonusInventory
	difficulty     DifficultyProgression
	shieldActive   bool
	continueCount int
	personalBestScore *int
	usedBonuses      map[QuestionID][]BonusType
	events []Event

	// Streak-based life regen (Marathon Momentum)
	streakCount   int // Current consecutive correct answers
	bestStreak    int // Best streak this session
	livesRestored int // Total times life was restored via streak
}
```

### Step 4: Update AnswerQuestionResultV2 to include new fields

In `marathon_game_aggregate_v2.go`, update the result struct (around line 148):
```go
type AnswerQuestionResultV2 struct {
	IsCorrect       bool
	CorrectAnswerID AnswerID
	TimeTaken       int64
	Score           int
	TotalQuestions  int
	DifficultyLevel DifficultyLevel
	LifeLost        bool
	ShieldConsumed  bool
	RemainingLives  int
	IsGameOver      bool
	StreakCount     int  // Current streak after this answer
	LifeRestored    bool // True if a life was restored by this answer's streak
	// Filled when IsGameOver = true
	GameOverData *GameOverData
}
```

### Step 5: Update NewMarathonGameV2 to initialize streak fields

In `NewMarathonGameV2` (around line 77-96), add to the game struct literal:
```go
game := &MarathonGameV2{
	// ... existing fields ...
	continueCount:       0,
	personalBestScore:   personalBestScore,
	usedBonuses:         make(map[QuestionID][]BonusType),
	events:              make([]Event, 0),
	// Streak
	streakCount:   0,
	bestStreak:    0,
	livesRestored: 0,
}
```

### Step 6: Update ReconstructMarathonGameV2 to include streak params

Update the function signature and body (around line 649):
```go
func ReconstructMarathonGameV2(
	id GameID,
	playerID UserID,
	category MarathonCategory,
	status GameStatus,
	startedAt int64,
	finishedAt int64,
	currentQuestion *quiz.Question,
	answeredQuestionIDs []QuestionID,
	recentQuestionIDs []QuestionID,
	score int,
	totalQuestions int,
	lives LivesSystem,
	bonusInventory BonusInventory,
	difficulty DifficultyProgression,
	shieldActive bool,
	continueCount int,
	personalBestScore *int,
	usedBonuses map[QuestionID][]BonusType,
	streakCount int,   // NEW
	bestStreak int,    // NEW
	livesRestored int, // NEW
) *MarathonGameV2 {
	return &MarathonGameV2{
		id:                  id,
		playerID:            playerID,
		category:            category,
		status:              status,
		startedAt:           startedAt,
		finishedAt:          finishedAt,
		currentQuestion:     currentQuestion,
		answeredQuestionIDs: answeredQuestionIDs,
		recentQuestionIDs:   recentQuestionIDs,
		score:               score,
		totalQuestions:       totalQuestions,
		lives:               lives,
		bonusInventory:      bonusInventory,
		difficulty:          difficulty,
		shieldActive:        shieldActive,
		continueCount:       continueCount,
		personalBestScore:   personalBestScore,
		usedBonuses:         usedBonuses,
		events:              make([]Event, 0),
		streakCount:         streakCount,   // NEW
		bestStreak:          bestStreak,    // NEW
		livesRestored:       livesRestored, // NEW
	}
}
```

### Step 7: Fix callers of ReconstructMarathonGameV2

Search for all callers:
```bash
cd /Users/barsukov/projects/quiz-sprint/backend
grep -rn "ReconstructMarathonGameV2" --include="*.go"
```

For each caller found, add the 3 trailing `0, 0, 0` arguments:
```go
ReconstructMarathonGameV2(
    // ... existing args ...
    usedBonuses,
    0, // streakCount
    0, // bestStreak
    0, // livesRestored
)
```

### Step 8: Run tests again (should compile, still fail on logic)

```bash
cd /Users/barsukov/projects/quiz-sprint/backend
go test ./internal/domain/solo_marathon/... -run TestMarathonGameV2_ -v 2>&1 | head -30
```
Expected: FAIL — StreakCount=0 instead of 1 (logic not implemented yet)

---

## Task 3: Implement streak tracking logic in AnswerQuestion

**Files:**
- Modify: `backend/internal/domain/solo_marathon/marathon_game_aggregate_v2.go`

### Step 1: Update the correct answer branch in AnswerQuestion

Find the correct answer section (around line 246 `if isCorrect {`). After incrementing `mg.score`, add streak logic:

```go
if isCorrect {
	// === CORRECT ANSWER ===

	// a. Increment score
	mg.score++
	result.Score = mg.score

	// b. Shield deactivates after question
	mg.shieldActive = false

	// c. Update difficulty
	questionIndex := mg.totalQuestions + 1
	previousLevel := mg.difficulty.Level()
	mg.difficulty = mg.difficulty.UpdateFromQuestionIndex(questionIndex)

	if mg.difficulty.Level() != previousLevel {
		mg.events = append(mg.events, NewDifficultyIncreasedEvent(
			mg.id, mg.playerID, mg.difficulty.Level(), questionIndex, answeredAt,
		))
	}
	result.DifficultyLevel = mg.difficulty.Level()

	// d. Streak: increment counter
	mg.streakCount++
	if mg.streakCount > mg.bestStreak {
		mg.bestStreak = mg.streakCount
	}

	// e. Check for life regen (every MarathonStreakForRegen correct answers)
	if mg.streakCount%MarathonStreakForRegen == 0 && mg.lives.CurrentLives() < mg.lives.MaxLives() {
		mg.lives = mg.lives.AddLives(1, answeredAt)
		mg.livesRestored++
		result.LifeRestored = true
	}

} else {
```

### Step 2: Update the incorrect answer branch to reset streak

In the `else` branch (wrong answer), after the existing shield check logic, reset the streak. The streak reset happens **regardless** of shield:

```go
} else {
	// === INCORRECT ANSWER ===

	// Streak always resets on wrong answer (even if shield protects life)
	mg.streakCount = 0

	if wasShieldActive {
		// Shield protects from life loss
		shieldConsumed = true
		mg.shieldActive = false
		result.ShieldConsumed = true
	} else {
		// No shield — lose life
		mg.lives = mg.lives.LoseLife(answeredAt)
		result.LifeLost = true
		result.RemainingLives = mg.lives.CurrentLives()

		// Publish LifeLost event
		mg.events = append(mg.events, NewLifeLostEvent(
			mg.id, mg.playerID, questionID,
			mg.lives.CurrentLives(), answeredAt,
		))

		// Check if game over
		if !mg.lives.HasLives() {
			// ... existing game over logic (unchanged) ...
		}
	}
}
```

### Step 3: Set StreakCount on result before returning

After the if/else block and before `mg.totalQuestions++`, add:
```go
// Set streak state on result
result.StreakCount = mg.streakCount
result.RemainingLives = mg.lives.CurrentLives()
```

**Note:** `result.RemainingLives` is set in multiple places; ensure the final value after regen is used. The easiest fix is to set it just before the return, overriding any earlier assignment:

Find the last line before `return result, nil` (around line 358) and add before it:
```go
result.RemainingLives = mg.lives.CurrentLives()
result.StreakCount = mg.streakCount
```

### Step 4: Run tests to verify all pass

```bash
cd /Users/barsukov/projects/quiz-sprint/backend
go test ./internal/domain/solo_marathon/... -run TestMarathonGameV2_ -v
```
Expected: all 6 tests PASS

### Step 5: Run full domain test suite

```bash
cd /Users/barsukov/projects/quiz-sprint/backend
go test ./internal/domain/solo_marathon/... -v 2>&1 | grep -E "^(--- |PASS|FAIL|ok)"
```
Expected: all PASS

### Step 6: Commit

```bash
cd /Users/barsukov/projects/quiz-sprint
git add backend/internal/domain/solo_marathon/marathon_game_aggregate_v2.go \
        backend/internal/domain/solo_marathon/marathon_game_aggregate_v2_test.go
git commit -m "feat: add streak-based life regen to MarathonGameV2

+1 life every 5 correct in a row (max 5 lives). Shield resets streak.

Co-Authored-By: Claude Sonnet 4.6 <noreply@anthropic.com>"
```

---

## Task 4: Update application layer DTOs + use case

**Files:**
- Modify: `backend/internal/application/marathon/dto.go:128-145`
- Modify: `backend/internal/application/marathon/submit_marathon_answer.go:92-106`

### Step 1: Add new fields to SubmitMarathonAnswerOutput

In `dto.go`, update `SubmitMarathonAnswerOutput` (around line 128):
```go
type SubmitMarathonAnswerOutput struct {
	IsCorrect       bool              `json:"isCorrect"`
	CorrectAnswerID string            `json:"correctAnswerId"`
	TimeTaken       int64             `json:"timeTaken"`
	Score           int               `json:"score"`
	TotalQuestions  int               `json:"totalQuestions"`
	DifficultyLevel string            `json:"difficultyLevel"`
	LifeLost        bool              `json:"lifeLost"`
	ShieldConsumed  bool              `json:"shieldConsumed"`
	Lives           LivesDTO          `json:"lives"`
	BonusInventory  BonusInventoryDTO `json:"bonusInventory"`
	IsGameOver      bool              `json:"isGameOver"`
	NextQuestion    *QuestionDTO      `json:"nextQuestion,omitempty"`
	NextTimeLimit   *int              `json:"nextTimeLimit,omitempty"`
	GameOverResult  *GameOverResultDTO `json:"gameOverResult,omitempty"`
	Milestone       *MilestoneDTO     `json:"milestone,omitempty"`
	StreakCount     int               `json:"streakCount"`   // NEW: current streak after this answer
	LifeRestored    bool              `json:"lifeRestored"`  // NEW: true if streak triggered life regen
}
```

### Step 2: Map new fields in use case

In `submit_marathon_answer.go`, update the output building block (around line 92-106):
```go
output := SubmitMarathonAnswerOutput{
	IsCorrect:       result.IsCorrect,
	CorrectAnswerID: result.CorrectAnswerID.String(),
	TimeTaken:       result.TimeTaken,
	Score:           result.Score,
	TotalQuestions:  result.TotalQuestions,
	DifficultyLevel: string(result.DifficultyLevel),
	LifeLost:        result.LifeLost,
	ShieldConsumed:  result.ShieldConsumed,
	Lives:           ToLivesDTO(game.Lives(), now),
	BonusInventory:  ToBonusInventoryDTO(game.BonusInventory()),
	IsGameOver:      result.IsGameOver,
	Milestone:       ToMilestoneDTO(result.Score),
	StreakCount:     result.StreakCount,   // NEW
	LifeRestored:    result.LifeRestored, // NEW
}
```

### Step 3: Run application tests

```bash
cd /Users/barsukov/projects/quiz-sprint/backend
go test ./internal/application/marathon/... -v 2>&1 | grep -E "^(--- |PASS|FAIL|ok)"
```
Expected: all PASS (existing tests should still work)

### Step 4: Run full backend build check

```bash
cd /Users/barsukov/projects/quiz-sprint/backend
go build ./...
```
Expected: no errors

### Step 5: Commit

```bash
cd /Users/barsukov/projects/quiz-sprint
git add backend/internal/application/marathon/dto.go \
        backend/internal/application/marathon/submit_marathon_answer.go
git commit -m "feat: propagate streak fields through application layer

SubmitMarathonAnswerOutput now includes streakCount and lifeRestored.

Co-Authored-By: Claude Sonnet 4.6 <noreply@anthropic.com>"
```

---

## Task 5: Update Swagger models + regenerate TypeScript types

**Files:**
- Modify: `backend/internal/infrastructure/http/handlers/swagger_models.go:646-665`
- Run: `pnpm run generate:all` in `tma/`

### Step 1: Add new fields to SubmitMarathonAnswerData

In `swagger_models.go`, update `SubmitMarathonAnswerData` (around line 646):
```go
// SubmitMarathonAnswerData contains answer submission result
type SubmitMarathonAnswerData struct {
	IsCorrect       bool                         `json:"isCorrect" validate:"required"`
	CorrectAnswerID string                       `json:"correctAnswerId" validate:"required"`
	TimeTaken       int64                        `json:"timeTaken" validate:"required"`
	Score           int                          `json:"score" validate:"required"`
	TotalQuestions  int                          `json:"totalQuestions" validate:"required"`
	DifficultyLevel string                       `json:"difficultyLevel" validate:"required"`
	LifeLost        bool                         `json:"lifeLost" validate:"required"`
	ShieldConsumed  bool                         `json:"shieldConsumed" validate:"required"`
	Lives           MarathonLivesDTO             `json:"lives" validate:"required"`
	BonusInventory  MarathonBonusInventoryDTO    `json:"bonusInventory" validate:"required"`
	IsGameOver      bool                         `json:"isGameOver" validate:"required"`
	NextQuestion    *QuestionDTO                 `json:"nextQuestion,omitempty"`
	NextTimeLimit   *int                         `json:"nextTimeLimit,omitempty"`
	GameOverResult  *MarathonGameOverResultDTO   `json:"gameOverResult,omitempty"`
	Milestone       *MarathonMilestoneDTO        `json:"milestone,omitempty"`
	StreakCount     int                          `json:"streakCount" validate:"required"`   // NEW
	LifeRestored    bool                         `json:"lifeRestored" validate:"required"`  // NEW
}
```

### Step 2: Build backend to confirm swagger models compile

```bash
cd /Users/barsukov/projects/quiz-sprint/backend
go build ./...
```
Expected: no errors

### Step 3: Regenerate Swagger + TypeScript types

```bash
cd /Users/barsukov/projects/quiz-sprint/tma
pnpm run generate:all
```
Expected: swagger.json updated, TypeScript types regenerated in `src/api/generated/`

Verify new fields appear in generated types:
```bash
grep -n "streakCount\|lifeRestored" /Users/barsukov/projects/quiz-sprint/tma/src/api/generated/types/marathonController/*.ts 2>/dev/null || \
grep -rn "streakCount\|lifeRestored" /Users/barsukov/projects/quiz-sprint/tma/src/api/generated/ 2>/dev/null | head -5
```
Expected: lines showing `streakCount` and `lifeRestored` in generated types

### Step 4: Commit

```bash
cd /Users/barsukov/projects/quiz-sprint
git add backend/internal/infrastructure/http/handlers/swagger_models.go \
        tma/src/api/generated/
git commit -m "feat: add streakCount+lifeRestored to swagger models, regen TypeScript

Co-Authored-By: Claude Sonnet 4.6 <noreply@anthropic.com>"
```

---

## Task 6: Update frontend composable

**Files:**
- Modify: `tma/src/composables/useMarathon.ts`

### Step 1: Add streak to MarathonState interface

Find the `interface MarathonState` (around line 25) and add two fields:
```ts
interface MarathonState {
  status: MarathonStatus
  game: InternalInfrastructureHttpHandlersMarathonGameDTO | null
  currentQuestion: InternalInfrastructureHttpHandlersQuestionDTO | null
  bonusInventory: InternalInfrastructureHttpHandlersMarathonBonusInventoryDTO
  score: number
  totalQuestions: number
  timeLimit: number
  shieldActive: boolean
  personalBest: number | null
  categoryId: string | null
  lastAnswerResult: InternalInfrastructureHttpHandlersSubmitMarathonAnswerData | null
  gameOverResult: InternalInfrastructureHttpHandlersMarathonGameOverResultDTO | null
  milestone: InternalInfrastructureHttpHandlersMarathonMilestoneDTO | null
  hiddenAnswerIds: string[]
  streakCount: number      // NEW: current streak
  lifeRestoredSignal: number // NEW: increments when life is restored (triggers animation)
}
```

### Step 2: Initialize new state fields

Find `const state = ref<MarathonState>({` (around line 54) and add:
```ts
const state = ref<MarathonState>({
  // ... existing fields ...
  hiddenAnswerIds: [],
  streakCount: 0,         // NEW
  lifeRestoredSignal: 0,  // NEW
})
```

### Step 3: Update submitAnswer to extract streak data

Find the `submitAnswer` function (around line 202). After `state.value.lastAnswerResult = answerData`, add:

```ts
// Update streak state
state.value.streakCount = answerData.streakCount ?? 0

// Signal life restoration for animation (increment = new event)
if (answerData.lifeRestored) {
  state.value.lifeRestoredSignal++
}
```

### Step 4: Reset streak on game start and continue

In `startGame` function (around line 163), after `state.value.lastAnswerResult = null`, add:
```ts
state.value.streakCount = 0
state.value.lifeRestoredSignal = 0
```

In `continueGame` function (after continue response), add:
```ts
state.value.streakCount = 0
```

### Step 5: Export streak state

Find the `return` statement at the end of `useMarathon` and add the new computed/state exports:
```ts
return {
  // ... existing exports ...
  streakCount: computed(() => state.value.streakCount),
  lifeRestoredSignal: computed(() => state.value.lifeRestoredSignal),
}
```

### Step 6: Build frontend to check for type errors

```bash
cd /Users/barsukov/projects/quiz-sprint/tma
pnpm build 2>&1 | tail -20
```
Expected: no TypeScript errors

### Step 7: Commit

```bash
cd /Users/barsukov/projects/quiz-sprint
git add tma/src/composables/useMarathon.ts
git commit -m "feat: track streakCount and lifeRestored signal in marathon composable

Co-Authored-By: Claude Sonnet 4.6 <noreply@anthropic.com>"
```

---

## Task 7: Update frontend play view

**Files:**
- Modify: `tma/src/views/Marathon/MarathonPlayView.vue`

### Step 1: Import new composable exports

In the `<script setup>` section, find where `useMarathon` is destructured (around line 23):
```ts
const {
  state,
  isPlaying,
  lives,
  canUseShield,
  canUseFiftyFifty,
  canUseSkip,
  canUseFreeze,
  submitAnswer,
  applyAnswerResult,
  useBonus,
  initialize,
  streakCount,        // NEW
  lifeRestoredSignal, // NEW
} = useMarathon(playerId)
```

### Step 2: Add local state for life-restored animation

After the existing local state declarations (around line 41), add:
```ts
const showLifeRestoredAnim = ref(false)
```

### Step 3: Watch lifeRestoredSignal to trigger animation

Add a watch after the local state declarations:
```ts
import { ref, computed, onMounted, onUnmounted, nextTick, watch } from 'vue'

// ... existing code ...

// Trigger ❤️+1 animation when life is restored
watch(lifeRestoredSignal, () => {
  showLifeRestoredAnim.value = true
  setTimeout(() => {
    showLifeRestoredAnim.value = false
  }, 1200)
})
```

### Step 4: Update template header — add streak counter + life animation

Find the header section in `<template>` (around line 277):

```html
<!-- Header: lives + score + timer -->
<div class="flex items-center gap-3">
  <!-- Lives -->
  <div class="flex gap-0.5 shrink-0 relative">
    <UIcon
      v-for="(filled, index) in livesDisplay"
      :key="index"
      :name="filled ? 'i-heroicons-heart-solid' : 'i-heroicons-heart'"
      :class="filled ? 'text-red-500' : 'text-gray-300 dark:text-gray-600'"
      class="size-4"
    />
    <!-- ❤️+1 animation when life is restored -->
    <Transition name="life-restore">
      <span
        v-if="showLifeRestoredAnim"
        class="absolute -top-5 left-0 text-xs font-bold text-green-500 pointer-events-none select-none"
      >
        ❤️+1
      </span>
    </Transition>
  </div>

  <!-- Score -->
  <span class="shrink-0 text-sm font-semibold text-primary tabular-nums">
    {{ state.score }}
  </span>

  <!-- Streak counter (only show when streak >= 2) -->
  <span
    v-if="streakCount >= 2"
    class="shrink-0 text-xs font-medium text-orange-500 tabular-nums"
  >
    🔥 {{ streakCount }}
  </span>

  <!-- Shield indicator -->
  <!-- ... existing shield icon ... -->
```

### Step 5: Add CSS transition for life-restore animation

At the bottom of the `<template>` section (or in a `<style>` block), add:
```html
<style scoped>
.life-restore-enter-active {
  animation: life-restore-pop 1.2s ease-out forwards;
}
@keyframes life-restore-pop {
  0%   { opacity: 0; transform: translateY(0); }
  20%  { opacity: 1; transform: translateY(-8px); }
  80%  { opacity: 1; transform: translateY(-14px); }
  100% { opacity: 0; transform: translateY(-20px); }
}
</style>
```

### Step 6: Build and type-check

```bash
cd /Users/barsukov/projects/quiz-sprint/tma
pnpm build 2>&1 | tail -20
```
Expected: no errors

### Step 7: Commit

```bash
cd /Users/barsukov/projects/quiz-sprint
git add tma/src/views/Marathon/MarathonPlayView.vue
git commit -m "feat: add streak counter and life-restored animation to marathon play view

Shows 🔥 N when streak >= 2. Shows ❤️+1 popup animation on life regen.

Co-Authored-By: Claude Sonnet 4.6 <noreply@anthropic.com>"
```

---

## Final Verification

### Run all backend tests

```bash
cd /Users/barsukov/projects/quiz-sprint/backend
go test ./... 2>&1 | grep -E "^(ok|FAIL|---)"
```
Expected: all `ok`, no `FAIL`

### Run frontend lint

```bash
cd /Users/barsukov/projects/quiz-sprint/tma
pnpm lint
```
Expected: no errors

### Manual smoke test checklist

Start backend:
```bash
cd /Users/barsukov/projects/quiz-sprint/backend
docker compose -f docker-compose.dev.yml up -d
```

Verify via API:
1. `POST /api/v1/marathon/start` → `game.lives.currentLives` should be **5** (not 3)
2. `POST /api/v1/marathon/{id}/answer` (correct) → `streakCount: 1`, `lifeRestored: false`
3. After 5 correct answers → `streakCount: 5`, `lifeRestored: true` (if lives < 5)
4. After wrong answer → `streakCount: 0`

### Push to remote

```bash
cd /Users/barsukov/projects/quiz-sprint
git push origin marathon-update
```
