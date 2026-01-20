# Swagger Generation Guide

## Quick Start

### From Backend Directory

```bash
cd backend

# Generate Swagger documentation
make swagger

# Or use go run directly
go run github.com/swaggo/swag/cmd/swag@latest init \
  --generalInfo cmd/api/main.go \
  --output docs \
  --parseDependency \
  --parseInternal
```

### From Frontend Directory

```bash
cd tma

# Generate Swagger only
pnpm run generate:swagger

# Generate Swagger + TypeScript types
pnpm run generate:all
```

---

## Available Commands

### Backend (Makefile)

```bash
make swagger          # Generate Swagger docs
make swagger-install  # Install swag CLI globally
make dev             # Generate swagger + run server
make all             # Generate swagger + build binary
make help            # Show all commands
```

### Frontend (pnpm scripts)

```bash
pnpm run generate:swagger    # Generate backend swagger.json
pnpm run generate:api        # Generate TypeScript from swagger.json
pnpm run generate:all        # Generate both (swagger + TypeScript)
```

---

## Output Files

After running `make swagger` or `pnpm run generate:swagger`:

```
backend/docs/
├── docs.go          # Go code for Swagger UI
├── swagger.json     # OpenAPI 2.0 spec (used by Kubb)
└── swagger.yaml     # YAML version
```

After running `pnpm run generate:api`:

```
tma/src/api/generated/
├── types/           # TypeScript types
├── schemas/         # Zod validation schemas
└── hooks/           # Vue Query hooks
```

---

## Workflow

### 1. After Changing Go Handlers

```bash
# Option A: Using Makefile
cd backend
make swagger

# Option B: Using pnpm from tma/
cd tma
pnpm run generate:swagger
```

### 2. Generate TypeScript Types

```bash
cd tma
pnpm run generate:api
```

### 3. One Command for Both

```bash
cd tma
pnpm run generate:all
```

---

## Swagger Annotations

### Example: Add New Endpoint

```go
// GetQuiz handles GET /api/v1/quiz/:id
// @Summary Get quiz by ID
// @Description Get detailed information about a specific quiz
// @Tags quiz
// @Accept json
// @Produce json
// @Param id path string true "Quiz ID"
// @Success 200 {object} handlers.GetQuizDetailsResponse "Quiz details"
// @Failure 404 {object} handlers.ErrorResponse "Quiz not found"
// @Router /quiz/{id} [get]
func (h *QuizHandler) GetQuiz(c *fiber.Ctx) error {
    // Implementation
}
```

After adding annotations:
1. Run `make swagger`
2. Run `pnpm run generate:api` (from tma/)
3. New TypeScript types will be available

---

## Required Fields

To mark fields as required in Swagger (generates non-optional TypeScript types):

```go
type QuizDTO struct {
    ID    string `json:"id" validate:"required"`       // ✅ Required
    Title string `json:"title" validate:"required"`    // ✅ Required
    Description string `json:"description"`             // ❌ Optional
}
```

This generates:

```typescript
export type QuizDTO = {
    id: string;          // required (no ?)
    title: string;       // required (no ?)
    description?: string; // optional (with ?)
}
```

---

## Troubleshooting

### Issue: Types have long names like `InternalInfrastructureHttpHandlersQuizDTO`

**Cause:** Swag uses full package path for uniqueness

**Solution:** This is expected behavior. Types are correctly generated with `required` fields.

### Issue: All fields have `| undefined` in JSDoc

**Cause:** Missing `validate:"required"` tags in Go structs

**Solution:** Add `validate:"required"` to required fields in `swagger_models.go`

### Issue: Swagger not regenerating

**Solution:**
```bash
# Clean and regenerate
cd backend
rm -rf docs/
make swagger
```

---

## CI/CD Integration

The GitHub Actions workflows automatically:
1. Generate Swagger on backend changes
2. Generate TypeScript types on swagger.json changes
3. Commit updated types to the repository

Manual generation is only needed during local development.

---

## Documentation

- [swaggo/swag Documentation](https://github.com/swaggo/swag)
- [Swagger 2.0 Spec](https://swagger.io/specification/v2/)
- [Kubb Documentation](https://kubb.dev/)
