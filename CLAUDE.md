# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Quiz Sprint TMA is a full-stack Telegram Mini App:
- **Frontend**: Vue 3 + TypeScript + Vite (in `tma/` subdirectory)
- **Backend**: Go + Fiber + DDD architecture (in `backend/` subdirectory)
- **Infrastructure**: VPS with nginx, systemd, PostgreSQL, Redis

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
2. `backend-deploy.yml` - Downloads binary, deploys to VPS, restarts systemd service

### Environments

| Environment | Frontend URL | Backend API | Backend Port | Database |
|-------------|-------------|-------------|--------------|----------|
| Development | `dev.quiz-sprint-tma.online` | Local tunnel | 5173 | In-memory |
| Staging | `staging.quiz-sprint-tma.online` | `/api`, `/ws` | 3001 | PostgreSQL staging |
| Production | `quiz-sprint-tma.online` | `/api`, `/ws` | 3000 | PostgreSQL production |

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

### Backend
- Go 1.23
- Fiber v2 (web framework)
- WebSocket support (gofiber/websocket)
- PostgreSQL 16 (database)
- Redis 7 (caching, optional)
- Docker + Docker Compose (database containers)
- systemd (process management)

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

### First-Time VPS Setup

```bash
# SSH into VPS
ssh root@your-vps-ip

# Clone repo and run setup
git clone <repo-url>
cd quiz-sprint/infrastructure/scripts
sudo ./setup-backend.sh

# Edit database passwords
nano /opt/quiz-sprint/.env

# Start databases
cd /opt/quiz-sprint
docker compose up -d

# Verify
docker compose ps
systemctl status quiz-sprint-api-staging
systemctl status quiz-sprint-api-production
```

### Deployment via GitHub Actions

1. Go to Actions tab
2. Run "Build Backend" workflow (builds binary)
3. Run "Deploy Backend" workflow:
   - Select environment (staging/production)
   - Select artifact (or leave empty for latest)
4. Health check will run automatically

### Manual Deployment (if needed)

```bash
# On VPS
cd /opt/quiz-sprint/staging  # or production
systemctl stop quiz-sprint-api-staging
# Replace binary
systemctl start quiz-sprint-api-staging
systemctl status quiz-sprint-api-staging

# Check logs
journalctl -u quiz-sprint-api-staging -f
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
