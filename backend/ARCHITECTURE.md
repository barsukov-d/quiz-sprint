# Architecture Guidelines: DDD + Clean Architecture

This document defines the architectural rules for Quiz Sprint Backend. **ALL code MUST follow these rules.**

## Core Principles

### 1. The Dependency Rule (Clean Architecture)

**Dependencies ONLY point inward:**

```
External (Fiber, HTTP)
    ↓ depends on
Infrastructure (Repositories impl, HTTP handlers)
    ↓ depends on
Application (Use Cases, DTOs)
    ↓ depends on
Domain (Entities, Value Objects, Interfaces)
    ↑ NOTHING depends on domain
```

**Rule:** Domain layer has ZERO external dependencies. Not even `context.Context` from stdlib.

### 2. Domain-Driven Design Tactical Patterns

**Always use:**
- **Entities** - objects with identity (Quiz, Question)
- **Value Objects** - immutable objects without identity (QuizID, Score, Email)
- **Aggregates** - clusters of entities with one root (Quiz is aggregate root)
- **Domain Events** - something that happened (QuizStarted, AnswerSubmitted)
- **Repository Interfaces** - defined in domain, implemented in infrastructure
- **Domain Services** - when logic doesn't belong to entity (e.g., QuizScoringService)

## Layer Rules

### Domain Layer (`internal/domain/`)

**Allowed:**
- ✅ Pure Go structs and interfaces
- ✅ Business logic (methods on entities)
- ✅ Validation (invariants)
- ✅ Domain errors
- ✅ Repository interfaces (contracts)
- ✅ Domain events (structs)

**NOT Allowed:**
- ❌ `context.Context`
- ❌ Database imports (sql, gorm, etc)
- ❌ HTTP imports (fiber, net/http)
- ❌ External libraries (except uuid for IDs)
- ❌ JSON/XML tags
- ❌ Framework-specific code

**Example Structure:**
```
domain/
├── quiz/
│   ├── aggregate.go        # Quiz aggregate root
│   ├── entity.go           # Question, Answer entities
│   ├── value_object.go     # QuizID, Score, TimeLimit
│   ├── events.go           # QuizStarted, AnswerSubmitted
│   ├── errors.go           # Domain errors
│   ├── repository.go       # Repository interface
│   └── service.go          # Domain services (if needed)
└── shared/
    └── value_objects.go    # Shared VOs (UserID, etc)
```

**Rules:**
1. All IDs are Value Objects (not raw strings/UUIDs)
2. All entities have business methods (not anemic models)
3. Aggregates enforce invariants (e.g., Quiz.CanStart())
4. Use factory methods (NewQuiz, NewQuestion)
5. Return domain errors (not generic errors)

### Application Layer (`internal/application/`)

**Purpose:** Orchestrate domain objects to fulfill use cases.

**Allowed:**
- ✅ Use Case structs
- ✅ Input/Output DTOs (data transfer objects)
- ✅ Orchestration logic
- ✅ Transaction coordination
- ✅ Event publishing
- ✅ `context.Context` (for timeouts, cancellation)

**NOT Allowed:**
- ❌ HTTP concerns (request/response)
- ❌ Database implementation details
- ❌ Business logic (belongs in domain)
- ❌ Direct access to external services

**Example Structure:**
```
application/
├── quiz/
│   ├── start_quiz.go       # StartQuizUseCase + DTOs
│   ├── submit_answer.go    # SubmitAnswerUseCase + DTOs
│   └── get_leaderboard.go  # GetLeaderboardUseCase + DTOs
└── common/
    └── dto.go              # Common DTOs
```

**Use Case Pattern:**
```go
// Input DTO
type StartQuizInput struct {
    QuizID string
    UserID string
}

// Output DTO
type StartQuizOutput struct {
    SessionID string
    Quiz      QuizDTO
    Error     error  // Optional: for Result pattern
}

// Use Case
type StartQuizUseCase struct {
    quizRepo    domain.QuizRepository
    sessionRepo domain.SessionRepository
    eventBus    domain.EventBus
}

func (uc *StartQuizUseCase) Execute(ctx context.Context, input StartQuizInput) (StartQuizOutput, error) {
    // 1. Validate & convert to domain types
    // 2. Load aggregates
    // 3. Execute business logic (call domain methods)
    // 4. Persist changes
    // 5. Publish events
    // 6. Return DTO (not domain models!)
}
```

**Rules:**
1. Use Cases are the ONLY entry point to domain logic
2. Input/Output are always DTOs (never domain models)
3. Use Cases don't contain business logic (delegate to domain)
4. Use Cases coordinate multiple aggregates
5. Use Cases publish domain events

### Infrastructure Layer (`internal/infrastructure/`)

**Purpose:** Implement technical details (HTTP, DB, cache, etc).

**Allowed:**
- ✅ HTTP handlers (Fiber)
- ✅ Repository implementations (PostgreSQL, in-memory)
- ✅ External service clients
- ✅ Serialization (JSON)
- ✅ Framework-specific code

**Example Structure:**
```
infrastructure/
├── http/
│   ├── handlers/
│   │   └── quiz_handler.go    # Thin adapter
│   ├── dto/
│   │   └── request.go          # HTTP-specific DTOs
│   └── routes/
│       └── routes.go
├── persistence/
│   ├── postgres/
│   │   └── quiz_repository.go # PostgreSQL implementation
│   └── memory/
│       └── quiz_repository.go # In-memory implementation
└── messaging/
    └── event_bus.go            # Event bus implementation
```

**Handler Pattern (Thin Adapter):**
```go
type QuizHandler struct {
    startQuizUC *application.StartQuizUseCase
}

func (h *QuizHandler) StartQuiz(c *fiber.Ctx) error {
    // 1. Parse HTTP request
    var req StartQuizHTTPRequest
    if err := c.BodyParser(&req); err != nil {
        return fiber.NewError(fiber.StatusBadRequest, "Invalid request")
    }

    // 2. Convert to Use Case Input (HTTP → Application)
    input := application.StartQuizInput{
        QuizID: c.Params("id"),
        UserID: req.UserID,
    }

    // 3. Execute Use Case
    ctx := c.Context()
    output, err := h.startQuizUC.Execute(ctx, input)
    if err != nil {
        return h.mapError(err)  // Domain error → HTTP error
    }

    // 4. Return HTTP response (Application → HTTP)
    return c.Status(fiber.StatusCreated).JSON(fiber.Map{
        "data": output,
    })
}
```

**Rules:**
1. Handlers are THIN adapters (no business logic)
2. Convert HTTP → Application DTOs
3. Map domain errors to HTTP status codes
4. Repository implementations implement domain interfaces
5. Use dependency injection (don't create dependencies inside)

## Tactical Patterns

### Value Objects

**When to use:** Always for identifiers, measurements, or descriptive objects.

**Rules:**
1. Immutable (no setters)
2. No identity (equality by value)
3. Self-validating (constructor validates)
4. Use factory methods (NewQuizID)

**Example:**
```go
// domain/quiz/value_object.go

type QuizID struct {
    value uuid.UUID
}

func NewQuizID(value string) (QuizID, error) {
    if value == "" {
        return QuizID{}, ErrInvalidQuizID
    }

    id, err := uuid.Parse(value)
    if err != nil {
        return QuizID{}, ErrInvalidQuizID
    }

    return QuizID{value: id}, nil
}

func (id QuizID) String() string {
    return id.value.String()
}

func (id QuizID) Equals(other QuizID) bool {
    return id.value == other.value
}

// Value Objects are comparable
func (id QuizID) IsZero() bool {
    return id.value == uuid.Nil
}
```

**Common Value Objects:**
- IDs: `QuizID`, `QuestionID`, `UserID`, `SessionID`
- Measurements: `Score`, `TimeLimit`, `Points`
- Descriptive: `QuizTitle`, `QuestionText`, `Email`

### Entities

**When to use:** Objects with identity that change over time.

**Rules:**
1. Have identity (ID value object)
2. Have lifecycle (created, modified, deleted)
3. Contain business logic (not anemic)
4. Validate invariants

**Example:**
```go
// domain/quiz/entity.go

type Question struct {
    id       QuestionID
    quizID   QuizID
    text     QuestionText
    answers  []Answer
    points   Points
    position int
}

// Factory method
func NewQuestion(id QuestionID, quizID QuizID, text QuestionText, points Points) (*Question, error) {
    if text.IsEmpty() {
        return nil, ErrInvalidQuestionText
    }

    if points.Value() <= 0 {
        return nil, ErrInvalidPoints
    }

    return &Question{
        id:      id,
        quizID:  quizID,
        text:    text,
        points:  points,
        answers: []Answer{},
    }, nil
}

// Business method
func (q *Question) AddAnswer(answer Answer) error {
    if len(q.answers) >= 4 {
        return ErrTooManyAnswers
    }

    q.answers = append(q.answers, answer)
    return nil
}

// Getters (no setters - modify through business methods)
func (q *Question) ID() QuestionID { return q.id }
func (q *Question) Text() QuestionText { return q.text }
```

### Aggregates

**When to use:** Group of entities that must be consistent together.

**Rules:**
1. One aggregate root (e.g., Quiz)
2. Only root is referenced from outside
3. Root enforces invariants for whole aggregate
4. Root is unit of persistence (save entire aggregate)
5. Small aggregates (prefer 1-3 entities)

**Example:**
```go
// domain/quiz/aggregate.go

type Quiz struct {
    id           QuizID
    title        QuizTitle
    questions    []Question     // Part of aggregate
    timeLimit    TimeLimit
    passingScore PassingScore
    createdAt    time.Time
    updatedAt    time.Time
}

// Aggregate root enforces invariants
func (q *Quiz) AddQuestion(question Question) error {
    // Invariant: Quiz must have at least 1 question to start
    if len(q.questions) >= 50 {
        return ErrTooManyQuestions
    }

    q.questions = append(q.questions, question)
    q.updatedAt = time.Now()
    return nil
}

// Aggregate root protects internal state
func (q *Quiz) CanStart() error {
    if len(q.questions) == 0 {
        return ErrNoQuestions
    }

    if q.timeLimit.IsZero() {
        return ErrInvalidTimeLimit
    }

    return nil
}

// Return copies, not internal state
func (q *Quiz) Questions() []Question {
    copies := make([]Question, len(q.questions))
    copy(copies, q.questions)
    return copies
}
```

**Aggregate boundaries:**
- `Quiz` (root) → `Question` → `Answer`
- `QuizSession` (root) → `UserAnswer`
- `User` (root) → `UserProfile`

### Domain Events

**When to use:** Something happened in the domain that other parts care about.

**Rules:**
1. Past tense naming (QuizStarted, AnswerSubmitted)
2. Immutable (no setters)
3. Contain all necessary data
4. Created by aggregates
5. Published by use cases

**Example:**
```go
// domain/quiz/events.go

type QuizStarted struct {
    quizID     QuizID
    sessionID  SessionID
    userID     UserID
    startedAt  time.Time
}

func NewQuizStarted(quizID QuizID, sessionID SessionID, userID UserID) QuizStarted {
    return QuizStarted{
        quizID:    quizID,
        sessionID: sessionID,
        userID:    userID,
        startedAt: time.Now(),
    }
}

// Getters only
func (e QuizStarted) QuizID() QuizID { return e.quizID }
func (e QuizStarted) SessionID() SessionID { return e.sessionID }
func (e QuizStarted) OccurredAt() time.Time { return e.startedAt }

// Implement Event interface
func (e QuizStarted) EventName() string {
    return "quiz.started"
}
```

**Event Flow:**
1. Aggregate method creates event (but doesn't publish)
2. Use case collects events from aggregate
3. Use case publishes events via EventBus
4. Event handlers react (async)

### Repository Pattern

**Rules:**
1. Interface defined in domain
2. Implementation in infrastructure
3. Work with aggregates (not individual entities)
4. No query methods in domain (use read models/CQRS)

**Example:**
```go
// domain/quiz/repository.go

type QuizRepository interface {
    // Commands (write model)
    Save(quiz *Quiz) error
    Delete(id QuizID) error

    // Queries (read model) - return aggregates
    FindByID(id QuizID) (*Quiz, error)

    // No: FindByTitleContaining - use read model
    // No: context.Context - keep domain pure
}

// infrastructure/persistence/postgres/quiz_repository.go

type PostgresQuizRepository struct {
    db *sql.DB
}

func (r *PostgresQuizRepository) Save(quiz *Quiz) error {
    ctx := context.Background()  // Infrastructure creates context

    // Save aggregate as unit
    tx, err := r.db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    defer tx.Rollback()

    // Save quiz
    if err := r.saveQuiz(tx, quiz); err != nil {
        return err
    }

    // Save questions (part of aggregate)
    for _, question := range quiz.Questions() {
        if err := r.saveQuestion(tx, question); err != nil {
            return err
        }
    }

    return tx.Commit()
}
```

## Anti-Patterns to Avoid

### ❌ Anemic Domain Model
```go
// BAD: No business logic
type Quiz struct {
    ID        string
    Title     string
    Questions []Question
}

// Business logic in service (wrong layer!)
func (s *QuizService) CanStart(quiz Quiz) bool {
    return len(quiz.Questions) > 0
}
```

```go
// GOOD: Rich domain model
type Quiz struct {
    id        QuizID
    questions []Question
}

func (q *Quiz) CanStart() error {
    if len(q.questions) == 0 {
        return ErrNoQuestions
    }
    return nil
}
```

### ❌ Leaking Domain Models
```go
// BAD: Use case returns domain model
func (uc *StartQuizUseCase) Execute(input StartQuizInput) (*domain.Quiz, error) {
    quiz, _ := uc.repo.FindByID(input.QuizID)
    return quiz, nil  // Leaked!
}

// BAD: Handler uses domain model directly
func (h *Handler) GetQuiz(c *fiber.Ctx) error {
    quiz, _ := h.useCase.Execute(...)
    return c.JSON(quiz)  // Domain model in HTTP response!
}
```

```go
// GOOD: Use case returns DTO
type QuizDTO struct {
    ID    string
    Title string
}

func (uc *StartQuizUseCase) Execute(input StartQuizInput) (QuizDTO, error) {
    quiz, _ := uc.repo.FindByID(input.QuizID)
    return toDTO(quiz), nil  // Convert domain → DTO
}
```

### ❌ Framework in Domain
```go
// BAD: JSON tags in domain
type Quiz struct {
    ID    QuizID `json:"id"`     // ❌ Framework leak
    Title string `json:"title"`  // ❌
}

// BAD: context.Context in domain
type QuizRepository interface {
    FindByID(ctx context.Context, id QuizID) (*Quiz, error)  // ❌
}
```

```go
// GOOD: Pure domain
type Quiz struct {
    id    QuizID  // No tags
    title QuizTitle
}

type QuizRepository interface {
    FindByID(id QuizID) (*Quiz, error)  // No context
}

// Infrastructure adds context
type PostgresQuizRepository struct{}

func (r *PostgresQuizRepository) FindByID(id QuizID) (*Quiz, error) {
    ctx := context.Background()  // ✅ Infrastructure concern
    // ... query with ctx
}
```

### ❌ Business Logic in Use Cases
```go
// BAD: Business logic in use case
func (uc *StartQuizUseCase) Execute(input StartQuizInput) error {
    quiz, _ := uc.repo.FindByID(input.QuizID)

    // ❌ Business logic doesn't belong here!
    if len(quiz.Questions) == 0 {
        return errors.New("no questions")
    }

    if quiz.TimeLimit <= 0 {
        return errors.New("invalid time")
    }

    // ...
}
```

```go
// GOOD: Delegate to domain
func (uc *StartQuizUseCase) Execute(input StartQuizInput) error {
    quiz, _ := uc.repo.FindByID(input.QuizID)

    // ✅ Domain enforces business rules
    if err := quiz.CanStart(); err != nil {
        return err
    }

    // Use case only orchestrates
}
```

## Testing Strategy

### Domain Tests (Pure Unit Tests)
```go
func TestQuiz_CanStart(t *testing.T) {
    quiz := NewQuiz(...)

    err := quiz.CanStart()

    assert.Error(t, err)  // No mocks needed!
}
```

### Use Case Tests (With Mocks)
```go
func TestStartQuizUseCase_Execute(t *testing.T) {
    mockRepo := &MockQuizRepository{}
    useCase := NewStartQuizUseCase(mockRepo)

    output, err := useCase.Execute(ctx, input)

    assert.NoError(t, err)
    assert.NotEmpty(t, output.SessionID)
}
```

### Handler Tests (Integration)
```go
func TestQuizHandler_StartQuiz(t *testing.T) {
    app := fiber.New()
    // Setup routes

    req := httptest.NewRequest("POST", "/api/quiz/123/start", body)
    resp, _ := app.Test(req)

    assert.Equal(t, 201, resp.StatusCode)
}
```

## File Naming Conventions

```
domain/quiz/
├── aggregate.go         # Aggregate roots
├── entity.go            # Entities
├── value_object.go      # Value objects
├── events.go            # Domain events
├── errors.go            # Domain errors
├── repository.go        # Repository interface
└── service.go           # Domain services

application/quiz/
├── start_quiz.go        # UseCase + Input + Output DTO
├── submit_answer.go
└── dto.go               # Shared DTOs

infrastructure/
├── http/handlers/quiz_handler.go
├── persistence/postgres/quiz_repository.go
└── messaging/event_bus.go
```

## Summary: Quick Reference

| Concept | Location | Dependencies | Contains |
|---------|----------|--------------|----------|
| **Value Object** | domain | None | Validation, equality |
| **Entity** | domain | Value Objects | Identity, business methods |
| **Aggregate** | domain | Entities, VOs | Invariants, consistency |
| **Domain Event** | domain | Value Objects | Past tense, immutable |
| **Repository Interface** | domain | None | Contract only |
| **Use Case** | application | Domain interfaces | Orchestration, DTOs |
| **DTO** | application | None | Data transfer |
| **Repository Impl** | infrastructure | Domain interface | SQL, cache, etc |
| **Handler** | infrastructure | Use Cases | HTTP adapter |

## Enforcement

**These rules are NOT optional.** All code reviews MUST check:
- [ ] Domain has no external dependencies
- [ ] Value Objects for all IDs
- [ ] Use Cases use DTOs (no domain model leaks)
- [ ] Handlers are thin adapters
- [ ] Business logic in domain (not use cases)
- [ ] Repository interface in domain
- [ ] Domain events for important actions

**When in doubt:** Ask "Does this follow DDD + Clean Architecture?" If unsure, refactor.
