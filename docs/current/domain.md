# Domain Model - Current Implementation

> **Для архитектурного overview см.** [`ARCHITECTURE.md`](../ARCHITECTURE.md)
> **Для словаря терминов см.** [`UBIQUITOUS_LANGUAGE.md`](../UBIQUITOUS_LANGUAGE.md)
> **Для будущих фич см.** [`future/ROADMAP.md`](../future/ROADMAP.md)

---

## 📋 Содержание

1. [Core Domain: Quiz Taking](#core-domain-quiz-taking)
2. [Supporting Domain: Quiz Catalog](#supporting-domain-quiz-catalog)
3. [Supporting Domain: Leaderboard](#supporting-domain-leaderboard)
4. [Supporting Domain: User Stats](#supporting-domain-user-stats)
5. [Domain Events](#domain-events)
6. [Repository Interfaces](#repository-interfaces)

---

## Core Domain: Quiz Taking

### Aggregate: QuizSession

**Ответственность:**

- Управление прохождением квиза одним пользователем
- Принятие и валидация ответов
- Подсчет очков с учетом скорости и серий
- Отслеживание прогресса

**Entities внутри:**

- `UserAnswer` - ответ пользователя на вопрос

**Value Objects:**

- `SessionID` - уникальный идентификатор сессии
- `QuizID` - ссылка на квиз (из Catalog Context)
- `UserID` - ссылка на пользователя
- `Points` - итоговые очки
- `SessionStatus` - статус сессии (Active, Completed, Abandoned)
- `CorrectAnswerStreak` - счетчик правильных ответов подряд
- `CurrentQuestionIndex` - текущий вопрос (0-indexed)

**Бизнес-правила (Invariants):**

1. Пользователь может иметь только одну активную сессию на квиз
2. Нельзя ответить на вопрос дважды
3. Нельзя отправить ответ после завершения сессии
4. Ответ должен принадлежать текущему вопросу
5. Очки увеличиваются только при правильном ответе
6. `CorrectAnswerStreak` сбрасывается при неверном ответе

**Domain Events:**

- `QuizStartedEvent` - когда создается новая сессия
- `AnswerSubmittedEvent` - когда пользователь отвечает на вопрос
- `QuizCompletedEvent` - когда все вопросы отвечены

---

### Use Cases

#### StartQuizUseCase

```go
StartQuizUseCase(quizID, userID) → (session, firstQuestion) | error

// Входные данные
- quizID: UUID квиза из каталога
- userID: UUID пользователя

// Бизнес-логика
1. Проверить отсутствие активной сессии для (userID, quizID)
2. Если есть активная сессия → вернуть 409 Conflict
3. Создать новую сессию в статусе Active
4. Инициализировать streak = 0, score = 0, currentQuestionIndex = 0
5. Опубликовать QuizStartedEvent
6. Вернуть сессию и первый вопрос

// Возвращаемые данные
- session: { sessionID, quizID, userID, status, score, currentQuestionIndex }
- firstQuestion: { questionID, text, answers[] }
```

**Status:** ✅ Реализовано

---

#### GetActiveSessionUseCase

```go
GetActiveSessionUseCase(quizID, userID) → (session, currentQuestion) | 404

// Входные данные
- quizID: UUID квиза
- userID: UUID пользователя

// Бизнес-логика
1. Найти активную сессию для (userID, quizID)
2. Если не найдена → вернуть 404 Not Found
3. Загрузить текущий вопрос по currentQuestionIndex
4. Вернуть сессию и вопрос

// Возвращаемые данные
- session: { sessionID, quizID, score, currentQuestionIndex, streak }
- currentQuestion: { questionID, text, answers[], timeLimit }
```

**Status:** ✅ Реализовано
**Use Case:** Восстановление активной сессии при возврате пользователя

---

#### SubmitAnswerUseCase

```go
SubmitAnswerUseCase(sessionID, questionID, answerID, userID, timeTaken)
  → (isCorrect, pointsEarned, streakInfo, nextQuestion?) | error

// Входные данные
- sessionID: UUID сессии
- questionID: UUID вопроса
- answerID: UUID выбранного ответа
- userID: UUID пользователя (для авторизации)
- timeTaken: int (секунды, затраченные на ответ)

// Бизнес-логика
1. Загрузить сессию по sessionID
2. Проверить авторизацию: session.userID == userID
3. Проверить status == Active
4. Проверить: вопрос не был отвечен ранее
5. Проверить: questionID соответствует currentQuestionIndex
6. Загрузить Quiz и Question
7. Проверить правильность ответа

Если ПРАВИЛЬНЫЙ:
  a. Рассчитать базовые очки (quiz.basePoints)
  b. Рассчитать time bonus:
     timeBonus = maxTimeBonus × (timeRemaining / timeLimitPerQuestion)
     где timeRemaining = timeLimitPerQuestion - timeTaken
  c. Увеличить streak++
  d. Если streak >= streakThreshold:
     - Начислить streakBonus ОДИН РАЗ при достижении порога
  e. totalPoints = basePoints + timeBonus + streakBonus (если применимо)
  f. session.score += totalPoints

Если НЕПРАВИЛЬНЫЙ:
  a. streak = 0 (сброс)
  b. totalPoints = 0

8. Сохранить UserAnswer(questionID, answerID, isCorrect, points, timeTaken)
9. session.currentQuestionIndex++
10. Опубликовать AnswerSubmittedEvent

Если это был последний вопрос:
  11. session.status = Completed
  12. Опубликовать QuizCompletedEvent
  13. nextQuestion = null
Иначе:
  14. Загрузить следующий вопрос
  15. nextQuestion = {...}

// Возвращаемые данные
- isCorrect: boolean
- pointsEarned: { base, timeBonus, streakBonus, total }
- streakInfo: { current, threshold, bonusEarned }
- nextQuestion: { questionID, text, answers[] } | null
```

**Status:** ✅ Реализовано
**Last Updated:** v1.3 - добавлен time bonus и streak bonus

---

#### AbandonSessionUseCase

```go
AbandonSessionUseCase(sessionID, userID) → void | error

// Входные данные
- sessionID: UUID сессии
- userID: UUID пользователя

// Бизнес-логика
1. Загрузить сессию
2. Проверить авторизацию: session.userID == userID
3. Проверить: status == Active (нельзя удалить завершенную)
4. Удалить сессию из БД

// Сценарий использования
Пользователь хочет начать квиз заново, не продолжая текущую сессию
```

**Status:** ✅ Реализовано

---

#### GetSessionResultsUseCase

```go
GetSessionResultsUseCase(sessionID) → (results) | 404

// Входные данные
- sessionID: UUID сессии

// Бизнес-логика
1. Загрузить сессию
2. Загрузить quiz (для метаданных)
3. Рассчитать статистику:
   - totalQuestions = len(quiz.questions)
   - correctAnswers = count(userAnswers where isCorrect == true)
   - timeSpent = sum(userAnswers.timeTaken)
   - passed = (score >= quiz.passingScore)
   - scorePercentage = (score / quiz.totalPoints) × 100
   - longestStreak = max streak во время сессии
   - avgAnswerTime = timeSpent / totalQuestions

// Возвращаемые данные
- session: { sessionID, quizID, score, status, startedAt, completedAt }
- quiz: { title, description, totalPoints, passingScore }
- statistics: {
    totalQuestions,
    correctAnswers,
    timeSpent,
    passed,
    scorePercentage,
    longestStreak,
    avgAnswerTime
  }
```

**Status:** ✅ Реализовано (v1.2)
**Use Case:** Отображение Results экрана

---

#### GetUserActiveSessionsUseCase

```go
GetUserActiveSessionsUseCase(userID) → (sessions[])

// Входные данные
- userID: UUID пользователя

// Бизнес-логика
1. Найти все сессии где userID == userID AND status == Active
2. Для каждой сессии загрузить quiz.title
3. Вернуть массив SessionSummary

// Возвращаемые данные
- sessions: [{
    sessionID,
    quizID,
    quizTitle,
    currentQuestion,    // Индекс (например, 3)
    totalQuestions,     // Всего (например, 10)
    score,
    startedAt
  }]
```

**Status:** ✅ Реализовано (v1.4)
**Use Case:** "Continue Playing" секция на главной странице

---

## Supporting Domain: Quiz Catalog

### Aggregate: Quiz

**Ответственность:**

- Хранение структуры и правил квиза
- Валидация правил квиза
- Предоставление вопросов для игры

**Entities внутри:**

- `Question` - вопрос с вариантами ответов
  - `questionID`: UUID
  - `text`: string (макс 500 символов)
  - `answers`: Answer[] (2-4 варианта)
  - `correctAnswerID`: UUID (ровно один)

**Value Objects:**

- `QuizID`, `QuestionID`, `AnswerID` (UUID)
- `QuizTitle` - название (макс 200 символов)
- `QuizDescription` - описание (опционально)
- `PassingScore` - минимальный процент (0-100%)
- `BasePoints` - базовые очки за правильный ответ
- `TimeLimitPerQuestion` - время на ответ (секунды, 5-60)
- `MaxTimeBonus` - максимальный бонус за скорость
- `StreakThreshold` - порог для бонуса за серию (например, 3)
- `StreakBonus` - очки за достижение серии

**Бизнес-правила:**

1. Квиз должен иметь минимум 5 вопросов
2. Квиз может иметь максимум 50 вопросов
3. Каждый вопрос должен иметь ровно 1 правильный ответ
4. Вопрос должен иметь от 2 до 4 вариантов ответов
5. `basePoints` × количество вопросов = `totalPoints`

---

### Use Cases (Quiz Catalog)

#### ListQuizzesUseCase

```go
ListQuizzesUseCase() → (quizzes[])

// Возвращает список всех доступных квизов
// Используется на главной странице для выбора квиза
```

#### GetQuizDetailsUseCase

```go
GetQuizDetailsUseCase(quizID) → (quiz) | 404

// Возвращает детальную информацию о квизе
// Используется на экране Quiz Details перед стартом
```

#### GetQuizzesByCategoryUseCase

```go
GetQuizzesByCategoryUseCase(categoryID) → (quizzes[])

// Фильтрация квизов по категории
// Используется при клике на категорию в навигации
```

#### GetDailyQuizUseCase

```go
GetDailyQuizUseCase(userID, date) → (quiz, completionStatus, userResult?)

// Входные данные
- userID: UUID пользователя
- date: string (YYYY-MM-DD)

// Бизнес-логика
1. Детерминированный выбор квиза:
   quizIndex = hash(date) % totalQuizzesCount
2. Загрузить quiz по индексу
3. Проверить, завершил ли пользователь этот квиз сегодня:
   - Найти session где quizID == quiz.ID AND userID == userID
     AND completedAt IN [date 00:00, date 23:59]
4. Если завершил:
   completionStatus = "completed"
   userResult = { score, rank, completedAt }
5. Иначе:
   completionStatus = "not_attempted"
   userResult = null

// Возвращаемые данные
- quiz: { id, title, description, questionCount, estimatedTime }
- completionStatus: "not_attempted" | "completed"
- userResult: { score, rank, completedAt } | null
```

**Status:** ✅ Реализовано (v1.4)
**Use Case:** Daily Challenge секция на главной

---

#### GetRandomQuizUseCase

```go
GetRandomQuizUseCase(categoryID?, excludeCompleted?) → (quiz)

// Входные данные
- categoryID: UUID (опционально) - фильтр по категории
- excludeCompleted: boolean - исключить пройденные квизы

// Бизнес-логика
1. Получить список квизов (с фильтром по categoryID если указан)
2. Если excludeCompleted == true:
   - Исключить квизы, которые пользователь уже завершал
3. Выбрать случайный квиз из списка

// Возвращаемые данные
- quiz: { id, title, description, questionCount }
```

**Status:** ✅ Реализовано (v1.4)
**Use Case:** Random Quiz кнопка на главной

---

### Aggregate: Category

**Ответственность:**

- Организация квизов по тематикам
- Навигация и фильтрация контента

**Value Objects:**

- `CategoryID` - UUID
- `CategoryName` - название (макс 100 символов, уникальное)
- `CategorySlug` - URL-friendly (auto-generated)
- `CategoryDescription` - описание (опционально, макс 200 символов)
- `CategoryIcon` - эмодзи для визуальной идентификации

**Бизнес-правила:**

1. Название категории уникально (case-insensitive)
2. Slug автогенерируется: "General Knowledge" → "general-knowledge"
3. Удаление категории не удаляет квизы (category_id → NULL)

**Связь с Quiz:**

- Quiz → CategoryID (optional foreign key)
- Один квиз = одна категория (или NULL)

---

### Use Cases (Category)

```go
ListCategoriesUseCase() → (categories[])
GetCategoryUseCase(categoryID) → (category) | 404
CreateCategoryUseCase(name, description, icon) → (categoryID)
GetCategoryWithQuizCountUseCase(categoryID) → (category, quizCount)
```

---

### Aggregate: Tag

**Ответственность:**

- Гибкая система метаданных для классификации квизов
- Множественные метки на квиз (many-to-many)

**Value Objects:**

- `TagID` - уникальный идентификатор (derived from name)
- `TagName` - значение тега в формате `{category}:{value}`

**Формат тега:**

```
{category}:{value}

Примеры:
- language:go
- difficulty:easy
- topic:concurrency
- domain:web-development
```

**Validation Rules:**

- Lowercase only: `^[a-z0-9-:]+$`
- Max length: 100 chars
- Required format: `category:value`
- No spaces (use hyphens)

**Бизнес-правила:**

1. Имя тега уникально
2. Тег immutable (для изменения - удалить и создать новый)
3. Квиз может иметь 0-10 тегов
4. Many-to-many связь через junction table `quiz_tags`

---

### Use Cases (Tag)

```go
ListTagsUseCase() → (tags[])
GetQuizzesByTagUseCase(tagName, limit, offset) → (quizzes[])
AssignTagsToQuizUseCase(quizID, tags[]) → void
```

---

### Quiz Import

```go
ImportQuizFromJSONUseCase(jsonData, format) → (quizID) | error

// format: "verbose" | "compact"
// Поддержка двух форматов для импорта квизов
// Compact format экономит 64% токенов для LLM generation

ImportQuizBatchUseCase(batchData) → (quizIDs[], errors[])

// Импорт нескольких квизов одновременно
// Возвращает массив успешных quizID и массив ошибок
```

**См. подробнее:** `backend/IMPORT.md`, `backend/data/quizzes/SCHEMA.md`

---

## Supporting Domain: Leaderboard

<!-- TODO: нужно пересмотреть реализацию мне не очень понятно как это будет по производительность -->

### Read Model: LeaderboardEntry

**Ответственность:**

- Отображение результатов пользователей
- Вычисление рангов
- Real-time обновления через WebSocket

**Структура:**

```go
type LeaderboardEntry struct {
    UserID      UUID
    Username    string
    Score       int
    Rank        int
    QuizID      UUID
    CompletedAt int64  // Unix timestamp
}
```

**Бизнес-правила:**

1. Один пользователь = одна запись (лучшая попытка)
2. Сортировка: по Score DESC, затем по CompletedAt ASC (при равенстве)
3. Rank рассчитывается динамически при запросе

**CQRS Pattern:**

- Leaderboard - это Read Model
- Обновляется асинхронно через `QuizCompletedEvent`
- Event Handler: `OnQuizCompleted` → Update/Insert leaderboard entry

---

### Use Cases (Leaderboard)

```go
GetLeaderboardUseCase(quizID, limit) → (entries[])

// Входные данные
- quizID: UUID квиза
- limit: int (default 50, max 100)

// Бизнес-логика
1. SELECT * FROM leaderboard_entries
   WHERE quiz_id = quizID
   ORDER BY score DESC, completed_at ASC
   LIMIT limit
2. Рассчитать rank для каждой записи

// Возвращаемые данные
- entries: [{
    userID,
    username,
    score,
    rank,
    completedAt
  }]


GetUserRankUseCase(quizID, userID) → (rank, entry) | null

// Находит позицию пользователя в leaderboard
// Если пользователь не проходил квиз → null
```

**WebSocket Support:**

- Endpoint: `wss://<domain>/ws/leaderboard/:quizId`
- Event: `LeaderboardUpdatedEvent` → broadcast топ-50 всем подключенным
- Real-time обновления при завершении квизов

---

<!-- TODO: нужно реализовать сейчас вообще не работает на фронтенде -->

## Supporting Domain: User Stats

### Aggregate: UserStats

**Ответственность:**

- Отслеживание прогресса пользователя
- Daily Quiz streak tracking
- Мотивационные метрики

**Value Objects:**

- `UserID` - UUID пользователя
- `CurrentStreak` - текущая серия дней подряд
- `LongestStreak` - лучшая серия за все время
- `LastDailyQuizDate` - дата последнего Daily Quiz (YYYY-MM-DD)
- `TotalQuizzesCompleted` - всего завершено квизов

**Бизнес-правила:**

1. Streak увеличивается ТОЛЬКО при завершении Daily Quiz
2. Streak сбрасывается при пропуске дня (gap > 1 день)
3. Повторное прохождение Daily Quiz в тот же день НЕ увеличивает streak
4. `LongestStreak` обновляется только если `CurrentStreak` > `LongestStreak`

**Streak Calculation:**

```go
func UpdateStreak(userID, completedAt) {
    stats := GetUserStats(userID)
    today := DateOf(completedAt)  // YYYY-MM-DD

    if stats.LastDailyQuizDate == today {
        // Уже проходили сегодня, ничего не меняем
        return
    }

    yesterday := today.AddDays(-1)

    if stats.LastDailyQuizDate == yesterday {
        // Продолжаем серию
        stats.CurrentStreak++
        if stats.CurrentStreak > stats.LongestStreak {
            stats.LongestStreak = stats.CurrentStreak
        }
    } else {
        // Пропустили день(и), сброс
        stats.CurrentStreak = 1
    }

    stats.LastDailyQuizDate = today
    SaveUserStats(stats)
}
```

---

### Use Cases (User Stats)

```go
GetUserStatsUseCase(userID) → (stats)

// Возвращаемые данные
- stats: {
    currentStreak,
    longestStreak,
    lastDailyQuizDate,
    totalQuizzesCompleted
  }


UpdateUserStatsOnQuizCompletionUseCase(userID, quizID, isDaily) → void

// Event Handler для QuizCompletedEvent
// Бизнес-логика:
1. Increment totalQuizzesCompleted
2. Если isDaily == true:
   - UpdateStreak(userID, completedAt)
```

**Status:** ✅ Реализовано (v1.4)

---

## Domain Events

### Event Flow

```
User starts quiz
    → QuizStartedEvent
        → [Analytics] Track quiz start

User submits answer
    → AnswerSubmittedEvent
        → [Analytics] Track answer submission

User completes quiz
    → QuizCompletedEvent
        → [Leaderboard] Update leaderboard entry
        → [User Stats] Update stats (increment total, update streak if daily)
        → [Analytics] Track completion
        → [WebSocket] Broadcast leaderboard update
```

---

### Events Catalog

| Event                       | Payload                                                    | Subscribers                                  |
| --------------------------- | ---------------------------------------------------------- | -------------------------------------------- |
| **QuizStartedEvent**        | quizID, sessionID, userID, timestamp                       | Analytics                                    |
| **AnswerSubmittedEvent**    | sessionID, questionID, answerID, isCorrect, points, streak | Analytics                                    |
| **QuizCompletedEvent**      | quizID, sessionID, userID, finalScore, timestamp, isDaily  | Leaderboard, UserStats, Analytics, WebSocket |
| **LeaderboardUpdatedEvent** | quizID, topEntries[]                                       | WebSocket clients                            |
| **QuizImportedEvent**       | quizID, categoryID, tags[], source                         | Analytics, SearchIndex                       |

---

## Repository Interfaces

### QuizRepository

```go
type QuizRepository interface {
    FindByID(quizID QuizID) (*Quiz, error)
    FindAll() ([]*Quiz, error)
    FindByCategory(categoryID CategoryID) ([]*Quiz, error)
    FindByTag(tagName string) ([]*Quiz, error)
    Save(quiz *Quiz) error
    Delete(quizID QuizID) error
}
```

### SessionRepository

```go
type SessionRepository interface {
    FindByID(sessionID SessionID) (*QuizSession, error)
    FindActiveByUserAndQuiz(userID UserID, quizID QuizID) (*QuizSession, error)
    FindActiveByUser(userID UserID) ([]*QuizSession, error)
    Save(session *QuizSession) error
    Delete(sessionID SessionID) error
}
```

### LeaderboardRepository

```go
type LeaderboardRepository interface {
    FindByQuizID(quizID QuizID, limit int) ([]*LeaderboardEntry, error)
    FindUserRank(quizID QuizID, userID UserID) (*LeaderboardEntry, error)
    Upsert(entry *LeaderboardEntry) error
}
```

### UserStatsRepository

```go
type UserStatsRepository interface {
    FindByUserID(userID UserID) (*UserStats, error)
    Save(stats *UserStats) error
    IncrementQuizzesCompleted(userID UserID) error
}
```

### CategoryRepository

```go
type CategoryRepository interface {
    FindByID(categoryID CategoryID) (*Category, error)
    FindAll() ([]*Category, error)
    Save(category *Category) error
}
```

### TagRepository

```go
type TagRepository interface {
    FindByName(name string) (*Tag, error)
    FindByNames(names []string) ([]*Tag, error)
    FindAll() ([]*Tag, error)
    FindByQuizID(quizID QuizID) ([]*Tag, error)
    Save(tag *Tag) error
}
```

---

## Aggregate Design Principles

Следуем Pragmatic DDD:

1. ✅ **One Repository per Aggregate** - каждый aggregate имеет свой репозиторий
2. ✅ **Protect Invariants** - все бизнес-правила внутри агрегатов
3. ⚠️ **Small Aggregates** - Quiz содержит Questions[] для performance (pragmatic choice)
4. ✅ **Reference by ID** - QuizSession → QuizID, не полный Quiz object

---

**Дата создания:** 2026-01-21
**Последнее обновление:** 2026-01-21
**Версия:** 1.0 (extracted from DOMAIN.md v1.5)
**Проект:** Quiz Sprint TMA
