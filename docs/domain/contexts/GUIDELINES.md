# Guidelines for Domain Manifests (YAML)

This document defines the strict structure and rules for YAML files in `docs/domain/contexts/`. These manifests serve as the **Technical Blueprint** for domain implementation.

## Purpose

**Manifests describe HOW to implement** the technical structure of aggregates, entities, and domain logic.

- **Audience:** Developers, code generators, architects
- **Focus:** Data structures, types, methods, events, repositories
- **Language:** Technical (Go types, DDD patterns)
- **Format:** YAML (.yaml)

**NOT included:** Business logic explanations, user stories, formulas rationale.

---

## Manifest Schema

Each YAML file represents a single **Bounded Context**.

### 1. Bounded Context Metadata

```yaml
bounded_context:
  name: "ContextName"
  type: "Core | Supporting | Generic"
  description: "High-level responsibility of this context."
```

**Types:**
- **Core:** Critical business domain (ClassicGame, DuelGame, DailyMarathon)
- **Supporting:** Enables core (Identity, Leaderboard)
- **Generic:** Shared utilities (SharedKernel)

---

### 2. Aggregates

Aggregates are consistency boundaries and entry points for domain operations.

```yaml
aggregates:
  - name: "AggregateName"
    description: "What this aggregate represents and manages."

    state:
      - name: "FieldName"
        type: "GoType"
        description: "Purpose of this field."
      - name: "AnotherField"
        type: "uuid.UUID"
        description: "Reference to another entity."

    methods:
      - name: "MethodName"
        description: "What this method does."
        params:
          - { name: "paramName", type: "string" }
          - { name: "count", type: "int" }
        returns:
          - { type: "error" }
        logic: |
          Reference to business rule in specs/feature_name.md Section 3.
          Example: "Apply Multiplier based on Streak (see classic_game.md Rule 2)"

      - name: "AnotherMethod"
        params: []
        returns:
          - { type: "int" }
        logic: "Calculate total score using formula from spec."
```

**State Field Types:**
- `uuid.UUID` - for IDs
- `int64` - for Unix timestamps
- `int` - for counters, scores
- `float64` - for multipliers, percentages
- `string` - for text, enums
- `bool` - for flags
- `[]Type` - for collections

---

### 3. Entities

Entities are objects with unique identity but managed within an aggregate.

```yaml
entities:
  - name: "EntityName"
    description: "Purpose of this entity."
    state:
      - name: "ID"
        type: "uuid.UUID"
        description: "Unique identifier."
      - name: "Field"
        type: "string"
        description: "Entity attribute."
```

---

### 4. Value Objects

Immutable objects defined by their attributes, not identity.

```yaml
value_objects:
  - name: "ValueObjectName"
    type: "string | int | struct"
    description: "What this value represents."
    allowed_values: ["Value1", "Value2", "Value3"]  # For enums
    invariants:
      - "Must be greater than 0"
      - "Cannot exceed 100"
```

**Examples:**
- Enum: `GameStatus: ["Pending", "InProgress", "Completed"]`
- Struct: `TimeWindow { Start: int64, End: int64 }`
- Primitive: `Score: int (0-10000)`

---

### 5. Domain Events

Events published when significant domain changes occur.

```yaml
domain_events:
  - name: "EventName"
    description: "When and why this event is published."
    attributes:
      - { name: "AggregateID", type: "uuid.UUID" }
      - { name: "Timestamp", type: "int64" }
      - { name: "Data", type: "int" }
```

**Naming Convention:** Past tense verb (e.g., `GameStarted`, `AnswerSubmitted`, `PlayerDefeated`)

---

### 6. Repositories

Interfaces for persisting and retrieving aggregates.

```yaml
repositories:
  - name: "RepositoryName"
    aggregate: "AggregateName"
    methods:
      - name: "FindByID"
        params:
          - { name: "ctx", type: "context.Context" }
          - { name: "id", type: "uuid.UUID" }
        returns:
          - { type: "*Aggregate" }
          - { type: "error" }

      - name: "Save"
        params:
          - { name: "ctx", type: "context.Context" }
          - { name: "aggregate", type: "*Aggregate" }
        returns:
          - { type: "error" }
```

---

## Strict Rules

### 1. Naming Conventions

- **Aggregates, Entities, Value Objects:** `PascalCase`
- **Fields, Methods:** `PascalCase` (Go convention)
- **Method parameters:** `camelCase`
- **Events:** Past tense, `PascalCase` (e.g., `GameFinished`)

### 2. Type Safety (Go Compatibility)

✅ **Allowed Types:**
- `uuid.UUID`
- `int64` (timestamps)
- `int`
- `float64`
- `string`
- `bool`
- `[]Type` (slices)
- Custom types defined in `value_objects`

❌ **Forbidden:**
- `any` or `interface{}`
- `map[string]interface{}`
- `time.Time` (use `int64` Unix timestamps)
- Non-Go types

### 3. Domain Purity

❌ **NO infrastructure concerns:**
- No JSON/database tags (`json:`, `db:`, `bson:`)
- No HTTP status codes
- No SQL queries
- No external library types (`sql.DB`, `*http.Request`)

✅ **Pure domain:**
- Only business concepts
- Only domain types
- Only domain behavior

### 4. Logic Traceability

Every `methods.logic` block must reference the corresponding business rule from `docs/domain/specs/*.md`.

**Example:**
```yaml
logic: |
  1. Validate answer correctness
  2. Update Streak (see classic_game.md Rule 3)
  3. Calculate Multiplier (see classic_game.md Rule 2)
  4. Apply scoring formula (see classic_game.md Rule 1)
```

---

## Workflow

1. **Read Spec:** Understand business requirements from `docs/domain/specs/*.md`
2. **Design Manifest:** Create/update YAML with aggregates, entities, value objects
3. **Define State:** Add fields based on what needs to be tracked
4. **Define Methods:** Add domain methods that implement business rules
5. **Define Events:** Add events for significant state changes
6. **Generate Code:** Use manifest to generate Go structs and interfaces

---

## Good Examples

### ✅ Aggregate Definition
```yaml
aggregates:
  - name: "ClassicGame"
    description: "Solo quiz session with streak-based scoring"
    state:
      - name: "SessionID"
        type: "uuid.UUID"
        description: "Unique session identifier"
      - name: "CurrentStreak"
        type: "int"
        description: "Consecutive correct answers"
      - name: "CurrentMultiplier"
        type: "float64"
        description: "Score multiplier: 1.0, 1.5, or 2.0"
    methods:
      - name: "SubmitAnswer"
        params:
          - { name: "answerID", type: "uuid.UUID" }
          - { name: "responseTime", type: "int" }
        returns:
          - { type: "error" }
        logic: |
          Apply scoring rules from classic_game.md:
          - Rule 1: Calculate score with formula
          - Rule 2: Update multiplier based on streak
          - Rule 3: Reset streak on incorrect answer
```

### ✅ Value Object (Enum)
```yaml
value_objects:
  - name: "GameStatus"
    type: "string"
    description: "Current state of a game session"
    allowed_values: ["Pending", "InProgress", "Completed", "Failed"]
```

### ✅ Domain Event
```yaml
domain_events:
  - name: "GameFinished"
    description: "Published when a game session completes (success or failure)"
    attributes:
      - { name: "SessionID", type: "uuid.UUID" }
      - { name: "FinalScore", type: "int" }
      - { name: "MaxStreak", type: "int" }
      - { name: "CompletedAt", type: "int64" }
```

---

## Bad Examples (What NOT to do)

### ❌ Mixing Business Logic
```yaml
# WRONG - business rules belong in specs, not YAML
methods:
  - name: "CalculateScore"
    logic: |
      The score is calculated by taking base points and adding time bonus.
      Time bonus decreases linearly from maximum to zero.
      Then multiply by the multiplier which is 1.0 for streak 0-2...
```

**Correct approach:**
```yaml
methods:
  - name: "CalculateScore"
    logic: "Apply scoring formula from classic_game.md Rule 1"
```

### ❌ Using Infrastructure Types
```yaml
# WRONG - no database types
state:
  - name: "CreatedAt"
    type: "time.Time"  # Use int64 instead
  - name: "Data"
    type: "map[string]interface{}"  # Define proper struct
```

### ❌ Including HTTP/API Concerns
```yaml
# WRONG - this is infrastructure, not domain
methods:
  - name: "HandleRequest"
    params:
      - { name: "req", type: "*http.Request" }
```

---

## Related Documents

- **Business requirements:** See `docs/domain/specs/GUIDELINES.md`
- **Existing manifests:** See YAML files in this directory
- **Domain model:** See `docs/DOMAIN.md`
- **DDD patterns:** See `backend/internal/domain/` for implementation examples
