# 📚 GLOSSARY - Ubiquitous Language для Quiz Sprint

> **Цель:** Единый словарь терминов для всей команды, кодовой базы и документации.
> **Принцип:** Один термин = одно понятие. Избегаем синонимов и двусмысленности.

**Последнее обновление:** 2026-01-25
**Версия:** 1.0

---

## 🎯 Основные концепции (Core Concepts)

### Quiz (Квиз-контент)
**Domain:** `quiz.Quiz` aggregate root
**Определение:** Набор вопросов с настройками (категория, лимит времени, проходной балл)
**Пример:** "Квиз по географии" содержит 10 вопросов о странах и столицах
**НЕ путать с:** Процесс прохождения (это `Game`)
**Код:**
```go
type Quiz struct {
    id           QuizID
    title        QuizTitle
    questions    []Question
    timeLimit    TimeLimit
    passingScore PassingScore
}
```
**Синонимы (❌ избегать):** Test, Questionnaire, Assessment

---

### Question (Вопрос)
**Domain:** `quiz.Question` entity (часть `Quiz` aggregate)
**Определение:** Один вопрос с 4 вариантами ответов, из которых 1 правильный
**Пример:** "Столица Франции?" → [Париж ✓, Лондон, Берлин, Мадрид]
**Свойства:**
- Текст вопроса
- 4 варианта ответа
- Правильный ответ (индекс)
- Сложность (easy, medium, hard)
- Категория

**Код:**
```go
type Question struct {
    id       QuestionID
    text     QuestionText
    answers  []Answer      // Всегда 4 ответа
    points   Points
}
```

---

### Answer (Вариант ответа)
**Domain:** `quiz.Answer` entity
**Определение:** Один вариант ответа на вопрос
**НЕ путать с:** `UserAnswer` (ответ игрока)
**Код:**
```go
type Answer struct {
    id        AnswerID
    text      AnswerText
    isCorrect bool
    position  int  // 0-3
}
```

---

### Game (Игра)
**Domain:** Разные aggregates в зависимости от режима
**Определение:** Процесс прохождения Quiz'а игроком в конкретном режиме
**Типы:**
- `solo_marathon.MarathonGame` - Solo Marathon
- `daily_challenge.DailyGame` - Daily Challenge
- `quick_duel.DuelGame` - Quick Duel
- `party_mode.PartyGame` - Party Mode

**НЕ путать с:** `Quiz` (контент вопросов), `Session` (чистая логика геймплея)
**Синонимы (❌ избегать):** Match, Round, Run

---

### Session (Сессия геймплея)
**Domain:** `kernel.QuizGameplaySession`
**Определение:** Чистая логика прохождения вопросов без mode-specific правил (shared kernel)
**Назначение:** Переиспользуется всеми игровыми режимами
**Что хранит:**
- Ответы игрока
- Текущий индекс вопроса
- Базовый счёт (без бонусов режима)

**Код:**
```go
type QuizGameplaySession struct {
    id           SessionID
    quiz         *quiz.Quiz
    userAnswers  map[QuestionID]AnswerData
    baseScore    Points
}
```
**НЕ путать с:** `Game` (содержит mode-specific логику)

---

### Category (Категория)
**Domain:** `quiz.Category` aggregate root
**Определение:** Тематическая категория для вопросов (География, История, Наука, и т.д.)
**Пример:** Категория "География" содержит вопросы о странах, городах, реках
**Свойства:**
- Название
- Иконка
- Цвет (для UI)

**Код:**
```go
type Category struct {
    id   CategoryID
    name CategoryName
    icon string
}
```

---

## 🎮 Игровые режимы (Game Modes)

### Solo Marathon (Бесконечный марафон)
**Domain:** `solo_marathon.MarathonGame` aggregate root
**Определение:** Одиночный режим "до первой ошибки" с системой жизней
**Ключевые механики:**
- Lives (жизни с регенерацией)
- Hints (подсказки)
- Adaptive Difficulty (адаптивная сложность)
- Personal Record (личный рекорд)

**Код:**
```go
type MarathonGame struct {
    id            GameID
    playerID      UserID
    category      MarathonCategory
    currentStreak int
    lives         LivesSystem
    hints         HintsSystem
}
```

**Терминология:**
- ✅ `MarathonGame` - процесс игры
- ✅ `currentStreak` - текущая серия правильных ответов
- ❌ НЕ `MarathonSession`, НЕ `SoloGame`, НЕ `MarathonRun`

**API:** `/api/v1/marathon/*`
**Database:** `marathon_games` table

---

### Daily Challenge (Ежедневный вызов)
**Domain:** `daily_challenge.DailyGame` aggregate root + `daily_challenge.DailyQuiz` aggregate
**Определение:** Один набор вопросов для всех игроков мира каждый день
**Ключевые механики:**
- Daily Streak (серия дней подряд)
- Global Leaderboard (глобальный рейтинг)
- One Attempt (одна попытка в день)
- Same Questions (все получают одинаковые вопросы)

**Код:**
```go
// Набор вопросов дня (один для всех)
type DailyQuiz struct {
    id        DailyQuizID
    date      Date          // 2026-01-25
    questions []QuestionID  // 10 вопросов
    expiresAt int64
}

// Игра конкретного игрока
type DailyGame struct {
    id          GameID
    playerID    UserID
    dailyQuizID DailyQuizID
    streak      StreakSystem
    score       int
    rank        *int  // Позиция в leaderboard
}
```

**Терминология:**
- ✅ `DailyGame` - прохождение игроком
- ✅ `DailyQuiz` - набор вопросов дня
- ✅ `dailyStreak` - дней подряд играл
- ❌ НЕ `DailyChallenge` как aggregate (это название режима)
- ❌ НЕ `DailySession`

**API:** `/api/v1/daily/*`
**Database:** `daily_quizzes`, `daily_games` tables

---

### Quick Duel (Быстрая дуэль)
**Domain:** `quick_duel.DuelGame` aggregate root
**Определение:** PvP 1v1 в реальном времени с синхронизированными вопросами
**Ключевые механики:**
- Matchmaking (поиск соперника по ELO)
- ELO Rating (шахматный рейтинг)
- Synchronized Questions (одинаковые вопросы для обоих)
- Real-time (WebSocket)

**Код:**
```go
type DuelGame struct {
    id           GameID
    player1      DuelPlayer
    player2      DuelPlayer
    questions    []QuestionID  // 7 вопросов
    currentRound int
    player1ELO   EloRating
    player2ELO   EloRating
}

type DuelPlayer struct {
    userID    UserID
    score     int
    connected bool
}
```

**Терминология:**
- ✅ `DuelGame` - процесс дуэли
- ✅ `DuelPlayer` - игрок в дуэли
- ✅ `currentRound` - текущий раунд (1-7)
- ❌ НЕ `Match`, НЕ `PvPGame`, НЕ `QuickDuel` как aggregate

**API:** WebSocket `/ws/duel`
**Database:** `duel_games` table

---

### Party Mode (Режим вечеринки)
**Domain:** `party_mode.PartyRoom` + `party_mode.PartyGame` aggregate roots
**Определение:** Мультиплеер 2-8 игроков в приватной комнате
**Два агрегата:**
1. `PartyRoom` - лобби перед игрой
2. `PartyGame` - активная игра

**Ключевые механики:**
- Room Code (код комнаты ABC-123)
- Host Permissions (права хоста)
- Custom Settings (настройки комнаты)
- Real-time (WebSocket)

**Код:**
```go
// Лобби
type PartyRoom struct {
    id       RoomID
    code     RoomCode  // ABC-123
    hostID   UserID
    players  []RoomPlayer
    settings RoomSettings
    status   RoomStatus  // Lobby, Playing, Finished
}

// Активная игра
type PartyGame struct {
    id       GameID
    roomID   RoomID
    players  []PartyPlayer
    questions []QuestionID
    currentRound int
}
```

**Терминология:**
- ✅ `PartyRoom` - комната (лобби)
- ✅ `PartyGame` - активная игра
- ✅ `RoomCode` - код комнаты (ABC-123)
- ✅ `hostID` - ID хоста комнаты
- ❌ НЕ `Lobby` как aggregate (это статус `PartyRoom`)
- ❌ НЕ `MultiplayerGame`

**API:** WebSocket `/ws/party`
**Database:** `party_rooms`, `party_games` tables

---

## ⚙️ Игровые механики (Game Mechanics)

### Lives (Жизни)
**Domain:** `solo_marathon.LivesSystem` value object
**Определение:** Система жизней с автоматической регенерацией по времени
**Правила:**
- Максимум 3 жизни
- Теряется 1 жизнь за неправильный ответ
- Регенерируется 1 жизнь каждые 4 часа
- При 0 жизней - game over

**Код:**
```go
type LivesSystem struct {
    maxLives      int
    currentLives  int
    regenInterval int64  // 4 hours in seconds
    lastUpdate    int64  // Unix timestamp
}
```

**Терминология:**
- ✅ `Lives` - жизни
- ✅ `LivesSystem` - система жизней
- ❌ НЕ `HP`, НЕ `Health`, НЕ `Hearts`

**Используется в:** Solo Marathon

---

### Streak (Серия)
**Определение:** Последовательность успешных действий
**Контексты (разные значения!):**

1. **Solo Marathon: Current Streak**
   - Подряд правильных ответов в текущей игре
   - Сбрасывается на 0 при неправильном ответе
   - Влияет на сложность вопросов
   ```go
   type MarathonGame struct {
       currentStreak int  // Текущая серия
       maxStreak     int  // Лучшая серия в этой игре
   }
   ```

2. **Daily Challenge: Daily Streak**
   - Дней подряд играл в Daily Challenge
   - НЕ сбрасывается при неправильных ответах
   - Сбрасывается если пропустил день
   ```go
   type StreakSystem struct {
       currentStreak int  // Дней подряд
       bestStreak    int  // Лучшая серия всех времён
       lastPlayedDate Date
   }
   ```

**Терминология:**
- ✅ `currentStreak` - текущая серия
- ✅ `maxStreak` / `bestStreak` - лучшая серия
- ❌ НЕ `combo`, НЕ `chain`

---

### ELO Rating (Рейтинг ELO)
**Domain:** `quick_duel.EloRating` value object
**Определение:** Шахматный рейтинг для Quick Duel режима
**Правила:**
- Начальный рейтинг: 1000
- K-фактор: 32 (первые 30 игр) → 16
- Минимум: 100
- Изменение: ±12 в среднем

**Код:**
```go
type EloRating struct {
    rating      int
    gamesPlayed int
}

func (e EloRating) KFactor() int {
    if e.gamesPlayed < 30 {
        return 32
    }
    return 16
}
```

**Терминология:**
- ✅ `EloRating` - рейтинг игрока
- ✅ `eloChange` - изменение рейтинга после игры
- ❌ НЕ `MMR`, НЕ `Rank`, НЕ `Rating` без уточнения

**Используется в:** Quick Duel

---

### Hint (Подсказка)
**Domain:** `solo_marathon.HintType` enum + `solo_marathon.HintsSystem` value object
**Определение:** Помощь игроку во время ответа на вопрос
**Типы:**
- `fifty_fifty` - убрать 2 неправильных ответа
- `extra_time` - добавить 10 секунд к таймеру
- `skip` - пропустить вопрос без потери жизни

**Код:**
```go
type HintType string

const (
    HintFiftyFifty HintType = "fifty_fifty"
    HintExtraTime  HintType = "extra_time"
    HintSkip       HintType = "skip"
)

type HintsSystem struct {
    fiftyFifty int  // Количество доступных 50/50
    extraTime  int
    skip       int
}
```

**Терминология:**
- ✅ `Hint` - подсказка
- ✅ `HintType` - тип подсказки
- ✅ `HintsSystem` - система подсказок
- ❌ НЕ `PowerUp`, НЕ `Boost`, НЕ `Help`

**Используется в:** Solo Marathon

---

### Difficulty (Сложность)
**Domain:** `quiz.Difficulty` enum + `solo_marathon.DifficultyProgression` value object
**Уровни:**
- `easy` - лёгкие вопросы
- `medium` - средние вопросы
- `hard` - сложные вопросы

**Контексты:**

1. **Question Difficulty (сложность вопроса):**
   ```go
   type Question struct {
       difficulty string  // "easy", "medium", "hard"
   }
   ```

2. **Adaptive Difficulty (адаптивная сложность в Marathon):**
   ```go
   type DifficultyProgression struct {
       level DifficultyLevel  // Beginner → Master
   }

   type DifficultyDistribution struct {
       Easy   float64  // 0.8 = 80% лёгких вопросов
       Medium float64
       Hard   float64
   }
   ```

**Терминология:**
- ✅ `difficulty` - сложность
- ✅ `DifficultyProgression` - прогрессия сложности
- ❌ НЕ `level` в значении сложности (level = уровень игрока)

---

### Leaderboard (Таблица лидеров)
**Domain:** Read model (CQRS)
**Определение:** Рейтинг игроков по результатам
**Типы:**

1. **Quiz Leaderboard** - топ игроков в конкретном Quiz
2. **Global Leaderboard** - топ игроков по всем Quiz
3. **Daily Leaderboard** - топ игроков в Daily Challenge
4. **Marathon Leaderboard** - топ по категориям в Marathon

**Код:**
```go
type LeaderboardEntry struct {
    userID   UserID
    username string
    score    Points
    rank     int
}
```

**Хранение:** Redis Sorted Sets
**Терминология:**
- ✅ `Leaderboard` - таблица лидеров
- ✅ `LeaderboardEntry` - запись в таблице
- ❌ НЕ `Ranking`, НЕ `TopScores`

---

## 🏗️ Технические термины DDD (Domain-Driven Design)

### Aggregate Root (Корень агрегата)
**Определение:** Главная сущность, контролирующая инварианты и границы транзакций
**Примеры:**
- `Quiz` - корень агрегата вопросов
- `MarathonGame` - корень игры в Marathon
- `DuelGame` - корень дуэли
- `PartyRoom` - корень комнаты
- `User` - корень пользователя

**Правило:** Все изменения в aggregate идут ТОЛЬКО через aggregate root
**Код:**
```go
// ✅ Правильно
quiz.AddQuestion(question)

// ❌ Неправильно (прямое изменение)
quiz.questions = append(quiz.questions, question)
```

---

### Entity (Сущность)
**Определение:** Объект с уникальной идентичностью
**Примеры:**
- `Question` - сущность внутри `Quiz` aggregate
- `Answer` - сущность внутри `Question`
- `DuelPlayer` - сущность внутри `DuelGame`

**Отличие от Aggregate Root:** Entity не имеет смысла вне своего aggregate
**Код:**
```go
type Question struct {
    id   QuestionID  // Идентичность
    text QuestionText
}
```

---

### Value Object (Объект-значение)
**Определение:** Иммутабельный объект без идентичности, определяемый своими атрибутами
**Примеры:**
- `QuizID`, `QuestionID`, `UserID` - идентификаторы
- `Points` - очки
- `LivesSystem` - система жизней
- `EloRating` - рейтинг
- `RoomCode` - код комнаты

**Правило:** Value Objects иммутабельны (методы возвращают новый объект)
**Код:**
```go
type LivesSystem struct {
    currentLives int
    maxLives     int
}

// ✅ Возвращает новый объект
func (ls LivesSystem) LoseLife() LivesSystem {
    return LivesSystem{
        currentLives: ls.currentLives - 1,
        maxLives:     ls.maxLives,
    }
}

// ❌ НЕ мутируем!
func (ls *LivesSystem) LoseLife() {
    ls.currentLives--  // WRONG!
}
```

---

### Domain Service (Доменный сервис)
**Определение:** Бизнес-логика, которая не принадлежит одному aggregate
**Когда использовать:** Операция требует координации нескольких aggregates
**Примеры:**
- `DailyQuizSelector` - выбор вопросов для Daily Challenge
- `MatchmakingService` - поиск соперника для Duel

**Код:**
```go
type DailyQuizSelector struct {
    questionRepo quiz.QuestionRepository
}

func (s *DailyQuizSelector) SelectQuestionsForDate(date Date) ([]quiz.QuestionID, error) {
    // Business logic:
    // 1. Get questions from all categories
    // 2. Exclude questions from last 30 days
    // 3. Balance categories
    // 4. Sort by difficulty
}
```

---

### Repository (Репозиторий)
**Определение:** Интерфейс для персистентности aggregate roots
**Правило:** Repository определяется в DOMAIN layer, реализуется в INFRASTRUCTURE
**Примеры:**
```go
// domain/solo_marathon/repository.go
type Repository interface {
    Save(game *MarathonGame) error
    FindByID(id GameID) (*MarathonGame, error)
    FindActiveByUser(userID UserID) (*MarathonGame, error)
}

// infrastructure/persistence/postgres/marathon_repository.go
type PostgresMarathonRepository struct {
    db *sql.DB
}

func (r *PostgresMarathonRepository) Save(game *MarathonGame) error {
    // SQL implementation
}
```

**Терминология:**
- ✅ `Repository` - интерфейс в domain
- ✅ `PostgresMarathonRepository` - реализация в infrastructure
- ❌ НЕ `DAO`, НЕ `Storage`

---

### Domain Event (Доменное событие)
**Определение:** Факт, произошедший в domain (прошедшее время!)
**Примеры:**
- `QuizStartedEvent` - Quiz был начат
- `AnswerSubmittedEvent` - Ответ был отправлен
- `GameOverEvent` - Игра завершена
- `MatchFoundEvent` - Соперник найден

**Код:**
```go
type GameOverEvent struct {
    gameID      GameID
    maxStreak   int
    isNewRecord bool
    occurredAt  int64
}

// Aggregate собирает события
func (mg *MarathonGame) AnswerQuestion(...) {
    // Business logic

    if !mg.lives.HasLives() {
        mg.events = append(mg.events, NewGameOverEvent(mg.id, mg.maxStreak))
    }
}

// Application публикует события
events := game.Events()
eventBus.Publish(events...)
```

**Терминология:**
- ✅ Прошедшее время: `GameStartedEvent`, `AnswerSubmittedEvent`
- ❌ НЕ настоящее время: `StartGameEvent`, `SubmitAnswerEvent`

---

### Shared Kernel (Общее ядро)
**Определение:** Общая domain-логика, используемая несколькими bounded contexts
**В Quiz Sprint:**
- `kernel.QuizGameplaySession` - чистая логика геймплея
- Используется всеми режимами: Marathon, Daily Challenge, Duel, Party

**Код:**
```go
// kernel/quiz_gameplay_session.go
type QuizGameplaySession struct {
    id          SessionID
    quiz        *quiz.Quiz
    userAnswers map[QuestionID]AnswerData
    baseScore   Points  // БЕЗ mode-specific бонусов
}

// solo_marathon/marathon_game.go
type MarathonGame struct {
    session *kernel.QuizGameplaySession  // Композиция
    lives   LivesSystem                  // Mode-specific
    hints   HintsSystem                  // Mode-specific
}
```

---

## 📝 Naming Conventions (Соглашения об именовании)

### Go Domain Layer

#### Aggregates
```go
// Существительное в единственном числе
type Quiz struct { ... }
type MarathonGame struct { ... }
type DuelGame struct { ... }
type PartyRoom struct { ... }
```

#### Value Objects
```go
// Существительное
type QuizID struct { ... }
type Points struct { ... }
type LivesSystem struct { ... }
type EloRating struct { ... }
```

#### Domain Services
```go
// Существительное + Service
type MatchmakingService struct { ... }
type DailyQuizSelector struct { ... }
```

#### Methods (aggregate methods)
```go
// Глагол в повелительном наклонении
func (mg *MarathonGame) AnswerQuestion(...) { ... }
func (mg *MarathonGame) UseHint(...) { ... }
func (pr *PartyRoom) AddPlayer(...) { ... }
func (pr *PartyRoom) StartGame(...) { ... }
```

#### Factory Methods
```go
// New + AggregateRoot
func NewMarathonGame(...) (*MarathonGame, error) { ... }
func NewDuelGame(...) (*DuelGame, error) { ... }

// Reconstruct + AggregateRoot (для загрузки из БД)
func ReconstructMarathonGame(...) *MarathonGame { ... }
```

#### Domain Events
```go
// Прошедшее время + Event
type GameStartedEvent struct { ... }
type AnswerSubmittedEvent struct { ... }
type GameOverEvent struct { ... }
type MatchFoundEvent struct { ... }
```

---

### Database Tables

#### Основные правила
- snake_case
- Множественное число
- Aggregate root = одна таблица

```sql
-- Aggregates
CREATE TABLE quizzes (...);
CREATE TABLE marathon_games (...);
CREATE TABLE duel_games (...);
CREATE TABLE party_rooms (...);
CREATE TABLE users (...);

-- Junction tables (связующие)
CREATE TABLE party_room_players (...);  -- party_room + players
CREATE TABLE quiz_tags (...);           -- quiz + tags

-- Child entities (если не JSONB)
CREATE TABLE questions (...);
CREATE TABLE answers (...);
```

#### Индексы
```sql
-- idx_ + таблица + колонки
CREATE INDEX idx_marathon_games_player_active
    ON marathon_games(player_id, is_active);

CREATE INDEX idx_duel_games_started
    ON duel_games(started_at DESC);
```

---

### API Endpoints

#### REST API
```
Pattern: /api/v1/{режим}/{ресурс}/{действие}

Solo Marathon:
POST   /api/v1/marathon/start
POST   /api/v1/marathon/{gameId}/answer
POST   /api/v1/marathon/{gameId}/hint
DELETE /api/v1/marathon/{gameId}
GET    /api/v1/marathon/leaderboard

Daily Challenge:
POST   /api/v1/daily/start
POST   /api/v1/daily/{gameId}/answer
GET    /api/v1/daily/leaderboard

Quick Duel:
WebSocket: /ws/duel

Party Mode:
WebSocket: /ws/party
```

#### WebSocket Messages
```json
// type в snake_case
{
    "type": "find_match",
    "elo": 1200
}

{
    "type": "match_found",
    "gameId": "..."
}

{
    "type": "submit_answer",
    "answerId": "..."
}
```

---

### Frontend (Vue/TypeScript)

#### Views (страницы)
```
PascalCase + режим

MarathonHome.vue
MarathonGame.vue
DailyChallenge.vue
DuelGame.vue
PartyRoom.vue
PartyLobby.vue
```

#### Composables
```typescript
// camelCase + use prefix
useSoloMarathon()
useDailyChallenge()
useQuickDuel()
usePartyMode()
useLives()
useHints()
```

#### Components
```
PascalCase

QuestionCard.vue
AnswerButton.vue
LivesIndicator.vue
TimerBar.vue
LeaderboardTable.vue
```

#### API Calls (generated from Swagger)
```typescript
// hooks/marathon.ts (auto-generated)
useStartMarathon()
useSubmitMarathonAnswer()
useGetMarathonLeaderboard()
```

---

## ❌ Anti-patterns (Что НЕ использовать)

### Избегать синонимов

| ❌ НЕ использовать | ✅ Использовать | Контекст |
|-------------------|----------------|----------|
| `Test`, `Questionnaire` | `Quiz` | Контент вопросов |
| `Match` | `DuelGame` | Quick Duel |
| `Run` | `MarathonGame` | Solo Marathon |
| `Challenge` (без уточнения) | `DailyGame` или `DailyQuiz` | Daily Challenge |
| `Session` (для режимов) | `Game` | Процесс прохождения |
| `HP`, `Health` | `Lives` | Жизни в Marathon |
| `PowerUp`, `Boost` | `Hint` | Подсказки |
| `Combo`, `Chain` | `Streak` | Серия |
| `Ranking`, `TopScores` | `Leaderboard` | Таблица лидеров |
| `Lobby` (как aggregate) | `PartyRoom` (со статусом Lobby) | Party Mode |

### Избегать двусмысленности

```go
// ❌ ПЛОХО - неясно, это контент или процесс?
type Quiz struct {
    userScore int  // ???
}

// ✅ ХОРОШО - чёткое разделение
type Quiz struct {
    questions []Question  // Контент
}

type MarathonGame struct {
    quiz   *Quiz  // Ссылка на контент
    score  int    // Процесс
}
```

### Избегать generic названий

```go
// ❌ ПЛОХО
type GameSession struct { ... }  // Какой игры?
type Player struct { ... }       // В каком контексте?

// ✅ ХОРОШО
type MarathonGame struct { ... }
type DuelPlayer struct { ... }
type PartyPlayer struct { ... }
```

---

## 📖 Примеры использования

### Пример 1: Обсуждение задачи

❌ **Плохо:**
> "Нужно добавить систему HP в соло режим"

✅ **Хорошо:**
> "Нужно добавить LivesSystem в MarathonGame aggregate"

---

### Пример 2: Код review

❌ **Плохо:**
```go
type SoloSession struct {
    health int
}
```

✅ **Хорошо:**
```go
type MarathonGame struct {
    lives LivesSystem
}
```

---

### Пример 3: Документация

❌ **Плохо:**
> "Когда пользователь начинает test, создаётся session в базе"

✅ **Хорошо:**
> "Когда пользователь начинает MarathonGame, создаётся запись в таблице marathon_games"

---

### Пример 4: Commit message

❌ **Плохо:**
```
feat: add HP system to solo mode
```

✅ **Хорошо:**
```
feat(marathon): add LivesSystem to MarathonGame aggregate
```

---

## 📅 Changelog

### 2026-01-25 - v1.0
- ✅ Создан начальный глоссарий
- ✅ Добавлены все 4 игровых режима
- ✅ Добавлены игровые механики (Lives, Streak, ELO, Hints)
- ✅ Добавлены DDD термины (Aggregate, Entity, Value Object, Domain Service)
- ✅ Добавлены naming conventions для всех слоёв
- ✅ Добавлены anti-patterns

---

## 🔗 Связанные документы

- **CLAUDE.md** - Инструкции для Claude Code (ссылается на этот глоссарий)
- **DOMAIN.md** - Описание domain model
- **docs/01_quick_duel.md** - Спецификация Quick Duel
- **docs/02_daily_challenge.md** - Спецификация Daily Challenge
- **docs/03_solo_marathon.md** - Спецификация Solo Marathon
- **docs/04_party_mode.md** - Спецификация Party Mode

---

## 💡 Как использовать этот глоссарий

### Для разработчиков:
1. Читайте перед написанием кода
2. Используйте точные термины из глоссария
3. При сомнениях - ищите термин здесь
4. Предлагайте изменения через PR

### Для LLM (Claude Code):
1. **ВСЕГДА** читать перед генерацией кода
2. Использовать только термины из глоссария
3. Следовать naming conventions
4. Избегать anti-patterns

### Для документации:
1. Использовать единую терминологию
2. Ссылаться на глоссарий при первом упоминании термина
3. Обновлять глоссарий при добавлении новых концепций

---

**Вопросы или предложения?**
Создайте issue с меткой `glossary` в репозитории.
