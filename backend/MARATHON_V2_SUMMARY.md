# Marathon Mode V2 - Implementation Summary

**Date:** 2026-01-26
**Status:** ✅ Domain + Application Layers Complete

---

## 🎯 Problem Solved

**Original Issue:** Marathon mode used `kernel.QuizGameplaySession` which contained a fixed `Quiz` aggregate with 10-20 questions. This doesn't work for an **endless mode** that needs infinite questions with **adaptive difficulty**.

**Solution:** Migrated to V2 architecture with dynamic question loading from a shared Question Pool.

---

## 📦 What Was Created

### 1. Domain Layer - Question Infrastructure

**File:** `backend/internal/domain/quiz/question_repository.go`
```go
type QuestionRepository interface {
    FindRandomQuestions(filter QuestionFilter, limit int) ([]*Question, error)
    FindByFilter(filter QuestionFilter) ([]*Question, error)
    CountByFilter(filter QuestionFilter) (int, error)
    // ... more methods
}

type QuestionFilter struct {
    CategoryID *CategoryID
    Difficulty *string
    ExcludeIDs []QuestionID  // Key for Marathon!
}
```

**Purpose:** Single source of questions for ALL game modes (Marathon, Daily, Duel, Party)

---

### 2. Domain Layer - QuestionSelector Service

**File:** `backend/internal/domain/solo_marathon/question_selector.go`

```go
type QuestionSelector struct {
    questionRepo quiz.QuestionRepository
}

func (qs *QuestionSelector) SelectNextQuestion(
    category MarathonCategory,
    difficulty DifficultyProgression,
    recentIDs []QuestionID,  // Exclude last 20 questions
) (*quiz.Question, error)
```

**Business Logic:**
- ✅ Weighted random selection based on difficulty distribution
  - Beginner: 80% easy, 20% medium, 0% hard
  - Master: 0% easy, 30% medium, 70% hard
- ✅ Excludes recently shown questions (sliding window of 20)
- ✅ Falls back if no questions available

**Tests:** `question_selector_test.go` validates weighted distribution

---

### 3. Domain Layer - MarathonGameV2 Aggregate

**File:** `backend/internal/domain/solo_marathon/marathon_game_aggregate_v2.go`

**Key Changes:**

| Old (V1) | New (V2) |
|----------|----------|
| `session *kernel.QuizGameplaySession` | `currentQuestion *quiz.Question` |
| Session contains fixed Quiz | Questions loaded dynamically |
| `baseScore` via session | `baseScore int` (direct storage) |
| No question history | `recentQuestionIDs []QuestionID` |
| - | `answeredQuestionIDs []QuestionID` |

**New Methods:**
```go
// Load next question using QuestionSelector
func (mg *MarathonGameV2) LoadNextQuestion(
    questionSelector *QuestionSelector,
) error

// Answer clears currentQuestion (will be loaded next time)
func (mg *MarathonGameV2) AnswerQuestion(...) (*AnswerQuestionResultV2, error)
```

**Workflow:**
```
1. Create game (no question yet)
2. LoadNextQuestion() → sets currentQuestion
3. Player answers
4. AnswerQuestion() → clears currentQuestion
5. If game continues: LoadNextQuestion() again
```

---

### 4. Application Layer - Updated Use Cases

**Updated Files:**
- ✅ `start_marathon.go` - Uses QuestionSelector to load first question
- ✅ `submit_marathon_answer.go` - Loads next question after correct answer
- ✅ `abandon_marathon.go` - Uses V2 baseScore
- ✅ `get_marathon_status.go` - Uses V2 mapper
- ✅ `mapper.go` - Added `ToMarathonGameDTOV2()`

**Updated Repository Interface:**
```go
// backend/internal/domain/solo_marathon/repository.go
type Repository interface {
    Save(game *MarathonGameV2) error
    FindByID(id GameID) (*MarathonGameV2, error)
    FindActiveByPlayer(playerID UserID) (*MarathonGameV2, error)
}
```

---

## 🗄️ Database Schema Changes

**New Table Structure (for Infrastructure layer to implement):**

```sql
CREATE TABLE marathon_games (
    id UUID PRIMARY KEY,
    player_id UUID NOT NULL,
    category_id UUID,
    status VARCHAR(20) NOT NULL,
    started_at BIGINT NOT NULL,
    finished_at BIGINT,

    -- V2 specific fields
    current_question_id UUID REFERENCES questions(id),  -- Current question
    answered_question_ids JSONB,  -- ["uuid1", "uuid2", ...] (all answered)
    recent_question_ids JSONB,    -- ["uuid1", ...] (last 20)

    current_streak INT DEFAULT 0,
    max_streak INT DEFAULT 0,
    base_score INT DEFAULT 0,  -- Direct storage (not via session)

    -- Lives, hints, difficulty (same as before)
    current_lives INT DEFAULT 3,
    lives_last_update BIGINT NOT NULL,
    hints_fifty_fifty INT DEFAULT 3,
    hints_extra_time INT DEFAULT 2,
    hints_skip INT DEFAULT 1,
    difficulty_level VARCHAR(20) DEFAULT 'beginner',
    personal_best_streak INT,

    INDEX idx_marathon_player_active (player_id, status),
    INDEX idx_marathon_current_question (current_question_id)
);
```

**Key Changes:**
- ✅ `current_question_id` - FK to questions table
- ✅ `answered_question_ids` - JSONB array (statistics)
- ✅ `recent_question_ids` - JSONB array (exclusion logic)
- ✅ `base_score` - Direct INT column (no session)

---

## 🔄 Architecture Flow

### Question Selection Flow

```
User → StartMarathon
  ↓
CreateMarathonGameV2() [no questions yet]
  ↓
QuestionSelector.SelectNextQuestion()
  ↓
DifficultyProgression.GetDistribution()
  → {"easy": 0.8, "medium": 0.2}
  ↓
Weighted Random Selection → "easy" (80% chance)
  ↓
QuestionFilter {
    Difficulty: "easy",
    CategoryID: geography,
    ExcludeIDs: [q1, q2, ..., q20]
}
  ↓
QuestionRepository.FindRandomQuestions(filter, 1)
  ↓
game.currentQuestion = question
game.recentQuestionIDs.append(question.ID)
  ↓
Return first question to user
```

### Answer Submission Flow

```
User → SubmitMarathonAnswer
  ↓
Load MarathonGameV2 from DB
  ↓
game.AnswerQuestion() [domain logic]
  ↓
IF correct:
  - Increment streak
  - Add base points
  - Update difficulty
  - currentQuestion = nil
  ↓
IF incorrect:
  - Lose life
  - Reset streak
  - currentQuestion = nil
  ↓
IF game_over:
  - Update PersonalBest
  - Return game over result
  ↓
ELSE (game continues):
  - game.LoadNextQuestion(questionSelector)
  - Return next question
```

---

## 📊 Comparison: V1 vs V2

| Aspect | V1 (Old) | V2 (New) |
|--------|----------|----------|
| **Question Source** | Fixed Quiz (10-20 questions) | Dynamic from Question Pool |
| **Session** | `kernel.QuizGameplaySession` | No session, just `currentQuestion` |
| **Question Loading** | All at start | One by one, on-demand |
| **Adaptive Difficulty** | ❌ Not really (fixed quiz) | ✅ Yes (real-time) |
| **Endless Mode** | ❌ Limited to quiz size | ✅ Truly endless |
| **Recent Exclusion** | ❌ No | ✅ Yes (last 20 questions) |
| **Base Score** | Via `session.BaseScore()` | Direct `baseScore int` |
| **Repository** | Returns `MarathonGame` | Returns `MarathonGameV2` |

---

## ✅ Testing Done

1. **✅ QuestionSelector Tests**
   - Weighted distribution (1000 samples)
   - Edge cases (empty distribution, single option)

2. **✅ Code Compilation**
   - All use cases compile
   - No type errors

---

## ⏳ Next Steps (Infrastructure Layer)

### 1. PostgreSQL Repository (High Priority)
**File:** `backend/internal/infrastructure/persistence/postgres/marathon_repository.go`

Tasks:
- [ ] Implement `Save()` with JSONB marshaling
- [ ] Implement `FindByID()` with question loading
- [ ] Implement `FindActiveByPlayer()`
- [ ] Implement `ReconstructMarathonGameV2()` helper

**Key Challenge:** Loading `currentQuestion` when reconstructing game
```go
// Pseudo-code
func (r *PostgresMarathonRepository) FindByID(id GameID) (*MarathonGameV2, error) {
    // 1. Load game row
    var row GameRow
    db.QueryRow("SELECT * FROM marathon_games WHERE id = $1", id).Scan(&row)

    // 2. Load current question if exists
    var currentQuestion *quiz.Question
    if row.CurrentQuestionID != nil {
        currentQuestion = r.questionRepo.FindByID(row.CurrentQuestionID)
    }

    // 3. Unmarshal JSONB arrays
    recentIDs := unmarshalQuestionIDs(row.RecentQuestionIDs)
    answeredIDs := unmarshalQuestionIDs(row.AnsweredQuestionIDs)

    // 4. Reconstruct aggregate
    return ReconstructMarathonGameV2(
        id, playerID, category, status,
        startedAt, finishedAt,
        currentQuestion,  // ← Key: loaded question
        answeredIDs, recentIDs,
        currentStreak, maxStreak, baseScore,
        lives, hints, difficulty,
        personalBestStreak, usedHints,
    )
}
```

---

### 2. HTTP Handlers + Swagger
**File:** `backend/internal/infrastructure/http/handlers/marathon_handlers.go`

Endpoints:
- [ ] `POST /api/v1/marathon/start`
- [ ] `POST /api/v1/marathon/{gameId}/answer`
- [ ] `POST /api/v1/marathon/{gameId}/hint`
- [ ] `DELETE /api/v1/marathon/{gameId}` (abandon)
- [ ] `GET /api/v1/marathon/status`
- [ ] `GET /api/v1/marathon/personal-bests`
- [ ] `GET /api/v1/marathon/leaderboard`

---

### 3. Database Migrations
**File:** `backend/migrations/XXX_create_marathon_tables.sql`

- [ ] Create `marathon_games` table
- [ ] Create `personal_bests` table
- [ ] Add indexes

---

### 4. Frontend (Vue)
- [ ] Generate TypeScript types from Swagger
- [ ] Create `MarathonHome.vue` component
- [ ] Create `MarathonGame.vue` component
- [ ] Create `useMarathon()` composable
- [ ] Add routes to Vue Router

---

## 📚 Documentation

**Created:**
- ✅ `ARCHITECTURE.md` - V2 architecture deep dive
- ✅ `README.md` - Updated with V2 changes
- ✅ `question_selector_test.go` - Test documentation
- ✅ `MARATHON_V2_SUMMARY.md` - This file

**Location:**
- Domain: `backend/internal/domain/solo_marathon/`
- Application: `backend/internal/application/marathon/`
- Docs: `docs/03_solo_marathon.md`

---

## 🎉 Summary

**✅ Completed:**
- Domain Layer: QuestionSelector, MarathonGameV2, QuestionRepository interface
- Application Layer: All 7 use cases updated for V2
- Documentation: Architecture guide, README updates
- Tests: QuestionSelector weighted distribution tests

**⏳ Next:**
- Infrastructure Layer: PostgreSQL repository, HTTP handlers
- Database: Migrations
- Frontend: Vue components, composables, routes

**Estimated Time to Complete:**
- Infrastructure: ~3-4 hours
- Frontend: ~2-3 hours
- Testing: ~1 hour
- **Total: 6-8 hours of work remaining**

---

**Ready for Infrastructure Layer implementation! 🚀**
