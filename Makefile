.PHONY: help init generate build test test-unit test-integration test-short test-all test-coverage-check lint fmt clean docker-build docker-up docker-down migrate-up migrate-down dev run build-health build-user build-post build-all build-custom test-db-up test-db-down test-db-reset test-db-shell install-tools test-coverage

# Variables
GO := go
GOFLAGS := -v
GOPROXY := direct
BINARY_NAME := service
BINARY_DIR := ./bin
BUILD_DIR := ./cmd/server

# Database configuration
DB_HOST := localhost
DB_PORT := 5432
DB_NAME := postgres
DB_USER := postgres
DB_PASSWORD := postgres
DB_SSL_MODE := disable
DB_URL := postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSL_MODE)

# Migration paths
MIGRATION_PATH := sqlc/migrations

# Coverage thresholds
COVERAGE_THRESHOLD_DEFAULT := 75
COVERAGE_THRESHOLD_INTEGRATION := 45
COVERAGE_THRESHOLD_MOCK := 40
COVERAGE_THRESHOLD_INFRA := 65

# Test configuration
TEST_PARALLEL := 4

# Container Runtime (Podman with Docker compatibility)
ifneq ($(shell command -v podman 2>/dev/null),)
  CONTAINER_RUNTIME := podman
  COMPOSE_CMD := podman-compose
else ifneq ($(shell command -v docker 2>/dev/null),)
  CONTAINER_RUNTIME := docker
  COMPOSE_CMD := docker-compose
else
  $(error "Neither podman nor docker found. Please install one of them.")
endif

# Docker compose files
COMPOSE_FILE := docker-compose.yml
COMPOSE_TEST_FILE := docker-compose.test.yml

# Build tags
BUILDTAGS_HEALTH := -tags=health
BUILDTAGS_USER := -tags=user
BUILDTAGS_POST := -tags=post
BUILDTAGS_ALL := -tags=all

# Help target
help:
	@echo "Available targets:"
	@echo ""
	@echo "Project Setup:"
	@echo "  init              - Initialize project dependencies"
	@echo "  install-tools    - Install development tools"
	@echo "  generate          - Generate sqlc code and mocks"
	@echo ""
	@echo "Build:"
	@echo "  build             - Build application (includes all handlers)"
	@echo "  build-health      - Build with health handler only"
	@echo "  build-user        - Build with user handler only"
	@echo "  build-post        - Build with post handler only"
	@echo "  build-all         - Build with all handlers explicitly"
	@echo "  build-custom      - Build with custom tags (use TAGS='tag1,tag2')"
	@echo "  dev               - Run in development mode"
	@echo "  run               - Run the compiled binary"
	@echo "  clean             - Clean build artifacts"
	@echo ""
	@echo "Testing:"
	@echo "  test-unit         - Run unit tests only (fast, no DB)"
	@echo "  test-integration  - Run integration tests (uses testcontainers)"
	@echo "  test-short        - Run fast tests only (-short flag)"
	@echo "  test-all          - Run all tests with coverage"
	@echo "  test-coverage     - Run tests with coverage (single file)"
	@echo "  test-coverage-check - Run tests and enforce 75% coverage threshold"
	@echo ""
	@echo "Code Quality:"
	@echo "  lint              - Run golangci-lint"
	@echo "  fmt               - Format code with gofmt"
	@echo ""
	@echo "Docker:"
	@echo "  docker-build      - Build Docker image"
	@echo "  docker-up         - Start Docker containers (PostgreSQL 18)"
	@echo "  docker-down       - Stop Docker containers"
	@echo ""
	@echo "Migrations:"
	@echo "  migrate-up        - Run database migrations"
	@echo "  migrate-down      - Rollback database migrations"
	@echo ""
	@echo "Note: Integration tests use testcontainers-go with automatic Podman/Docker detection"

# Initialize project
init:
	@echo "Initializing project..."
	$(GO) mod tidy
	$(GO) mod download

# Install tools
install-tools:
	@echo "Installing development tools..."
	$(GO) install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
	$(GO) install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	$(GO) install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	$(GO) install github.com/swaggo/swag/cmd/swag@latest

# Build the application (default - includes all handlers)
build: clean
	@echo "Building application..."
	@mkdir -p $(BINARY_DIR)
	$(GO) build $(GOFLAGS) -o $(BINARY_DIR)/$(BINARY_NAME) $(BUILD_DIR)

# Generic build function with tags
define BUILD_WITH_TAGS
	@echo "Building with $(2)..."
	@mkdir -p $(BINARY_DIR)
	$(GO) build $(GOFLAGS) $(1) -o $(BINARY_DIR)/$(BINARY_NAME) $(BUILD_DIR)
	@echo "Binary built: $(BINARY_DIR)/$(BINARY_NAME)"
endef

# Build with specific tags
build-health: clean
	$(call BUILD_WITH_TAGS,$(BUILDTAGS_HEALTH),health handler)

build-user: clean
	$(call BUILD_WITH_TAGS,$(BUILDTAGS_USER),user handler)

build-post: clean
	$(call BUILD_WITH_TAGS,$(BUILDTAGS_POST),post handler)

build-all: clean
	$(call BUILD_WITH_TAGS,$(BUILDTAGS_ALL),all handlers)

build-custom:
	$(call BUILD_WITH_TAGS,-tags="$(TAGS)",custom tags: $(TAGS))

# Run the application
run:
	@echo "Running application..."
	./$(BINARY_DIR)/$(BINARY_NAME)

# Development mode
dev:
	@echo "Running in development mode..."
	$(GO) run $(BUILD_DIR)

# Generate sqlc code and mocks
generate:
	@echo "Generating sqlc code..."
	go run github.com/sqlc-dev/sqlc/cmd/sqlc@latest generate
	@echo "Generating swagger docs..."
	swag init -g cmd/server/main.go -o docs --parseDependency --parseInternal
	@echo "Generating mocks..."
	$(GO) generate ./...

# Run linter
lint:
	@echo "Running golangci-lint..."
	GOPROXY=$(GOPROXY) golangci-lint run ./...

# Format code
fmt:
	@echo "Formatting code..."
	$(GO) fmt ./...
	gofmt -s -w .

# Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GO) clean
	rm -rf $(BINARY_DIR)
	rm -f coverage.out coverage.html coverage_unit.out coverage_integration.out

# Build container image
docker-build:
	@echo "Building container image using $(CONTAINER_RUNTIME)..."
	$(CONTAINER_RUNTIME) build -t zercle-go-template:latest .

# Start containers with compose
docker-up:
	@echo "Starting containers using $(CONTAINER_RUNTIME)..."
	$(COMPOSE_CMD) -f $(COMPOSE_FILE) up -d

# Stop containers
docker-down:
	@echo "Stopping containers using $(CONTAINER_RUNTIME)..."
	$(COMPOSE_CMD) -f $(COMPOSE_FILE) down

# Run database migrations
migrate-up:
	@echo "Running migrations..."
	migrate -path $(MIGRATION_PATH) -database "$(DB_URL)" up

# Rollback database migrations
migrate-down:
	@echo "Rolling back migrations..."
	migrate -path $(MIGRATION_PATH) -database "$(DB_URL)" down

# ============================================
# Testing Targets
# ============================================
test-unit:
	@echo "🧪 Running unit tests..."
	$(GO) test -v -race -short -count=1 ./internal/domain/... ./pkg/... ./internal/infrastructure/...
	@echo "✅ Unit tests complete"

test-integration:
	@echo "🧪 Running integration tests with testcontainers..."
	@echo "Note: testcontainers will automatically use Podman if available"
	$(GO) test -v -race -count=1 ./test/integration/...
	@echo "✅ Integration tests complete"

test-short:
	@echo "⚡ Running fast tests only..."
	$(GO) test -v -race -short -count=1 ./...
	@echo "✅ Fast tests complete"

# Run tests with coverage (single file)
test-coverage:
	@echo "🧪 Running tests with coverage..."
	$(GO) test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "📊 Coverage report: coverage.html"

# Run all tests with separate coverage profiles
test-all:
	@echo "🧪 Running all tests with coverage..."
	@echo "Running unit tests..."
	$(GO) test -v -race -coverprofile=coverage_unit.out -covermode=atomic -parallel $(TEST_PARALLEL) $$(go list ./... | grep -v '^\.$$' | grep -vE '/test/integration$$')
	@echo "Running integration tests..."
	$(GO) test -v -race -coverprofile=coverage_integration.out -covermode=atomic -parallel $(TEST_PARALLEL) github.com/zercle/zercle-go-template/test/integration
	@echo "Merging coverage profiles..."
	@echo "mode: atomic" > coverage.out
	@tail -q -n +2 coverage_unit.out coverage_integration.out 2>/dev/null >> coverage.out || true
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "📊 Coverage report: coverage.html"

# Check coverage threshold per package
test-coverage-check: test-all
	@echo "📊 Checking coverage threshold..."
	@echo ""
	@FAILED=0; \
	for pkg in $$(go list ./... | grep -v -E '/(docs|mock)$$' | grep -v '^$$'); do \
		if [ ! -d "$$pkg" ] || ! ls "$$pkg"/*_test.go 1>/dev/null 2>&1; then \
			continue; \
		fi; \
		result=$$(go test -covermode=atomic $$pkg 2>&1 | head -5); \
		coverage=$$(echo "$$result" | grep -E 'coverage:' | grep -oE '[0-9]+\.[0-9]+' | head -1); \
		if [ -z "$$coverage" ]; then \
			continue; \
		fi; \
		threshold=$(COVERAGE_THRESHOLD_DEFAULT); \
		echo "$$pkg" | grep -qE '/test/integration$$' && threshold=$(COVERAGE_THRESHOLD_INTEGRATION); \
		echo "$$pkg" | grep -qE '/test/mock$$' && threshold=$(COVERAGE_THRESHOLD_MOCK); \
		echo "$$pkg" | grep -qE '/(infrastructure/db|cmd/server)$$' && threshold=$(COVERAGE_THRESHOLD_INFRA); \
		if [ "$$(echo "$$coverage < $$threshold" | bc -l 2>/dev/null || echo 0)" -eq 1 ]; then \
			echo "❌ $$pkg: $$coverage% (below $$threshold%)"; \
			FAILED=1; \
		else \
			echo "✅ $$pkg: $$coverage%"; \
		fi; \
	done; \
	echo ""; \
	if [ "$$FAILED" -eq 1 ]; then \
		echo "❌ Some packages are below coverage threshold"; \
		exit 1; \
	fi; \
	echo "✅ All packages with tests meet coverage threshold"

# Legacy test target
test: test-all

.DEFAULT_GOAL := help
