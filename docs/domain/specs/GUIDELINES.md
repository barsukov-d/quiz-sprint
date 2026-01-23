# Guidelines for Creating Domain Specifications (Domain Specs)

This document describes the process of transforming business ideas into technical documentation that serves as the "source of truth" for code generation and manifest updates.

## Workflow

1. **Intent (User Input):** Describe a feature or game mode in simple words.
2. **Specification (Spec):** Create or update a file in `docs/domain/specs/` using the template below. This is the "Design Document".
3. **Manifest (YAML):** Based on the Spec, update the corresponding file in `docs/domain/contexts/`. This is the "Technical Blueprint".
4. **Implementation (Code):** Based on the YAML manifest, Go code is generated or updated.

---

## Universal Specification Template

When creating a new file in `docs/domain/specs/`, use the following format:

```markdown
# Specification: [Name]
**Context:** [Identity | QuizCatalog | ClassicGame | DailyMarathon | DuelGame | Leaderboard]
**Status:** [Draft | Approved | Implemented]

## 1. Business Goal (User Story)
*Who, what, and why.*
> As [role], I want [action], so that [value].

## 2. Ubiquitous Language
*Dictionary of terms that will become class/variable names. Focus on new terms.*
- **Term:** Description and usage context.

## 3. Business Rules & Logic
*The core of the domain. Describe complex logic, formulas, and constraints.*
1. **[Rule Name]:** Detailed description of the logic (e.g., scoring formulas, win conditions).
2. **Invariants:** Conditions that must always be true (e.g., "Lives cannot exceed 3").

## 4. Manifest Updates (Intent)
*Instructions for updating the corresponding YAML file in docs/domain/contexts/.*
- **New Fields:** Fields to add to Aggregates, Entities, or Value Objects.
- **New Methods:** Domain methods/actions to be added to Aggregates.
- **New Events:** Domain events to be published.

## 5. Scenarios (User Flows / BDD)
*Concrete examples of system behavior. These serve as the basis for unit tests.*
- **Scenario: [Name]**
    - **Given:** [Initial conditions]
    - **When:** [Action]
    - **Then:** [Expected result and state changes]
```

---

## Rules for LLM When Generating Specs

1. **Naming:** Always suggest names in English using CamelCase (for Go compatibility).
2. **Logic First:** Focus on *behavior* and *rules* in Section 3. YAML is for structure; Spec is for logic.
3. **Conciseness:** Do not duplicate the entire data model if only a few fields are changing. Use Section 4 to list specific delta changes for the YAML manifest.
4. **Types:** When mentioning fields in Section 4, use Go-compatible types (uuid.UUID, int64 for timestamps, string, int).
5. **DDD-Centric:** Focus on the domain layer. Avoid mentioning database tables, HTTP status codes, or UI implementation details unless they directly represent a business rule.