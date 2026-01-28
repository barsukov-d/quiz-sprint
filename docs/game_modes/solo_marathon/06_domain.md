# Solo Marathon - Domain Model

## Bounded Context
`solo_marathon` - PvE endless mode with lives and bonuses.

---

## Aggregates

### MarathonGame (Root)

**Purpose:** Player's marathon run with lives and bonus management.

```go
type MarathonGame struct {
    id              GameID
    playerID        UserID
    status          GameStatus          // in_progress, game_over, completed

    session         *kernel.QuizGameplaySession  // Question flow
    livesSystem     LivesSystem         // 3 lives management
    bonusInventory  BonusInventory      // Available bonuses

    continueCount   int                 // Times continued
    personalBest    int                 // Player's record

    events          []Event
}
```

**Invariants:**
- ✅ Lives: 0 ≤ lives ≤ 3
- ✅ Game over when lives == 0
- ✅ Cannot use bonus if inventory == 0
- ✅ Score = correct answers count

**Factory:**
```go
func NewMarathonGame(
    playerID UserID,
    personalBest int,
    bonuses BonusInventory,
    questions *quiz.Quiz,
    startedAt int64,
) (*MarathonGame, error)
```

**Key Methods:**
```go
func (mg *MarathonGame) AnswerQuestion(
    questionID QuestionID,
    answerID AnswerID,
    timeTaken int64,
    shieldActive bool,
    answeredAt int64,
) (*AnswerResult, error)

func (mg *MarathonGame) UseBonus(
    bonusType BonusType,
    questionID QuestionID,
) error

func (mg *MarathonGame) Continue(
    paymentMethod PaymentMethod,
) error

func (mg *MarathonGame) IsGameOver() bool
func (mg *MarathonGame) GetScore() int
```

**Repository:**
```go
type MarathonGameRepository interface {
    GetByID(id GameID) (*MarathonGame, error)
    GetActiveByPlayer(playerID UserID) (*MarathonGame, error)
    Save(mg *MarathonGame) error
}
```

---

## Value Objects

### LivesSystem

```go
type LivesSystem struct {
    current int  // 0-3
    max     int  // Always 3
}

func NewLivesSystem() LivesSystem {
    return LivesSystem{current: 3, max: 3}
}

func (ls LivesSystem) LoseLife() LivesSystem {
    if ls.current > 0 {
        return LivesSystem{current: ls.current - 1, max: ls.max}
    }
    return ls
}

func (ls LivesSystem) Reset() LivesSystem {
    return LivesSystem{current: 1, max: ls.max}  // Continue gives 1 life
}

func (ls LivesSystem) IsGameOver() bool {
    return ls.current == 0
}

func (ls LivesSystem) Current() int {
    return ls.current
}

func (ls LivesSystem) Label() string {
    hearts := strings.Repeat("❤️", ls.current)
    lost := strings.Repeat("🖤", ls.max - ls.current)
    return hearts + lost
}
```

**Immutable:** All methods return NEW instance.

---

### BonusInventory

```go
type BonusInventory struct {
    shield      int
    fiftyFifty  int
    skip        int
    freeze      int
}

func NewBonusInventory(shield, fiftyFifty, skip, freeze int) BonusInventory {
    return BonusInventory{
        shield:     max(0, shield),
        fiftyFifty: max(0, fiftyFifty),
        skip:       max(0, skip),
        freeze:     max(0, freeze),
    }
}

func (bi BonusInventory) Use(bonusType BonusType) (BonusInventory, error) {
    switch bonusType {
    case BonusShield:
        if bi.shield <= 0 {
            return bi, ErrInsufficientBonuses
        }
        return BonusInventory{...bi, shield: bi.shield - 1}, nil
    // ... similar for other types
    }
}

func (bi BonusInventory) Has(bonusType BonusType) bool {
    switch bonusType {
    case BonusShield:
        return bi.shield > 0
    // ... similar for other types
    }
}

func (bi BonusInventory) Count(bonusType BonusType) int {
    // Returns quantity for specific bonus
}
```

---

### BonusType

```go
type BonusType string

const (
    BonusShield     BonusType = "shield"
    BonusFiftyFifty BonusType = "fifty_fifty"
    BonusSkip       BonusType = "skip"
    BonusFreeze     BonusType = "freeze"
)
```

---

### GameStatus

```go
type GameStatus string

const (
    GameStatusInProgress GameStatus = "in_progress"
    GameStatusGameOver   GameStatus = "game_over"
    GameStatusCompleted  GameStatus = "completed"
)
```

---

### PaymentMethod

```go
type PaymentMethod string

const (
    PaymentCoins PaymentMethod = "coins"
    PaymentAd    PaymentMethod = "ad"
)
```

---

## Domain Services

### DifficultyCalculator

```go
type DifficultyCalculator struct{}

func (dc *DifficultyCalculator) GetTimeLimit(questionIndex int) int {
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

func (dc *DifficultyCalculator) SelectDifficulty(questionIndex int) string {
    // Returns "easy", "medium", "hard"
}
```

---

### ContinueCostCalculator

```go
type ContinueCostCalculator struct{}

func (ccc *ContinueCostCalculator) GetCost(continueCount int) int {
    return 200 + (continueCount * 200)
}

func (ccc *ContinueCostCalculator) HasAdOption(continueCount int) bool {
    return continueCount < 3  // No ads after 3rd continue
}
```

---

### PersonalBestTracker

```go
type PersonalBestTracker struct {
    repo PersonalBestRepository
}

func (pbt *PersonalBestTracker) UpdateIfBetter(
    playerID UserID,
    score int,
    gameID GameID,
) (isNewRecord bool, bonus int, error) {
    current := pbt.repo.GetPersonalBest(playerID)

    if score > current {
        pbt.repo.Save(playerID, score, gameID)
        return true, 500, nil  // 500 coin bonus
    }

    return false, 0, nil
}
```

---

## Domain Events

```go
type MarathonGameStartedEvent struct {
    GameID         GameID
    PlayerID       UserID
    PersonalBest   int
    InitialBonuses BonusInventory
    Timestamp      int64
}

type MarathonQuestionAnsweredEvent struct {
    GameID         GameID
    PlayerID       UserID
    QuestionID     QuestionID
    AnswerID       AnswerID
    IsCorrect      bool
    ShieldActive   bool
    ShieldConsumed bool
    LivesRemaining int
    CurrentScore   int
    Timestamp      int64
}

type MarathonLifeLostEvent struct {
    GameID         GameID
    PlayerID       UserID
    LivesRemaining int
    CurrentScore   int
    Timestamp      int64
}

type MarathonBonusUsedEvent struct {
    GameID     GameID
    PlayerID   UserID
    BonusType  BonusType
    QuestionID QuestionID
    Timestamp  int64
}

type MarathonGameOverEvent struct {
    GameID          GameID
    PlayerID        UserID
    FinalScore      int
    TotalQuestions  int
    PersonalBest    int
    IsNewRecord     bool
    ContinueCount   int
    BonusesUsed     map[BonusType]int
    Timestamp       int64
}

type MarathonContinueUsedEvent struct {
    GameID        GameID
    PlayerID      UserID
    ContinueCount int
    PaymentMethod PaymentMethod
    CostCoins     int
    Timestamp     int64
}

type MarathonNewRecordEvent struct {
    GameID       GameID
    PlayerID     UserID
    NewRecord    int
    OldRecord    int
    BonusCoins   int
    Timestamp    int64
}
```

---

## Integration with Other Contexts

### User Context
```
MarathonGame --[PlayerID]--> user.User
Continue payment → user.Inventory (coins)
Rewards → user.Inventory
```

### Quiz Context
```
MarathonGame uses quiz.Quiz via kernel.QuizGameplaySession
Questions filtered by difficulty
```

### Daily Challenge Context
```
Daily Chest → BonusInventory (source of bonuses)
```

---

## Database Schema

### Table: marathon_games
```sql
CREATE TABLE marathon_games (
    id VARCHAR(36) PRIMARY KEY,
    player_id VARCHAR(36) NOT NULL,
    status VARCHAR(20) NOT NULL,

    -- Session data
    session_data JSONB NOT NULL,

    -- Lives & bonuses
    lives_remaining INT DEFAULT 3,
    bonus_inventory JSONB NOT NULL,

    -- Progress
    current_score INT DEFAULT 0,
    total_questions INT DEFAULT 0,
    continue_count INT DEFAULT 0,

    -- Metadata
    personal_best INT,
    started_at TIMESTAMP NOT NULL,
    completed_at TIMESTAMP,

    INDEX idx_player_active (player_id, status),
    INDEX idx_completed (completed_at DESC)
);
```

### Table: marathon_personal_best
```sql
CREATE TABLE marathon_personal_best (
    player_id VARCHAR(36) PRIMARY KEY,
    best_score INT NOT NULL,
    game_id VARCHAR(36),
    achieved_at TIMESTAMP NOT NULL,

    INDEX idx_best_score (best_score DESC)
);
```

### Table: marathon_bonus_usage
```sql
CREATE TABLE marathon_bonus_usage (
    id VARCHAR(36) PRIMARY KEY,
    game_id VARCHAR(36),
    player_id VARCHAR(36),
    bonus_type VARCHAR(20),
    question_id VARCHAR(36),
    used_at TIMESTAMP,

    INDEX idx_game (game_id)
);
```

---

## Redis Structures

### Weekly Leaderboard
```
Key: marathon:leaderboard:weekly:{week_id}
Type: Sorted Set

Score: correctAnswers * 1000000 - totalQuestions
Member: playerID:gameID

ZADD marathon:leaderboard:weekly:2026-W04 87000087 "user_123:mg_abc"
ZREVRANK marathon:leaderboard:weekly:2026-W04 "user_123:mg_abc"
ZREVRANGE marathon:leaderboard:weekly:2026-W04 0 99 WITHSCORES
```

### All-Time Leaderboard
```
Key: marathon:leaderboard:alltime
Type: Sorted Set

Score: bestScore * 1000000 - totalQuestions
Member: playerID

ZADD marathon:leaderboard:alltime 187000187 "user_123"
```

---

## Error Types

```go
var (
    ErrGameNotFound          = errors.New("game not found")
    ErrGameNotActive         = errors.New("game not active")
    ErrGameOver              = errors.New("game already over")
    ErrActiveGameExists      = errors.New("active game already exists")
    ErrInsufficientBonuses   = errors.New("insufficient bonuses")
    ErrInsufficientCoins     = errors.New("insufficient coins")
    ErrInvalidTimeTaken      = errors.New("invalid time taken")
    ErrQuestionNotFound      = errors.New("question not found")
    ErrAnswerNotFound        = errors.New("answer not found")
)
```

Mapped to HTTP in infrastructure layer.

---

## Shared Kernel Usage

Uses `kernel.QuizGameplaySession` for question flow:
- Question navigation
- Answer tracking
- Score calculation (base)

**MarathonGame adds:**
- Lives system
- Bonus mechanics
- Continue logic
- Adaptive difficulty
