# Go REST API Template Makefile
# Provides convenient commands for building, testing, linting, and deployment

# =============================================================================
# Variables
# =============================================================================

# Application name
APP_NAME := zercle-go-template

# Go module path
MODULE_PATH := zercle-go-template

# Binary output directory
BIN_DIR := bin

# Build output name
BINARY := $(BIN_DIR)/$(APP_NAME)

# Go version (should match go.mod)
GO_VERSION := 1.24

# Docker image configuration
DOCKER_REGISTRY ?=
DOCKER_IMAGE := $(if $(DOCKER_REGISTRY),$(DOCKER_REGISTRY)/$(APP_NAME),$(APP_NAME))
DOCKER_TAG ?= latest

# Build flags for optimization and versioning
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
GIT_BRANCH := $(shell git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")

# LDFLAGS for embedding version info into the binary
LDFLAGS := -ldflags "\
	-s -w \
	-X main.Version=$(GIT_BRANCH)-$(GIT_COMMIT) \
	-X main.BuildTime=$(BUILD_TIME) \
	-X main.GitCommit=$(GIT_COMMIT) \
"

# Go build flags
GOFLAGS := -trimpath

# Test flags with race detection
TEST_FLAGS := -v -race -count=1

# Coverage output file
COVERAGE_FILE := coverage.out
COVERAGE_HTML := coverage.html

# =============================================================================
# Default Target
# =============================================================================

.PHONY: all
all: clean deps fmt vet lint test build ## Run all checks and build the application

# =============================================================================
# Build Targets
# =============================================================================

.PHONY: build
build: deps ## Build the application binary
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 go build $(GOFLAGS) $(LDFLAGS) -o $(BINARY) ./cmd/api
	@echo "Binary created: $(BINARY)"

.PHONY: build-linux
build-linux: deps ## Cross-compile for Linux AMD64
	@echo "Building $(APP_NAME) for Linux AMD64..."
	@mkdir -p $(BIN_DIR)
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build $(GOFLAGS) $(LDFLAGS) -o $(BINARY)-linux-amd64 ./cmd/api
	@echo "Binary created: $(BINARY)-linux-amd64"

.PHONY: build-all
build-all: build build-linux ## Build for all supported platforms

.PHONY: run
run: deps ## Run the application locally (requires config)
	@echo "Running $(APP_NAME)..."
	go run ./cmd/api

# =============================================================================
# Test Targets
# =============================================================================

.PHONY: test
test: ## Run all tests with race detection
	@echo "Running tests..."
	go test $(TEST_FLAGS) ./...

.PHONY: test-unit
test-unit: ## Run unit tests only (excludes integration tests)
	@echo "Running unit tests..."
	go test $(TEST_FLAGS) ./...

.PHONY: test-integration
test-integration: ## Run integration tests only (requires database)
	@echo "Running integration tests..."
	go test -v -race -tags=integration ./...

.PHONY: test-coverage
test-coverage: ## Generate test coverage report
	@echo "Running tests with coverage..."
	go test -race -coverprofile=$(COVERAGE_FILE) -covermode=atomic ./...
	@echo "Coverage report generated: $(COVERAGE_FILE)"
	@go tool cover -func=$(COVERAGE_FILE) | tail -1

.PHONY: test-coverage-html
test-coverage-html: test-coverage ## Generate HTML coverage report
	@echo "Generating HTML coverage report..."
	go tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "HTML coverage report generated: $(COVERAGE_HTML)"

.PHONY: test-short
test-short: ## Run short tests only (skips integration tests)
	@echo "Running short tests..."
	go test -short -v ./...

.PHONY: benchmark
benchmark: ## Run all benchmarks
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./...

# =============================================================================
# Code Quality Targets
# =============================================================================

.PHONY: lint
lint: ## Run golangci-lint (requires installation: https://golangci-lint.run/usage/install/)
	@echo "Running golangci-lint..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run --timeout=5m ./...; \
	else \
		echo "golangci-lint not found. Install with:"; \
		echo "  brew install golangci-lint  # macOS"; \
		echo "  or visit: https://golangci-lint.run/usage/install/"; \
		exit 1; \
	fi

.PHONY: fmt
fmt: ## Format code with gofmt
	@echo "Formatting code..."
	gofmt -s -w .

.PHONY: vet
vet: ## Run go vet
	@echo "Running go vet..."
	go vet ./...

.PHONY: fmt-check
fmt-check: ## Check if code is formatted (useful for CI)
	@echo "Checking code formatting..."
	@if [ "$$(gofmt -s -l . | wc -l)" -gt 0 ]; then \
		echo "The following files are not formatted:"; \
		gofmt -s -l .; \
		exit 1; \
	else \
		echo "All files are formatted correctly"; \
	fi

.PHONY: security
security: ## Run security scanner with gosec (requires installation)
	@echo "Running security scan..."
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "gosec not found. Install with:"; \
		echo "  go install github.com/securego/gosec/v2/cmd/gosec@latest"; \
		exit 1; \
	fi

.PHONY: check
check: fmt vet lint test ## Run all code quality checks (fmt, vet, lint, test)

# =============================================================================
# Dependency Management
# =============================================================================

.PHONY: deps
deps: ## Download and verify dependencies
	@echo "Downloading dependencies..."
	go mod download
	go mod verify

.PHONY: deps-update
deps-update: ## Update dependencies
	@echo "Updating dependencies..."
	go get -u ./...
	go mod tidy

.PHONY: deps-graph
deps-graph: ## Show dependency graph (requires graphviz)
	go mod graph | head -30

.PHONY: tidy
tidy: ## Run go mod tidy to clean up dependencies
	@echo "Running go mod tidy..."
	go mod tidy

# =============================================================================
# Docker Targets
# =============================================================================

.PHONY: docker-build
docker-build: ## Build Docker image
	@echo "Building Docker image $(DOCKER_IMAGE):$(DOCKER_TAG)..."
	docker build \
		--build-arg APP_NAME=$(APP_NAME) \
		--build-arg GO_VERSION=$(GO_VERSION) \
		-t $(DOCKER_IMAGE):$(DOCKER_TAG) \
		-t $(DOCKER_IMAGE):latest \
		.

.PHONY: docker-run
docker-run: ## Run Docker container locally
	@echo "Running Docker container..."
	docker run --rm \
		-p 8080:8080 \
		-e APP_ENVIRONMENT=production \
		-e APP_LOG_LEVEL=info \
		$(DOCKER_IMAGE):$(DOCKER_TAG)

.PHONY: docker-push
docker-push: ## Push Docker image to registry
	@echo "Pushing Docker image $(DOCKER_IMAGE):$(DOCKER_TAG)..."
	docker push $(DOCKER_IMAGE):$(DOCKER_TAG)
	docker push $(DOCKER_IMAGE):latest

.PHONY: docker-scan
docker-scan: docker-build ## Scan Docker image for vulnerabilities
	@echo "Scanning Docker image for vulnerabilities..."
	@if command -v docker >/dev/null 2>&1 && docker scan --version >/dev/null 2>&1; then \
		docker scan $(DOCKER_IMAGE):$(DOCKER_TAG); \
	else \
		echo "Docker scan not available. Consider using Trivy:"; \
		echo "  brew install trivy  # macOS"; \
		echo "  trivy image $(DOCKER_IMAGE):$(DOCKER_TAG)"; \
	fi

# =============================================================================
# Development Tools
# =============================================================================

.PHONY: swagger
swagger: ## Generate Swagger documentation (requires swag tool)
	@echo "Generating Swagger documentation..."
	@if command -v swag >/dev/null 2>&1; then \
		swag init -g ./cmd/api/main.go --output ./api/docs; \
	else \
		echo "swag not found. Install with:"; \
		echo "  go install github.com/swaggo/swag/cmd/swag@latest"; \
		exit 1; \
	fi

.PHONY: install-tools
install-tools: ## Install development tools (golangci-lint, swag, mockgen, sqlc, etc.)
	@echo "Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/swaggo/swag/cmd/swag@latest
	go install github.com/securego/gosec/v2/cmd/gosec@latest
	go install go.uber.org/mock/mockgen@latest
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# =============================================================================
# Mock Generation Targets
# =============================================================================

# Mock generation directories
MOCK_DIRS := \
	internal/feature/user/usecase \
	internal/feature/auth/usecase \
	internal/feature/user/repository \
	internal/logger

.PHONY: mock-deps
mock-deps: ## Install mockgen tool if not present
	@if ! command -v mockgen >/dev/null 2>&1; then \
		echo "Installing mockgen..."; \
		go install go.uber.org/mock/mockgen@latest; \
	fi

.PHONY: mock
mock: mock-deps ## Generate all mocks using mockgen
	@echo "Generating mocks with mockgen..."
	@go generate ./...
	@echo "Mocks generated successfully"

.PHONY: mock-clean
mock-clean: ## Remove all generated mocks
	@echo "Cleaning generated mocks..."
	@rm -rf internal/feature/user/usecase/mocks
	@rm -rf internal/feature/auth/usecase/mocks
	@rm -rf internal/feature/user/repository/mocks
	@rm -f internal/logger/mocks/*.go
	@echo "Generated mocks cleaned"

.PHONY: mock-verify
mock-verify: mock ## Verify mocks are up to date
	@echo "Verifying mocks are up to date..."
	@if [ -n "$$(git status --porcelain '**/mocks/*.go')" ]; then \
		echo "Error: Mocks are out of date. Run 'make mock' to regenerate."; \
		git status --short '**/mocks/*.go'; \
		exit 1; \
	fi
	@echo "Mocks are up to date"

# =============================================================================
# SQLC Targets
# =============================================================================

SQLC_DIR := internal/infrastructure/db/sqlc
MIGRATIONS_DIR := internal/infrastructure/db/migrations
QUERIES_DIR := internal/infrastructure/db/queries

.PHONY: sqlc-deps
sqlc-deps: ## Install sqlc tool if not present
	@if ! command -v sqlc >/dev/null 2>&1; then \
		echo "Installing sqlc..."; \
		go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest; \
	fi

.PHONY: sqlc
sqlc: sqlc-deps ## Generate sqlc Go code from SQL queries
	@echo "Generating sqlc code..."
	@sqlc generate
	@echo "sqlc code generated successfully"

.PHONY: sqlc-clean
sqlc-clean: ## Remove all generated sqlc code
	@echo "Cleaning generated sqlc code..."
	@rm -f $(SQLC_DIR)/db.go
	@rm -f $(SQLC_DIR)/models.go
	@rm -f $(SQLC_DIR)/querier.go
	@rm -f $(SQLC_DIR)/users.sql.go
	@echo "Generated sqlc code cleaned"

.PHONY: sqlc-verify
sqlc-verify: sqlc ## Verify sqlc code is up to date
	@echo "Verifying sqlc code is up to date..."
	@if [ -n "$$(git status --porcelain '$(SQLC_DIR)/*.go')" ]; then \
		echo "Error: sqlc code is out of date. Run 'make sqlc' to regenerate."; \
		git status --short '$(SQLC_DIR)/*.go'; \
		exit 1; \
	fi
	@echo "sqlc code is up to date"

# =============================================================================
# Database Migration Targets
# =============================================================================

.PHONY: migrate-deps
migrate-deps: ## Install golang-migrate tool if not present
	@if ! command -v migrate >/dev/null 2>&1; then \
		echo "Installing golang-migrate..."; \
		go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest; \
	fi

.PHONY: migrate
migrate: migrate-deps ## Run database migrations up
	@echo "Running database migrations..."
	@migrate -path $(MIGRATIONS_DIR) -database "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)" up
	@echo "Migrations completed"

.PHONY: migrate-down
migrate-down: migrate-deps ## Rollback one database migration
	@echo "Rolling back last migration..."
	@migrate -path $(MIGRATIONS_DIR) -database "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)" down 1
	@echo "Migration rolled back"

.PHONY: migrate-reset
migrate-reset: migrate-deps ## Rollback all migrations and re-run
	@echo "Resetting all migrations..."
	@migrate -path $(MIGRATIONS_DIR) -database "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)" down
	@migrate -path $(MIGRATIONS_DIR) -database "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)" up
	@echo "Migrations reset completed"

.PHONY: migrate-create
migrate-create: ## Create a new migration file (usage: make migrate-create name=add_users_table)
	@if [ -z "$(name)" ]; then \
		echo "Error: Migration name required. Usage: make migrate-create name=add_users_table"; \
		exit 1; \
	fi
	@echo "Creating new migration: $(name)"
	@migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $(name)

# =============================================================================
# Cleanup Targets
# =============================================================================

.PHONY: clean
clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	rm -rf $(BIN_DIR)
	rm -f $(COVERAGE_FILE) $(COVERAGE_HTML)
	rm -f results.sarif
	rm -rf tmp/
	go clean -cache

.PHONY: clean-all
clean-all: clean ## Clean everything including downloaded dependencies
	@echo "Cleaning all artifacts and dependencies..."
	rm -rf vendor/
	go clean -modcache

# =============================================================================
# Help
# =============================================================================
# Git Hooks Targets (Pre-commit)
# =============================================================================

.PHONY: hooks-install
hooks-install: ## Install pre-commit hooks
	@echo "Installing pre-commit hooks..."
	@if command -v pre-commit >/dev/null 2>&1; then \
		pre-commit install; \
		echo "Pre-commit hooks installed successfully!"; \
		echo "Hooks will run automatically on 'git commit'"; \
	else \
		echo "pre-commit not found. Install with:"; \
		echo "  pip install pre-commit"; \
		echo "  or: brew install pre-commit"; \
		echo ""; \
		echo "Or visit: https://pre-commit.com/#install"; \
		exit 1; \
	fi

.PHONY: hooks-uninstall
hooks-uninstall: ## Remove pre-commit hooks
	@echo "Removing pre-commit hooks..."
	@if command -v pre-commit >/dev/null 2>&1; then \
		pre-commit uninstall; \
		echo "Pre-commit hooks removed successfully!"; \
	else \
		echo "pre-commit not found. Nothing to uninstall."; \
	fi

.PHONY: hooks-run
hooks-run: ## Run all pre-commit hooks manually on all files
	@echo "Running all pre-commit hooks..."
	@if command -v pre-commit >/dev/null 2>&1; then \
		pre-commit run --all-files; \
	else \
		echo "pre-commit not found. Install with:"; \
		echo "  pip install pre-commit"; \
		echo "  or: brew install pre-commit"; \
		exit 1; \
	fi

.PHONY: hooks-update
hooks-update: ## Update pre-commit hook versions
	@echo "Updating pre-commit hooks..."
	@if command -v pre-commit >/dev/null 2>&1; then \
		pre-commit autoupdate; \
	else \
		echo "pre-commit not found. Install with:"; \
		echo "  pip install pre-commit"; \
		echo "  or: brew install pre-commit"; \
		exit 1; \
	fi

# =============================================================================

.PHONY: help
help: ## Show available targets
	@echo "Available targets:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "Variables:"
	@echo "  DOCKER_REGISTRY      - Docker registry URL (optional)"
	@echo "  DOCKER_TAG          - Docker image tag (default: latest)"
	@echo ""
	@echo "Examples:"
	@echo "  make build                    - Build the binary"
	@echo "  make docker-build             - Build Docker image"
	@echo "  make test-coverage            - Generate coverage report"
	@echo "  make DOCKER_TAG=v1.0.0 build  - Build with specific version"

# =============================================================================
# CI/CD Convenience Targets
# =============================================================================

.PHONY: ci
-ci: fmt-check vet lint test build ## Full CI pipeline (format check, vet, lint, test, build)

.PHONY: release
release: clean check build-all ## Prepare a release build
	@echo "Release build complete. Binaries in $(BIN_DIR)/"

# =============================================================================
# Local Development
# =============================================================================

.PHONY: dev
dev: ## Run the application in development mode with hot reload (requires air)
	@if command -v air >/dev/null 2>&1; then \
		air; \
	else \
		echo "air not found. Install with:"; \
		echo "  go install github.com/air-verse/air@latest"; \
		echo "  or visit: https://github.com/air-verse/air"; \
		exit 1; \
	fi

.PHONY: setup
setup: deps install-tools ## Initial project setup for new developers
	@echo "Setup complete! You can now run 'make run' to start the application."
