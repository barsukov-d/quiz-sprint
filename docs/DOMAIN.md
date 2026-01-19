# Domain Documentation - Quiz Sprint TMA

## 📋 Оглавление

1. [Описание доменной области](#описание-доменной-области)
2. [Bounded Contexts (Ограниченные контексты)](#bounded-contexts)
3. [Context Map (Карта контекстов)](#context-map)
4. [Ubiquitous Language (Общий язык)](#ubiquitous-language)
5. [Core Domain: Quiz Taking](#core-domain-quiz-taking)
6. [Supporting Domains](#supporting-domains)
7. [Domain Events](#domain-events)
8. [Aggregates & Entities](#aggregates--entities)

---

## Описание доменной области

**Quiz Sprint TMA** - это Telegram Mini Application для прохождения интерактивных викторин в режиме реального времени.

### Бизнес-цели:
- Предоставить пользователям увлекательный опыт прохождения викторин
- Создать соревновательную среду через таблицу лидеров
- Мотивировать пользователей проходить квизы быстро и точно
- Интеграция с Telegram для легкого доступа и социального взаимодействия

### Ключевые характеристики домена:
- **Ограничение по времени**: Каждый квиз имеет временной лимит
- **Мгновенная обратная связь**: Пользователь сразу узнает правильность ответа
- **Подсчет очков**: Баллы начисляются за правильные ответы
- **Соревнование**: Результаты сравниваются в реальном времени
- **Неизменяемость**: Ответы нельзя изменить после отправки

---

## Bounded Contexts

### 1. Quiz Taking Context (Core Domain) 🎯

**Ответственность:**
- Процесс прохождения квизов
- Управление игровыми сессиями
- Отслеживание ответов пользователя
- Подсчет очков и времени

**Ubiquitous Language:**
- Quiz Session (Игровая сессия)
- User Answer (Ответ пользователя)
- Score (Очки)
- Time Limit (Временной лимит)

**Почему Core Domain?**
Это сердце бизнес-логики. Именно здесь происходит основное взаимодействие пользователя с системой.

---

### 2. Quiz Catalog Context (Supporting) 📚

**Ответственность:**
- Хранение и управление контентом квизов
- Управление вопросами и ответами
- Категоризация квизов
- Публикация квизов

**Ubiquitous Language:**
- Quiz (Квиз)
- Question (Вопрос)
- Answer (Вариант ответа)
- Category (Категория)

**Почему Supporting?**
Необходим для работы Core Domain, но не является уникальным конкурентным преимуществом.

---

### 3. Leaderboard Context (Supporting) 🏆

**Ответственность:**
- Отображение рейтинга игроков
- Вычисление позиций в таблице
- Real-time обновления результатов
- Хранение исторических данных

**Ubiquitous Language:**
- Leaderboard (Таблица лидеров)
- Rank (Позиция/Ранг)
- Leaderboard Entry (Запись в таблице)

**Особенность:**
Использует CQRS pattern - это Read Model, обновляется через Domain Events.

---

### 4. Identity Context (Generic) 👤

**Ответственность:**
- Управление пользователями
- Авторизация через Telegram
- Профили пользователей

**Ubiquitous Language:**
- User (Пользователь)
- Telegram User (Telegram пользователь)
- User Profile (Профиль пользователя)

**Почему Generic?**
Типовая функциональность, не специфичная для квиз-приложения.

---

## Context Map

```
┌─────────────────────────────────────────────────────────────┐
│                    QUIZ SPRINT SYSTEM                        │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────┐
│   Identity Context      │
│   (Generic Subdomain)   │
│                         │
│ • User                  │
│ • TelegramAuth          │
│ • UserProfile           │
└────────────┬────────────┘
             │ ACL (Anti-Corruption Layer)
             │ Exposes: UserID, Username
             │
    ┌────────┴──────────────────────────────────┐
    │                                            │
    ▼                                            ▼
┌─────────────────────────┐        ┌─────────────────────────┐
│  Quiz Catalog Context   │        │  Quiz Taking Context    │
│  (Supporting Subdomain) │◄───────│  (Core Domain) 🎯       │
│                         │ Uses   │                         │
│ • Quiz                  │        │ • QuizSession           │
│ • Question              │        │ • UserAnswer            │
│ • Answer                │        │ • SessionProgress       │
│ • Category              │        │                         │
└─────────────────────────┘        └──────────┬──────────────┘
                                              │
                                              │ Domain Events:
                                              │ • QuizStarted
                                              │ • AnswerSubmitted
                                              │ • QuizCompleted
                                              │
                                              ▼
                                   ┌─────────────────────────┐
                                   │ Leaderboard Context     │
                                   │ (Supporting Subdomain)  │
                                   │                         │
                                   │ • LeaderboardEntry      │
                                   │ • Ranking               │
                                   │ • EventHandlers         │
                                   └─────────────────────────┘
```

### Типы взаимодействий:

1. **Shared Kernel**: Quiz Catalog ↔ Quiz Taking
   - Делят QuizID, QuestionID
   - Quiz Taking читает Quiz (read-only)

2. **Published Language**: Identity → All
   - UserID - общий идентификатор
   - Username - для отображения

3. **Event-Driven**: Quiz Taking → Leaderboard
   - Асинхронное обновление через Domain Events
   - Eventual consistency

4. **ACL (Anti-Corruption Layer)**: Quiz Taking → Identity
   - Защита от изменений в Identity Context
   - Минимальная зависимость (только UserID)

---

## Ubiquitous Language

### Core Domain (Quiz Taking)

| Термин | Описание | Инварианты |
|--------|----------|-----------|
| **Quiz Session** | Активная попытка пользователя пройти квиз | Одна активная сессия на квиз на пользователя |
| **User Answer** | Ответ пользователя на конкретный вопрос | Нельзя ответить на один вопрос дважды |
| **Score** | Очки, набранные в сессии | Неотрицательное число, увеличивается только |
| **Session Status** | Состояние сессии (Active, Completed, Abandoned) | Можно отвечать только в Active |
| **Current Question** | Индекс текущего вопроса в сессии | 0-indexed, от 0 до количества вопросов |
| **Time Limit** | Ограничение времени на весь квиз (в секундах) | Положительное число, макс 3600 секунд |
| **Passing Score** | Минимальный процент для прохождения | От 0 до 100% |

### Supporting Domain (Quiz Catalog)

| Термин | Описание | Инварианты |
|--------|----------|-----------|
| **Quiz** | Набор вопросов с правилами прохождения | Минимум 5 вопросов, максимум 50 |
| **Question** | Вопрос с вариантами ответов | Ровно 1 правильный ответ из 2-4 вариантов |
| **Answer** | Вариант ответа на вопрос | Текст не пустой, макс 200 символов |
| **Points** | Баллы за правильный ответ | Неотрицательное число, макс 1000 |
| **Category** | Тематическая категория квиза | Уникальное название |

### Supporting Domain (Leaderboard)

| Термин | Описание | Инварианты |
|--------|----------|-----------|
| **Leaderboard** | Таблица лидеров для конкретного квиза | Упорядочена по Score (DESC), затем по времени |
| **Leaderboard Entry** | Запись о прохождении квиза пользователем | Один пользователь = одна лучшая попытка |
| **Rank** | Позиция в таблице лидеров | Положительное число, начинается с 1 |

### Generic Domain (Identity)

| Термин | Описание | Инварианты |
|--------|----------|-----------|
| **User** | Пользователь системы | Уникальный UserID |
| **Telegram User** | Пользователь из Telegram | Уникальный TelegramID |
| **Username** | Имя пользователя для отображения | Строка, может быть пустой (anonymous) |

---

## Core Domain: Quiz Taking

### Aggregates

#### 1. QuizSession (Aggregate Root)

**Ответственность:**
- Управление прохождением квиза одним пользователем
- Принятие и валидация ответов
- Подсчет очков
- Отслеживание прогресса

**Entities внутри:**
- `UserAnswer` - ответ пользователя на вопрос

**Value Objects:**
- `SessionID` - уникальный идентификатор сессии
- `QuizID` - ссылка на квиз (из Catalog Context)
- `UserID` - ссылка на пользователя
- `Points` - очки
- `SessionStatus` - статус сессии

**Бизнес-правила (Invariants):**
1. Пользователь может иметь только одну активную сессию на квиз
2. Нельзя ответить на вопрос дважды
3. Нельзя отправить ответ после завершения сессии
4. Ответ должен принадлежать текущему вопросу
5. Очки увеличиваются только при правильном ответе

**Domain Events:**
- `QuizStartedEvent` - когда создается новая сессия
- `AnswerSubmittedEvent` - когда пользователь отвечает на вопрос
- `QuizCompletedEvent` - когда все вопросы отвечены

**Use Cases:**
```go
// ✅ Реализовано
StartQuizUseCase(quizID, userID) → (session, firstQuestion)
  • Проверяет отсутствие активной сессии (409 если есть)
  • Создает новую сессию в статусе Active

// ✅ Реализовано
GetActiveSessionUseCase(quizID, userID) → (session, currentQuestion, totalQuestions, timeLimit)
  • Находит активную сессию пользователя для квиза
  • Возвращает текущий вопрос и прогресс
  • 404 если нет активной сессии

// ✅ Реализовано
SubmitAnswerUseCase(sessionID, questionID, answerID, userID) → (isCorrect, pointsEarned, nextQuestion | finalResult)
  • Проверяет авторизацию (userID должен совпадать)
  • Не позволяет ответить на один вопрос дважды
  • Автоматически завершает квиз после последнего вопроса

// ✅ Реализовано
AbandonSessionUseCase(sessionID, userID) → (void)
  • Удаляет активную сессию
  • Проверяет авторизацию (только владелец)
  • Позволяет начать квиз заново

// ⚠️ TODO
GetSessionStatusUseCase(sessionID) → (progress, score)
  • Получить статус любой сессии (включая завершенные)
```

---

## Supporting Domains

### Quiz Catalog Domain

#### Aggregate: Quiz

**Ответственность:**
- Хранение структуры квиза
- Валидация правил квиза
- Предоставление вопросов для игры

**Entities внутри:**
- `Question` - вопрос
- `Answer` - вариант ответа

**Value Objects:**
- `QuizID`, `QuestionID`, `AnswerID`
- `QuizTitle` - название (макс 200 символов)
- `QuestionText` - текст вопроса (макс 500 символов)
- `AnswerText` - текст ответа (макс 200 символов)
- `TimeLimit` - лимит времени (1-3600 секунд)
- `PassingScore` - минимальный процент (0-100%)

**Бизнес-правила:**
1. Квиз должен иметь минимум 5 вопросов
2. Квиз может иметь максимум 50 вопросов
3. Каждый вопрос должен иметь ровно 1 правильный ответ
4. Вопрос должен иметь от 2 до 4 вариантов ответов

**Use Cases:**
```go
ListQuizzesUseCase() → (quizzes[])
GetQuizDetailsUseCase(quizID) → (quiz)
GetQuizzesByCategoryUseCase(categoryID) → (quizzes[])
CreateQuizUseCase(title, description, questions) → (quizID)
```

---

#### Aggregate: Category

**Ответственность:**
- Организация квизов по тематикам
- Навигация и фильтрация контента
- Подсчет квизов в категории

**Value Objects:**
- `CategoryID` - уникальный идентификатор (UUID)
- `CategoryName` - название категории (макс 100 символов)
- `CategorySlug` - URL-friendly идентификатор (lowercase, hyphenated)
- `CategoryDescription` - описание категории (опциональное, макс 200 символов)
- `CategoryIcon` - эмодзи или иконка для визуальной идентификации

**Бизнес-правила:**
1. Название категории должно быть уникальным (case-insensitive)
2. Slug автогенерируется из названия: "General Knowledge" → "general-knowledge"
3. Категория может содержать 0 или более квизов
4. Удаление категории не удаляет квизы (category_id → NULL)

**Use Cases:**
```go
ListCategoriesUseCase() → (categories[])
GetCategoryUseCase(categoryID) → (category)
CreateCategoryUseCase(name, description) → (categoryID)
GetCategoryWithQuizCountUseCase(categoryID) → (category, quizCount)
```

**Связь с Quiz:**
- Quiz → CategoryID (optional foreign key)
- Квиз может принадлежать только одной категории
- При удалении категории, квизы остаются (category_id = NULL)

---

### Leaderboard Domain (CQRS Read Model)

#### Read Model: LeaderboardEntry

**Ответственность:**
- Отображение результатов пользователей
- Вычисление рангов
- Real-time обновления

**Структура:**
```go
type LeaderboardEntry struct {
    UserID      UserID
    Username    string
    Score       Points
    Rank        int
    QuizID      QuizID
    CompletedAt timestamp
}
```

**Event Handlers:**
- `OnQuizCompleted` → Обновить leaderboard
- `OnBetterScoreAchieved` → Пересчитать ранги

**Use Cases:**
```go
GetLeaderboardUseCase(quizID, limit) → (entries[])
GetUserRankUseCase(quizID, userID) → (rank, entry)
```

---

## Domain Events

### Event Flow

```
User starts quiz
    → QuizStartedEvent
        → [Analytics] Track quiz start
        → [Notification] Send welcome message

User submits answer
    → AnswerSubmittedEvent
        → [Analytics] Track answer submission
        → (No leaderboard update yet)

User completes quiz
    → QuizCompletedEvent
        → [Leaderboard] Update leaderboard ⭐
        → [Notification] Send completion message
        → [Analytics] Track completion
```

### Events List

| Event | Context | Payload | Subscribers |
|-------|---------|---------|-------------|
| `QuizStartedEvent` | Quiz Taking | quizID, sessionID, userID, timestamp | Analytics |
| `AnswerSubmittedEvent` | Quiz Taking | sessionID, questionID, answerID, isCorrect, points | Analytics |
| `QuizCompletedEvent` | Quiz Taking | quizID, sessionID, userID, finalScore, timestamp | Leaderboard, Notification, Analytics |
| `LeaderboardUpdatedEvent` | Leaderboard | quizID, topEntries[] | WebSocket (real-time) |

---

## Aggregates & Entities

### Aggregate Design Rules (Pragmatic DDD)

1. **One Repository per Aggregate** ✅
   - `QuizRepository` для Quiz
   - `SessionRepository` для QuizSession
   - `LeaderboardRepository` для Read Model

2. **Protect Invariants** ✅
   - Все бизнес-правила внутри агрегатов
   - Валидация в конструкторах Value Objects

3. **Small Aggregates** ⚠️ (Pragmatic)
   - Quiz содержит Questions[] (для производительности)
   - QuizSession содержит UserAnswers[] (в рамках транзакции)

4. **Reference by ID** ✅
   - QuizSession → QuizID (не полный Quiz)
   - UserAnswer → QuestionID (не полный Question)

### Entity Lifecycle

```
Quiz:
Create → Add Questions → Validate → Publish → (Read-Only)

QuizSession:
Create → Submit Answers → Complete/Abandon

LeaderboardEntry:
(Projection from QuizCompletedEvent)
```

---

## Диаграмма зависимостей

```
┌────────────────────────────────────────────────────────┐
│                   Application Layer                     │
│  (Use Cases - orchestration only)                      │
├────────────────────────────────────────────────────────┤
│ • StartQuizUseCase                                     │
│ • SubmitAnswerUseCase                                  │
│ • GetLeaderboardUseCase                                │
└────────────────┬───────────────────────────────────────┘
                 │ depends on
                 ▼
┌────────────────────────────────────────────────────────┐
│                    Domain Layer                         │
│  (Business logic, rules, invariants)                   │
├────────────────────────────────────────────────────────┤
│ Aggregates:                                            │
│ • Quiz (with Questions, Answers)                       │
│ • QuizSession (with UserAnswers)                       │
│                                                         │
│ Value Objects:                                         │
│ • IDs, Points, TimeLimit, etc                          │
│                                                         │
│ Domain Events:                                         │
│ • QuizStarted, AnswerSubmitted, QuizCompleted         │
│                                                         │
│ Interfaces (defined, not implemented):                │
│ • QuizRepository                                       │
│ • SessionRepository                                    │
│ • EventBus                                             │
└────────────────┬───────────────────────────────────────┘
                 │ implemented by
                 ▼
┌────────────────────────────────────────────────────────┐
│                Infrastructure Layer                     │
│  (HTTP, Database, WebSocket, External services)        │
├────────────────────────────────────────────────────────┤
│ • Fiber HTTP Handlers                                  │
│ • PostgreSQL Repository Implementations                │
│ • Redis Cache                                          │
│ • WebSocket Hub                                        │
│ • In-Memory Event Bus                                  │
└────────────────────────────────────────────────────────┘
```

---

## Следующие шаги

1. ✅ Определить Bounded Contexts
2. ✅ Создать Ubiquitous Language
3. ✅ Описать Aggregates и инварианты
4. ✅ Определить Domain Events
5. 🔄 Реализовать недостающие Use Cases
6. 🔄 Добавить тесты для бизнес-правил
7. 🔄 Создать Event Handlers для Leaderboard
8. 🔄 Реализовать Frontend интеграцию

---

**Дата создания:** 2026-01-15
**Методология:** Pragmatic DDD (по мотивам Vernon Vaughn IDDD)
**Проект:** Quiz Sprint TMA
