# ==========================================
# Stage 1: Build
# ==========================================
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files first (better caching)
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build binary
# - CGO_ENABLED=0: Static binary (no C dependencies)
# - -ldflags="-s -w": Strip debug info (smaller binary)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -o /app/quiz-sprint-api \
    ./cmd/api

# ==========================================
# Stage 2: Runtime
# ==========================================
FROM alpine:3.19

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user for security
RUN addgroup -g 1000 app && \
    adduser -u 1000 -G app -s /bin/sh -D app

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/quiz-sprint-api .

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
