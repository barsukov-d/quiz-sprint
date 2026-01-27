# GLOSSARY - Ubiquitous Language

> **Principle:** One term = one concept. Avoid synonyms and ambiguity.

## Core Domain Concepts

### Quiz (Quiz Content)
- **Aggregate Root:** `quiz.Quiz`
- **Definition:** Set of questions with settings (not the gameplay process)
- **Code:** `type Quiz struct { id QuizID; questions []Question; ... }`
- **❌ Avoid:** Test, Questionnaire, Assessment

### Question
- **Entity:** `quiz.Question` (part of Quiz aggregate)
- **Definition:** One question with 4 answer options (1 correct)
- **Code:** `type Question struct { id QuestionID; text string; answers []Answer; difficulty string }`

### Answer
- **Entity:** `quiz.Answer`
- **Not to be confused:** `UserAnswer` (player's response)
- **Code:** `type Answer struct { id AnswerID; text string; isCorrect bool; position int }`

### Game
- **Definition:** Process of taking a Quiz in specific mode
- **Types:** `MarathonGame`, `DailyGame`, `DuelGame`, `PartyGame`
- **Not to be confused:** `Quiz` (content), `Session` (pure gameplay logic)
- **❌ Avoid:** Match, Round, Run

### Session (Gameplay Session)
- **Shared Kernel:** `kernel.QuizGameplaySession`
- **Definition:** Pure question-answering logic without mode-specific rules
- **Usage:** Reused by all game modes via composition

### Category
- **Aggregate Root:** `quiz.Category`
- **Definition:** Thematic category for questions (Geography, History, etc.)

---

## Bounded Contexts (Game Modes)

### Daily Challenge Context
**Aggregates:**
- `daily_challenge.DailyQuiz` - Daily question set (one for all players, 10 questions, refreshes 00:00 UTC)
- `daily_challenge.DailyGame` - Individual player's attempt

**Key Concepts:**
- **Daily Chest** - Main reward (🪵 Wooden/🥈 Silver/🏆 Golden) based on score (0-4/5-7/8-10 correct)
- **Daily Streak** - Consecutive days played → resource multiplier (+10%/+25%/+50%)
- **Chest Contents:** PvP Tickets, Coins, Marathon Bonuses (Shield, 50/50, Skip, Freeze)

**Mechanics:**
- One free attempt per day
- Global/Friends/Country leaderboard
- Streak recovery (monetization)
- Second attempt (monetization)

**API:** `/api/v1/daily/*`
**DB:** `daily_quizzes`, `daily_games`
**❌ Avoid:** DailyChallenge as aggregate, DailySession

---

### Solo Marathon Context
**Aggregate Root:** `solo_marathon.MarathonGame`

**Key Concepts:**
- **Lives System** - 3 lives, -1 per wrong answer, game over at 0
- **Bonuses** (earned from Daily Chest):
  - 🛡️ Shield - One free mistake without losing life
  - 🔀 50/50 - Remove 2 wrong answers
  - ⏭️ Skip - Skip question without penalty
  - ❄️ Freeze - Add 10 seconds to timer
- **Score** - Number of correct answers in single run
- **Adaptive Difficulty** - Harder questions as player progresses

**Mechanics:**
- Endless questions until 3 lives lost
- Strategic bonus usage for record runs
- Weekly leaderboard (top 100 rewards)
- All-time hall of fame
- Continue run (monetization: rewarded ad or premium currency)

**API:** `/api/v1/marathon/*`
**DB:** `marathon_games`
**❌ Avoid:** MarathonSession, SoloGame, MarathonRun

---

### PvP Duel Context (Ranked)
**Aggregate Root:** `quick_duel.DuelGame`

**Key Concepts:**
- **PvP Ticket** - Entry cost (earned from Daily Challenge)
- **MMR/ELO Rating** - Skill-based matchmaking and ranking
- **League System:** 🥉 Bronze → 🥈 Silver → 🥇 Gold → 💍 Platinum → 💎 Diamond → 👑 Legend
- **Season** - 1 month duration, partial rating reset, exclusive cosmetic rewards

**Mechanics:**
- 1v1 synchronized gameplay (7 identical questions)
- Winner: more correct answers OR faster time (tiebreaker)
- **NO bonuses allowed** (pure skill)
- Matchmaking by MMR

**API:** WebSocket `/ws/duel`
**DB:** `duel_games`
**❌ Avoid:** Match, PvPGame, QuickDuel as aggregate

---

### Party Mode Context (Arcade PvP)
**Aggregate Root:** `party_mode.PartyRoom` (lobby) + `party_mode.PartyGame` (active game)

**Key Concepts:**
- **Room Code** - ABC-123 format for private rooms
- **Host Permissions** - Room creator controls
- **Weekly Modifiers** - Changing rules each week (Knockout, Speed, Themed, Bonuses Allowed, True/False)

**Mechanics:**
- 4 players per match
- Last man standing
- Quick matchmaking (no MMR)
- Private rooms with friends
- PvP Ticket entry cost
- Small coin reward for winner

**API:** WebSocket `/ws/party`
**DB:** `party_rooms`, `party_games`
**❌ Avoid:** Lobby as aggregate, MultiplayerGame

---

## Value Objects

### Lives System
- **Context:** Solo Marathon
- **Value Object:** `solo_marathon.LivesSystem`
- **Rules:** Max 3, -1 per error, game over at 0
- **❌ Avoid:** HP, Health, Hearts

### Streak
**Context-dependent meanings:**
1. **Marathon Streak:** `currentStreak` - consecutive correct answers (resets on error)
2. **Daily Streak:** `dailyStreak` - consecutive days played (NOT reset on errors)
- **❌ Avoid:** combo, chain

### ELO Rating
- **Context:** PvP Duel
- **Value Object:** `quick_duel.EloRating`
- **Rules:** Start 1000, K-factor 32→16, min 100
- **❌ Avoid:** MMR, Rank without clarification

### Bonus Types
- **Context:** Solo Marathon
- **Value Object:** `solo_marathon.BonusType`
- **Types:** `shield`, `fifty_fifty`, `skip`, `freeze`
- **❌ Avoid:** PowerUp, Boost, Help

### Daily Chest
- **Context:** Daily Challenge
- **Value Object:** `daily_challenge.ChestType`
- **Types:** `wooden` (0-4 correct), `silver` (5-7), `golden` (8-10)
- **Contents:** PvP Tickets, Coins, Marathon Bonuses

### PvP Ticket
- **Value Object:** `pvp.Ticket`
- **Usage:** Entry cost for PvP Duel and Party Mode
- **Source:** Earned from Daily Challenge Chest

### Difficulty
- **Levels:** `easy`, `medium`, `hard`
- **Contexts:** Question difficulty (property) vs Adaptive difficulty (Marathon progression)
- **❌ Avoid:** level as difficulty

### Leaderboard
- **Read Model (CQRS)**
- **Types:** Daily Global/Friends/Country, Marathon Weekly/All-Time, PvP Seasonal
- **Storage:** Redis Sorted Sets
- **❌ Avoid:** Ranking, TopScores

---

## DDD Patterns

### Aggregate Root
**Definition:** Main entity controlling invariants and transaction boundaries
**Examples:** `Quiz`, `MarathonGame`, `DuelGame`, `PartyRoom`, `DailyQuiz`, `User`
**Rule:** All changes ONLY through aggregate root

### Entity
**Definition:** Object with unique identity (meaningless outside aggregate)
**Examples:** `Question`, `Answer`, `DuelPlayer`, `PartyPlayer`

### Value Object
**Definition:** Immutable object without identity
**Examples:** `QuizID`, `Points`, `LivesSystem`, `EloRating`, `ChestType`, `Ticket`
**Rule:** Methods return new object (NO mutation)

### Domain Service
**Definition:** Business logic coordinating multiple aggregates
**Examples:** `DailyQuizSelector`, `MatchmakingService`, `ChestRewardCalculator`

### Repository
**Definition:** Interface for aggregate root persistence
**Rule:** Defined in DOMAIN, implemented in INFRASTRUCTURE
**❌ Avoid:** DAO, Storage

### Domain Event
**Definition:** Fact that happened in domain (past tense!)
**Examples:** `GameStartedEvent`, `AnswerSubmittedEvent`, `GameOverEvent`, `ChestOpenedEvent`
**❌ Avoid:** Present tense (StartGameEvent)

### Shared Kernel
**Definition:** Common domain logic shared by bounded contexts
**In project:** `kernel.QuizGameplaySession` - used by all game modes

---

## Naming Conventions

### Go Domain
```go
// Aggregates - singular noun
type MarathonGame struct { ... }

// Value Objects - noun
type LivesSystem struct { ... }
type ChestType string

// Domain Services - noun + Service
type MatchmakingService struct { ... }

// Methods - imperative verb
func (mg *MarathonGame) AnswerQuestion(...) { ... }
func (mg *MarathonGame) UseBonus(...) { ... }

// Factory Methods
func NewMarathonGame(...) (*MarathonGame, error) { ... }
func ReconstructMarathonGame(...) *MarathonGame { ... }  // for DB loading

// Domain Events - past tense + Event
type GameStartedEvent struct { ... }
type ChestOpenedEvent struct { ... }

// Enums
type BonusType string
const (
    BonusShield     BonusType = "shield"
    BonusFiftyFifty BonusType = "fifty_fifty"
    BonusSkip       BonusType = "skip"
    BonusFreeze     BonusType = "freeze"
)
```

### Database
```sql
-- snake_case, plural
CREATE TABLE marathon_games (...);
CREATE TABLE daily_quizzes (...);
CREATE TABLE daily_games (...);
CREATE TABLE duel_games (...);
CREATE TABLE party_rooms (...);

-- Indexes: idx_ + table + columns
CREATE INDEX idx_marathon_games_player_active ON marathon_games(player_id, is_active);
CREATE INDEX idx_daily_games_date ON daily_games(date DESC);
```

### API
```
REST: /api/v1/{mode}/{resource}/{action}
WebSocket messages: type in snake_case {"type": "find_match"}
```

### Frontend
```typescript
// Views - PascalCase
MarathonGame.vue
DailyChallenge.vue

// Composables - camelCase + use prefix
useSoloMarathon()
useDailyChallenge()

// Components - PascalCase
QuestionCard.vue
ChestReward.vue
```

---

## Anti-patterns

### Avoid Synonyms
| ❌ DON'T use | ✅ DO use |
|-------------|----------|
| Test, Questionnaire | Quiz |
| Match | DuelGame |
| Run | MarathonGame |
| Session (for modes) | Game |
| HP, Health | Lives |
| PowerUp, Boost | Bonus |
| Combo, Chain | Streak |
| Challenge (without context) | DailyGame or DailyQuiz |

### Avoid Ambiguity
```go
// ❌ BAD - mixing content and process
type Quiz struct { userScore int }

// ✅ GOOD - clear separation
type Quiz struct { questions []Question }
type MarathonGame struct { quiz *Quiz; score int }
```

### Avoid Generic Names
```go
// ❌ BAD
type GameSession struct { ... }  // Which game?

// ✅ GOOD
type MarathonGame struct { ... }
type DuelPlayer struct { ... }
```

---

## Cross-Context Integration

### Resource Flow
```
Daily Challenge → Daily Chest → Resources:
  ├─ PvP Tickets → PvP Duel / Party Mode
  ├─ Coins → Shop, Marathon Continue, Streak Recovery
  └─ Marathon Bonuses → Solo Marathon strategic usage
```

### Monetization Points
- **Daily Challenge:** Second attempt, Streak recovery, Premium (chest upgrade)
- **Solo Marathon:** Continue run, Bonus packs
- **PvP Duel:** Ticket purchase, Cosmetics
- **Party Mode:** Ticket purchase, Rewarded ad for free ticket
