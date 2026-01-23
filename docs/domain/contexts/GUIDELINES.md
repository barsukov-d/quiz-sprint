# Guidelines for Domain Manifests (YAML)

This document defines the strict structure and rules for YAML files in `docs/domain/contexts/`. These manifests serve as the "Technical Blueprint" for the domain layer and are used as the primary source for code generation.

## 1. Manifest Schema

Each YAML file represents a single **Bounded Context**. The structure is divided into the following sections:

### 1.1. Bounded Context Metadata
```yaml
bounded_context:
  name: "Context Name"
  type: "Core | Supporting | Generic"
  description: "High-level responsibility of this context."
```

### 1.2. Aggregates & Entities
Aggregates are the primary entry points for domain logic.
```yaml
aggregates:
  - name: "AggregateName"
    description: "What this aggregate represents."
    state:
      - name: "FieldName"
        type: "GoType" # uuid.UUID, int64, string, int, bool
        description: "Field purpose."
    methods:
      - name: "MethodName"
        params:
          - { name: "param", type: "Type" }
        returns:
          - { type: "Type" }
        logic: |
          Pseudo-code or reference to Business Rules in Spec.md.
```

### 1.3. Value Objects
Immutable objects defined by their attributes.
```yaml
value_objects:
  - name: "ValueObjectName"
    type: "string | int | struct"
    allowed_values: ["Enum1", "Enum2"] # Optional for enums
    invariants:
      - "Description of validation rule."
```

### 1.4. Domain Events
Events that occur within the domain.
```yaml
domain_events:
  - name: "EventName"
    description: "When this event is published."
    attributes:
      - { name: "ID", type: "uuid.UUID" }
```

### 1.5. Repositories
Interfaces for persisting aggregates.
```yaml
repositories:
  - name: "RepositoryName"
    methods:
      - name: "FindByID"
        params: [{ name: "id", type: "uuid.UUID" }]
        returns: ["*Aggregate", "error"]
```

## 2. Strict Rules

1.  **Naming Convention:**
    *   Use `PascalCase` for all names (Aggregates, Fields, Methods, Events).
    *   Use `camelCase` for method parameters.
2.  **Go Type Safety:**
    *   **IDs:** Always use `uuid.UUID`.
    *   **Timestamps:** Always use `int64` (Unix timestamps).
    *   **Collections:** Use `[]Type`.
    *   **No `any`:** Every field and parameter must have a concrete type.
3.  **Domain Purity:**
    *   No database tags (`json:`, `db:`).
    *   No HTTP status codes.
    *   No infrastructure dependencies (e.g., `sql.DB`).
4.  **Logic Traceability:**
    *   The `logic` block in methods must directly reference the rules defined in the corresponding `docs/domain/specs/*.md`.

## 3. Workflow

1.  **Update Spec:** First, define the logic in `docs/domain/specs/feature_name.md`.
2.  **Update Manifest:** Reflect the data model and method changes in the corresponding `.yaml` file.
3.  **Generate Code:** Use the manifest to generate Go interfaces, structs, and domain logic skeletons.