# GLOSSARY - Ubiquitous Language

> **Principle:** One term = one concept. No synonyms.

---

## Core Domain

### Quiz
- `quiz.Quiz` - Content (questions), NOT gameplay
- ❌ Avoid: Test, Questionnaire

### Question
- `quiz.Question` - One question + 4 answers
- Entity (part of Quiz aggregate)

### Answer
- `quiz.Answer` - Answer option (text + isCorrect)
- NOT `UserAnswer` (player's response)

### Game
- Process of playing Quiz in specific mode
- Types: `DailyGame`, `MarathonGame`
- ❌ Avoid: Session (for modes), Match, Round

### Session
- `kernel.QuizGameplaySession` - Shared Kernel
- Pure Q&A logic (reused by all modes)

---

## Daily Challenge Context

### Aggregates
- `daily_challenge.DailyQuiz` - Daily question set (10q, 00:00 UTC)
- `daily_challenge.DailyGame` - Player's attempt

### Value Objects
- `StreakSystem` - Consecutive days (immutable)
- `ChestType` - Enum: `wooden`, `silver`, `golden`
- `GameStatus` - Enum: `in_progress`, `completed`, `abandoned`

### Key Terms
- **Daily Chest** - Reward (0-4→Wooden, 5-7→Silver, 8-10→Golden)
- **Daily Streak** - Multiplier: 3d→1.1x, 7d→1.25x, 30d→1.5x
- **Second Attempt** - Retry: 100 coins OR Ad

### Anti-patterns
- ❌ `DailyChallenge` (aggregate name)
- ❌ `DailySession` (use DailyGame)

---

## Solo Marathon Context

### Aggregates
- `solo_marathon.MarathonGame` - PvE endless run

### Value Objects
- `LivesSystem` - 3 lives, immutable
- `BonusInventory` - 4 bonus types + quantities
- `BonusType` - Enum: `shield`, `fifty_fifty`, `skip`, `freeze`
- `PaymentMethod` - Enum: `coins`, `ad`

### Key Terms
- **Lives** - 3❤️ start, -1 per wrong, 0→Game Over
  - ❌ Avoid: HP, Health, Hearts
- **Bonuses** - Shield🛡️, 50/50🔀, Skip⏭️, Freeze❄️
  - ❌ Avoid: PowerUp, Boost, Help
- **Continue** - +1 life at game over: 200/400/600 coins OR Ad
- **Score** - Correct answers count (NO time bonus)

### Anti-patterns
- ❌ `MarathonSession` (use MarathonGame)
- ❌ `SoloGame`, `MarathonRun`

---

## Domain Services

**Daily Challenge:**
- `DailyQuizSelector` - Selects 10 questions (deterministic)
- `ChestRewardCalculator` - Calculates chest contents

**Solo Marathon:**
- `DifficultyCalculator` - Time limit + difficulty by question index
- `ContinueCostCalculator` - Cost: `200 + (count * 200)`
- `PersonalBestTracker` - Updates best score + awards 500 coins

---

## Domain Events

**Naming:** Past tense + Event

**Daily Challenge:**
```
DailyGameStartedEvent
DailyGameCompletedEvent
ChestOpenedEvent
StreakMilestoneReachedEvent
```

**Solo Marathon:**
```
MarathonGameStartedEvent
MarathonGameOverEvent
MarathonBonusUsedEvent
MarathonContinueUsedEvent
MarathonNewRecordEvent
```

---

## Naming Conventions

### Go Code

**Aggregates:** `DailyGame`, `MarathonGame`
**Value Objects:** `StreakSystem`, `LivesSystem`, `ChestType`
**Services:** `ChestRewardCalculator`, `DifficultyCalculator`
**Methods:** `AnswerQuestion()`, `UseBonus()` (imperative verb)
**Factory:** `NewDailyGame()`, `ReconstructDailyGame()` (for DB)
**Events:** `GameStartedEvent` (past tense)

### Database
**Tables:** `daily_games`, `marathon_games` (snake_case, plural)
**Indexes:** `idx_daily_games_player_date`

### API
**Routes:** `/api/v1/daily/start`, `/api/v1/marathon/:gameId/answer`

### Frontend
**Views:** `DailyChallenge.vue`, `SoloMarathon.vue`
**Composables:** `useDailyChallenge()`, `useMarathonGame()`

---

## DDD Patterns

**Aggregate Root:** Main entity. Examples: `DailyQuiz`, `DailyGame`, `MarathonGame`

**Value Object:** Immutable, no identity. Examples: `StreakSystem`, `LivesSystem`
```go
// ✅ Returns new instance
func (s StreakSystem) Update(date Date) StreakSystem

// ❌ Mutates (WRONG)
func (s *StreakSystem) Update(date Date)
```

**Domain Service:** Coordinates aggregates. Examples: `ChestRewardCalculator`

**Repository:** Persistence interface (domain), implementation (infrastructure)

**Shared Kernel:** `kernel.QuizGameplaySession` (used by all modes)

---

## Anti-Patterns

| ❌ DON'T | ✅ DO |
|---------|------|
| Test, Questionnaire | Quiz |
| Session (for mode) | Game |
| HP, Health, Hearts | Lives |
| PowerUp, Boost | Bonus |
| Combo, Chain | Streak |

---

## Integration

```
Daily Challenge → Chest → Resources:
  ├─ Coins → Marathon Continue
  └─ Bonuses → Marathon strategic usage
```

**Anti-Corruption:** Contexts do NOT import each other. Use Domain Events + Application Layer.

---

## Context-Specific Terms

**Daily Challenge Only:**
- DailyQuiz, DailyGame, StreakSystem, ChestType, Second Attempt

**Solo Marathon Only:**
- MarathonGame, LivesSystem, BonusInventory, Continue, Personal Best

**Shared:**
- Quiz, Question, Answer, Session (Shared Kernel), Coins

---

## PvP Duel Context

### Aggregates
- `pvp_duel.DuelGame` - Single 1v1 duel between two players
  - ❌ Avoid: DuelMatch, DuelSession
- `pvp_duel.PlayerRating` - Player's MMR and league standing
- `pvp_duel.DuelChallenge` - Friend challenge request
- `pvp_duel.Referral` - Track friend invitations and rewards

### Value Objects
- `DuelPlayer` - Player state within a duel (answers, score, time)
- `PlayerAnswer` - Single answer submission
- `League` - Enum: `bronze`, `silver`, `gold`, `platinum`, `diamond`, `legend`
- `GameStatus` - Enum: `waiting`, `countdown`, `in_progress`, `completed`, `cancelled`
  - ❌ Avoid: MatchStatus
- `WinReason` - Enum: `score`, `time`, `forfeit`
- `Milestone` - Referral milestone: `registered`, `played_5_duels`, `reached_silver`, etc.

### Domain Services
- `MMRCalculator` - ELO-based rating calculation (K=32)
- `MatchmakingService` - Queue management and opponent finding
- `ChallengeService` - Friend challenge creation and acceptance
- `ReferralService` - Milestone tracking and reward distribution

### Key Terms
- **Duel** - 1v1 real-time quiz game (7 questions, 10s each)
  - ❌ Avoid: Match, Battle, Fight
- **MMR** - Matchmaking Rating (ELO-based, starts at 1000)
- **League** - Rank tier (Bronze → Legend)
- **Division** - Sub-rank within league (IV → I)
- **PvP Ticket** - Entry cost for duel (1 ticket per game)
- **Challenge** - Friend duel invitation (direct or via link)
- **Rematch** - Immediate re-duel with same opponent

### Domain Events
```
DuelGameCreatedEvent
DuelGameCompletedEvent
DuelAnswerSubmittedEvent
PlayerPromotedEvent
PlayerDemotedEvent
DuelChallengeCreatedEvent
ReferralMilestoneAchievedEvent
SeasonEndedEvent
```

### Anti-patterns
- ❌ `DuelMatch` → `DuelGame`
- ❌ `MatchStatus` → `GameStatus`
- ❌ `isFriendMatch` → `isFriendGame`
- ❌ `match_found` → `game_found`
- ❌ `match_complete` → `game_complete`

### Context-Specific Terms
**PvP Duel Only:**
- DuelGame, DuelChallenge, PlayerRating, League, Division, MMR, PvP Ticket, Rematch, Referral
