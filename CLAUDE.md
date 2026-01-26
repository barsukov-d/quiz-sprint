# CLAUDE.md

This file provides guidance to Claude Code when working with code in this repository.

## 📚 IMPORTANT: Read Before Coding

**ALWAYS read these documents before generating code:**
- **[docs/GLOSSARY.md](docs/GLOSSARY.md)** - Ubiquitous Language (единый словарь терминов)
  - Используй ТОЛЬКО термины из глоссария
  - Следуй naming conventions
  - Избегай anti-patterns

## Project Overview

Quiz Sprint TMA - Telegram Mini App for quizzes:
- **Frontend**: Vue 3 + TypeScript + Vite (`tma/`)
- **Backend**: Go + Fiber + DDD architecture (`backend/`)
- **Infrastructure**: Docker, PostgreSQL, Redis, nginx

## Quick Start

### Frontend (from `tma/`)
```bash
pnpm dev                    # Dev server (port 5173)
pnpm build                  # Build for production
pnpm run generate:all       # Generate Swagger + TypeScript types
pnpm lint                   # Lint code
pnpm test:unit              # Run tests
```

### Backend (from `backend/`)
```bash
# Development (Docker - recommended)
docker compose -f docker-compose.dev.yml up        # Start all services
docker compose -f docker-compose.dev.yml logs -f api  # View logs

# Services: http://localhost:3000 (API), http://localhost:8080 (Adminer)
# PostgreSQL: localhost:5432 (quiz_user/quiz_password_dev/quiz_sprint_dev)

# Swagger
make swagger                # Generate Swagger docs
pnpm run generate:all       # From tma/ - generates Swagger + TypeScript

# Testing
go test ./...               # Run all tests

# Quiz Import
make import-quiz FILE=data/quizzes/my-quiz.json
make import-all-quizzes
# See backend/IMPORT.md for details
```

## Architecture

### Monorepo Structure
```
quiz-sprint/
├── tma/                    # Vue 3 frontend
├── backend/                # Go backend (DDD)
├── infrastructure/         # VPS configs (nginx, systemd)
├── dev-tunnel/             # SSH tunnels for HTTPS dev
└── docs/                   # Domain docs (DOMAIN.md, USER_FLOW.md)
```

### Backend DDD Layers (`backend/internal/`)
```
domain/                     # Pure business logic (NO external deps)
├── quiz/                   # Quiz, Question, Answer, Session
├── user/                   # User aggregate
└── shared/                 # Shared value objects

application/                # Use cases (DTOs, orchestration)
├── quiz/                   # StartQuiz, SubmitAnswer, etc.
└── user/                   # RegisterUser, GetUser, etc.

infrastructure/             # Technical implementations
├── http/handlers/          # Fiber handlers + Swagger
├── persistence/postgres/   # PostgreSQL repositories
└── persistence/memory/     # In-memory fallback
```

### Frontend Structure (`tma/src/`)
```
main.ts                     # Vue app init
App.vue                     # Root component
router/                     # Vue Router
views/                      # Page components
api/client.ts               # Axios client (runtime hostname detection)
api/generated/              # Auto-generated from Swagger
├── types/                  # TypeScript interfaces
├── hooks/                  # Vue Query hooks (useGetQuiz, etc.)
└── schemas/                # Zod validation
```

## Tech Stack

**Frontend**: Vue 3.5, TypeScript 5.9, Vite, Vue Router, Vitest, Playwright, @telegram-apps/sdk

**Backend**: Go 1.25, Fiber v3, PostgreSQL 16, Redis 7, swaggo/swag, Air (hot reload)

**Infrastructure**: Docker, nginx, Let's Encrypt, GitHub Actions

## DDD Guidelines (Backend)

### Domain Layer (`internal/domain/`)
**Pure business logic - NO external dependencies:**
- ✅ Use: Value Objects, Factory methods (`NewQuiz`), `ReconstructEntity()` for DB loading
- ❌ NO: `context.Context`, JSON tags, database imports, `time.Time` (use `int64` Unix timestamps)

### Application Layer (`internal/application/`)
**Use Cases:**
- ✅ Use: Input/Output DTOs, `context.Context`, orchestration
- ❌ NO: Business logic (delegate to domain), HTTP concerns

### Infrastructure Layer (`internal/infrastructure/`)
**Technical implementations:**
- ✅ Use: HTTP handlers (thin adapters), Repository implementations, DB/SQL
- ❌ NO: Business logic

**Error Mapping**: Each handler has domain-specific error mapper (e.g., `quiz.ErrQuizNotFound` → HTTP 404)

## Swagger/OpenAPI Workflow

**Code-first approach:**
```
Go Handlers (@annotations) → swag → swagger.json → kubb → TypeScript types + Vue Query hooks
```

**After changing backend API:**
1. Update Go handler annotations in `backend/internal/infrastructure/http/handlers/`
2. Define DTOs in `swagger_models.go` (use concrete types, never `map[string]interface{}`)
3. Run `pnpm run generate:all` from `tma/` (generates Swagger + TypeScript)
4. Use generated hooks: `import { useGetQuizId } from '@/api/generated/hooks/quizController'`

**Required fields**: Use `validate:"required"` tag in Go structs → non-optional TypeScript types

## Environments

| Environment | URL | API Port | Database |
|-------------|-----|----------|----------|
| Development | `dev.quiz-sprint-tma.online` | 3000 (local) | PostgreSQL (Docker) |
| Staging | `staging.quiz-sprint-tma.online` | 3001 (Docker) | PostgreSQL (Docker) |
| Production | `quiz-sprint-tma.online` | 3000 (Docker) | PostgreSQL (Docker) |

**API Endpoints**: `https://<domain>/api/v1/*`, WebSocket: `wss://<domain>/ws/leaderboard/:id`

## Development with HTTPS (Telegram requires it)

1. Start backend: `cd backend && docker compose -f docker-compose.dev.yml up`
2. Start frontend: `cd tma && pnpm dev`
3. Start tunnels:
   ```bash
   ./dev-tunnel/start-backend-tunnel.sh   # localhost:3000 → VPS:3000
   ./dev-tunnel/start-frontend-tunnel.sh  # localhost:5173 → VPS:5173
   ```
4. Access: `https://dev.quiz-sprint-tma.online`

**How it works**: nginx on VPS proxies `/api/*` → localhost:3000, `/` → localhost:5173. Frontend detects hostname at runtime (`window.location.hostname`).

## Database

**Tables** (PostgreSQL):
- `users` - User profiles (Telegram auth)
- `quizzes`, `questions`, `answers` - Quiz data
- `quiz_sessions` - User attempts (⚠️ TODO: migrate from in-memory)
- `categories` - Quiz categories
- `tags` (⚠️ planned) - Quiz tags

**Viewing DB**:
- Web: http://localhost:8080 (Adminer: postgres/quiz_user/quiz_password_dev/quiz_sprint_dev)
- CLI: `docker compose -f docker-compose.dev.yml exec postgres psql -U quiz_user -d quiz_sprint_dev`

## Quiz Import

See `backend/IMPORT.md` for detailed guide.

**Quick commands**:
```bash
make import-quiz FILE=data/quizzes/my-quiz.json   # Import single quiz
make import-all-quizzes                           # Import all from data/quizzes/
```

**Formats**: Verbose (full field names) and Compact (LLM-optimized, 64% token reduction). See `backend/data/quizzes/SCHEMA.md`.

## Documentation

See `docs/` for domain model and user flows:
- `DOMAIN.md` - DDD patterns, aggregates, use cases
- `USER_FLOW.md` - User journeys, wireframes, UI spec
- `DOCUMENTATION_WORKFLOW.md` - When/how to update docs

**Workflow**: Update docs BEFORE code → Commit together

## Code Style

- No semicolons, single quotes, 100 char line width
- Path alias: `@` → `./src`
- No `any` types (TypeScript strict mode)

## Telegram Authentication

**Security**:
- ✅ Cryptographic signature validation (server-side)
- ✅ 1-hour expiration (prevents replay attacks)
- ✅ Base64-encoded init data in Authorization header: `Authorization: tma <base64>`
- ✅ Client cannot forge user data

**Flow**: Frontend SDK → Base64 encode → Auth header → Backend middleware validates signature → Handler uses validated data

## Deployment

**Frontend**: GitHub Actions → Build → Deploy to VPS

**Backend**: GitHub Actions → Docker build → Push to GHCR → Deploy via docker-compose

**Manual restart** (on VPS):
```bash
cd /opt/quiz-sprint/staging  # or production
docker compose restart api
docker compose logs -f api
```

## Key API Endpoints

- **Quiz**: `GET /api/v1/quiz`, `GET /api/v1/quiz/:id`, `POST /api/v1/quiz/:id/start`
- **Session**: `POST /api/v1/quiz/session/:sessionId/answer`, `DELETE /api/v1/quiz/session/:sessionId`
- **User**: `POST /api/v1/user/register`, `GET /api/v1/user/:id`
- **Categories**: `GET /api/v1/categories`, `POST /api/v1/categories`
- **Docs**: `GET /swagger/index.html`, `GET /health`

Full API docs: http://localhost:3000/swagger/index.html

## Workflow Requirements

Before completing session:
1. Track issues with `bd` (beads) tool
2. Always push to remote: `git push` required
- to memorize general