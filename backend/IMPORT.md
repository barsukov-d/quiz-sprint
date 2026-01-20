# Quiz Import Guide

This guide explains how to import quizzes from JSON files into the Quiz Sprint database.

## Quick Start

```bash
# Import a single quiz
make import-quiz FILE=data/quizzes/programming-basics.json

# Import all quizzes from directory
make import-all-quizzes

# Validate without importing (dry run)
make import-quiz-dry-run FILE=data/quizzes/my-quiz.json
```

## Table of Contents

1. [Installation](#installation)
2. [Basic Usage](#basic-usage)
3. [Creating Quiz JSON Files](#creating-quiz-json-files)
4. [Import Commands](#import-commands)
5. [Advanced Usage](#advanced-usage)
6. [Troubleshooting](#troubleshooting)

## Installation

No additional installation required if you have the Quiz Sprint backend set up. The import tool is part of the project.

**Prerequisites:**
- Go 1.25+
- PostgreSQL database (running via Docker or locally)
- Database connection configured (see `pkg/database/database.go`)

## Basic Usage

### 1. Create a Quiz JSON File

Create a JSON file following the schema defined in `data/quizzes/SCHEMA.md`:

```json
{
  "title": "My Quiz",
  "description": "A quiz about something interesting",
  "categoryId": null,
  "timeLimit": 600,
  "passingScore": 70,
  "questions": [
    {
      "text": "What is 2+2?",
      "points": 10,
      "answers": [
        { "text": "3", "isCorrect": false },
        { "text": "4", "isCorrect": true },
        { "text": "5", "isCorrect": false }
      ]
    }
  ]
}
```

### 2. Validate Your Quiz (Optional)

Before importing, validate your JSON file:

```bash
make import-quiz-dry-run FILE=data/quizzes/my-quiz.json
```

This will check for:
- Valid JSON syntax
- Required fields
- Correct answer count (exactly 1 per question)
- Valid value ranges

### 3. Import the Quiz

```bash
make import-quiz FILE=data/quizzes/my-quiz.json
```

Expected output:
```
✓ Connected to database
Found 1 file(s) to import

--- Processing: my-quiz.json ---
  Title: My Quiz
  Questions: 1
  Time Limit: 600 seconds
  Passing Score: 70%
  Quiz ID: 550e8400-e29b-41d4-a716-446655440000
✓ Success

=== Import Summary ===
Total: 1
Success: 1
Errors: 0
```

## Creating Quiz JSON Files

### File Structure

Place your quiz JSON files in the `data/quizzes/` directory:

```
backend/
├── data/
│   └── quizzes/
│       ├── SCHEMA.md              # Schema documentation
│       ├── programming-basics.json
│       ├── world-geography.json
│       └── your-quiz.json
```

### JSON Schema

See `data/quizzes/SCHEMA.md` for complete schema documentation.

**Required fields:**
- `title` (3-200 chars)
- `timeLimit` (positive integer, seconds)
- `passingScore` (0-100)
- `questions` (array with at least 1 question)
  - `text` (5-500 chars)
  - `points` (positive integer)
  - `answers` (array with at least 2 answers)
    - `text` (1-200 chars)
    - `isCorrect` (boolean, exactly 1 must be true)

**Optional fields:**
- `description`
- `categoryId` (UUID format)

### Examples

See these example quizzes:
- `data/quizzes/programming-basics.json` - Programming concepts
- `data/quizzes/world-geography.json` - Geography trivia
- `data/quizzes/javascript-advanced.json` - Advanced JavaScript
- `data/quizzes/movie-trivia.json` - Movie knowledge

## Import Commands

### Makefile Commands (Recommended)

| Command | Description | Usage |
|---------|-------------|-------|
| `make import-quiz` | Import single quiz | `make import-quiz FILE=path/to/quiz.json` |
| `make import-quiz-dry-run` | Validate single quiz | `make import-quiz-dry-run FILE=path/to/quiz.json` |
| `make import-all-quizzes` | Import all quizzes from directory | `make import-all-quizzes` |
| `make import-all-quizzes-dry-run` | Validate all quizzes | `make import-all-quizzes-dry-run` |

### Direct Go Commands

You can also run the import tool directly:

```bash
# Import single file
go run cmd/import/main.go -file=data/quizzes/my-quiz.json

# Import directory
go run cmd/import/main.go -dir=data/quizzes

# Dry run (validate only)
go run cmd/import/main.go -file=data/quizzes/my-quiz.json -dry-run
```

## Compact Format (LLM-Optimized)

### Why Compact Format?

The compact format is designed for efficient batch quiz generation using Large Language Models (LLMs) like Claude or ChatGPT.

**Token Savings:**
- **Verbose format:** ~640 tokens per quiz
- **Compact format:** ~230 tokens per quiz
- **Savings:** **64%** fewer tokens

**For batch LLM generation of 100 quizzes:**
- Verbose: 64,000 tokens (~$1.28 with Claude Opus)
- Compact: 23,000 tokens (~$0.46 with Claude Opus)
- **Saved: $0.82 + faster generation**

### Format Comparison

**Verbose Format (Legacy):**
```json
{
  "title": "Programming Basics",
  "description": "Test your knowledge of programming fundamentals",
  "categoryId": null,
  "timeLimit": 600,
  "passingScore": 70,
  "questions": [
    {
      "text": "What is a variable?",
      "points": 10,
      "answers": [
        { "text": "A container for data", "isCorrect": true },
        { "text": "A function", "isCorrect": false },
        { "text": "A loop", "isCorrect": false }
      ]
    }
  ]
}
```

**Compact Format (New):**
```json
{
  "t": "Programming Basics",
  "d": "Test your knowledge of programming fundamentals",
  "cat": "programming",
  "tags": ["domain:programming", "difficulty:easy"],
  "q": [
    {
      "t": "What is a variable?",
      "a": ["A container for data", "A function", "A loop"],
      "c": 0
    }
  ]
}
```

**Field Mapping:**
| Verbose | Compact | Notes |
|---------|---------|-------|
| `title` | `t` | Quiz title |
| `description` | `d` | Description |
| `categoryId` | `cat` | Category name (or inferred from tags) |
| - | `tags` | Array of tags (new!) |
| `timeLimit` | `l` | Omit if 60 (default) |
| `passingScore` | `p` | Omit if 70 (default) |
| `questions` | `q` | Questions array |
| `text` | `t` | Question text |
| `answers` | `a` | Array of answer strings |
| `isCorrect` | `c` | Correct answer index (0-based) |

### Batch Format

Import multiple quizzes from a single file:

```json
{
  "batch": {
    "version": 1,
    "cat": "programming",
    "tags": ["language:go", "difficulty:medium"]
  },
  "quizzes": [
    {
      "t": "Go Basics",
      "tags": ["topic:variables"],
      "q": [...]
    },
    {
      "t": "Go Concurrency",
      "tags": ["topic:goroutines"],
      "l": 120,
      "q": [...]
    }
  ]
}
```

**Batch-level fields are merged with each quiz:**
- Tags: `batch.tags + quiz.tags` (deduplicated)
- Category: `quiz.cat` or `batch.cat` or inferred from tags

### Category + Tags System

**Category** (one per quiz):
- Used for main navigation (CategoriesView in UI)
- Examples: programming, history, science, movies
- Can be explicit (`cat: "programming"`) or inferred from tags

**Tags** (multiple per quiz):
- Used for filtering and metadata
- Format: `{category}:{value}`
- Examples: `language:go`, `difficulty:easy`, `topic:concurrency`

**Category Inference:**
If `cat` field is omitted, it's inferred from tags:
- `language:*` → "programming"
- `domain:history` → "history"
- `domain:science` → "science"
- Fallback: "general"

### Import Commands

```bash
# Single compact quiz
make import-quiz FILE=data/quizzes/compact-quiz.json

# Batch of 10+ quizzes
make import-quiz FILE=data/quizzes/batches/2024-01/go-basics.json

# Validate batch before importing
make import-quiz-dry-run FILE=data/quizzes/batches/2024-01/go-basics.json

# Import all batches from a directory
make import-quiz FILE=data/quizzes/batches/2024-01/*.json
```

### Templates

**Single Quiz Template:**
```bash
cp data/quizzes/TEMPLATE-COMPACT.json data/quizzes/my-quiz.json
```

**Batch Template:**
```bash
cp data/quizzes/TEMPLATE-BATCH.json data/quizzes/batches/my-batch.json
```

### LLM Generation Guide

For detailed instructions on generating quizzes with Claude/ChatGPT:
- **Full guide:** `backend/docs/LLM_GENERATION_GUIDE.md`
- Includes prompt templates, validation rules, and best practices

**Quick Example Prompt:**
```
Generate 5 programming quizzes about Go in compact JSON format.

{
  "batch": {
    "version": 1,
    "cat": "programming",
    "tags": ["language:go"]
  },
  "quizzes": [
    {
      "t": "Quiz Title",
      "tags": ["difficulty:easy"],
      "q": [
        {
          "t": "Question?",
          "a": ["A", "B", "C", "D"],
          "c": 0
        }
      ]
    }
  ]
}

Rules:
- 5-7 questions per quiz
- 4 answers per question
- "c" = 0-based correct answer index
- Use valid tags: language:*, difficulty:*, topic:*
```

### File Organization

**Recommended structure:**
```
data/quizzes/
├── batches/
│   ├── 2024-01/
│   │   ├── programming-go-basics.json       (10 quizzes)
│   │   ├── programming-go-concurrency.json  (10 quizzes)
│   │   └── history-world-war-2.json         (8 quizzes)
│   └── 2024-02/
│       └── ...
├── legacy/
│   └── *.json (old verbose format)
└── templates/
    ├── TEMPLATE.json (verbose)
    ├── TEMPLATE-COMPACT.json (compact single)
    └── TEMPLATE-BATCH.json (batch)
```

**Naming convention:** `{domain}-{topic}-{variant}.json`

### Validation

The import tool automatically:
1. **Detects format** (batch, compact, or verbose)
2. **Converts** compact → verbose internally
3. **Validates** all fields and business rules
4. **Infers category** from tags if needed
5. **Merges tags** from batch and quiz levels

**Format detection:**
```go
// Automatically detects:
if has "batch" key → batch format
else if has "t" key → compact single quiz
else if has "title" key → verbose single quiz
```

### Token Savings Breakdown

| Scenario | Verbose | Compact | Savings |
|----------|---------|---------|---------|
| 1 quiz (5 questions) | ~640 tokens | ~230 tokens | 64% |
| 10 quizzes | 6,400 tokens | 2,300 tokens | 64% |
| 10 quizzes (batch) | 6,400 tokens | 2,100 tokens | 67% |
| 100 quizzes (batch) | 64,000 tokens | 21,000 tokens | 67% |

**Additional benefits:**
- ✅ Faster LLM generation (less output to produce)
- ✅ Easier to fit in context window
- ✅ Lower API costs
- ✅ Better organization (batch files)

### Migration from Verbose Format

**All existing verbose format files still work!**

The import tool supports both formats:
```bash
# Old format (still works)
make import-quiz FILE=data/quizzes/programming-basics.json

# New format
make import-quiz FILE=data/quizzes/compact-programming-basics.json
```

**To convert manually:**
1. Copy TEMPLATE-COMPACT.json
2. Map fields: title→t, description→d, etc.
3. Convert answers: `[{text, isCorrect}]` → `["text1", "text2"], c: index`
4. Add tags for better filtering

**Or keep both!** Use verbose for hand-written quizzes, compact for LLM-generated.

## Advanced Usage

### Assigning Categories

To assign a quiz to a category, you need the category UUID:

1. **Get category ID from database:**
   ```sql
   SELECT id, name FROM categories;
   ```

2. **Use in JSON:**
   ```json
   {
     "title": "My Quiz",
     "categoryId": "550e8400-e29b-41d4-a716-446655440000",
     ...
   }
   ```

### Batch Import

Import all quizzes from a directory at once:

```bash
make import-all-quizzes
```

This will:
- Find all `.json` files in `data/quizzes/`
- Validate each file
- Import successful files
- Show summary with success/error counts

### Custom Import Location

Import from a different directory:

```bash
go run cmd/import/main.go -dir=/path/to/my/quizzes
```

### Building Import Binary

For production use, build a standalone binary:

```bash
go build -o quiz-import cmd/import/main.go

# Use the binary
./quiz-import -file=data/quizzes/my-quiz.json
```

## Troubleshooting

### Database Connection Error

```
Failed to connect to database: connection refused
```

**Solution:**
1. Ensure PostgreSQL is running:
   ```bash
   docker compose -f docker-compose.dev.yml up -d postgres
   ```

2. Check database configuration in `.env` or environment variables

### Validation Errors

#### Multiple Correct Answers
```
Error: question 2: exactly one answer must be correct (found 2)
```
**Solution:** Each question must have exactly one answer with `"isCorrect": true`

#### Not Enough Answers
```
Error: question 1: at least 2 answers required
```
**Solution:** Each question needs at least 2 answer options

#### Invalid Title Length
```
Error: invalid title: title must be between 3 and 200 characters
```
**Solution:** Adjust title length to be within the valid range

### JSON Syntax Errors

```
Error: failed to parse JSON: invalid character '}' after object key:value pair
```

**Solution:**
1. Use a JSON validator (e.g., https://jsonlint.com/)
2. Check for:
   - Missing commas
   - Extra commas (especially after last item in arrays/objects)
   - Unmatched brackets/braces
   - Missing quotes around strings

### Import Succeeds but Quiz Not Visible

**Possible causes:**
1. Quiz was assigned to wrong category
2. Frontend not refreshing data

**Solutions:**
1. Check the quiz in database:
   ```sql
   SELECT id, title, category_id FROM quizzes ORDER BY created_at DESC LIMIT 5;
   ```

2. Refresh frontend or clear cache

## Best Practices

### 1. Validate Before Importing

Always run dry-run first:
```bash
make import-quiz-dry-run FILE=data/quizzes/new-quiz.json
```

### 2. Use Version Control

Keep quiz JSON files in Git to track changes:
```bash
git add data/quizzes/my-quiz.json
git commit -m "Add new quiz: My Quiz"
```

### 3. Organize by Category

Create subdirectories for different quiz types:
```
data/quizzes/
├── programming/
│   ├── javascript.json
│   └── python.json
├── geography/
│   └── world-capitals.json
└── movies/
    └── classics.json
```

Import from subdirectory:
```bash
go run cmd/import/main.go -dir=data/quizzes/programming
```

### 4. Meaningful Filenames

Use descriptive filenames that match quiz titles:
- ✅ `programming-basics.json`
- ✅ `world-geography-capitals.json`
- ❌ `quiz1.json`
- ❌ `test.json`

### 5. Points Distribution

- Easy questions: 5-10 points
- Medium questions: 10-15 points
- Hard questions: 15-20 points

Total points should relate to passing score:
```
If total = 100 points and passingScore = 70
→ User needs 70 points to pass
```

### 6. Time Limits

General guidelines:
- ~60 seconds per question
- Add 120 seconds buffer for reading instructions
- Example: 10 questions → 600-720 seconds (10-12 minutes)

## Database Schema

The import tool creates records in these tables:

```
quizzes
├── id (UUID, auto-generated)
├── title
├── description
├── category_id (FK to categories, optional)
├── time_limit
├── passing_score
└── created_at

questions
├── id (UUID, auto-generated)
├── quiz_id (FK to quizzes)
├── text
├── points
└── order_index

answers
├── id (UUID, auto-generated)
├── question_id (FK to questions)
├── text
├── is_correct
└── order_index
```

## API Endpoints

After importing, quizzes are available via API:

- `GET /api/v1/quiz` - List all quizzes
- `GET /api/v1/quiz/:id` - Get quiz details
- `POST /api/v1/quiz/:id/start` - Start quiz session

See Swagger docs: http://localhost:3000/swagger/index.html

## Need Help?

- **Schema documentation**: `data/quizzes/SCHEMA.md`
- **Example quizzes**: `data/quizzes/`
- **Source code**: `cmd/import/main.go`
- **Report issues**: GitHub Issues

## Future Enhancements

Planned features:
- [ ] CSV import support
- [ ] Bulk update existing quizzes
- [ ] Image support for questions
- [ ] Quiz export (database → JSON)
- [ ] Web UI for quiz creation
- [ ] Import from Google Forms/Typeform
