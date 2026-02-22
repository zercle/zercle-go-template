# =============================================================================
# Multi-stage Dockerfile for Go REST API Template
# =============================================================================
# This Dockerfile uses a multi-stage build to create a minimal, secure
# production image containing only the compiled binary and necessary files.
#
# Build stages:
#   1. builder: Compiles the Go application with optimizations
#   2. runtime: Creates the final minimal image with non-root user
# =============================================================================

# -----------------------------------------------------------------------------
# Stage 1: Builder
# -----------------------------------------------------------------------------
FROM mirror.gcr.io/golang AS builder

# Build arguments for versioning (passed via --build-arg)
ARG APP_NAME=zercle-go-template
ARG VERSION=dev
ARG BUILD_TIME
ARG GIT_COMMIT

# Install required build dependencies:
#   - git: For fetching Go modules from private repos or git-based dependencies
#   - ca-certificates: For HTTPS requests during build
#   - tzdata: For timezone support
RUN apt-get update && apt-get install -y --no-install-recommends \
    git ca-certificates tzdata \
    && rm -rf /var/lib/apt/lists/*

# Set working directory inside the container
WORKDIR /build

# Copy go.mod and go.sum first to leverage Docker layer caching
# This layer will be cached unless these files change
COPY go.mod go.sum ./

# Download and verify Go dependencies
# This step is cached as long as go.mod/go.sum don't change
RUN go mod download && go mod verify

# Copy the rest of the source code
COPY . .

# Build the application with optimizations:
#   - CGO_ENABLED=0: Disable CGO for static binary (no libc dependency)
#   - GOOS=linux: Target Linux OS
#   - ldflags:
#       -s: Omit symbol table and debug info
#       -w: Omit DWARF symbol table
#       -extldflags '-static': Create fully static binary
#       -X: Inject version variables into the binary
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w \
    -extldflags '-static' \
    -X main.Version=${VERSION} \
    -X main.BuildTime=${BUILD_TIME} \
    -X main.GitCommit=${GIT_COMMIT}" \
    -o /build/bin/${APP_NAME} \
    ./cmd/api

# -----------------------------------------------------------------------------
# Stage 2: Runtime (Distroless)
# -----------------------------------------------------------------------------
FROM gcr.io/distroless/base:nonroot AS runtime

# Build argument for app name (must match builder stage)
ARG APP_NAME=zercle-go-template

# Set working directory
WORKDIR /app

# Copy the compiled binary from the builder stage
# The distroless nonroot user (UID 65532) owns this by default
COPY --from=builder /build/bin/${APP_NAME} /app/server

# Copy any additional required files (config templates, migrations, etc.)
# Uncomment if you have static files or configs to include
# COPY --from=builder /build/configs/ /app/configs/

# Use the distroless nonroot user (UID 65532)
USER nonroot:nonroot

# Expose the application port
# This should match the port configured in your application
EXPOSE 8080

# Set environment variables
# These can be overridden at runtime with -e flags
ENV APP_ENVIRONMENT=production \
    APP_LOG_LEVEL=info \
    APP_LOG_FORMAT=json \
    APP_SERVER_PORT=8080

# Use ENTRYPOINT to run the Go binary directly
# Go binaries handle signals properly without needing an init system
ENTRYPOINT ["/app/server"]

# CMD can be used to provide default arguments (optional)
# CMD ["--config", "/app/configs/config.yaml"]

# -----------------------------------------------------------------------------
# Metadata Labels (following OCI annotation conventions)
# -----------------------------------------------------------------------------
LABEL org.opencontainers.image.title="${APP_NAME}" \
      org.opencontainers.image.description="Go REST API Template" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.source="https://github.com/zercle/${APP_NAME}" \
      org.opencontainers.image.licenses="MIT"
