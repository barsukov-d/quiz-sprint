# Quiz Sprint Backend Makefile

.PHONY: help
help: ## Show this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

.PHONY: run
run: ## Run the API server
	go run cmd/api/main.go

.PHONY: build
build: ## Build the API binary
	go build -o quiz-sprint-api cmd/api/main.go

.PHONY: test
test: ## Run all tests
	go test ./...

.PHONY: test-verbose
test-verbose: ## Run tests with verbose output
	go test -v ./...

.PHONY: swagger
swagger: ## Generate Swagger documentation
	@echo "Generating Swagger documentation..."
	go run github.com/swaggo/swag/v2/cmd/swag@latest init \
		--generalInfo cmd/api/main.go \
		--output docs \
		--parseDependency \
		--parseInternal
	@echo "✅ Swagger generated: docs/swagger.json"

.PHONY: swagger-install
swagger-install: ## Install swag v2 CLI globally
	go install github.com/swaggo/swag/v2/cmd/swag@latest

.PHONY: deps
deps: ## Download dependencies
	go mod download

.PHONY: tidy
tidy: ## Clean up dependencies
	go mod tidy

.PHONY: fmt
fmt: ## Format code
	go fmt ./...

.PHONY: lint
lint: ## Run linter (requires golangci-lint)
	golangci-lint run

.PHONY: clean
clean: ## Clean build artifacts
	rm -f quiz-sprint-api
	rm -rf docs/

.PHONY: dev
dev: swagger run ## Generate swagger and run server

.PHONY: all
all: swagger build ## Generate swagger and build binary

# ============================================
# Quiz Import Commands
# ============================================

.PHONY: import-quiz
import-quiz: ## Import a single quiz from JSON file (usage: make import-quiz FILE=path/to/quiz.json)
	@if [ -z "$(FILE)" ]; then \
		echo "Error: FILE parameter is required"; \
		echo "Usage: make import-quiz FILE=data/quizzes/example.json"; \
		exit 1; \
	fi
	go run cmd/import/main.go -file=$(FILE)

.PHONY: import-quiz-dry-run
import-quiz-dry-run: ## Validate a quiz JSON file without importing (usage: make import-quiz-dry-run FILE=path/to/quiz.json)
	@if [ -z "$(FILE)" ]; then \
		echo "Error: FILE parameter is required"; \
		echo "Usage: make import-quiz-dry-run FILE=data/quizzes/example.json"; \
		exit 1; \
	fi
	go run cmd/import/main.go -file=$(FILE) -dry-run

.PHONY: import-all-quizzes
import-all-quizzes: ## Import all quiz JSON files from data/quizzes directory
	go run cmd/import/main.go -dir=data/quizzes

.PHONY: import-all-quizzes-dry-run
import-all-quizzes-dry-run: ## Validate all quiz JSON files without importing
	go run cmd/import/main.go -dir=data/quizzes -dry-run

# ============================================
# Docker Import Commands (staging/production)
# ============================================

.PHONY: docker-import-all
docker-import-all: ## Import all quizzes via Docker (for staging/production)
	docker compose exec api /app/quiz-sprint-import -dir=/app/data/quizzes

.PHONY: docker-import-quiz
docker-import-quiz: ## Import a single quiz via Docker (usage: make docker-import-quiz FILE=legacy/movie-trivia.json)
	@if [ -z "$(FILE)" ]; then \
		echo "Error: FILE parameter is required"; \
		echo "Usage: make docker-import-quiz FILE=legacy/movie-trivia.json"; \
		exit 1; \
	fi
	docker compose exec api /app/quiz-sprint-import -file=/app/data/quizzes/$(FILE)

# ============================================
# Quiz Export Commands
# ============================================

.PHONY: export-all-quizzes
export-all-quizzes: ## Export all quizzes from DB to JSON files (individual)
	go run cmd/export/main.go -dir=data/quizzes/exported

.PHONY: export-all-quizzes-batch
export-all-quizzes-batch: ## Export all quizzes as a single batch JSON file
	go run cmd/export/main.go -dir=data/quizzes/exported -batch

.PHONY: export-quiz
export-quiz: ## Export a single quiz by ID (usage: make export-quiz ID=<uuid>)
	@if [ -z "$(ID)" ]; then \
		echo "Error: ID parameter is required"; \
		echo "Usage: make export-quiz ID=<quiz-uuid>"; \
		exit 1; \
	fi
	go run cmd/export/main.go -id=$(ID) -dir=data/quizzes/exported

.PHONY: list-quizzes
list-quizzes: ## List all quizzes in the database
	go run cmd/export/main.go -list

# ============================================
# Docker Export Commands (staging/production)
# ============================================

.PHONY: docker-export-all
docker-export-all: ## Export all quizzes via Docker (individual files)
	docker compose exec api /app/quiz-sprint-export -dir=/app/data/quizzes/exported

.PHONY: docker-export-all-batch
docker-export-all-batch: ## Export all quizzes via Docker (single batch file)
	docker compose exec api /app/quiz-sprint-export -dir=/app/data/quizzes/exported -batch

.PHONY: docker-list-quizzes
docker-list-quizzes: ## List all quizzes via Docker
	docker compose exec api /app/quiz-sprint-export -list

# ============================================
# Daily Challenge Debug Commands
# ============================================

.PHONY: daily-debug
daily-debug: ## Show daily challenge debug info (usage: make daily-debug [USER=123] [DATE=YYYY-MM-DD])
	@./scripts/debug-daily.sh $(USER) $(DATE)

.PHONY: daily-reset
daily-reset: ## Reset daily challenge with confirmation (usage: make daily-reset [USER=123] [DATE=YYYY-MM-DD])
	@./scripts/reset-daily.sh $(USER) $(DATE)

.PHONY: daily-quick-reset
daily-quick-reset: ## Quick reset without confirmation (usage: make daily-quick-reset [USER=123] [DATE=YYYY-MM-DD])
	@./scripts/quick-reset.sh $(USER) $(DATE)

.PHONY: daily-reset-all
daily-reset-all: ## Reset all users for date (usage: make daily-reset-all [DATE=YYYY-MM-DD])
	@./scripts/reset-daily.sh --all $(DATE)
