# Daily Challenge - Domain Model

## Bounded Context
`daily_challenge` - independent context for daily competitions.

## Aggregates

### 1. DailyQuiz (Root)

**Purpose:** Single daily quiz shared by all players.

```go
type DailyQuiz struct {
    id       DailyQuizID     // "dq_20260128"
    date     Date            // 2026-01-28
    quizID   quiz.QuizID     // Reference to quiz content
    seed     int             // Deterministic selection
    isActive bool
}
```

**Invariants:**
- ✅ ONE active DailyQuiz per date
- ✅ Must have exactly 10 questions
- ✅ Seed ensures same questions for all players

**Repository:**
```go
type DailyQuizRepository interface {
    GetByDate(date Date) (*DailyQuiz, error)
    GetOrCreate(date Date) (*DailyQuiz, error)
    Save(dq *DailyQuiz) error
}
```

---

### 2. DailyGame (Root)

**Purpose:** Player's attempt at daily challenge.

```go
type DailyGame struct {
    id          GameID
    playerID    UserID
    dailyQuizID DailyQuizID
    date        Date
    status      GameStatus      // in_progress, completed, abandoned

    session     *kernel.QuizGameplaySession  // Composition
    streak      StreakSystem                 // Value Object
    rank        *int                         // Set by app layer

    events      []Event
}
```

**Invariants:**
- ✅ Max 1 free attempt per player per day
- ✅ Cannot modify after completion
- ✅ Time taken: 0 < t ≤ 15 sec
- ✅ Streak updated only on completion

**Factory Methods:**
```go
func NewDailyGame(
    playerID UserID,
    dailyQuizID DailyQuizID,
    date Date,
    quizAggregate *quiz.Quiz,
    currentStreak StreakSystem,
    startedAt int64,
) (*DailyGame, error)

func ReconstructDailyGame(...) *DailyGame  // DB loading
```

**Key Methods:**
```go
func (dg *DailyGame) AnswerQuestion(
    questionID QuestionID,
    answerID AnswerID,
    timeTaken int64,
    answeredAt int64,
) (*AnswerQuestionResult, error)

func (dg *DailyGame) GetFinalScore() int
func (dg *DailyGame) GetCorrectAnswersCount() int
func (dg *DailyGame) SetRank(rank int)
```

**Repository:**
```go
type DailyGameRepository interface {
    GetByID(id GameID) (*DailyGame, error)
    GetByPlayerAndDate(playerID UserID, date Date) (*DailyGame, error)
    HasPlayedToday(playerID UserID, date Date) (bool, error)
    Save(dg *DailyGame) error
}
```

---

## Value Objects

### StreakSystem
```go
type StreakSystem struct {
    currentStreak   int
    lastPlayedDate  Date
}

func NewStreakSystem(streak int, lastDate Date) StreakSystem
func (s StreakSystem) UpdateForDate(date Date) StreakSystem
func (s StreakSystem) GetBonus() float64
func (s StreakSystem) CurrentStreak() int
func (s StreakSystem) LastPlayedDate() Date
```

**Immutable:** All methods return NEW instance.

---

### ChestType
```go
type ChestType string

const (
    ChestTypeWooden ChestType = "wooden"
    ChestTypeSilver ChestType = "silver"
    ChestTypeGolden ChestType = "golden"
)

func DetermineChestType(correctAnswers int) ChestType
```

---

### GameStatus
```go
type GameStatus string

const (
    GameStatusInProgress GameStatus = "in_progress"
    GameStatusCompleted  GameStatus = "completed"
    GameStatusAbandoned  GameStatus = "abandoned"
)

func (gs GameStatus) CanTransitionTo(next GameStatus) bool
func (gs GameStatus) IsTerminal() bool
```

---

## Domain Services

### DailyQuizSelector

**Purpose:** Select questions for daily quiz.

```go
type DailyQuizSelector struct {
    quizRepo quiz.QuizRepository
}

func (s *DailyQuizSelector) SelectQuestionsForDate(
    date Date,
) (*quiz.Quiz, error) {
    // Uses date as seed for deterministic selection
    // All players get same questions
}
```

---

### ChestRewardCalculator

**Purpose:** Calculate chest rewards.

```go
type ChestRewardCalculator struct{}

func (c *ChestRewardCalculator) CalculateRewards(
    chestType ChestType,
    streak int,
    isPremium bool,
) ChestContents {
    // Base rewards by chest type
    // Apply streak multiplier
    // Apply premium upgrade if applicable
}
```

---

## Domain Events

```go
type DailyGameStartedEvent struct {
    GameID        GameID
    PlayerID      UserID
    DailyQuizID   DailyQuizID
    Date          Date
    CurrentStreak int
    Timestamp     int64
}

type DailyGameCompletedEvent struct {
    GameID         GameID
    PlayerID       UserID
    DailyQuizID    DailyQuizID
    Date           Date
    FinalScore     int
    CorrectAnswers int
    TotalQuestions int
    CurrentStreak  int
    StreakBonus    float64
    Rank           *int
    Timestamp      int64
}

type DailyQuestionAnsweredEvent struct {
    GameID     GameID
    PlayerID   UserID
    QuestionID QuestionID
    AnswerID   AnswerID
    TimeTaken  int64
    Timestamp  int64
}

type StreakMilestoneReachedEvent struct {
    GameID       GameID
    PlayerID     UserID
    Streak       int
    BonusPercent int
    Timestamp    int64
}

type ChestEarnedEvent struct {
    PlayerID   UserID
    GameID     GameID
    ChestType  ChestType
    Rewards    ChestContents
    StreakBonus float64
    Timestamp  int64
}
```

---

## Integration with Other Contexts

### User Context
```
DailyGame --[PlayerID]--> user.User
ChestEarnedEvent --> Update user.Inventory
```

### Quiz Context
```
DailyQuiz --[QuizID]--> quiz.Quiz (read-only)
DailyGame uses quiz.Quiz via kernel.QuizGameplaySession
```

### PvP/Marathon Contexts
```
ChestContents --> PvP Tickets --> pvp_duel.DuelGame
ChestContents --> Bonuses --> solo_marathon.MarathonGame
```

**Anti-Corruption Layer:**
- Daily Challenge does NOT import other game modes
- Integration via Domain Events + Application Layer

---

## Shared Kernel

Uses `kernel.QuizGameplaySession` for pure Q&A logic:
```go
type QuizGameplaySession struct {
    id              SessionID
    quiz            *quiz.Quiz
    currentIndex    int
    answers         []UserAnswer
    startedAt       int64
    finishedAt      *int64
    status          SessionStatus
}
```

**Benefits:**
- Reusable across all game modes
- No duplication of answer validation logic
- Consistent scoring calculation

---

## Database Schema

### Table: daily_quizzes
```sql
CREATE TABLE daily_quizzes (
    id VARCHAR(36) PRIMARY KEY,
    date DATE UNIQUE NOT NULL,
    quiz_id VARCHAR(36) NOT NULL,
    seed INTEGER NOT NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    INDEX idx_daily_quizzes_date (date DESC),
    INDEX idx_daily_quizzes_active (is_active, date)
);
```

### Table: daily_games
```sql
CREATE TABLE daily_games (
    id VARCHAR(36) PRIMARY KEY,
    player_id VARCHAR(36) NOT NULL,
    daily_quiz_id VARCHAR(36) NOT NULL,
    date DATE NOT NULL,
    status VARCHAR(20) NOT NULL,

    -- Session data (JSONB for flexibility)
    session_data JSONB NOT NULL,

    -- Streak
    current_streak INTEGER DEFAULT 0,
    last_played_date DATE,

    -- Results
    final_score INTEGER,
    correct_answers INTEGER,
    rank INTEGER,

    started_at TIMESTAMP NOT NULL,
    completed_at TIMESTAMP,

    FOREIGN KEY (daily_quiz_id) REFERENCES daily_quizzes(id),
    INDEX idx_daily_games_player_date (player_id, date),
    INDEX idx_daily_games_leaderboard (date, final_score DESC, completed_at),
    UNIQUE (player_id, date)
);
```

---

## Redis Structures

### Leaderboard
```
Key: daily:leaderboard:{date}
Type: Sorted Set

Score: finalScore * 1000000 - completedAtTimestamp
Member: playerID

Commands:
ZADD daily:leaderboard:2026-01-28 920000001706428800 user_123
ZREVRANK daily:leaderboard:2026-01-28 user_123
ZREVRANGE daily:leaderboard:2026-01-28 0 99 WITHSCORES
```

---

## Error Types

```go
var (
    ErrDailyQuizNotFound     = errors.New("daily quiz not found")
    ErrGameNotFound          = errors.New("game not found")
    ErrAlreadyPlayedToday    = errors.New("already played today")
    ErrGameAlreadyCompleted  = errors.New("game already completed")
    ErrGameNotActive         = errors.New("game not active")
    ErrInvalidGameStatus     = errors.New("invalid game status")
    ErrInvalidTimeTaken      = errors.New("invalid time taken")
    ErrInvalidDate           = errors.New("invalid date")
    ErrInvalidGameID         = errors.New("invalid game id")
    ErrInvalidDailyQuizID    = errors.New("invalid daily quiz id")
)
```

Mapped to HTTP in infrastructure layer.
