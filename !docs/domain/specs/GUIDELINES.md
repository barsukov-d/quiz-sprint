# Guidelines for Domain Specifications

This document defines the format and rules for creating business requirement specifications in `docs/domain/specs/`.

## Purpose

**Specifications describe WHAT the system does and WHY** from a business perspective.

- **Audience:** Product owners, analysts, stakeholders, QA, developers
- **Focus:** Business logic, rules, formulas, user value
- **Language:** Domain terminology (Ubiquitous Language)
- **Format:** Markdown (.md)

**NOT included:** Technical implementation details, data structures, method signatures, field types.

---

## Specification Template

```markdown
# Specification: [Feature Name]
**Context:** [Identity | QuizCatalog | ClassicGame | DailyMarathon | DuelGame | Leaderboard]
**Status:** [Draft | Approved | Implemented]

## 1. Business Goal (User Story)
> As [role], I want [action], so that [value/benefit].

## 2. Ubiquitous Language
- **Term:** Clear definition and usage context in this domain.
- **Term:** Description (avoid technical jargon).

## 3. Business Rules & Logic
1. **[Rule Name]:** Detailed description of behavior, formulas, constraints.
2. **[Formula]:** Mathematical expressions (e.g., `Score = (Base + Bonus) * Multiplier`).
3. **[Invariants]:** Conditions that must always be true (e.g., "Lives cannot exceed 3").

## 4. Scenarios (User Flows)
- **Scenario: [Descriptive Name]**
    - **Given:** [Initial state with specific values]
    - **When:** [Action or event occurs]
    - **Then:** [Expected outcome with concrete results]
```

---

## Writing Rules

### ✅ DO:
1. **Focus on behavior:** Describe what happens when, not how it's implemented
2. **Use domain language:** All terms from Section 2
3. **Provide concrete examples:** Scenarios with real numbers, not placeholders
4. **Document all formulas:** Exact calculations with all steps
5. **Be specific:** "Streak increases by 1" not "Streak changes"
6. **Keep it business-focused:** Avoid mentioning databases, APIs, UI components

### ❌ DON'T:
1. **No technical types:** Don't write `CurrentStreak: int` (that's for YAML manifests)
2. **No method signatures:** Don't write "Method SubmitAnswer(answerID uuid.UUID)"
3. **No data structures:** Don't describe aggregate fields and types
4. **No event schemas:** Don't list event attributes and types
5. **No placeholders in scenarios:** Don't write "Score = X", use actual values
6. **No implementation details:** Don't mention database tables, HTTP codes, JSON tags

---

## Good Examples

### ✅ Business Rule (Correct)
```
**Multiplier Levels:**
- Streak 0-2: Multiplier = x1.0
- Streak 3-5: Multiplier = x1.5
- Streak 6+: Multiplier = x2.0

**Reset Condition:** Any incorrect answer or timeout resets Streak to 0 and Multiplier to x1.0.
```

### ✅ Scenario (Correct)
```
- **Scenario: Entering Flow State**
    - **Given:** Player has Streak = 2, Multiplier = x1.0
    - **When:** Player answers question correctly within 5 seconds
    - **Then:** Streak becomes 3, Multiplier increases to x1.5, UI shows "On Fire" effects
```

### ✅ Formula (Correct)
```
**Damage Calculation:**
BaseDamage → Apply ComboMultiplier → Apply Critical/Block → Round down
Example: 15 → 19 (×1.3 combo) → 28 (+50% crit) → 28 HP
```

---

## Bad Examples (What NOT to do)

### ❌ Technical Structure (Wrong - belongs in YAML)
```
**New Fields:**
- CurrentStreak: int
- MaxStreak: int
- CurrentMultiplier: float64

**New Methods:**
- SubmitAnswer(answerID uuid.UUID, responseTime int) error
```

### ❌ Event Schema (Wrong - belongs in YAML)
```
**ClassicGameFinished Event:**
- GameID: uuid.UUID
- FinalScore: int
- MaxStreak: int
- CompletedAt: int64
```

### ❌ Vague Scenario (Wrong - not specific)
```
- **Scenario: User plays game**
    - **Given:** User is playing
    - **When:** User does something
    - **Then:** Score changes
```

---

## Document Structure Best Practices

1. **Section 1 (Business Goal):** One clear user story. Focus on user value, not features.
2. **Section 2 (Ubiquitous Language):** Define ALL terms before using them. 5-10 key terms max.
3. **Section 3 (Business Rules):** 4-8 core rules. Use formulas, thresholds, conditions.
4. **Section 4 (Scenarios):** 3-5 scenarios covering happy path, edge cases, failures.

---

## Workflow

1. **Write Spec First:** Start here to clarify business requirements
2. **Review with Stakeholders:** Validate business logic without technical jargon
3. **Update YAML Manifest:** Technical structure in `docs/domain/contexts/*.yaml`
4. **Implement Code:** Based on YAML manifest structure

---

## Related Documents

- **Technical structure:** See `docs/domain/contexts/GUIDELINES.md`
- **Implementation examples:** See existing specs in this directory
- **Domain model:** See `docs/DOMAIN.md`
