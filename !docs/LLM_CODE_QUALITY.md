# 🤖 Практики для улучшения качества кода, генерируемого LLM

> **Цель:** Стандартизировать подход к генерации кода для Domain Layer (`backend/internal/domain`)
> **Для кого:** Claude Code и другие LLM помощники

**Последнее обновление:** 2026-01-25

---

## 📚 Обязательное чтение перед генерацией кода

**ВСЕГДА читай в этом порядке:**

1. **[GLOSSARY.md](GLOSSARY.md)** - Единый словарь терминов
2. **[LLM_CODE_QUALITY.md](LLM_CODE_QUALITY.md)** (этот файл) - Практики и шаблоны
3. **[ADR/](adr/)** - Архитектурные решения (если есть)

---

## 🎯 Scope: Только Domain Layer

**Фокус:** `backend/internal/domain/`

```
backend/internal/domain/
├── quiz/                    # Quiz aggregate (базовый контент)
├── user/                    # User aggregate
├── kernel/                  # Shared kernel (геймплей)
├── game_modes/              # Игровые режимы (фокус!)
│   ├── solo_marathon/
│   ├── daily_challenge/
│   ├── quick_duel/
│   └── party_mode/
└── shared/                  # Shared domain logic
```

**НЕ включает:**
- ❌ Application layer (`internal/application/`)
- ❌ Infrastructure layer (`internal/infrastructure/`)

---

## 1️⃣ Code Templates для Domain Layer

### 1.1 Template: Aggregate Root

**Файл:** `docs/templates/domain/aggregate_root.go.template`

```go
package {{DOMAIN}}

import (
    "github.com/barsukov/quiz-sprint/backend/internal/domain/shared"
)

// {{AGGREGATE}} is the aggregate root for {{DESCRIPTION}}
//
// Business Invariants:
// - {{INVARIANT_1}}
// - {{INVARIANT_2}}
type {{AGGREGATE}} struct {
    // Identity
    id {{AGGREGATE}}ID

    // Core domain fields
    // TODO: add domain fields

    // Lifecycle
    createdAt int64 // Unix timestamp (no time.Time to keep domain pure)
    updatedAt int64

    // Domain events (transient, not persisted)
    events []Event
}

// New{{AGGREGATE}} creates a new {{AGGREGATE}} aggregate
//
// Business Rules:
// - {{RULE_1}}
// - {{RULE_2}}
func New{{AGGREGATE}}(
    id {{AGGREGATE}}ID,
    // TODO: add parameters
    createdAt int64,
) (*{{AGGREGATE}}, error) {
    // 1. Validate inputs
    if id.IsZero() {
        return nil, ErrInvalid{{AGGREGATE}}ID
    }

    // 2. Create aggregate
    agg := &{{AGGREGATE}}{
        id:        id,
        createdAt: createdAt,
        updatedAt: createdAt,
        events:    make([]Event, 0),
    }

    // 3. Record domain event
    agg.events = append(agg.events, New{{AGGREGATE}}CreatedEvent(id, createdAt))

    return agg, nil
}

// Reconstruct{{AGGREGATE}} reconstructs a {{AGGREGATE}} from persistence (no validation)
// Used by repository when loading from database
func Reconstruct{{AGGREGATE}}(
    id {{AGGREGATE}}ID,
    // TODO: add all fields
    createdAt int64,
    updatedAt int64,
) *{{AGGREGATE}} {
    return &{{AGGREGATE}}{
        id:        id,
        createdAt: createdAt,
        updatedAt: updatedAt,
        events:    make([]Event, 0), // Don't replay events from DB
    }
}

// Business Methods (commands)
// ============================================================================

// DoSomething performs a business operation
//
// Business Rules:
// - {{RULE}}
func (a *{{AGGREGATE}}) DoSomething(params ...) error {
    // 1. Validate preconditions
    if a.someCondition {
        return ErrPreconditionFailed
    }

    // 2. Apply business logic
    // ... domain logic here ...

    // 3. Update state
    a.updatedAt = getCurrentTimestamp()

    // 4. Record domain event
    a.events = append(a.events, NewSomethingHappenedEvent(...))

    return nil
}

// Getters (immutable access)
// ============================================================================

func (a *{{AGGREGATE}}) ID() {{AGGREGATE}}ID { return a.id }
func (a *{{AGGREGATE}}) CreatedAt() int64     { return a.createdAt }
func (a *{{AGGREGATE}}) UpdatedAt() int64     { return a.updatedAt }

// Events returns collected domain events and clears them
func (a *{{AGGREGATE}}) Events() []Event {
    events := a.events
    a.events = make([]Event, 0)
    return events
}
```

**Использование:**
```
1. Копируй template
2. Замени {{PLACEHOLDERS}}
3. Добавь domain-specific поля
4. Реализуй бизнес-методы
```

---

### 1.2 Template: Value Object

**Файл:** `docs/templates/domain/value_object.go.template`

```go
package {{DOMAIN}}

// {{VALUE_OBJECT}} is a value object representing {{DESCRIPTION}}
//
// Invariants:
// - {{INVARIANT_1}}
// - Immutable after creation
type {{VALUE_OBJECT}} struct {
    value {{TYPE}}
}

// New{{VALUE_OBJECT}} creates a new {{VALUE_OBJECT}} with validation
func New{{VALUE_OBJECT}}(value {{TYPE}}) ({{VALUE_OBJECT}}, error) {
    // Validation
    if !isValid(value) {
        return {{VALUE_OBJECT}}{}, ErrInvalid{{VALUE_OBJECT}}
    }

    return {{VALUE_OBJECT}}{value: value}, nil
}

// Reconstruct{{VALUE_OBJECT}} reconstructs from persistence (no validation)
func Reconstruct{{VALUE_OBJECT}}(value {{TYPE}}) {{VALUE_OBJECT}} {
    return {{VALUE_OBJECT}}{value: value}
}

// Business methods (return new value object, don't mutate!)
// ============================================================================

// Transform applies a transformation and returns NEW value object
func (vo {{VALUE_OBJECT}}) Transform() {{VALUE_OBJECT}} {
    // Calculate new value
    newValue := // ... transformation logic ...

    // Return NEW value object (immutable!)
    return {{VALUE_OBJECT}}{value: newValue}
}

// Query methods
// ============================================================================

func (vo {{VALUE_OBJECT}}) Value() {{TYPE}} { return vo.value }

func (vo {{VALUE_OBJECT}}) IsZero() bool {
    return vo.value == {{ZERO_VALUE}}
}

func (vo {{VALUE_OBJECT}}) Equals(other {{VALUE_OBJECT}}) bool {
    return vo.value == other.value
}
```

**Ключевое правило:** Value Objects ИММУТАБЕЛЬНЫ!
```go
// ✅ ПРАВИЛЬНО - возвращает новый объект
func (ls LivesSystem) LoseLife() LivesSystem {
    return LivesSystem{currentLives: ls.currentLives - 1}
}

// ❌ НЕПРАВИЛЬНО - мутирует!
func (ls *LivesSystem) LoseLife() {
    ls.currentLives--  // WRONG!
}
```

---

### 1.3 Template: Domain Service

**Файл:** `docs/templates/domain/domain_service.go.template`

```go
package {{DOMAIN}}

// {{SERVICE}} is a domain service for {{DESCRIPTION}}
//
// When to use Domain Service:
// - Business logic doesn't belong to a single aggregate
// - Operation requires coordination between aggregates
// - Stateless operation
//
// Examples:
// - DailyQuizSelector (selects questions from multiple quizzes)
// - MatchmakingService (matches players)
type {{SERVICE}} struct {
    // Dependencies (repository interfaces from domain)
    repo1 Repository1
    repo2 Repository2
}

// New{{SERVICE}} creates a new domain service
func New{{SERVICE}}(repo1 Repository1, repo2 Repository2) *{{SERVICE}} {
    return &{{SERVICE}}{
        repo1: repo1,
        repo2: repo2,
    }
}

// DoOperation performs a domain operation
//
// Business Rules:
// - {{RULE_1}}
// - {{RULE_2}}
func (s *{{SERVICE}}) DoOperation(params ...) (result, error) {
    // 1. Load aggregates
    agg1, err := s.repo1.FindByID(id1)
    if err != nil {
        return nil, err
    }

    agg2, err := s.repo2.FindByID(id2)
    if err != nil {
        return nil, err
    }

    // 2. Coordinate business logic
    // ... domain logic using agg1 and agg2 ...

    // 3. Return result
    return result, nil
}
```

---

### 1.4 Template: Repository Interface

**Файл:** `docs/templates/domain/repository.go.template`

```go
package {{DOMAIN}}

import (
    "github.com/barsukov/quiz-sprint/backend/internal/domain/shared"
)

// Repository defines the interface for {{AGGREGATE}} persistence
//
// NOTE: Interface is defined in DOMAIN layer (dependency inversion)
// Implementation is in INFRASTRUCTURE layer
//
// IMPORTANT: No context.Context - domain layer is pure!
// Infrastructure implementations add context internally
type Repository interface {
    // FindByID retrieves an aggregate by its ID
    FindByID(id {{AGGREGATE}}ID) (*{{AGGREGATE}}, error)

    // FindAll retrieves all aggregates
    FindAll() ([]{{AGGREGATE}}, error)

    // Save persists an aggregate (create or update)
    Save(agg *{{AGGREGATE}}) error

    // Delete removes an aggregate by ID
    Delete(id {{AGGREGATE}}ID) error

    // Domain-specific queries
    FindActiveByUser(userID shared.UserID) (*{{AGGREGATE}}, error)
}
```

**Правила для Repository:**
- ✅ Только методы для Aggregate Root (не для entities)
- ✅ Без `context.Context` (добавляется в infrastructure)
- ✅ Возвращает domain types, не DTOs
- ❌ Без SQL, без JSON, без HTTP

---

### 1.5 Template: Domain Events

**Файл:** `docs/templates/domain/events.go.template`

```go
package {{DOMAIN}}

import (
    "github.com/barsukov/quiz-sprint/backend/internal/domain/shared"
)

// Event is a marker interface for domain events
type Event interface {
    EventName() string
    OccurredAt() int64
}

// {{AGGREGATE}}CreatedEvent is emitted when {{AGGREGATE}} is created
type {{AGGREGATE}}CreatedEvent struct {
    aggregateID {{AGGREGATE}}ID
    occurredAt  int64
}

// New{{AGGREGATE}}CreatedEvent creates a new event
func New{{AGGREGATE}}CreatedEvent(id {{AGGREGATE}}ID, occurredAt int64) {{AGGREGATE}}CreatedEvent {
    return {{AGGREGATE}}CreatedEvent{
        aggregateID: id,
        occurredAt:  occurredAt,
    }
}

func (e {{AGGREGATE}}CreatedEvent) EventName() string { return "{{domain}}.{{aggregate}}.created" }
func (e {{AGGREGATE}}CreatedEvent) OccurredAt() int64  { return e.occurredAt }
func (e {{AGGREGATE}}CreatedEvent) AggregateID() {{AGGREGATE}}ID { return e.aggregateID }

// SomethingHappenedEvent is emitted when something happens
type SomethingHappenedEvent struct {
    aggregateID {{AGGREGATE}}ID
    // Event-specific data
    data        string
    occurredAt  int64
}

func NewSomethingHappenedEvent(id {{AGGREGATE}}ID, data string, occurredAt int64) SomethingHappenedEvent {
    return SomethingHappenedEvent{
        aggregateID: id,
        data:        data,
        occurredAt:  occurredAt,
    }
}

func (e SomethingHappenedEvent) EventName() string { return "{{domain}}.something.happened" }
func (e SomethingHappenedEvent) OccurredAt() int64  { return e.occurredAt }
```

**Правила для Events:**
- ✅ Прошедшее время: `GameStartedEvent`, `AnswerSubmittedEvent`
- ❌ Не настоящее: `StartGameEvent`, `SubmitAnswerEvent`
- ✅ Иммутабельны
- ✅ Содержат факты, произошедшие в domain

---

### 1.6 Template: Domain Errors

**Файл:** `docs/templates/domain/errors.go.template`

```go
package {{DOMAIN}}

import "errors"

// Domain errors represent business rule violations
// These are PUBLIC - used by application layer

var (
    // Validation errors
    ErrInvalid{{AGGREGATE}}ID = errors.New("invalid {{aggregate}} id")

    // Business rule violations
    ErrPreconditionFailed = errors.New("precondition failed")
    ErrInvariantViolated  = errors.New("invariant violated")

    // Not found errors
    Err{{AGGREGATE}}NotFound = errors.New("{{aggregate}} not found")
)
```

**Правила для Errors:**
- ✅ Описывают нарушения бизнес-правил
- ✅ Без упоминания технических деталей (SQL, HTTP)
- ✅ Используют domain terminology из GLOSSARY.md
- ✅ Public (экспортируемые)

---

## 2️⃣ Testing Strategy для Domain Layer

### 2.1 Что тестируем

```
Domain Layer Tests = 80% всех тестов

✅ Aggregate business logic
✅ Value Object validation
✅ Domain Service coordination
✅ Domain Events emission
```

### 2.2 Template: Aggregate Tests

```go
package {{DOMAIN}}_test

import (
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
    "github.com/barsukov/quiz-sprint/backend/internal/domain/game_modes/{{DOMAIN}}"
)

// Test naming: Test{Type}_{Method}_{Scenario}
// Examples:
// - TestMarathonGame_AnswerQuestion_CorrectAnswer
// - TestMarathonGame_AnswerQuestion_WrongAnswer_LosesLife
// - TestMarathonGame_AnswerQuestion_NoLives_ReturnsError

func TestNew{{AGGREGATE}}_Success(t *testing.T) {
    // Arrange
    id := {{DOMAIN}}.New{{AGGREGATE}}ID()
    // ... other params

    // Act
    agg, err := {{DOMAIN}}.New{{AGGREGATE}}(id, ..., time.Now().Unix())

    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, agg)
    assert.Equal(t, id, agg.ID())
}

func TestNew{{AGGREGATE}}_InvalidID_ReturnsError(t *testing.T) {
    // Arrange
    invalidID := {{DOMAIN}}.{{AGGREGATE}}ID{} // Zero value

    // Act
    agg, err := {{DOMAIN}}.New{{AGGREGATE}}(invalidID, ..., time.Now().Unix())

    // Assert
    assert.ErrorIs(t, err, {{DOMAIN}}.ErrInvalid{{AGGREGATE}}ID)
    assert.Nil(t, agg)
}

func Test{{AGGREGATE}}_BusinessMethod_Success(t *testing.T) {
    // Arrange
    agg := createValidAggregate(t)

    // Act
    err := agg.DoSomething(...)

    // Assert
    assert.NoError(t, err)
    // Assert state changes
    // Assert events emitted
    events := agg.Events()
    assert.Len(t, events, 1)
    assert.Equal(t, "{{domain}}.something.happened", events[0].EventName())
}

func Test{{AGGREGATE}}_BusinessMethod_ViolatesInvariant_ReturnsError(t *testing.T) {
    // Arrange
    agg := createInvalidStateAggregate(t)

    // Act
    err := agg.DoSomething(...)

    // Assert
    assert.ErrorIs(t, err, {{DOMAIN}}.ErrInvariantViolated)
}

// Test helpers
func createValidAggregate(t *testing.T) *{{DOMAIN}}.{{AGGREGATE}} {
    t.Helper()

    id := {{DOMAIN}}.New{{AGGREGATE}}ID()
    agg, err := {{DOMAIN}}.New{{AGGREGATE}}(id, ..., time.Now().Unix())
    assert.NoError(t, err)

    return agg
}
```

### 2.3 Template: Value Object Tests

```go
func TestNew{{VALUE_OBJECT}}_Valid_Success(t *testing.T) {
    // Arrange
    validValue := "valid-value"

    // Act
    vo, err := {{DOMAIN}}.New{{VALUE_OBJECT}}(validValue)

    // Assert
    assert.NoError(t, err)
    assert.Equal(t, validValue, vo.Value())
}

func TestNew{{VALUE_OBJECT}}_Invalid_ReturnsError(t *testing.T) {
    // Arrange
    invalidValue := ""

    // Act
    vo, err := {{DOMAIN}}.New{{VALUE_OBJECT}}(invalidValue)

    // Assert
    assert.ErrorIs(t, err, {{DOMAIN}}.ErrInvalid{{VALUE_OBJECT}})
}

func Test{{VALUE_OBJECT}}_Transform_Immutable(t *testing.T) {
    // Arrange
    original, _ := {{DOMAIN}}.New{{VALUE_OBJECT}}("original")

    // Act
    transformed := original.Transform()

    // Assert
    assert.NotEqual(t, original.Value(), transformed.Value())
    assert.Equal(t, "original", original.Value()) // Original unchanged!
}
```

### 2.4 Test Coverage Requirements

```
Domain Layer Coverage: >= 80%

✅ Aggregates: 90%+ (критичны!)
✅ Value Objects: 85%+
✅ Domain Services: 80%+
✅ Events: 70%+ (простые структуры)
```

---

## 3️⃣ Error Handling в Domain Layer

### 3.1 Стратегия

```
Domain Layer Errors = Business Rule Violations

✅ DO: errors.New("no lives remaining")
❌ DON'T: errors.New("sql: no rows in result set")

✅ DO: ErrGameNotActive
❌ DON'T: ErrDatabaseConnectionFailed
```

### 3.2 Категории Domain Errors

```go
// 1. Validation Errors (invalid input)
var (
    ErrInvalidGameID      = errors.New("invalid game id")
    ErrInvalidUserID      = errors.New("invalid user id")
    ErrInvalidCategory    = errors.New("invalid category")
)

// 2. Business Rule Violations (preconditions not met)
var (
    ErrNoLivesRemaining   = errors.New("no lives remaining")
    ErrGameNotActive      = errors.New("game is not active")
    ErrHintNotAvailable   = errors.New("hint not available")
)

// 3. Not Found (entity doesn't exist)
var (
    ErrGameNotFound     = errors.New("game not found")
    ErrQuestionNotFound = errors.New("question not found")
)

// 4. Already Exists (uniqueness violation)
var (
    ErrGameAlreadyExists = errors.New("game already exists")
)
```

### 3.3 Error Wrapping

```go
// ✅ ПРАВИЛЬНО - используй fmt.Errorf с %w
func (s *DomainService) DoSomething(id GameID) error {
    game, err := s.repo.FindByID(id)
    if err != nil {
        return fmt.Errorf("failed to find game: %w", err)
    }
    // ...
}

// ❌ НЕПРАВИЛЬНО - не используй %v (теряется цепочка)
func (s *DomainService) DoSomething(id GameID) error {
    game, err := s.repo.FindByID(id)
    if err != nil {
        return fmt.Errorf("failed to find game: %v", err) // WRONG!
    }
    // ...
}
```

### 3.4 Error Testing

```go
func TestGame_AnswerQuestion_NoLives_ReturnsError(t *testing.T) {
    // Arrange
    game := createGameWithNoLives(t)

    // Act
    _, err := game.AnswerQuestion(...)

    // Assert
    assert.ErrorIs(t, err, solo_marathon.ErrNoLivesRemaining)
}
```

---

## 4️⃣ Domain-Specific Best Practices

### 4.1 Aggregate Boundaries

**Правило:** Одна транзакция = один aggregate

```go
// ✅ ПРАВИЛЬНО - операция внутри одного aggregate
func (mg *MarathonGame) AnswerQuestion(...) error {
    // All changes within MarathonGame aggregate
    mg.currentStreak++
    mg.lives = mg.lives.LoseLife()
    return nil
}

// ❌ НЕПРАВИЛЬНО - изменяет несколько aggregates
func (mg *MarathonGame) AnswerQuestion(user *User) error {
    mg.currentStreak++
    user.UpdateStats(...)  // WRONG! Crossing aggregate boundary
    return nil
}
```

**Решение:** Используй Domain Events для координации между aggregates
```go
// ✅ ПРАВИЛЬНО
func (mg *MarathonGame) AnswerQuestion(...) error {
    mg.currentStreak++

    // Emit event - User aggregate will handle separately
    mg.events = append(mg.events, NewCorrectAnswerEvent(...))

    return nil
}
```

### 4.2 Инварианты Aggregate

**Инвариант** = правило, которое ВСЕГДА истинно для aggregate

```go
// MarathonGame invariants:
// - currentLives >= 0 AND currentLives <= maxLives
// - currentStreak >= 0
// - isActive = false IF currentLives = 0

func (mg *MarathonGame) validateInvariants() error {
    if mg.lives.Current() < 0 || mg.lives.Current() > mg.lives.Max() {
        return ErrInvariantViolated
    }

    if mg.currentStreak < 0 {
        return ErrInvariantViolated
    }

    if mg.lives.Current() == 0 && mg.isActive {
        return ErrInvariantViolated
    }

    return nil
}
```

### 4.3 Value Object Composition

```go
// ✅ ХОРОШО - композиция value objects
type MarathonGame struct {
    id       GameID
    lives    LivesSystem      // Value Object
    hints    HintsSystem      // Value Object
    difficulty DifficultyProgression  // Value Object
}

// Value Objects можно заменять целиком (immutable)
func (mg *MarathonGame) UseHint(hintType HintType) error {
    // Replace entire value object
    mg.hints = mg.hints.UseHint(hintType)
    return nil
}
```

### 4.4 Domain Events Best Practices

```go
// ✅ ПРАВИЛЬНО - событие описывает ЧТО произошло
type CorrectAnswerEvent struct {
    gameID        GameID
    questionID    QuestionID
    currentStreak int
    occurredAt    int64
}

// ❌ НЕПРАВИЛЬНО - событие содержит команду
type UpdateStreakEvent struct {  // WRONG! Event name is imperative
    gameID GameID
    newStreak int
}
```

**Правила Events:**
- ✅ Прошедшее время (AnswerSubmittedEvent)
- ✅ Иммутабельны
- ✅ Содержат всю информацию о произошедшем
- ❌ Не содержат команды/инструкции

### 4.5 Timestamps в Domain

```go
// ✅ ПРАВИЛЬНО - int64 Unix timestamp
type MarathonGame struct {
    startedAt int64  // Unix timestamp
    endedAt   int64
}

// ❌ НЕПРАВИЛЬНО - time.Time (external dependency)
type MarathonGame struct {
    startedAt time.Time  // WRONG! Breaks domain purity
}
```

**Почему int64?**
- ✅ Domain остаётся pure (без external dependencies)
- ✅ Легко сериализуется в JSON/DB
- ✅ Application layer может конвертировать в time.Time

---

## 5️⃣ ADR (Architecture Decision Records)

### 5.1 Формат ADR

```markdown
# ADR-XXX: Title

**Статус:** [Proposed | Accepted | Deprecated | Superseded]
**Дата:** YYYY-MM-DD
**Авторы:** Team

## Контекст

Описание проблемы/ситуации

## Решение

Что мы решили делать

## Альтернативы

1. Alternative 1
   - Плюсы: ...
   - Минусы: ...

2. Alternative 2
   - Плюсы: ...
   - Минусы: ...

## Последствия

**Плюсы:**
- ✅ ...

**Минусы:**
- ⚠️ ...
```

### 5.2 Пример ADR

```markdown
# ADR-001: Use Shared Kernel for Gameplay Logic

**Статус:** Accepted
**Дата:** 2026-01-25

## Контекст

У нас 4 игровых режима: Marathon, Daily Challenge, Duel, Party.

Каждый режим:
- Показывает вопросы
- Принимает ответы
- Считает базовые очки

Без shared kernel → дублирование кода 4 раза.

## Решение

Создать `kernel.QuizGameplaySession` с чистой логикой геймплея.

Каждый режим композирует kernel:
```go
type MarathonGame struct {
    session *kernel.QuizGameplaySession  // Shared
    lives   LivesSystem                  // Mode-specific
}
```

## Альтернативы

1. **Дублировать логику** ❌
   - Минусы: Code duplication, hard to maintain

2. **Inheritance** ❌
   - Минусы: Go doesn't support inheritance

3. **Shared Kernel** ✅

## Последствия

**Плюсы:**
- ✅ Нет дублирования
- ✅ Kernel тестируется отдельно
- ✅ Режимы независимы

**Минусы:**
- ⚠️ Kernel не должен знать о mode-specific логике
```

---

## 6️⃣ Bounded Context Map

```markdown
# Bounded Contexts в Quiz Sprint

## Контексты

1. **Quiz Context** (`domain/quiz/`)
   - Ответственность: Контент вопросов, категории
   - Aggregates: Quiz, Category
   - Entities: Question, Answer

2. **User Context** (`domain/user/`)
   - Ответственность: Профили пользователей
   - Aggregates: User

3. **Shared Kernel** (`domain/kernel/`)
   - Ответственность: Чистая логика геймплея
   - Используется: Всеми game modes

4. **Marathon Context** (`domain/game_modes/solo_marathon/`)
   - Ответственность: Solo Marathon режим
   - Aggregates: MarathonGame
   - Зависит от: Quiz Context, User Context, Shared Kernel

5. **Daily Challenge Context** (`domain/game_modes/daily_challenge/`)
   - Ответственность: Daily Challenge режим
   - Aggregates: DailyGame, DailyQuiz
   - Зависит от: Quiz Context, User Context, Shared Kernel

## Relationships

```
Marathon Context ---(uses)---> Shared Kernel
Daily Context    ---(uses)---> Shared Kernel
Duel Context     ---(uses)---> Shared Kernel
Party Context    ---(uses)---> Shared Kernel

Marathon Context ---(uses)---> Quiz Context
Daily Context    ---(uses)---> Quiz Context

Marathon Context ---(uses)---> User Context
Daily Context    ---(uses)---> User Context
```

**Правило:** Contexts не зависят друг от друга напрямую!
Только через Shared Kernel или Domain Events.
```

---

## 7️⃣ Checklist перед генерацией кода

### Pre-Generation Checklist

```
Перед написанием кода ответь на вопросы:

Domain Understanding:
□ Прочитал ли я GLOSSARY.md?
□ Прочитал ли я соответствующие ADR?
□ Знаю ли я bounded context для этого кода?
□ Понимаю ли я business invariants?

Code Structure:
□ Это aggregate root, entity, или value object?
□ Какие бизнес-правила нужно реализовать?
□ Какие domain events должны эмитироваться?
□ Нужен ли domain service?

Testing:
□ Какие unit tests нужно написать?
□ Какие edge cases нужно покрыть?
□ Какие errors могут возникнуть?

Purity:
□ Нет ли external dependencies (time.Time, context.Context)?
□ Value Objects иммутабельны?
□ Aggregate boundaries соблюдены?
```

### Post-Generation Checklist

```
После генерации кода проверь:

Code Quality:
□ Следует naming conventions из GLOSSARY.md?
□ Использует правильные термины?
□ Нет anti-patterns?

Domain Purity:
□ Нет time.Time, context.Context?
□ Нет JSON tags?
□ Нет database imports?
□ Нет HTTP imports?

Business Logic:
□ Бизнес-правила в aggregate methods?
□ Инварианты соблюдены?
□ Domain events эмитируются?

Testing:
□ Unit tests написаны?
□ Coverage >= 80%?
□ Edge cases покрыты?
```

---

## 8️⃣ Common Patterns & Anti-Patterns

### ✅ Good Patterns

```go
// 1. Factory Method
func NewMarathonGame(...) (*MarathonGame, error) {
    // Validation + creation
}

// 2. Reconstruct Method (for DB loading)
func ReconstructMarathonGame(...) *MarathonGame {
    // No validation
}

// 3. Immutable Value Objects
func (ls LivesSystem) LoseLife() LivesSystem {
    return LivesSystem{currentLives: ls.currentLives - 1}
}

// 4. Domain Events Collection
func (mg *MarathonGame) AnswerQuestion(...) error {
    // ... business logic
    mg.events = append(mg.events, NewEvent(...))
}

// 5. Getters (no setters!)
func (mg *MarathonGame) ID() GameID { return mg.id }
```

### ❌ Anti-Patterns

```go
// 1. Direct field mutation from outside ❌
game.currentStreak = 10  // WRONG!
// ✅ Use: game.AnswerQuestion(...)

// 2. Mutable value objects ❌
func (ls *LivesSystem) LoseLife() {
    ls.currentLives--  // WRONG!
}

// 3. External dependencies in domain ❌
import "time"
type Game struct {
    startedAt time.Time  // WRONG!
}

// 4. Crossing aggregate boundaries ❌
func (game *Game) Update(user *User) {
    user.stats.Update(...)  // WRONG!
}

// 5. No validation in constructors ❌
func NewGame(...) *Game {
    return &Game{...}  // WRONG! No validation
}
```

---

## 9️⃣ Quick Reference Card

```
┌─────────────────────────────────────────────────────────┐
│             Domain Layer Quick Reference                │
├─────────────────────────────────────────────────────────┤
│                                                         │
│  Aggregate Root:                                        │
│  • Factory method: New{Aggregate}(...)                 │
│  • Reconstruct:    Reconstruct{Aggregate}(...)         │
│  • Business methods in aggregate                        │
│  • Emit domain events                                   │
│  • Return errors, not panic                             │
│                                                         │
│  Value Object:                                          │
│  • IMMUTABLE (methods return new value object)         │
│  • Validation in constructor                            │
│  • No ID (defined by attributes)                        │
│                                                         │
│  Domain Service:                                        │
│  • Use when logic doesn't belong to one aggregate      │
│  • Stateless                                            │
│  • Coordinates aggregates                               │
│                                                         │
│  Repository:                                            │
│  • Interface in domain                                  │
│  • Only for aggregate roots                             │
│  • No context.Context                                   │
│                                                         │
│  Domain Events:                                         │
│  • Past tense (GameStartedEvent)                       │
│  • Immutable                                            │
│  • Collected in aggregate, cleared on Events()         │
│                                                         │
│  Errors:                                                │
│  • Business rule violations                             │
│  • No technical details                                 │
│  • Use domain terminology                               │
│                                                         │
│  NO in Domain:                                          │
│  ❌ time.Time (use int64)                               │
│  ❌ context.Context (add in infrastructure)             │
│  ❌ JSON tags (use in DTOs)                             │
│  ❌ SQL, HTTP, external libs                            │
│                                                         │
└─────────────────────────────────────────────────────────┘
```

---

## 🎯 Example: Complete Domain Implementation

Полный пример реализации см. в:
- `backend/internal/domain/quiz/` - готовая реализация
- `backend/internal/domain/kernel/` - shared kernel пример
- `backend/internal/domain/classic_mode/` - пример mode-specific aggregate

---

## 📚 Дополнительные ресурсы

- **GLOSSARY.md** - Единый словарь терминов
- **docs/adr/** - Architecture Decision Records
- **docs/templates/domain/** - Code templates
- **Domain-Driven Design** by Eric Evans
- **Implementing Domain-Driven Design** by Vaughn Vernon

---

**Вопросы?** Создай issue с меткой `domain-layer` или `llm-code-quality`
