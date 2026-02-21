# Solo Marathon - Business Rules

## Lives System

### Initial State
```
lives = 5
```

### Life Loss Rules
```
Wrong answer && !shieldActive:
    lives -= 1

Wrong answer && shieldActive:
    lives -= 0
    shield consumed (was activated before answering)
```

### Game Over Condition
```
if lives == 0:
    status = GAME_OVER
    offer_continue()
```

### Marathon Momentum (Streak-based Life Regen)
```
streakForRegen = 5  // correct answers in a row

On correct answer:
    streakCount++
    if streakCount % streakForRegen == 0 && lives < maxLives:
        lives += 1  // restore 1 life (capped at maxLives)

On wrong answer:
    streakCount = 0  // always resets, even if shield saved the life
```

**Example:** 5 correct in a row → +1 life. 10 correct in a row → +2 lives total.

### Continue Mechanic
```go
func GetContinueCost(continueCount int) int {
    return 200 + (continueCount * 200)
    // 1st: 200, 2nd: 400, 3rd: 600, ...
}

func Continue(continueCount int) {
    lives = 1  // NOT +1, reset to 1
    status = IN_PROGRESS
}
```

**Max continues:** Unlimited (but escalating cost).

---

## Scoring System

### Base Score
```
score = correctAnswersCount
```

**NO time bonus** (unlike Daily Challenge).
**NO streak multiplier** (simplicity).

### Leaderboard Tiebreaker
Primary: `score DESC`
Secondary: `questionCount ASC` (fewer total questions = better efficiency)
Tertiary: `completedAt ASC` (earlier = better)

### Continue Flag
Games with continues:
- **Shown in leaderboard:** Yes (with asterisk *)
- **Separate leaderboard:** No (same pool)
- **UI indicator:** "Продолжений: 2"

**Philosophy:** Continues allowed, but tracked for transparency.

---

## Adaptive Difficulty

### Timer Calculation
```go
func GetTimeLimit(questionIndex int) int {
    switch {
    case questionIndex <= 10:
        return 15
    case questionIndex <= 25:
        return 12
    case questionIndex <= 50:
        return 10
    default:
        return 8
    }
}
```

### Question Difficulty Distribution
```go
func SelectDifficulty(questionIndex int) string {
    switch {
    case questionIndex <= 10:
        return weightedRandom(["easy": 0.8, "medium": 0.2])
    case questionIndex <= 30:
        return "medium"
    case questionIndex <= 50:
        return weightedRandom(["medium": 0.7, "hard": 0.3])
    default:
        return "hard"
    }
}
```

### Question Selection
- **No repeats** within same game
- **Random from pool** (filtered by difficulty)
- **Category variety:** No more than 3 consecutive questions from same category

---

## Bonus Mechanics

### Bonus Types & Effects

#### 🛡️ Shield
```
Activation: Before answering — bonus consumed immediately on activation
Effect: Next wrong answer doesn't cost life
Consumption: On activation (regardless of answer outcome)
Deactivation: After question ends (correct or wrong), shield deactivates
Carry-over: NO — shield does NOT carry to next question
Cooldown: None (can activate again immediately on next question)
```

**Logic:**
```go
func ActivateShield(questionID) {
    consumeBonus(BonusShield)  // consumed immediately on activation
    shieldActive = true
}

func AnswerQuestion(answerID) {
    isCorrect := validate(answerID)

    if !isCorrect && shieldActive {
        // Wrong answer but shield saves — bonus already consumed on activation
        // lives unchanged
    } else if !isCorrect {
        // Wrong answer, no shield
        lives -= 1
    }

    shieldActive = false  // deactivates after every question
}
```

#### 🔀 50/50
```
Activation: Before answering
Effect: Remove 2 wrong answers
Consumption: Immediately on use
```

**Logic:**
```go
func Use5050(questionID) []AnswerID {
    correctAnswer := getCorrectAnswer(questionID)
    wrongAnswers := getWrongAnswers(questionID) // 3 answers

    // Keep 1 random wrong + 1 correct
    keep := [correctAnswer, randomChoice(wrongAnswers)]

    consumeBonus(Bonus5050)
    return keep
}
```

#### ⏭️ Skip
```
Activation: Before answering
Effect: Skip to next question
Consumption: Immediately
Score: No increment (doesn't count as question)
Lives: No change
```

**Logic:**
```go
func SkipQuestion() {
    consumeBonus(BonusSkip)
    currentQuestionIndex++
    // totalQuestionsAsked++ (for efficiency metric)
    // correctAnswers unchanged
    // lives unchanged
}
```

#### ❄️ Freeze
```
Activation: During question (anytime)
Effect: +10 seconds to current timer
Consumption: Immediately
Stackable: Yes (can use multiple per question)
```

**Logic:**
```go
func UseFreeze() {
    timeRemaining += 10
    consumeBonus(BonusFreeze)
}
```

### Bonus Inventory

**Default allocation at game start:**
```
Shield:    2
50/50:     1
Skip:      0
Freeze:    3
```
Plus accumulated bonuses from player's wallet (earned from Daily Challenge chests).
Total = defaults + wallet.

**Source:** Default allocation on every run + earned from Daily Challenge chests.

**Storage:**
```sql
CREATE TABLE user_inventory (
    user_id VARCHAR(36),
    bonus_shield INT DEFAULT 0,
    bonus_fifty_fifty INT DEFAULT 0,
    bonus_skip INT DEFAULT 0,
    bonus_freeze INT DEFAULT 0
);
```

**Validation:**
```go
func UseBonus(bonusType BonusType) error {
    if inventory[bonusType] <= 0 {
        return ErrInsufficientBonuses
    }
    inventory[bonusType]--
    return nil
}
```

---

## Weekly Leaderboard

### Week Definition
```
Start: Monday 00:00 UTC
End: Sunday 23:59 UTC
```

### Reset Logic
```
On Monday 00:00 UTC:
1. Archive previous week's scores
2. Distribute rewards to top 100
3. Clear weekly leaderboard (Redis)
4. All players start fresh
```

### Ranking
```
Key: marathon:leaderboard:weekly:{week_id}
Score: correctAnswers * 1000000 - totalQuestions
Member: playerID

Higher score = Better rank
```

**Example:**
```
Player A: 87 correct, 87 total → 87000000 - 87 = 86999913
Player B: 87 correct, 90 total → 87000000 - 90 = 86999910
Rank: A > B (more efficient)
```

---

## All-Time Leaderboard

### Purpose
Hall of Fame (no rewards, pure prestige).

### Ranking
```
Primary: correctAnswers DESC
Secondary: totalQuestions ASC
Tertiary: completedAt ASC
```

**Stored in PostgreSQL** (permanent).

```sql
CREATE TABLE marathon_hall_of_fame (
    player_id VARCHAR(36),
    best_score INT,
    total_questions INT,
    continue_count INT,
    completed_at TIMESTAMP,
    INDEX idx_best_score (best_score DESC, total_questions ASC)
);
```

---

## Validations

### Time Taken
```
0 < timeTaken ≤ timeLimit
```

Violation: `ErrInvalidTimeTaken`

### Answer Once
Cannot change answer after submission.
Violation: `ErrQuestionAlreadyAnswered`

### Bonus Availability
```
if inventory[bonusType] <= 0:
    return ErrInsufficientBonuses
```

### Game State
```
if status != IN_PROGRESS:
    return ErrGameNotActive
```

---

## Personal Best Tracking

```sql
CREATE TABLE marathon_personal_best (
    player_id VARCHAR(36) PRIMARY KEY,
    best_score INT,
    achieved_at TIMESTAMP,
    game_id VARCHAR(36)
);
```

**Update logic:**
```go
func UpdatePersonalBest(playerID, score int) {
    current := getPersonalBest(playerID)
    if score > current {
        savePersonalBest(playerID, score)
        rewardCoins(playerID, 500) // New record bonus
    }
}
```

---

## Anti-Cheat

### Impossible Times
```
if timeTaken < 0.5:
    flagSuspiciousActivity(playerID, "too_fast")
```

### Question Skipping Pattern
```
if skipCount > 50:
    flagSuspiciousActivity(playerID, "excessive_skips")
```

### Continue Abuse
```
if continueCount > 10:
    flagSuspiciousActivity(playerID, "excessive_continues")
```

**Note:** Flags for review, not auto-ban.
