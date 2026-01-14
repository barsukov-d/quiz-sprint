# Quick Start Guide

## For New Developers

### 1. Clone & Setup

```bash
git clone <repo-url>
cd quiz-sprint/backend
```

### 2. Install Go

Requires Go 1.23+

```bash
go version  # Should show 1.23 or higher
```

### 3. Install Dependencies

```bash
go mod download
```

### 4. Run Locally

```bash
# Copy environment template
cp .env.example .env

# Run server
go run cmd/api/main.go
```

Server starts at `http://localhost:3000`

### 5. Test API

```bash
# Health check
curl http://localhost:3000/health

# Get all quizzes
curl http://localhost:3000/api/v1/quiz
```

## Architecture

This project follows **DDD + Clean Architecture**.

**MUST READ before coding:**
- `ARCHITECTURE.md` - Complete architecture rules
- `REFACTORING.md` - Current refactoring status

**Quick Rules:**
1. ✅ Domain layer has ZERO dependencies
2. ✅ Use Value Objects for all IDs
3. ✅ Use Cases work with DTOs (not domain models)
4. ✅ Handlers are thin adapters
5. ✅ Business logic ONLY in domain

## Directory Structure

```
backend/
├── cmd/api/           # Entry point
├── internal/
│   ├── domain/        # 🔴 Pure business logic (NO dependencies)
│   ├── application/   # 🟡 Use cases (orchestration)
│   └── infrastructure/# 🟢 HTTP, DB, external services
├── ARCHITECTURE.md    # 📖 MUST READ
├── REFACTORING.md     # 🚧 Current status
└── README.md          # Full documentation
```

## Common Commands

```bash
# Run
go run cmd/api/main.go

# Build
go build -o quiz-sprint-api cmd/api/main.go

# Test
go test ./...

# Format
go fmt ./...

# Clean dependencies
go mod tidy
```

## Need Help?

1. Read `ARCHITECTURE.md` first
2. Check `REFACTORING.md` for current status
3. See `README.md` for full docs
4. Ask in team chat

## Before Committing

- [ ] Code follows `ARCHITECTURE.md` rules
- [ ] No business logic in Use Cases
- [ ] No domain models leaked to infrastructure
- [ ] Tests pass: `go test ./...`
- [ ] Formatted: `go fmt ./...`
