# Guidelines for Creating Domain Specifications (Domain Specs)

This document describes the process of transforming business ideas into technical documentation that serves as the "source of truth" for code generation.

## Workflow

1. **Intent (User Input):** You describe a feature in simple words (e.g., "Add difficulty level system for quizzes").
2. **Specification (Spec):** LLM transforms the description into a structured file using the template below.
3. **Manifest (YAML):** Based on Spec, the corresponding file in `docs/domain/contexts/` is updated.
4. **Implementation (Code):** Based on YAML and Spec, Go code is generated or updated.

---

## Specification Template

When creating a new file in `docs/domain/specs/`, use the following format:

```markdown
# Specification: [Name]
**Context:** [Identity | QuizCatalog | ClassicMode | DuelMode | Leaderboard]
**Status:** [Draft | Approved | Implemented]

## 1. Business Goal (User Story)
*Who, what, and why.*
> As [role], I want [action], so that [value].

## 2. Terminology (Ubiquitous Language)
*Dictionary of terms that will become class/variable names.*
- **Term:** Description and usage context.

## 3. Business Rules and Invariants
*Critical validation and logic rules.*
1. [Rule 1]
2. [Rule 2]

## 4. Data Model Changes
*Sketch of what will change in Aggregates, Entities, or Events.*
- **Aggregate [Name]**:
    - Field: Type (Description)
- **Domain Events**:
    - Event name (Attributes)

## 5. Scenarios (User Flows / BDD)
*Concrete examples of system behavior.*
- **Scenario: [Name]**
    - **Given:** [Initial conditions]
    - **When:** [Action]
    - **Then:** [Expected result]
```

---

## Rules for LLM When Generating Specs

1. **Variable Names:** Always suggest names in English using CamelCase (for Go).
2. **Data Types:** Use types compatible with Go (uuid.UUID, int64 for time, string, int).
3. **DDD-Centric:** Focus on behavior (aggregate methods) and rules, not database tables.
4. **Relations:** If feature affects other contexts, be sure to mention it in "Data Model Changes" section.
5. **Conciseness:** Description should be sufficient for implementation, without unnecessary verbosity.
