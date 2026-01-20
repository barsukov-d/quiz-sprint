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
