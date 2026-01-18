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

# API Code Generation (from Swagger/OpenAPI)
pnpm run generate:swagger    # Generate Swagger docs from Go code (backend)
pnpm run generate:api        # Generate TypeScript types from Swagger
pnpm run generate:all        # Generate both (Swagger + TypeScript)

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
# Development (Recommended - Docker with hot reload)
docker compose -f docker-compose.dev.yml up        # Start all services
docker compose -f docker-compose.dev.yml up -d     # Start in background
docker compose -f docker-compose.dev.yml down      # Stop all services
docker compose -f docker-compose.dev.yml down -v   # Stop and remove volumes (clean slate)
docker compose -f docker-compose.dev.yml logs -f api  # View API logs

# Services available:
# - API: http://localhost:3000 (with hot reload via air)
# - PostgreSQL: localhost:5432 (quiz_user/quiz_password_dev/quiz_sprint_dev)
# - Redis: localhost:6379
# - Adminer (DB UI): http://localhost:8080
# - Redis Commander: http://localhost:8081
# - Swagger UI: http://localhost:3000/swagger/index.html

# Local Development (Alternative - without Docker)
go run cmd/api/main.go                      # Start dev server (port 3000)
# Note: Requires PostgreSQL and Redis running locally or via docker-compose.dev.yml

# Building
go build -o quiz-sprint-api cmd/api/main.go # Build binary
docker build -t quiz-sprint-api .           # Build production Docker image
docker build -t quiz-sprint-api --target development .  # Build dev image

# Testing
go test ./...                                # Run all tests
go test -v ./internal/domain/quiz           # Run specific package tests

# Dependencies
go mod download                              # Download dependencies
go mod tidy                                  # Clean up dependencies

# Formatting
go fmt ./...                                 # Format code

# Swagger Generation (Makefile commands)
make swagger                                 # Generate Swagger docs
make swagger-install                         # Install swag CLI globally
make dev                                     # Generate swagger + run server
make help                                    # Show all Makefile commands

# Manual Swagger generation (if needed)
go run github.com/swaggo/swag/v2/cmd/swag@latest init --generalInfo cmd/api/main.go --output docs --parseDependency --parseInternal

# Database
docker compose -f docker-compose.dev.yml exec postgres psql -U quiz_user -d quiz_sprint_dev
# Run migrations: docker compose -f docker-compose.dev.yml exec api go run migrations/migrate.go
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
- `api/client.ts` - Axios client with runtime hostname detection for API URLs
- `api/generated/` - Auto-generated TypeScript types and Vue Query hooks (from Swagger)
- `__tests__/` - Vitest unit tests

**API Code Generation:**
The frontend uses [Kubb v4](https://kubb.dev/) to auto-generate TypeScript types and Vue Query hooks from the backend's Swagger/OpenAPI spec:
- **Types**: `src/api/generated/types/` - TypeScript interfaces from Swagger schemas
- **Hooks**: `src/api/generated/hooks/` - Vue Query hooks (useGetQuiz, usePostQuizIdStart, etc.)
- **Schemas**: `src/api/generated/schemas/` - Zod validation schemas
- **Config**: `kubb.config.ts` - Generation configuration
- **Trigger**: Run `pnpm run generate:api` after backend Swagger changes

### Backend Structure (`backend/`)
- `cmd/api/` - Application entry point with DB connection setup
- `internal/domain/` - Domain models, business rules, repository interfaces (pure Go)
  - `quiz/` - Quiz aggregate (Quiz, Question, Answer entities)
  - `user/` - User aggregate (User entity with profile management)
  - `shared/` - Shared value objects (UserID, etc.)
- `internal/application/` - Use cases orchestrating domain logic
  - `quiz/` - Quiz use cases (StartQuiz, SubmitAnswer, GetLeaderboard)
  - `user/` - User use cases (RegisterUser, UpdateProfile, GetUser)
- `internal/infrastructure/` - Technical implementations
  - `http/handlers/` - Fiber HTTP handlers with Swagger annotations
  - `http/handlers/swagger_models.go` - Response DTOs for Swagger
  - `persistence/postgres/` - PostgreSQL repository implementations
  - `persistence/memory/` - In-memory repositories (fallback/testing)
  - `messaging/` - Event bus implementation
- `migrations/` - SQL database migrations (auto-applied on startup)
  - `init.sql` - Initial schema (quizzes, questions, answers, sessions, leaderboard VIEW)
  - `002_create_users_table.sql` - Users table
- `docs/` - Auto-generated Swagger documentation (swagger.json, swagger.yaml)
- `pkg/database/` - PostgreSQL connection utilities

Backend follows **Domain-Driven Design (DDD)** with clean architecture:
- **Domain layer**: Pure business logic, no external dependencies (not even context.Context)
- **Application layer**: Use cases coordinating domain objects, DTOs for data transfer
- **Infrastructure layer**: Fiber HTTP handlers, PostgreSQL repositories, WebSocket hub

**Data Persistence Architecture:**
```
PostgreSQL (Production):
✅ users            → postgres.UserRepository
✅ quizzes          → postgres.QuizRepository (with questions & answers)
⚠️  quiz_sessions   → memory.SessionRepository (TODO: migrate to PostgreSQL)
⚠️  user_answers    → (part of session, TODO)
⚠️  leaderboard     → memory.LeaderboardRepository (TODO: use PostgreSQL VIEW)

In-Memory (Development/Fallback):
- memory.QuizRepository (with seed data)
- memory.SessionRepository
- memory.LeaderboardRepository
```

**Repository Pattern:**
- Interfaces defined in domain layer (pure Go)
- Implementations in infrastructure/persistence/
- Routes automatically select PostgreSQL if available, fallback to memory
- `main.go` establishes DB connection, passes to `routes.SetupRoutes(app, db)`

**Swagger/OpenAPI Integration:**
- Uses [swaggo/swag](https://github.com/swaggo/swag) for Go annotations
- All handlers in `internal/infrastructure/http/handlers/` have Swagger comments
- Response types are defined in `swagger_models.go` for type safety
- Frontend auto-generates TypeScript types from `docs/swagger.json`

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
| Development | `dev.quiz-sprint-tma.online` | Local tunnel | 3000 (local) | PostgreSQL (Docker) |
| Staging | `staging.quiz-sprint-tma.online` | `/api`, `/ws` | 3001 (Docker) | PostgreSQL (Docker) |
| Production | `quiz-sprint-tma.online` | `/api`, `/ws` | 3000 (Docker) | PostgreSQL (Docker) |

**API Endpoints:**
- REST API: `https://<domain>/api/v1/*`
- WebSocket: `wss://<domain>/ws/leaderboard/:id`
- Health: `https://<domain>/api/health`

### Development Tunnel Setup

For TMA development, you need HTTPS (Telegram requires it). Use SSH tunnels to expose localhost to your VPS with nginx reverse proxy:

**Setup:**
1. Start backend locally: `cd backend && go run cmd/api/main.go` (port 3000)
2. Start frontend locally: `cd tma && pnpm dev` (port 5173)
3. Start backend tunnel: `./dev-tunnel/start-backend-tunnel.sh` (forwards localhost:3000 → VPS:3000)
4. Start frontend tunnel: `./dev-tunnel/start-frontend-tunnel.sh` (forwards localhost:5173 → VPS:5173)
5. Access TMA at: `https://dev.quiz-sprint-tma.online`

**How it works:**
- Frontend client (`src/api/client.ts`) detects `window.location.hostname` at runtime
- When accessed via `dev.quiz-sprint-tma.online`, it uses `https://dev.quiz-sprint-tma.online/api/v1`
- nginx on VPS proxies `/api/*` → `localhost:3000` and `/` → `localhost:5173`
- SSH tunnels forward VPS ports to your MacBook's localhost

**Environment Variables:**
- `.env.development` - Used when running `pnpm dev` locally (localhost URLs)
- `.env.local` - For tunnel development (overrides .env.development, but runtime detection is more reliable)
- `.env.staging` - Staging environment
- `.env.production` - Production environment

All env files include `/api/v1` and `/ws` suffixes in base URLs.

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
- Go 1.25
- Fiber v3 (web framework)
- WebSocket support (gofiber/contrib/websocket)
- PostgreSQL 16 (database, runs in Docker)
  - `lib/pq` - PostgreSQL driver
  - Automatic migrations on startup (`migrations/` folder)
  - Connection pooling and health checks
- Redis 7 (caching, runs in Docker) - **TODO: not yet integrated**
- Docker + Docker Compose (containerized deployment)
- **swaggo/swag** - Swagger/OpenAPI 2.0 documentation generator
- **kubb** - TypeScript code generator from OpenAPI (frontend)
- **Air** - Hot reload for Go development in Docker

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

## Swagger/OpenAPI Code Generation

### Overview

The project uses a **code-first** approach with automatic type generation:

```
Go Handlers (with @annotations)
    ↓ swag generates
Swagger/OpenAPI spec (swagger.json)
    ↓ kubb generates
TypeScript types + Vue Query hooks
```

### Quick Commands

**From `tma/` directory (Recommended):**
```bash
pnpm run generate:swagger    # Generate Swagger from Go code (backend)
pnpm run generate:api        # Generate TypeScript from Swagger (frontend)
pnpm run generate:all        # Generate both in one command ✨
```

**From `backend/` directory:**
```bash
make swagger                 # Generate Swagger documentation
make help                    # Show all Makefile commands
```

### Generated Files

**Backend:**
- `backend/docs/swagger.json` - OpenAPI 2.0 specification (used by frontend)
- `backend/docs/swagger.yaml` - YAML version
- `backend/docs/docs.go` - Embedded Swagger UI

**Frontend:**
- `tma/src/api/generated/types/` - TypeScript type definitions
- `tma/src/api/generated/schemas/` - Zod validation schemas
- `tma/src/api/generated/hooks/` - Vue Query hooks (useGetQuiz, etc.)

### Typical Workflow

1. **Add/modify Go handler annotations:**
```go
// @Summary Get quiz by ID
// @Tags quiz
// @Param id path string true "Quiz ID"
// @Success 200 {object} handlers.QuizDTO
// @Router /quiz/{id} [get]
func (h *QuizHandler) GetQuiz(c *fiber.Ctx) error {
    // implementation
}
```

2. **Generate Swagger + TypeScript:**
```bash
cd tma
pnpm run generate:all
```

3. **Use in Vue components:**
```typescript
import { useGetQuizId } from '@/api/generated/hooks/quizController'

const { data: quiz, isLoading } = useGetQuizId({ id: '123' })
// quiz is fully typed with IntelliSense!
```

### Required Fields

To mark fields as required (generates non-optional TypeScript types):

```go
type QuizDTO struct {
    ID    string `json:"id" validate:"required"`       // required in TypeScript
    Title string `json:"title" validate:"required"`    // required in TypeScript
    Description string `json:"description"`             // optional in TypeScript
}
```

Generated TypeScript:
```typescript
export type QuizDTO = {
    id: string;          // required (no ?)
    title: string;       // required (no ?)
    description?: string; // optional (with ?)
}
```

### Documentation

- Detailed guide: `backend/SWAGGER.md`
- Quick reference: `QUICKSTART.md`

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

**Quiz Endpoints:**
- `GET /api/v1/quiz` - List all quizzes
- `GET /api/v1/quiz/:id` - Get quiz details (with questions & top scores)
- `POST /api/v1/quiz/:id/start` - Start quiz session
- `POST /api/v1/quiz/session/:sessionId/answer` - Submit answer
- `GET /api/v1/quiz/:id/leaderboard` - Get leaderboard

**User Endpoints:**
- `POST /api/v1/user/register` - Register or update user from Telegram (idempotent)
- `GET /api/v1/user/:id` - Get user profile by Telegram ID
- `PUT /api/v1/user/:id` - Update user profile
- `GET /api/v1/user/username/:username` - Get user by Telegram @username
- `GET /api/v1/users` - List all users (admin, with pagination)

**WebSocket:**
- `GET /ws/leaderboard/:id` - Real-time leaderboard updates

**Health & Docs:**
- `GET /health` - Health check endpoint
- `GET /swagger/index.html` - Swagger UI documentation

## Swagger → TypeScript Type Generation Workflow

When you update backend API handlers, follow this workflow to keep frontend types in sync:

1. **Update Go Handler** - Add/modify endpoint in `backend/internal/infrastructure/http/handlers/`
2. **Add Swagger Annotations** - Use swaggo comments with concrete types from `swagger_models.go`
3. **Generate Swagger Docs** - Run `swag init -g cmd/api/main.go -o docs` in `backend/`
4. **Generate TypeScript Types** - Run `pnpm run generate:api` in `tma/`
5. **Use Generated Hooks** - Import from `@/api/generated/hooks/`

**Example:**
```go
// In swagger_models.go - Define response DTO
type GetQuizDetailsResponse struct {
    Data GetQuizDetailsData `json:"data"`
}

// In quiz_handler.go - Use in Swagger annotation
// @Success 200 {object} handlers.GetQuizDetailsResponse "Quiz details"
```

**Generated TypeScript:**
```typescript
// Auto-generated in src/api/generated/hooks/quizController/useGetQuizId.ts
import { useGetQuizId } from '@/api'

const { data, isLoading, error } = useGetQuizId({ id: quizId })
// data is typed as GetQuizDetailsResponse
```

**Important:**
- Always use concrete types in `swagger_models.go`, never `map[string]interface{}`
- Response wrapper format: `{ data: ActualData }` for consistency
- Kubb config uses `dataReturnType: 'data'` but response still has `.data` property
- Frontend components must access `response.data` (e.g., `quizzes?.data`)

## Domain-Driven Design (DDD) Guidelines

### Working with Domains

The backend strictly follows **DDD + Clean Architecture** principles:

#### Domain Layer (`internal/domain/`)
**Pure business logic - NO external dependencies:**
- ✅ Use: Pure Go structs, interfaces, business methods
- ✅ Use: Value Objects for all IDs, measurements, descriptive objects
- ✅ Use: Factory methods (`NewQuiz`, `NewUser`)
- ✅ Use: `ReconstructEntity()` methods for loading from database
- ❌ NO: `context.Context`, JSON tags, database imports, HTTP imports
- ❌ NO: `time.Time` (use `int64` Unix timestamps)

**Example:**
```go
// Value Object
type UserID struct { value string }
func NewUserID(value string) (UserID, error) { ... }

// Entity with business logic
type User struct {
    id UserID
    username Username
    // ...
}

func (u *User) UpdateProfile(...) error {
    // Business rules here
}

// Reconstruct from DB (no validation)
func ReconstructUser(...) *User { return &User{...} }
```

#### Application Layer (`internal/application/`)
**Use Cases orchestrating domain logic:**
- ✅ Use: Input/Output DTOs (never domain models!)
- ✅ Use: `context.Context` for timeouts/cancellation
- ✅ Use: Orchestration of multiple aggregates
- ❌ NO: Business logic (delegate to domain)
- ❌ NO: HTTP concerns, database details

**Example:**
```go
type RegisterUserInput struct {
    UserID string
    Username string
}

type RegisterUserOutput struct {
    User UserDTO
    IsNewUser bool
}

func (uc *RegisterUserUseCase) Execute(input RegisterUserInput) (RegisterUserOutput, error) {
    // 1. Convert to domain types
    // 2. Execute domain logic
    // 3. Save via repository
    // 4. Return DTO
}
```

#### Infrastructure Layer (`internal/infrastructure/`)
**Technical implementations:**
- ✅ Use: HTTP handlers (thin adapters)
- ✅ Use: Repository implementations (PostgreSQL, in-memory)
- ✅ Use: Database/SQL, JSON tags, framework code
- ❌ NO: Business logic

**Handler Pattern:**
```go
func (h *UserHandler) RegisterUser(c fiber.Ctx) error {
    // 1. Parse HTTP request
    // 2. Convert to Use Case Input
    // 3. Execute Use Case
    // 4. Map domain errors → HTTP errors
    // 5. Return HTTP response
}
```

### Repository Pattern

**Interface in Domain, Implementation in Infrastructure:**

```go
// domain/user/repository.go
type UserRepository interface {
    FindByID(id UserID) (*User, error)  // NO context.Context!
    Save(user *User) error
}

// infrastructure/persistence/postgres/user_repository.go
type UserRepository struct { db *sql.DB }

func (r *UserRepository) FindByID(id UserID) (*User, error) {
    ctx := context.Background()  // Infrastructure adds context
    // SQL query with ctx
    // Use ReconstructUser() to rebuild entity
}
```

### Error Handling

**Domain Errors → HTTP Status Codes:**

Each handler has its own error mapper:
```go
// quiz_handler.go
func mapError(err error) error {
    switch err {
    case quiz.ErrQuizNotFound:
        return fiber.NewError(404, "Quiz not found")
    case quiz.ErrInvalidQuizID:
        return fiber.NewError(400, "Invalid quiz ID")
    // ...
    }
}

// user_handler.go
func mapUserError(err error) error {
    switch err {
    case user.ErrUserNotFound:
        return fiber.NewError(404, "User not found")
    // ...
    }
}
```

**Separation by Domain** - Each domain has its own error mapper for clean separation of concerns.

## Implementation Status & Recent Changes

### ✅ Fully Implemented Features

#### User Authentication & Management
- **Telegram Auth Middleware** (`backend/internal/infrastructure/http/middleware/telegram_auth.go`)
  - ✅ Base64 decoding of init data from Authorization header
  - ✅ Cryptographic signature validation using bot token
  - ✅ Expiration check (1 hour window, prevents replay attacks)
  - ✅ Stores validated `InitData` in request context (as pointer)
  - ✅ Secure: Client cannot forge user data (server validates signature)
  - **Format:** `Authorization: tma <base64-encoded-init-data-raw>`

- **User Registration Flow** (End-to-end working)
  - ✅ Frontend: `useAuth.ts` composable retrieves init data from Telegram SDK
  - ✅ Frontend: Base64-encodes and sends in Authorization header
  - ✅ Backend: Middleware validates signature
  - ✅ Backend: Handler extracts user info from validated data
  - ✅ Backend: Creates/updates user in PostgreSQL (idempotent)
  - ✅ Frontend: Receives user DTO and stores in global state

- **User Domain** (`backend/internal/domain/user/`)
  - ✅ Value Objects: UserID, Username, TelegramUsername, Email, AvatarURL, LanguageCode
  - ✅ Entity: User with profile management methods
  - ✅ Repository: PostgreSQL implementation with all CRUD operations
  - ✅ Database: `users` table created (migration 002)

- **User Use Cases** (`backend/internal/application/user/`)
  - ✅ RegisterUser - Register/update from Telegram (idempotent)
  - ✅ GetUser - Get user by ID
  - ✅ UpdateUserProfile - Update profile fields
  - ✅ UpdateUserLanguage - Update language preference
  - ✅ ListUsers - Paginated user list (admin)
  - ✅ GetUserByTelegramUsername - Find by @username

#### Quiz Domain
- **Quiz Aggregate** (`backend/internal/domain/quiz/`)
  - ✅ Entities: Quiz, Question, Answer, Session, UserAnswer
  - ✅ PostgreSQL repository for Quiz (with questions & answers)
  - ✅ In-memory repositories for Session and Leaderboard (fallback)

- **Quiz Use Cases** (`backend/internal/application/quiz/`)
  - ✅ StartQuiz - Create quiz session
  - ✅ SubmitAnswer - Submit answer and get feedback
  - ✅ GetLeaderboard - Get top scores

#### Frontend
- **Composables** (`tma/src/composables/`)
  - ✅ `useAuth.ts` - Telegram authentication & user state management
  - ✅ Type-safe with custom `ParsedInitData` interface
  - ✅ Handles both SDK and hash-based init data extraction
  - ✅ Base64 encoding for Authorization header

- **API Integration**
  - ✅ Automatic type generation from Swagger
  - ✅ Vue Query hooks for all endpoints
  - ✅ Axios interceptor adds Authorization header automatically
  - ✅ Runtime hostname detection for multi-environment support

- **Code Quality**
  - ✅ TypeScript strict mode - zero errors
  - ✅ ESLint + Oxlint - zero errors
  - ✅ Generated files excluded from linting
  - ✅ No `any` types (all properly typed)

### ⚠️ In Progress / TODO

#### Database
- ⚠️ Migrate `SessionRepository` from in-memory to PostgreSQL
- ⚠️ Migrate `LeaderboardRepository` to PostgreSQL VIEW
- ⚠️ Add database indexes for performance optimization
- ⚠️ Quiz session timeout/cleanup mechanism

#### Backend
- ⚠️ Redis integration for caching
- ⚠️ WebSocket real-time updates (infrastructure exists, needs testing)
- ⚠️ Rate limiting middleware
- ⚠️ Admin endpoints for quiz CRUD operations
- ⚠️ File upload support for quiz images

#### Frontend
- ⚠️ Quiz playing UI components
- ⚠️ Leaderboard display with real-time updates
- ⚠️ User profile page
- ⚠️ Quiz results/statistics page
- ⚠️ Telegram theme integration

### 🔧 Recent Fixes (2026-01-18)

#### Backend Fixes
1. **Telegram Auth Context Storage Bug**
   - **Issue:** Middleware stored init data as value, getter expected pointer → type assertion failed → nil
   - **Fix:** Changed `c.Locals("telegram_init_data", parsedData)` to `c.Locals("telegram_init_data", &parsedData)`
   - **File:** `backend/internal/infrastructure/http/middleware/telegram_auth.go:66`
   - **Impact:** Handler now successfully retrieves validated init data

2. **Users Table Migration**
   - **Issue:** `002_create_users_table.sql` existed but wasn't applied (Docker entrypoint only runs on first DB init)
   - **Fix:** Manually applied migration via `docker compose exec postgres psql -f /docker-entrypoint-initdb.d/002_create_users_table.sql`
   - **Result:** Users table created with all indexes

3. **RegisterUserRequest Swagger Schema**
   - **Issue:** Swagger schema required `userId` field in body, but handler uses Authorization header (not body)
   - **Fix:** Made `RegisterUserRequest` an empty struct with comment explaining data comes from header
   - **File:** `backend/internal/infrastructure/http/handlers/swagger_models.go:218-225`
   - **Impact:** TypeScript types now correctly reflect that endpoint doesn't need body data

#### Frontend Fixes
1. **TypeScript Type Errors in useAuth.ts**
   - **Issue:** `launchParams.initData` had type `unknown`, causing assignment errors
   - **Fix:** Created `ParsedInitData` interface and added type assertions
   - **Result:** Strict typing without `any`

2. **User Field Names (snake_case vs camelCase)**
   - **Issue:** TelegramUser type uses `first_name`, code used `firstName`
   - **Fix:** Updated all references to use SDK's snake_case convention
   - **Files:** `useAuth.ts:71-78, 118`

3. **App.vue registerUser Call**
   - **Issue:** After Swagger regeneration, `registerUser()` became parameterless (void)
   - **Fix:** Changed from `registerUser({ data: {} })` to `registerUser()`
   - **File:** `tma/src/App.vue:53`

4. **ESLint Configuration**
   - **Issue:** Linter errors in auto-generated Kubb files
   - **Fix:** Added `src/api/generated/**` to `globalIgnores`
   - **File:** `tma/eslint.config.ts:20`

5. **Unused Type Parameter**
   - **Issue:** `TError` parameter in fetch function was unused
   - **Fix:** Renamed to `_TError` (convention for intentionally unused params)
   - **File:** `tma/src/api/client.ts:30`

6. **Incorrect Import in Generated Schema**
   - **Issue:** Kubb generated wrong import path: `GetUserUsername.ts` instead of `GetUserUsernameUsername.ts`
   - **Fix:** Manually corrected import path
   - **File:** `tma/src/api/generated/schemas/userController/getUserUsernameUsernameSchema.ts:6`

### 📊 Current Database Status

**Viewing Database:**
- Web UI: http://localhost:8080 (Adminer)
  - Server: `postgres`
  - Username: `quiz_user`
  - Password: `quiz_password_dev`
  - Database: `quiz_sprint_dev`

- CLI:
  ```bash
  docker compose -f docker-compose.dev.yml exec postgres psql -U quiz_user -d quiz_sprint_dev
  ```

**Tables:**
- ✅ `users` - User profiles (Telegram integration)
- ✅ `quizzes` - Quiz metadata
- ✅ `questions` - Quiz questions
- ✅ `answers` - Answer options
- ✅ `quiz_sessions` - User quiz attempts (schema exists, using in-memory for now)

**Migrations Applied:**
- ✅ `init.sql` - Initial schema
- ✅ `002_create_users_table.sql` - Users table

### 🚀 How to Test End-to-End

1. **Start Backend:**
   ```bash
   docker compose -f docker-compose.dev.yml up -d
   # Check logs: docker compose -f docker-compose.dev.yml logs -f api
   ```

2. **Start Frontend:**
   ```bash
   cd tma && pnpm dev
   ```

3. **Start Tunnels (for Telegram access):**
   ```bash
   ./dev-tunnel/start-backend-tunnel.sh
   ./dev-tunnel/start-frontend-tunnel.sh
   ```

4. **Open in Telegram:**
   - URL: `https://dev.quiz-sprint-tma.online`
   - User registration happens automatically on load
   - Check backend logs for auth flow
   - Check Adminer for new user in database

### 🔐 Security Notes

**Telegram Authentication:**
- ✅ All user data comes from Telegram-signed init data (not client input)
- ✅ Server validates cryptographic signature before trusting data
- ✅ 1-hour expiration window prevents replay attacks
- ✅ Client cannot forge user IDs or usernames
- ✅ Base64 encoding for safe HTTP header transmission

**Best Practices:**
- ✅ Separate error mappers per domain
- ✅ No sensitive data in logs (bot token truncated)
- ✅ Database connection pooling
- ✅ CORS configuration per environment
- ✅ Type-safe DTOs prevent data leaks

## Workflow Requirements (from AGENTS.md)

Before completing a session:
1. Track issues with `bd` (beads) tool
2. Always push to remote: `git push` is required
