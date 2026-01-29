# LLM Quiz Generation Guide

Guide for generating quizzes using Claude/ChatGPT in compact, token-optimized format.

## Quick Start

**Prompt Template:**
```
Generate 10 programming quizzes about Go concurrency in compact JSON format.

Use this exact structure:

{
  "batch": {
    "version": 1,
    "cat": "programming",
    "tags": ["language:go", "difficulty:medium", "topic:concurrency"]
  },
  "quizzes": [
    {
      "t": "Quiz Title (concise, descriptive)",
      "d": "Brief description of what this quiz tests",
      "tags": ["topic:goroutines"],
      "q": [
        {
          "t": "Question text?",
          "a": ["Option 1", "Option 2", "Option 3", "Option 4"],
          "c": 1
        }
      ]
    }
  ]
}

RULES:
1. Each quiz: 5-7 questions
2. Each question: exactly 4 answers
3. "c" = 0-based index of correct answer (0, 1, 2, or 3)
4. Omit "l" and "p" fields (defaults: 60s time limit, 70% passing score)
5. Use specific topic tags for each quiz
6. Ensure exactly ONE correct answer per question
7. No ambiguous questions - each should have a clear correct answer
8. Make questions practical and test real understanding

Generate valid JSON only (no markdown, no comments, no explanations).
```

**Expected Output:**
A single JSON file with 10 quizzes, ready to import:
```bash
# Save LLM output to file
pbpaste > data/quizzes/batches/2024-01/go-concurrency.json

# Validate
make import-quiz-dry-run FILE=data/quizzes/batches/2024-01/go-concurrency.json

# Import
make import-quiz FILE=data/quizzes/batches/2024-01/go-concurrency.json
```

---

## Format Specification

### Batch Structure

```json
{
  "batch": {
    "version": 1,
    "cat": "programming",
    "tags": ["language:go", "difficulty:medium"]
  },
  "quizzes": [
    // Array of quiz objects
  ]
}
```

**Batch Fields:**
- `version` (int, required): Format version, always `1`
- `cat` (string, optional): Default category for all quizzes
- `tags` (array, optional): Shared tags applied to all quizzes

### Quiz Structure

```json
{
  "t": "Quiz Title",
  "d": "Description of what this quiz covers",
  "cat": "programming",
  "tags": ["topic:specific-topic"],
  "l": 120,
  "p": 80,
  "q": [
    // Array of question objects
  ]
}
```

**Quiz Fields:**
- `t` (string, required): Quiz title (3-200 chars)
- `d` (string, optional): Description (max 500 chars)
- `cat` (string, optional): Category override (omit to inherit from batch or infer from tags)
- `tags` (array, optional): Quiz-specific tags (merged with batch tags)
- `l` (int, optional): Time limit in seconds (default: 60, omit if using default)
- `p` (int, optional): Passing score percentage 0-100 (default: 70, omit if using default)
- `q` (array, required): Questions (minimum 1, recommended 5-7)

### Question Structure

```json
{
  "t": "What is the output of this code?",
  "a": ["Answer 1", "Answer 2", "Answer 3", "Answer 4"],
  "c": 2
}
```

**Question Fields:**
- `t` (string, required): Question text (5-500 chars)
- `a` (array, required): Answer options (minimum 2, recommended 4)
- `c` (int, required): Correct answer index (0-based, 0 = first answer, 1 = second, etc.)

---

## Tag System

### Tag Format

**Pattern:** `{category}:{value}`

**Valid Examples:**
- `language:go`
- `difficulty:easy`
- `topic:web-development`
- `domain:programming`

### Tag Categories

| Category | Description | Examples |
|----------|-------------|----------|
| `language:*` | Programming languages | go, python, javascript, rust, java |
| `difficulty:*` | Quiz difficulty level | easy, medium, hard, expert |
| `topic:*` | Specific subject matter | goroutines, pointers, interfaces, error-handling |
| `domain:*` | General knowledge domain | programming, history, science, movies |
| `format:*` | Question format type | multiple-choice, true-false, code-completion |

### Tag Rules

1. **Lowercase only** - `language:go` ✅, `Language:Go` ❌
2. **Hyphens for multi-word** - `web-development` ✅, `web_development` ❌, `web development` ❌
3. **Max 100 characters**
4. **Pattern:** `^[a-z0-9-:]+$`
5. **1-10 tags per quiz**

### Category Inference

If you omit the `cat` field, the category is automatically inferred from tags:

| Tag Pattern | Inferred Category |
|-------------|-------------------|
| `language:*` | programming |
| `domain:history` | history |
| `domain:science` | science |
| `domain:movies` | movies |
| `domain:geography` | geography |
| (none match) | general |

**Example:**
```json
{
  "t": "Go Concurrency Patterns",
  "tags": ["language:go", "difficulty:medium", "topic:goroutines"]
  // cat: "programming" - auto-inferred from "language:go"
}
```

---

## Validation Rules

### Quiz-Level Validation

- ✅ Title: 3-200 characters, required
- ✅ Description: max 500 characters, optional
- ✅ Time limit: positive integer (seconds), defaults to 60
- ✅ Passing score: 0-100 (percentage), defaults to 70
- ✅ Questions: minimum 1 question required

### Question-Level Validation

- ✅ Text: 5-500 characters, required
- ✅ Answers: minimum 2 options required (recommended 4)
- ✅ Correct index: must be valid (0 ≤ c < number of answers)
- ✅ Each answer: 1-200 characters

### Common Validation Errors

**Error:** `"c" value out of bounds`
```json
// ❌ Wrong - only 3 answers, but c=3 (4th answer)
{
  "a": ["A", "B", "C"],
  "c": 3
}

// ✅ Correct - c must be 0, 1, or 2
{
  "a": ["A", "B", "C"],
  "c": 1
}
```

**Error:** `Invalid tag format`
```json
// ❌ Wrong - uppercase, missing category
"tags": ["Go", "Easy"]

// ✅ Correct - lowercase, with category
"tags": ["language:go", "difficulty:easy"]
```

**Error:** `Title too short`
```json
// ❌ Wrong - less than 3 chars
"t": "Go"

// ✅ Correct
"t": "Go Basics Quiz"
```

---

## Example Prompts

### Example 1: Single-Language Deep Dive

```
Generate 10 quizzes about Python programming, covering different difficulty levels.

Format: Compact JSON batch
Category: programming
Tags: language:python, plus difficulty and topic tags

Topics to cover:
1. Variables and data types (easy)
2. Functions and scope (easy)
3. List comprehensions (medium)
4. Decorators (medium)
5. Generators (medium)
6. Async/await (hard)
7. Metaclasses (hard)
8. Type hints (medium)
9. Context managers (medium)
10. Memory management (hard)

Each quiz: 5-7 questions, 4 answer options each.
```

### Example 2: Cross-Domain Collection

```
Generate 5 quizzes across different knowledge domains.

Format: Compact JSON batch
Omit batch-level category (each quiz has its own)

Quizzes:
1. World War II History
   - cat: history
   - tags: domain:history, difficulty:medium, period:20th-century

2. Solar System Basics
   - cat: science
   - tags: domain:science, difficulty:easy, topic:astronomy

3. Classic Movies 1980s
   - cat: movies
   - tags: domain:movies, difficulty:medium, decade:1980s

4. JavaScript ES6 Features
   - cat: programming
   - tags: language:javascript, difficulty:medium, topic:es6

5. World Capitals
   - cat: geography
   - tags: domain:geography, difficulty:easy

Each quiz: 6 questions, 4 answer options each.
```

### Example 3: Difficulty Progression

```
Generate 6 Go quizzes with progressive difficulty.

Format: Compact JSON batch
Category: programming
Base tags: language:go

Progression:
1. "Go Syntax Basics" - difficulty:easy, topic:syntax
2. "Variables and Types" - difficulty:easy, topic:variables
3. "Control Flow" - difficulty:medium, topic:control-flow
4. "Goroutines Intro" - difficulty:medium, topic:concurrency
5. "Channels Advanced" - difficulty:hard, topic:channels
6. "Concurrency Patterns" - difficulty:expert, topic:concurrency

Quizzes 1-2: 5 questions each
Quizzes 3-4: 7 questions each
Quizzes 5-6: 10 questions each (l: 120)
```

---

## Best Practices

### 1. Question Quality

**Good Questions:**
- ✅ Clear, unambiguous wording
- ✅ Test understanding, not memorization
- ✅ One obviously correct answer
- ✅ Plausible but incorrect distractors

**Bad Questions:**
- ❌ Ambiguous or trick questions
- ❌ Multiple correct interpretations
- ❌ Require external context
- ❌ Too easy (obvious) or impossibly hard

**Example:**
```json
// ❌ Bad - Ambiguous
{
  "t": "What's the best way to handle errors in Go?",
  "a": ["if err != nil", "panic", "ignore", "try-catch"],
  "c": 0
}

// ✅ Good - Clear, specific
{
  "t": "In Go, what is the idiomatic way to check for errors?",
  "a": ["if err != nil { return err }", "try { } catch(err) { }", "err.Check()", "panic(err)"],
  "c": 0
}
```

### 2. Answer Options

- Use 4 options for multiple choice (standard)
- Make all options similar length
- Avoid "all of the above" or "none of the above"
- Distractors should be plausible but clearly wrong

### 3. Tagging Strategy

**Be Specific:**
```json
// ❌ Too generic
"tags": ["programming", "easy"]

// ✅ Specific and useful
"tags": ["language:go", "difficulty:easy", "topic:goroutines"]
```

**Consistent Naming:**
```json
// ❌ Inconsistent
"tags": ["go-lang", "language:python", "JavaScript"]

// ✅ Consistent
"tags": ["language:go", "language:python", "language:javascript"]
```

### 4. Time Limits

**Recommended Formulas:**
- Easy quiz: `questions * 60s` (1 min per question)
- Medium quiz: `questions * 90s` (1.5 min per question)
- Hard quiz: `questions * 120s` (2 min per question)

**Examples:**
```json
// Easy: 5 questions = 300s (5 min)
{ "t": "Go Basics", "l": 300, "q": [...] }

// Medium: 7 questions = 630s (~10 min)
{ "t": "Go Concurrency", "l": 600, "q": [...] }

// Hard: 10 questions = 1200s (20 min)
{ "t": "Go Advanced", "l": 1200, "q": [...] }
```

### 5. Passing Scores

| Difficulty | Recommended Passing Score |
|------------|---------------------------|
| Easy | 60-70% |
| Medium | 70-80% |
| Hard | 80-85% |
| Expert | 85-90% |

### 6. Batch Organization

**File Naming:**
```
data/quizzes/batches/
├── 2024-01/
│   ├── go-basics.json           (10 easy Go quizzes)
│   ├── go-concurrency.json      (10 medium Go quizzes)
│   ├── python-fundamentals.json
│   └── javascript-es6.json
└── 2024-02/
    └── ...
```

**Batch Metadata:**
```json
{
  "batch": {
    "version": 1,
    "generated": "2024-01-15T10:30:00Z",
    "model": "claude-opus-4",
    "prompt": "Generate 10 quizzes about Go concurrency",
    "cat": "programming",
    "tags": ["language:go"]
  },
  "quizzes": [...]
}
```

---

## Troubleshooting

### Problem: LLM generates invalid JSON

**Symptoms:**
- Parse errors when importing
- Missing commas, brackets
- Comments in JSON

**Solution:**
1. Add to prompt: "Generate valid JSON only (no markdown, no comments)"
2. If LLM adds markdown code blocks, manually remove ` ```json ` and ` ``` `
3. Validate JSON before import: `make import-quiz-dry-run FILE=...`

### Problem: Category inference not working

**Symptoms:**
- Quizzes appear in "general" category instead of expected category

**Solution:**
1. Check tags include category-specific prefix (e.g., `language:go`, not just `go`)
2. Or explicitly set `cat` field:
   ```json
   {
     "t": "Quiz Title",
     "cat": "programming",  // Explicit category
     "tags": ["language:go"]
   }
   ```

### Problem: Validation errors during import

**Common Errors:**

1. **"exactly one answer must be correct (found 0)"**
   ```json
   // ❌ Wrong - no correct answer marked
   { "a": ["A", "B", "C", "D"], "c": null }

   // ✅ Fix - mark correct answer
   { "a": ["A", "B", "C", "D"], "c": 2 }
   ```

2. **"question N: at least 2 answers required"**
   ```json
   // ❌ Wrong - only 1 answer
   { "a": ["Only option"] }

   // ✅ Fix - add more answers
   { "a": ["Option 1", "Option 2", "Option 3", "Option 4"] }
   ```

3. **"invalid tag format"**
   ```json
   // ❌ Wrong - missing category prefix
   "tags": ["go", "easy"]

   // ✅ Fix - use category:value format
   "tags": ["language:go", "difficulty:easy"]
   ```

### Problem: Token limit exceeded during generation

**Symptoms:**
- LLM stops mid-generation
- Incomplete batch
- Only 3-5 quizzes instead of 10

**Solution:**
1. Reduce batch size: Ask for 5 quizzes instead of 10
2. Reduce questions per quiz: 5 instead of 7
3. Simplify prompt: Remove examples, keep rules concise
4. Use multiple prompts and merge JSON files manually

---

## Token Savings Summary

| Format | Tokens/Quiz | 10 Quizzes | 100 Quizzes | Savings |
|--------|-------------|------------|-------------|---------|
| Verbose | ~640 | 6,400 | 64,000 | Baseline |
| Compact | ~230 | 2,300 | 23,000 | **64%** |
| Batch | ~210 | 2,100 | 21,000 | **67%** |

**Cost Savings (Claude Opus 4.5):**
- 100 quizzes verbose: ~$1.28 (64k tokens)
- 100 quizzes compact: ~$0.46 (23k tokens)
- **Saved: $0.82 per 100 quizzes**

---

## Advanced Usage

### Multi-Model Generation

Generate with different LLMs and merge:

```bash
# Generate with Claude
claude-api generate --prompt "..." > batches/claude-go-quizzes.json

# Generate with GPT-4
openai-api generate --prompt "..." > batches/gpt4-go-quizzes.json

# Validate both
make import-quiz-dry-run FILE=batches/claude-go-quizzes.json
make import-quiz-dry-run FILE=batches/gpt4-go-quizzes.json

# Import both
make import-quiz FILE=batches/claude-go-quizzes.json
make import-quiz FILE=batches/gpt4-go-quizzes.json
```

### Template Customization

Create domain-specific templates:

```json
// data/quizzes/templates/programming-quiz-template.json
{
  "batch": {
    "version": 1,
    "cat": "programming",
    "tags": ["TO_BE_REPLACED"]
  },
  "quizzes": [
    {
      "t": "QUIZ_TITLE",
      "d": "DESCRIPTION",
      "tags": ["topic:TOPIC"],
      "q": [
        {
          "t": "QUESTION_TEXT",
          "a": ["OPTION_1", "OPTION_2", "OPTION_3", "OPTION_4"],
          "c": 0
        }
      ]
    }
  ]
}
```

### Validation Scripts

```bash
# Validate all quizzes in a directory
for file in data/quizzes/batches/2024-01/*.json; do
  echo "Validating $file..."
  make import-quiz-dry-run FILE="$file"
done

# Count total quizzes
jq '.quizzes | length' data/quizzes/batches/2024-01/*.json | awk '{sum+=$1} END {print "Total:", sum}'
```

---

## Next Steps

After generating quizzes:

1. **Validate:** `make import-quiz-dry-run FILE=your-batch.json`
2. **Review:** Check questions for quality and clarity
3. **Import:** `make import-quiz FILE=your-batch.json`
4. **Test:** Try quizzes in the TMA to verify UX
5. **Iterate:** Generate more quizzes with improved prompts

## Resources

- **Import Guide:** `backend/IMPORT.md`
- **JSON Schema:** `backend/data/quizzes/SCHEMA.md`
- **Templates:** `backend/data/quizzes/TEMPLATE-COMPACT.json`, `TEMPLATE-BATCH.json`
- **Domain Docs:** `docs/DOMAIN.md` (Tag Aggregate section)
- **Claude Code Guide:** `CLAUDE.md` (Quiz Import Formats section)
