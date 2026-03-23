# Makefile for zercle-go-template
# Go 2026 best practices

.PHONY: all build run test test-unit test-integration test-ci test-coverage lint fmt \
	sqlc-generate migrate-up migrate-down migrate-create migrate-force \
	docker-up docker-down docker-build clean help

# Application metadata
APP_NAME := zercle-go-template
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
BUILD_DIR := bin
MAIN_PATH := ./cmd/server

# Build flags
LDFLAGS := -X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)

# Go commands
GO := go
GOTEST := $(GO) test
GOBUILD := $(GO) build
GOFMT := gofmt
GOLINT := golangci-lint
SQLC := sqlc
MIGRATE := migrate

# Directories
DB_MIGRATION_DIR := internal/infrastructure/database/migrations
DB_SQLC_DIR := internal/infrastructure/database/sqlc

# Default target
all: build

## build: Build the binary
build:
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_PATH)

## run: Run the server
run: build
	@echo "Running $(APP_NAME)..."
	./$(BUILD_DIR)/$(APP_NAME)

## dev: Run the server in development mode
dev:
	@echo "Running in development mode..."
	$(GO) run $(MAIN_PATH)/main.go

## test: Run unit tests
test:
	@echo "Running unit tests..."
	$(GOTEST) -v -race -short ./...

## test-integration: Run integration tests
test-integration:
	@echo "Running integration tests..."
	$(GOTEST) -v -race -tags=integration ./test/integration/...

## test-unit: Run unit tests with coverage
test-unit:
	@echo "Running unit tests..."
	$(GOTEST) -short -race -coverprofile=coverage-unit.txt -covermode=atomic ./...

## test-ci: Run all tests for CI (unit and integration)
test-ci: test-unit test-integration

## test-coverage: Run tests with coverage report
test-coverage: test
	@echo "Generating coverage report..."
	$(GOTEST) -v -race -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

## lint: Run golangci-lint
lint:
	@echo "Running linter..."
	$(GOLINT) run ./...

## lint-fix: Run golangci-lint with auto-fix
lint-fix:
	@echo "Running linter with auto-fix..."
	$(GOLINT) run --fix ./...

## fmt: Format code
fmt:
	@echo "Formatting code..."
	$(GOFMT) -s -w .
	$(GO) mod tidy

## vet: Run go vet
vet:
	@echo "Running go vet..."
	$(GO) vet ./...

## sqlc-generate: Generate sqlc code
sqlc-generate:
	@echo "Generating sqlc code..."
	$(SQLC) generate

## migrate-up: Apply migrations
migrate-up:
	@echo "Applying migrations..."
	$(GO) run ./cmd/migrate up

## migrate-down: Rollback migrations
migrate-down:
	@echo "Rolling back migrations..."
	$(GO) run ./cmd/migrate down

## migrate-force: Force migrate to a specific version
migrate-force:
	@echo "Forcing migration to version $(VERSION)..."
	$(GO) run ./cmd/migrate force $(VERSION)

## migrate-create: Create a new migration
# Usage: make migrate-create name=add_users_table
migrate-create:
ifndef name
	$(error name is required. Usage: make migrate-create name=migration_name)
endif
	@echo "Creating migration: $(name)"
	mkdir -p $(DB_MIGRATION_DIR)
	$(MIGRATE) create -ext sql -dir $(DB_MIGRATION_DIR) -seq $(name)

## docker-up: Start development containers
docker-up:
	@echo "Starting development containers..."
	docker-compose up -d
	@echo "Waiting for database to be ready..."
	@sleep 5

## docker-down: Stop development containers
docker-down:
	@echo "Stopping development containers..."
	docker-compose down

## docker-build: Build Docker image
docker-build:
	@echo "Building Docker image..."
	docker build -t $(APP_NAME):$(VERSION) .

## docker-run: Run Docker container
docker-run:
	@echo "Running Docker container..."
	docker run -p 8080:8080 $(APP_NAME):$(VERSION)

## clean: Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html
	rm -rf .coverage

## tidy: Clean up go modules
tidy:
	@echo "Tidying go modules..."
	$(GO) mod tidy
	$(GO) mod verify

## deps: Install dependencies
deps:
	@echo "Installing dependencies..."
	$(GO) mod download
	$(GO) mod verify

## verify: Verify dependencies
verify:
	@echo "Verifying dependencies..."
	$(GO) mod verify

## generate: Run code generation
generate: sqlc-generate
	@echo "Running code generation..."

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@awk 'BEGIN { FS = ":.*?## " } /^[a-zA-Z_-]+:.*?## / { printf "  %-20s %s\n", $$1, $$2 }' $(MAKEFILE_LIST)
