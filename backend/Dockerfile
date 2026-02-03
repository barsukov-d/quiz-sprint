# ==========================================
# Stage 1: Base - Common dependencies
# ==========================================
FROM golang:1.25-alpine AS base

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files first (better caching)
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# ==========================================
# Stage 2: Development - Hot reload
# ==========================================
FROM base AS development

# Install air for hot reload
RUN go install github.com/air-verse/air@latest

# Copy source code (will be overridden by volume)
COPY . .

# Expose port
EXPOSE 3000

# Run with hot reload
CMD ["air", "-c", ".air.toml"]

# ==========================================
# Stage 3: Builder - Compile binary
# ==========================================
FROM base AS builder

# Copy source code
COPY . .

# Build binaries
# - CGO_ENABLED=0: Static binary (no C dependencies)
# - -ldflags="-s -w": Strip debug info (smaller binary)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -o /app/quiz-sprint-api \
    ./cmd/api

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -o /app/quiz-sprint-import \
    ./cmd/import

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -o /app/quiz-sprint-export \
    ./cmd/export

# ==========================================
# Stage 4: Production - Minimal runtime
# ==========================================
FROM alpine:3.19 AS production

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata wget

# Create non-root user for security
RUN addgroup -g 1000 app && \
    adduser -u 1000 -G app -s /bin/sh -D app

# Set working directory
WORKDIR /app

# Copy binaries, migrations, and quiz data from builder
COPY --from=builder /app/quiz-sprint-api .
COPY --from=builder /app/quiz-sprint-import .
COPY --from=builder /app/quiz-sprint-export .
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/data ./data

# Copy timezone data
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Set ownership
RUN chown -R app:app /app

# Switch to non-root user
USER app

# Expose port
EXPOSE 3000

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:3000/health || exit 1

# Run binary
ENTRYPOINT ["/app/quiz-sprint-api"]
