# Backend Refactoring: DDD + Clean Architecture

## Status: 🚧 IN PROGRESS

This document tracks the refactoring of Quiz Sprint Backend to follow strict DDD + Clean Architecture principles defined in `ARCHITECTURE.md`.

## Progress

### ✅ Completed

1. **Architecture Guidelines Created** (`ARCHITECTURE.md`)
   - Complete DDD + Clean Architecture rules
   - Anti-patterns documented
   - Examples for all patterns
   - Testing strategy defined

2. **Value Objects Created**
   - `shared/value_objects.go` - UserID, ID (base)
   - `quiz/value_objects.go` - QuizID, QuestionID, AnswerID, SessionID, QuizTitle, QuestionText, AnswerText, Points, TimeLimit, PassingScore
   - All IDs are now Value Objects (not raw UUIDs)
   - Self-validating (factory methods)
   - Immutable

3. **Domain Errors Expanded**
   - Added validation errors for all Value Objects
   - Clear, specific error messages

### 🚧 In Progress

4. **Refactor Domain Models**
   - [ ] Create `entity.go` for Question, Answer entities
   - [ ] Create `aggregate.go` for Quiz, QuizSession aggregates
   - [ ] Replace all `uuid.UUID` with Value Objects
   - [ ] Replace all `string` with proper Value Objects
   - [ ] Add business methods (not anemic models)
   - [ ] Add factory methods (NewQuiz, NewQuestion)
   - [ ] Enforce invariants in aggregate roots

### ⏳ TODO

5. **Domain Events**
   - [ ] Create `events.go` with:
     - `QuizStarted`
     - `AnswerSubmitted`
     - `QuizCompleted`
   - [ ] Add EventBus interface in domain
   - [ ] Aggregates collect events (but don't publish)

6. **Repository Refactoring**
   - [ ] Remove `context.Context` from domain interfaces
   - [ ] Update signatures to use Value Objects
   - [ ] Keep infrastructure implementations with context

7. **Application DTOs**
   - [ ] Create `application/quiz/dto.go`
   - [ ] Input DTOs for all use cases
   - [ ] Output DTOs for all use cases
   - [ ] Conversion functions (Domain ↔ DTO)

8. **Use Case Refactoring**
   - [ ] Update `StartQuizUseCase` to use DTOs
   - [ ] Update `SubmitAnswerUseCase` to use DTOs
   - [ ] Update `GetLeaderboardUseCase` to use DTOs
   - [ ] Add event publishing to use cases
   - [ ] Remove business logic (delegate to domain)

9. **Infrastructure Refactoring**
   - [ ] Update handlers to be thin adapters
   - [ ] HTTP Request → Application DTO conversion
   - [ ] Application DTO → HTTP Response conversion
   - [ ] Error mapping (Domain → HTTP status codes)
   - [ ] Update repository implementations

10. **EventBus Implementation**
    - [ ] Create EventBus interface in domain
    - [ ] In-memory implementation in infrastructure
    - [ ] Async event handling
    - [ ] Event handlers registration

## Migration Strategy

To avoid breaking everything at once:

1. **Phase 1: Domain Layer** (Current)
   - ✅ Value Objects
   - 🚧 Entities & Aggregates
   - ⏳ Domain Events

2. **Phase 2: Application Layer**
   - DTOs
   - Refactor Use Cases
   - EventBus integration

3. **Phase 3: Infrastructure Layer**
   - Thin handlers
   - Repository implementations
   - Event handlers

4. **Phase 4: Testing**
   - Unit tests for domain
   - Use case tests with mocks
   - Integration tests for handlers

## Breaking Changes

### For Frontend Developers

**After Phase 2 completes**, API responses will change slightly:

**Before:**
```json
{
  "data": {
    "id": "uuid-here",
    "title": "Quiz Title",
    "questions": [...]
  }
}
```

**After:**
```json
{
  "data": {
    "id": "uuid-here",
    "title": "Quiz Title",
    "description": "Quiz Description",
    "questionsCount": 10,
    "timeLimit": 30,
    "passingScore": 70
  }
}
```

DTOs will be cleaner and only expose what frontend needs (no internal domain structure leakage).

## Files Structure After Refactoring

```
backend/internal/
├── domain/
│   ├── shared/
│   │   ├── value_objects.go  ✅ Done
│   │   └── errors.go          ✅ Done
│   └── quiz/
│       ├── value_objects.go   ✅ Done
│       ├── entity.go          ⏳ TODO (Question, Answer)
│       ├── aggregate.go       ⏳ TODO (Quiz, QuizSession)
│       ├── events.go          ⏳ TODO
│       ├── errors.go          ✅ Done
│       ├── repository.go      ⏳ TODO (refactor)
│       └── event_bus.go       ⏳ TODO (interface)
│
├── application/
│   └── quiz/
│       ├── dto.go             ⏳ TODO
│       ├── start_quiz.go      ⏳ TODO (refactor)
│       ├── submit_answer.go   ⏳ TODO (refactor)
│       └── get_leaderboard.go ⏳ TODO (refactor)
│
└── infrastructure/
    ├── http/
    │   ├── handlers/
    │   │   ├── quiz_handler.go      ⏳ TODO (thin adapter)
    │   │   └── websocket_handler.go ⏳ TODO (refactor)
    │   ├── dto/
    │   │   └── request.go           ⏳ TODO (HTTP-specific DTOs)
    │   └── routes/
    │       └── routes.go            (no changes)
    ├── persistence/
    │   └── memory_repository.go     ⏳ TODO (refactor)
    └── messaging/
        └── event_bus.go             ⏳ TODO (implementation)
```

## Testing During Refactoring

Current backend uses in-memory repository, so:
- No database migrations needed
- Can test immediately after each phase
- Easy rollback if something breaks

## Timeline

- **Phase 1 (Domain)**: ~2-3 days
- **Phase 2 (Application)**: ~2-3 days
- **Phase 3 (Infrastructure)**: ~1-2 days
- **Phase 4 (Testing)**: ~1-2 days

**Total**: ~1-2 weeks for complete refactoring

## Notes

- Current backend still works (using old structure in `quiz.go`)
- New Value Objects created but not used yet
- No breaking changes until we update handlers
- Can continue working on frontend while refactoring happens

## Next Steps

1. Create `entity.go` with Question, Answer (using Value Objects)
2. Create `aggregate.go` with Quiz, QuizSession (using Value Objects)
3. Create Domain Events
4. Test domain layer (pure unit tests, no mocks)
5. Continue with Application layer...

## Questions?

See `ARCHITECTURE.md` for detailed guidelines.

For specific patterns, search for examples in that document.
