# Marathon Mode Architecture V2

## 🎯 Ключевые изменения в V2

### Что изменилось

**V1 (старая версия):**
- ❌ `MarathonGame` содержал `kernel.QuizGameplaySession`
- ❌ Session содержал полный `quiz.Quiz` с фиксированным набором вопросов
- ❌ Не подходит для бесконечного режима

**V2 (текущая версия):**
- ✅ `MarathonGameV2` содержит `currentQuestion *quiz.Question`
- ✅ Вопросы загружаются динамически через `QuestionSelector` Domain Service
- ✅ Поддерживает бесконечный поток вопросов
- ✅ Адаптивная сложность в реальном времени
- ✅ Исключение недавних вопросов (sliding window 20 вопросов)

---

## 🏗️ Архитектура

### Domain Layer

```
┌─────────────────────────────────────────────────┐
│         QuestionRepository                       │
│    (единый источник вопросов для всех режимов)  │
└────────────────┬────────────────────────────────┘
                 │
                 ├──────────────────┐
                 ↓                  ↓
┌────────────────────────┐  ┌──────────────────────┐
│  QuestionSelector      │  │  MarathonGameV2      │
│  (Domain Service)      │  │  (Aggregate Root)    │
│                        │  │                      │
│ - SelectNextQuestion() │  │ - currentQuestion    │
│ - Weighted random      │  │ - recentQuestionIDs  │
│ - Adaptive difficulty  │  │ - LoadNextQuestion() │
│ - Exclude recent       │  │ - AnswerQuestion()   │
└────────────────────────┘  └──────────────────────┘
```

### Поток выбора вопросов

```
1. MarathonGameV2.LoadNextQuestion(questionSelector)
   ↓
2. QuestionSelector.SelectNextQuestion(category, difficulty, recentIDs)
   ↓
3. DifficultyProgression.GetDistribution()
   → {"easy": 0.2, "medium": 0.5, "hard": 0.3}
   ↓
4. Weighted Random Selection
   → selectedDifficulty = "medium" (50% вероятность)
   ↓
5. QuestionRepository.FindRandomQuestions(filter, limit=1)
   ↓
6. MarathonGameV2.currentQuestion = question
   ↓
7. Update recentQuestionIDs (sliding window)
```

---

## 📦 Domain Model

### MarathonGameV2 Aggregate

```go
type MarathonGameV2 struct {
    // Identity
    id       GameID
    playerID UserID
    category MarathonCategory
    status   GameStatus

    // Timestamps
    startedAt  int64
    finishedAt int64

    // Current question (dynamic)
    currentQuestion *quiz.Question

    // Question history
    answeredQuestionIDs []QuestionID // All answered (for stats)
    recentQuestionIDs   []QuestionID // Last 20 (for exclusion)

    // Scoring
    currentStreak int
    maxStreak     int
    baseScore     int // Direct storage (no session)

    // Marathon mechanics
    lives      LivesSystem
    hints      HintsSystem
    difficulty DifficultyProgression

    // Personal best reference
    personalBestStreak *int

    // Events
    events []Event
}
```

### Key Methods

```go
// Load next question using Domain Service
func (mg *MarathonGameV2) LoadNextQuestion(
    questionSelector *QuestionSelector,
) error

// Answer current question
func (mg *MarathonGameV2) AnswerQuestion(
    questionID QuestionID,
    answerID AnswerID,
    timeTaken int64,
    answeredAt int64,
) (*AnswerQuestionResultV2, error)
```

---

## 🔄 Application Layer Flow

### StartMarathon

```go
1. Validate player has no active game
2. Determine category (all or specific)
3. Load PersonalBest (if exists)
4. Create MarathonGameV2 (WITHOUT questions)
5. Create QuestionSelector(questionRepo)
6. game.LoadNextQuestion(questionSelector)  // Load first question
7. Save game
8. Publish events
9. Return DTO with first question
```

### SubmitMarathonAnswer

```go
1. Load game from repository
2. Validate ownership
3. Get current question (for correct answer)
4. game.AnswerQuestion(questionID, answerID, timeTaken)
5. IF game continues:
   → game.LoadNextQuestion(questionSelector)  // Load next question
6. IF game over:
   → Update PersonalBest (if new record)
7. Save game
8. Publish events
9. Return DTO with next question OR game over result
```

---

## 🗄️ Database Schema

```sql
CREATE TABLE marathon_games (
    id UUID PRIMARY KEY,
    player_id UUID NOT NULL,
    category_id UUID,  -- NULL = all categories
    status VARCHAR(20) NOT NULL,
    started_at BIGINT NOT NULL,
    finished_at BIGINT,

    -- Current question (V2 specific)
    current_question_id UUID REFERENCES questions(id),

    -- Question history (JSONB for flexibility)
    answered_question_ids JSONB,  -- ["uuid1", "uuid2", ...]
    recent_question_ids JSONB,    -- Last 20 for exclusion

    -- Scoring
    current_streak INT DEFAULT 0,
    max_streak INT DEFAULT 0,
    base_score INT DEFAULT 0,

    -- Lives
    current_lives INT DEFAULT 3,
    lives_last_update BIGINT NOT NULL,

    -- Hints
    hints_fifty_fifty INT DEFAULT 3,
    hints_extra_time INT DEFAULT 2,
    hints_skip INT DEFAULT 1,

    -- Difficulty
    difficulty_level VARCHAR(20) DEFAULT 'beginner',

    -- Personal best reference
    personal_best_streak INT,

    -- Indexes
    INDEX idx_marathon_player_active (player_id, status),
    INDEX idx_marathon_current_question (current_question_id)
);

-- Questions table (shared across all modes)
CREATE TABLE questions (
    id UUID PRIMARY KEY,
    text TEXT NOT NULL,
    difficulty VARCHAR(10) NOT NULL,  -- 'easy', 'medium', 'hard'
    category_id UUID REFERENCES categories(id),
    points INT DEFAULT 100,
    created_at BIGINT NOT NULL,

    INDEX idx_questions_category_difficulty (category_id, difficulty)
);
```

---

## 📝 Use Cases Summary

### Обновленные Use Cases (V2)

1. **✅ StartMarathon**
   - Changed: Uses `MarathonGameV2`, loads question via `QuestionSelector`
   - Removed: Quiz loading logic

2. **✅ SubmitMarathonAnswer**
   - Changed: Loads next question after correct answer
   - Added: `QuestionSelector` dependency

3. **✅ UseMarathonHint**
   - No changes needed (works with currentQuestion)

4. **✅ AbandonMarathon**
   - Changed: Uses `game.BaseScore()` instead of `session.BaseScore()`

5. **✅ GetMarathonStatus**
   - Changed: Uses `ToMarathonGameDTOV2()` mapper

6. **✅ GetPersonalBests**
   - No changes needed

7. **✅ GetMarathonLeaderboard**
   - No changes needed

---

## 🚀 Migration Path (V1 → V2)

### For Infrastructure Layer

When implementing PostgreSQL repository:

1. **Save()**
   ```go
   // Save current_question_id separately
   currentQuestionID := game.CurrentQuestion().ID()

   // Save recent_question_ids as JSONB
   recentIDs := game.RecentQuestionIDs()
   recentJSON := jsonb.Marshal(recentIDs)

   // Save answered_question_ids as JSONB
   answeredIDs := game.AnsweredQuestionIDs()
   answeredJSON := jsonb.Marshal(answeredIDs)
   ```

2. **Reconstruct()**
   ```go
   // Load current question from repository
   var currentQuestion *quiz.Question
   if currentQuestionID != nil {
       currentQuestion = questionRepo.FindByID(currentQuestionID)
   }

   // Unmarshal JSONB arrays
   recentIDs := unmarshalQuestionIDs(recentJSON)
   answeredIDs := unmarshalQuestionIDs(answeredJSON)

   // Reconstruct game
   game := ReconstructMarathonGameV2(
       id, playerID, category, status,
       startedAt, finishedAt,
       currentQuestion,  // Pass loaded question
       answeredIDs, recentIDs,
       currentStreak, maxStreak, baseScore,
       lives, hints, difficulty,
       personalBestStreak, usedHints,
   )
   ```

---

## ✅ Testing Checklist

- [ ] QuestionSelector weighted random distribution
- [ ] Question exclusion (recent 20 questions)
- [ ] Adaptive difficulty progression
- [ ] Question loading after correct answer
- [ ] Game over when no questions available
- [ ] Personal best update logic
- [ ] Repository save/load cycle

---

## 📚 Related Files

**Domain Layer:**
- `marathon_game_aggregate_v2.go` - Main aggregate
- `question_selector.go` - Domain Service
- `question_selector_test.go` - Tests
- `repository.go` - Updated interface (uses V2)

**Application Layer:**
- `start_marathon.go` - ✅ Updated
- `submit_marathon_answer.go` - ✅ Updated
- `abandon_marathon.go` - ✅ Updated
- `get_marathon_status.go` - ✅ Updated
- `mapper.go` - Added `ToMarathonGameDTOV2()`

**Quiz Domain:**
- `quiz/question_repository.go` - New interface for question querying
