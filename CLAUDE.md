# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Quiz Sprint TMA is a full-stack Telegram Mini App:
- **Frontend**: Vue 3 + TypeScript + Vite (in `tma/` subdirectory)
- **Backend**: Go + Fiber + DDD architecture (in `backend/` subdirectory)
- **Infrastructure**: VPS with nginx, Docker, Docker Compose, PostgreSQL, Redis

## Commands

### Frontend (TMA)
All commands run from the `tma/` directory using pnpm:

```bash
# Development
pnpm dev              # Start dev server (port 5173)
pnpm preview          # Preview production build (port 4173)

# Building
pnpm build            # Type-check + build for production
pnpm build-only       # Build only (skip type-check)

# Type Checking
pnpm type-check       # Run vue-tsc

# Linting
pnpm lint             # Run all linters (oxlint + eslint)
pnpm format           # Format with Prettier

# Testing
pnpm test:unit                              # Run Vitest unit tests
pnpm test:e2e                               # Run Playwright E2E tests
pnpm test:e2e --project=chromium            # Run E2E on specific browser
npx playwright install                       # Install browser drivers (first run)

# Telegram Mini App Setup
pnpm add @telegram-apps/sdk @telegram-apps/sdk-vue    # Official TMA SDK
pnpm add -D eruda                                      # Mobile debugging console
pnpm add pinia                                         # State management
pnpm add @vueuse/core                                  # Vue composables
pnpm add -D tailwindcss postcss autoprefixer          # Utility CSS (optional)
pnpm add @iconify/vue                                  # Icons (optional)
```

### Backend (Go API)
All commands run from the `backend/` directory:

```bash
# Development
go run cmd/api/main.go                      # Start dev server (port 3000)

# Building
go build -o quiz-sprint-api cmd/api/main.go # Build binary

# Testing
go test ./...                                # Run all tests
go test -v ./internal/domain/quiz           # Run specific package tests

# Dependencies
go mod download                              # Download dependencies
go mod tidy                                  # Clean up dependencies

# Formatting
go fmt ./...                                 # Format code

# Swagger Documentation
~/go/bin/swag init -g cmd/api/main.go -o docs  # Generate Swagger docs
# Access Swagger UI at: http://localhost:3000/swagger/index.html
```

## Architecture

### Monorepo Structure
- `tma/` - Vue 3 frontend application
- `backend/` - Go backend API with DDD architecture
- `infrastructure/` - VPS server configurations, nginx, systemd services
- `dev-tunnel/` - SSH tunnel scripts for HTTPS development
- `.github/workflows/` - CI/CD pipelines for frontend and backend

### Frontend Structure (`tma/src/`)
- `main.ts` - Vue app initialization
- `App.vue` - Root component
- `router/` - Vue Router configuration
- `views/` - Page components
- `__tests__/` - Vitest unit tests

### Backend Structure (`backend/`)
- `cmd/api/` - Application entry point
- `internal/domain/` - Domain models, business rules, repository interfaces (pure Go)
- `internal/application/` - Use cases (StartQuiz, SubmitAnswer, GetLeaderboard)
- `internal/infrastructure/` - HTTP handlers, WebSocket, persistence implementations
- `pkg/` - Shared utilities

Backend follows **Domain-Driven Design (DDD)** with clean architecture:
- Domain layer: Pure business logic, no dependencies
- Application layer: Use cases coordinating domain objects
- Infrastructure layer: Fiber HTTP handlers, WebSocket hub, in-memory repository

### Build Once, Deploy Many
The CI/CD uses a two-stage workflow for both frontend and backend:

**Frontend:**
1. `build.yml` - Runs quality checks (type-check, lint), builds, and uploads artifact
2. `deploy.yml` - Downloads artifact and deploys to staging or production

**Backend:**
1. `backend-build.yml` - Builds Go binary, runs tests, uploads artifact
2. `backend-docker-deploy.yml` - Builds Docker image, pushes to GHCR, deploys via docker-compose

### Environments

| Environment | Frontend URL | Backend API | Backend Port | Database |
|-------------|-------------|-------------|--------------|----------|
| Development | `dev.quiz-sprint-tma.online` | Local tunnel | 5173 | In-memory |
| Staging | `staging.quiz-sprint-tma.online` | `/api`, `/ws` | 3001 (Docker) | PostgreSQL (Docker) |
| Production | `quiz-sprint-tma.online` | `/api`, `/ws` | 3000 (Docker) | PostgreSQL (Docker) |

**API Endpoints:**
- REST API: `https://<domain>/api/v1/*`
- WebSocket: `wss://<domain>/ws/leaderboard/:id`
- Health: `https://<domain>/api/health`

## Tech Stack

### Frontend
- Vue 3.5 with Composition API (`<script setup>`)
- TypeScript 5.9
- Vite (dev server and bundler)
- Vue Router 4
- Vitest + Vue Test Utils (unit testing)
- Playwright (E2E testing)
- ESLint + Oxlint + Prettier (code quality)
- pnpm 9 (package manager)
- Node.js ^20.19.0 || >=22.12.0

**Telegram Mini App (TMA) Stack:**
- `@telegram-apps/sdk` - Official Telegram Mini Apps SDK (initialization, theme, events)
- `@telegram-apps/sdk-vue` - Vue 3 bindings for TMA SDK (composables)
- `@vkruglikov/react-telegram-web-app` - Alternative TMA library (optional)
- `eruda` - Mobile debugging console (dev tool for testing in Telegram)
- `vconsole` - Alternative mobile console (optional)

**Recommended additions for TMA:**
- `@tanstack/vue-query` or `@vueuse/core` - Data fetching and state management
- `pinia` - State management (for quiz session, user data)
- `@vueuse/core` - Vue composables utilities (useLocalStorage, useWebSocket)
- `tailwindcss` or `unocss` - Utility-first CSS (better than plain CSS for TMA)
- `@iconify/vue` - Icon components (Telegram-style icons)

### Backend
- Go 1.23
- Fiber v2 (web framework)
- WebSocket support (gofiber/websocket)
- PostgreSQL 16 (database, runs in Docker)
- Redis 7 (caching, runs in Docker)
- Docker + Docker Compose (containerized deployment)

### Infrastructure
- nginx (reverse proxy, SSL termination)
- Let's Encrypt (SSL certificates)
- GitHub Actions (CI/CD)
- VPS (Ubuntu 20.04+)

## Code Style

- No semicolons
- Single quotes
- 100 character line width
- Path alias: `@` maps to `./src`

## Backend Deployment

### Deployment via GitHub Actions (Docker)

1. Go to Actions tab
2. Run "Deploy Backend (Docker)" workflow:
   - Select environment (staging/production)
   - Builds Docker image and pushes to GitHub Container Registry
   - Deploys via docker-compose on VPS
3. Health check will run automatically

The workflow automatically:
- Builds the Docker image with the Go API
- Pushes to `ghcr.io/<repo>/quiz-sprint-api`
- Generates `docker-compose.yml` with API + PostgreSQL + Redis
- Pulls and starts all containers on VPS
- Runs health check

### Manual Deployment (if needed)

```bash
# On VPS
cd /opt/quiz-sprint/staging  # or production

# Check running containers
docker compose ps

# View logs
docker compose logs -f api

# Restart services
docker compose restart api

# Full restart (including DB)
docker compose down
docker compose up -d
```

### Backend API Structure

**REST Endpoints:**
- `GET /api/v1/quiz` - List all quizzes
- `GET /api/v1/quiz/:id` - Get quiz by ID
- `POST /api/v1/quiz/:id/start` - Start quiz session
- `POST /api/v1/quiz/session/:sessionId/answer` - Submit answer
- `GET /api/v1/quiz/:id/leaderboard` - Get leaderboard

**WebSocket:**
- `GET /ws/leaderboard/:id` - Real-time leaderboard updates

## Workflow Requirements (from AGENTS.md)

Before completing a session:
1. Track issues with `bd` (beads) tool
2. Always push to remote: `git push` is required
