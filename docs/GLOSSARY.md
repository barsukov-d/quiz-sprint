# GLOSSARY - Ubiquitous Language

> **Principle:** One term = one concept. Avoid synonyms and ambiguity.

---

## Core Domain Concepts

### Quiz (Content)
**Aggregate Root:** `quiz.Quiz`
**Definition:** Set of questions (content), NOT the gameplay process.
**❌ Avoid:** Test, Questionnaire, Assessment

### Question
**Entity:** `quiz.Question` (part of Quiz aggregate)
**Definition:** Single question with 4 answers (1 correct).

### Answer
**Entity:** `quiz.Answer`
**Definition:** One answer option (text + isCorrect flag).
**Not confused with:** `UserAnswer` (player's submission).

### Game
**Definition:** Process of playing Quiz in specific mode.
**Types:** `DailyGame`, `MarathonGame`
**❌ Avoid:** Match, Round, Session (for modes)

### Session (Gameplay)
**Shared Kernel:** `kernel.QuizGameplaySession`
**Definition:** Pure Q&A logic (question navigation, answer tracking).
**Usage:** Composed into all Game aggregates.

### Category
**Aggregate Root:** `quiz.Category`
**Definition:** Thematic grouping (Geography, History, etc.)

---

## Bounded Context: Daily Challenge

### Aggregates

**DailyQuiz**
- `daily_challenge.DailyQuiz`
- Daily question set (10 questions, same for all players)
- One active per date (00:00 UTC refresh)
- Deterministic selection (seed-based)

**DailyGame**
- `daily_challenge.DailyGame`
- Individual player's attempt at daily quiz
- 1 free attempt/day
- Max 1 free + N paid retries

### Value Objects

**StreakSystem**
- `daily_challenge.StreakSystem`
- Consecutive days played
- Immutable (methods return new instance)
- Bonuses: 3d→1.1x, 7d→1.25x, 30d→1.5x

**ChestType**
- `daily_challenge.ChestType`
- Enum: `wooden`, `silver`, `golden`
- Determined by correct answers: 0-4/5-7/8-10

**GameStatus**
- `daily_challenge.GameStatus`
- Enum: `in_progress`, `completed`, `abandoned`

### Key Concepts

**Daily Chest**
- Main reward container
- 3 types: 🪵 Wooden / 🥈 Silver / 🏆 Golden
- Contains: Coins, PvP Tickets, Marathon Bonuses

**Daily Streak**
- Consecutive days played
- Multiplies chest rewards
- Recoverable (monetization)

**Second Attempt**
- Retry same day: 100 coins OR Rewarded Ad
- Best score counts for leaderboard

### API Routes
`/api/v1/daily/*`

### Database
Tables: `daily_quizzes`, `daily_games`

### Anti-patterns
❌ `DailyChallenge` (aggregate name)
❌ `DailySession` (use DailyGame)

---

## Bounded Context: Solo Marathon

### Aggregates

**MarathonGame**
- `solo_marathon.MarathonGame`
- PvE endless run until 0 lives
- Manages: lives, bonuses, score, continues

### Value Objects

**LivesSystem**
- `solo_marathon.LivesSystem`
- Start: 3 lives
- Wrong answer: -1 life
- Game over: 0 lives
- Immutable

**BonusInventory**
- `solo_marathon.BonusInventory`
- Tracks 4 bonus types + quantities
- Immutable (use deducts)

**BonusType**
- `solo_marathon.BonusType`
- Enum: `shield`, `fifty_fifty`, `skip`, `freeze`

**PaymentMethod**
- `solo_marathon.PaymentMethod`
- Enum: `coins`, `ad`

**GameStatus**
- `solo_marathon.GameStatus`
- Enum: `in_progress`, `game_over`, `completed`

### Key Concepts

**Lives**
- 3 ❤️❤️❤️ at start
- -1 per wrong answer (unless Shield active)
- 0 → Game Over
- ❌ Avoid: HP, Health, Hearts

**Bonuses** (4 types)

| Type | Icon | Effect |
|------|------|--------|
| Shield | 🛡️ | 1 free mistake (no life loss) |
| 50/50 | 🔀 | Remove 2 wrong answers |
| Skip | ⏭️ | Skip question (no penalty) |
| Freeze | ❄️ | +10 seconds to timer |

❌ Avoid: PowerUp, Boost, Help

**Score**
- Count of correct answers
- NO time bonus (unlike Daily Challenge)
- Tiebreaker: totalQuestions ASC, completedAt ASC

**Adaptive Difficulty**
- Timer: 15s → 12s → 10s → 8s
- Questions: easy → medium → hard

**Continue**
- At game over: +1 life (reset to 1, not +1)
- Cost: 200/400/600/800 coins OR Ad
- Unlimited continues (escalating cost)

**Weekly Leaderboard**
- Monday-Sunday UTC
- Top 100 get rewards
- Resets weekly

**All-Time Leaderboard**
- Hall of Fame (prestige only)
- No rewards

### API Routes
`/api/v1/marathon/*`

### Database
Tables: `marathon_games`, `marathon_personal_best`

### Anti-patterns
❌ `MarathonSession` (use MarathonGame)
❌ `SoloGame` (use MarathonGame)
❌ `MarathonRun` (use MarathonGame)

---

## Domain Services

### Daily Challenge

**DailyQuizSelector**
- Selects 10 questions for date (deterministic)

**ChestRewardCalculator**
- Calculates chest contents (coins, tickets, bonuses)

### Solo Marathon

**DifficultyCalculator**
- Returns time limit for question index
- Selects difficulty level

**ContinueCostCalculator**
- Calculates continue cost: `200 + (count * 200)`

**PersonalBestTracker**
- Updates personal best if score higher
- Awards 500 coin bonus

---

## Domain Events

### Daily Challenge Events

```go
DailyGameStartedEvent
DailyQuestionAnsweredEvent
DailyGameCompletedEvent
ChestOpenedEvent
StreakMilestoneReachedEvent
StreakBrokenEvent
```

### Solo Marathon Events

```go
MarathonGameStartedEvent
MarathonQuestionAnsweredEvent
MarathonLifeLostEvent
MarathonBonusUsedEvent
MarathonGameOverEvent
MarathonContinueUsedEvent
MarathonNewRecordEvent
```

**Naming:** Past tense + Event (e.g., `GameStartedEvent` not `StartGameEvent`)

---

## Value Object Patterns

All value objects are **immutable**:

```go
// ✅ GOOD: Returns new instance
func (s StreakSystem) UpdateForDate(date Date) StreakSystem {
    // logic
    return StreakSystem{...}
}

// ❌ BAD: Mutates
func (s *StreakSystem) UpdateForDate(date Date) {
    s.currentStreak++  // WRONG
}
```

---

## Naming Conventions

### Go Code

**Aggregates:** Singular noun
```go
type DailyGame struct { ... }
type MarathonGame struct { ... }
```

**Value Objects:** Noun
```go
type StreakSystem struct { ... }
type LivesSystem struct { ... }
type ChestType string
```

**Domain Services:** Noun + Service
```go
type DailyQuizSelector struct { ... }
type ChestRewardCalculator struct { ... }
```

**Methods:** Imperative verb
```go
func (dg *DailyGame) AnswerQuestion(...) { ... }
func (mg *MarathonGame) UseBonus(...) { ... }
```

**Factory Methods:**
```go
func NewDailyGame(...) (*DailyGame, error)
func ReconstructDailyGame(...) *DailyGame  // DB loading
```

**Events:** Past tense + Event
```go
type GameStartedEvent struct { ... }
type ChestOpenedEvent struct { ... }
```

**Enums:**
```go
type ChestType string
const (
    ChestTypeWooden ChestType = "wooden"
    ChestTypeSilver ChestType = "silver"
    ChestTypeGolden ChestType = "golden"
)
```

### Database

**Tables:** snake_case, plural
```sql
daily_quizzes
daily_games
marathon_games
marathon_personal_best
```

**Indexes:** `idx_` + table + columns
```sql
idx_daily_games_player_date
idx_marathon_games_player_active
```

### API

**REST:** `/api/v1/{mode}/{resource}/{action}`
```
/api/v1/daily/start
/api/v1/marathon/:gameId/answer
```

**Events:** snake_case
```json
{"type": "game_started"}
```

### Frontend

**Views:** PascalCase
```
DailyChallenge.vue
SoloMarathon.vue
```

**Composables:** camelCase + use prefix
```typescript
useDailyChallenge()
useMarathonGame()
```

**Components:** PascalCase
```
QuestionCard.vue
LivesDisplay.vue
ChestReward.vue
```

---

## Integration Between Contexts

### Resource Flow
```
Daily Challenge → Daily Chest → Resources:
  ├─ Coins → Shop, Marathon Continue, Streak Recovery
  ├─ PvP Tickets → (future: PvP Duel, Party Mode)
  └─ Marathon Bonuses → Solo Marathon strategic usage
```

### Anti-Corruption Layer
- Daily Challenge does NOT import Marathon
- Marathon does NOT import Daily Challenge
- Integration via: Domain Events + Application Layer

---

## Anti-Patterns

### Avoid Synonyms

| ❌ DON'T use | ✅ DO use |
|-------------|----------|
| Test, Questionnaire | Quiz |
| Session (for mode) | Game |
| HP, Health, Hearts | Lives |
| PowerUp, Boost, Help | Bonus |
| Combo, Chain | Streak |
| Continue cost formula duplication | Use ContinueCostCalculator |

### Avoid Ambiguity

```go
// ❌ BAD: Mixing content and process
type Quiz struct { userScore int }

// ✅ GOOD: Clear separation
type Quiz struct { questions []Question }
type DailyGame struct { quiz *Quiz; score int }
```

### Avoid Generic Names

```go
// ❌ BAD
type GameSession struct { ... }  // Which game?

// ✅ GOOD
type DailyGame struct { ... }
type MarathonGame struct { ... }
```

---

## DDD Pattern Summary

**Aggregate Root**
- Main entity controlling invariants
- Examples: `DailyQuiz`, `DailyGame`, `MarathonGame`
- Rule: All changes through aggregate root

**Entity**
- Object with identity (part of aggregate)
- Examples: `Question`, `Answer`

**Value Object**
- Immutable, no identity
- Examples: `StreakSystem`, `LivesSystem`, `ChestType`, `BonusInventory`
- Rule: Methods return NEW instance

**Domain Service**
- Business logic coordinating aggregates
- Examples: `ChestRewardCalculator`, `DifficultyCalculator`

**Repository**
- Interface for persistence
- Defined in DOMAIN, implemented in INFRASTRUCTURE
- Examples: `DailyGameRepository`, `MarathonGameRepository`

**Domain Event**
- Fact that happened (past tense)
- Examples: `GameStartedEvent`, `ChestOpenedEvent`

**Shared Kernel**
- Common logic across contexts
- Example: `kernel.QuizGameplaySession`

---

## Context-Specific Terms

### Daily Challenge Only
- DailyQuiz, DailyGame
- StreakSystem, ChestType
- Second Attempt, Streak Recovery

### Solo Marathon Only
- MarathonGame
- LivesSystem, BonusInventory, BonusType
- Shield, 50/50, Skip, Freeze
- Continue, Personal Best
- Weekly/All-Time Leaderboard

### Shared
- Quiz, Question, Answer, Category
- QuizGameplaySession (Shared Kernel)
- Score, Correct Answers
- Leaderboard (different types)
- Coins (currency)
