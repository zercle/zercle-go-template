# Multi-stage Dockerfile for Go Fiber Microservice

# Stage 1: Build stage
FROM golang:alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY cmd cmd/
COPY internal internal/
COPY pkg pkg/
COPY sql sql/

# Install swag for generating swagger docs
RUN go install github.com/swaggo/swag/cmd/swag@latest

# Generate swagger docs
RUN swag init -g cmd/server/main.go -o docs

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo \
    -ldflags="-w -s" \
    -o /app/server \
    ./cmd/server

# Stage 2: Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata tini

ENTRYPOINT ["/sbin/tini", "--"]

# Create app user
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder --chown=appuser:appuser /app/server .

# Copy configuration files
COPY --chown=appuser:appuser .env.example .
COPY --chown=appuser:appuser configs/ configs/

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 3000

# Health check (requires wget in alpine)
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --quiet --tries=1 --spider http://localhost:3000/health/live || exit 1

# Run the application
CMD ["./server"]
