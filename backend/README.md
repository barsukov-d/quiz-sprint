# Quiz Sprint Backend API

Go backend API for Quiz Sprint TMA built with **Fiber**, **DDD architecture**, and **WebSocket** support.

## Tech Stack

- **Go 1.23**
- **Fiber v2** - Fast web framework (6M+ req/sec)
- **WebSocket** - Real-time leaderboard updates
- **PostgreSQL 16** - Primary database
- **Redis 7** - Caching (optional)
- **Domain-Driven Design** - Clean architecture

## Project Structure

```
backend/
├── cmd/api/                    # Application entry point
│   └── main.go                 # Server initialization
├── internal/
│   ├── domain/                 # Domain layer (pure Go, no dependencies)
│   │   └── quiz/
│   │       ├── quiz.go         # Domain models (Quiz, Question, Answer, Session)
│   │       ├── errors.go       # Domain errors
│   │       └── repository.go   # Repository interface
│   ├── application/            # Application layer (use cases)
│   │   └── quiz/
│   │       ├── start_quiz.go   # StartQuiz use case
│   │       ├── submit_answer.go # SubmitAnswer use case
│   │       └── get_leaderboard.go # GetLeaderboard use case
│   └── infrastructure/         # Infrastructure layer
│       ├── http/
│       │   ├── handlers/       # Fiber HTTP handlers
│       │   └── routes/         # Route configuration
│       └── persistence/        # Repository implementations
│           └── memory_repository.go # In-memory (for dev)
├── pkg/                        # Shared utilities
├── .env.example                # Environment variables template
├── go.mod                      # Go modules
└── README.md                   # This file
```

## DDD Layers

### Domain Layer (`internal/domain/`)
Pure business logic with **zero external dependencies**:
- Domain models: `Quiz`, `Question`, `Answer`, `QuizSession`
- Business rules: `CanStart()`, `HasPassed()`, `GetNextQuestion()`
- Repository interfaces (contracts)
- Domain errors

### Application Layer (`internal/application/`)
Use cases orchestrating domain objects:
- `StartQuizUseCase` - Create quiz session
- `SubmitAnswerUseCase` - Process answer, calculate score
- `GetLeaderboardUseCase` - Query leaderboard

### Infrastructure Layer (`internal/infrastructure/`)
Framework-specific code:
- Fiber HTTP handlers
- WebSocket hub for real-time updates
- Repository implementations (in-memory, PostgreSQL)

## API Endpoints

### REST API

**Base URL:** `/api/v1`

#### Quiz Operations
```
GET    /api/v1/quiz              # List all quizzes
GET    /api/v1/quiz/:id          # Get quiz by ID
POST   /api/v1/quiz/:id/start    # Start quiz session
GET    /api/v1/quiz/:id/leaderboard # Get leaderboard
```

#### Session Operations
```
POST   /api/v1/quiz/session/:sessionId/answer # Submit answer
```

### WebSocket

```
GET    /ws/leaderboard/:id       # Real-time leaderboard updates
```

## Running Locally

### Prerequisites
- Go 1.23+
- PostgreSQL 16 (optional, uses in-memory by default)

### Steps

1. **Clone and navigate:**
```bash
cd backend
```

2. **Install dependencies:**
```bash
go mod download
```

3. **Create `.env` file:**
```bash
cp .env.example .env
# Edit .env with your settings
```

4. **Run server:**
```bash
go run cmd/api/main.go
```

Server starts on `http://localhost:3000`

### Test Endpoints

```bash
# Health check
curl http://localhost:3000/health

# Get all quizzes
curl http://localhost:3000/api/v1/quiz

# Start a quiz
curl -X POST http://localhost:3000/api/v1/quiz/<quiz-id>/start \
  -H "Content-Type: application/json" \
  -d '{"userId": "telegram_user_123"}'
```

## Development

### Building
```bash
go build -o quiz-sprint-api cmd/api/main.go
```

### Testing
```bash
go test ./...                   # All tests
go test -v ./internal/domain/quiz  # Domain tests
```

### Formatting
```bash
go fmt ./...
```

### Code Coverage
```bash
go test -cover ./...
```

## Deployment

See [CLAUDE.md](../CLAUDE.md#backend-deployment) for full deployment guide.

### Quick Deploy via GitHub Actions

1. Push code to `main` branch
2. Go to Actions → "Build Backend" → Run workflow
3. Go to Actions → "Deploy Backend" → Select environment → Run workflow

### VPS Setup (First Time)

```bash
# On VPS
sudo ./infrastructure/scripts/setup-backend.sh
```

This sets up:
- Docker + Docker Compose
- PostgreSQL + Redis containers
- systemd services
- nginx reverse proxy
- User and directories

## Environment Variables

```bash
# Server
PORT=3000
ENV=production

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=quiz_user
DB_PASSWORD=your_password
DB_NAME=quiz_sprint
DB_SSLMODE=disable

# CORS
CORS_ORIGINS=https://quiz-sprint-tma.online

# Telegram
TELEGRAM_BOT_TOKEN=your_bot_token
```

## Architecture Decisions

### Why Fiber?
- **Performance**: 6M+ req/sec, fastest Go framework
- **WebSocket**: Built-in support for real-time features
- **Express-like API**: Familiar, easy to learn
- **Middleware**: Rich ecosystem (CORS, logger, recovery)

### Why DDD?
- **Maintainability**: Clear separation of concerns
- **Testability**: Pure domain logic, easy to test
- **Scalability**: Easy to add features without breaking existing code
- **Portability**: Business logic independent of framework

### Why In-Memory First?
- **Simplicity**: No database setup needed for development
- **Speed**: Fast iteration during prototyping
- **Testing**: Easy to test without database
- **Migration**: Easy to swap to PostgreSQL later (repository pattern)

## Next Steps

- [ ] Add PostgreSQL repository implementation
- [ ] Add authentication/authorization
- [ ] Add rate limiting
- [ ] Add request validation
- [ ] Add API documentation (Swagger)
- [ ] Add monitoring (Prometheus/Grafana)
- [ ] Add unit tests for use cases
- [ ] Add integration tests

## License

Private project - Quiz Sprint TMA
