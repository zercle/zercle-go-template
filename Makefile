#!/usr/bin/env make

# =============================================================================
# Go Template Project - Comprehensive Makefile
# =============================================================================
# This Makefile provides convenient targets for building, testing, linting,
# and managing the Go application throughout its lifecycle.
#
# Usage:
#   make <target> [VAR=value]
#
# Examples:
#   make build                    # Build the application
#   make test                     # Run all tests
#   make docker-build            # Build Docker image
#   make migrate-create NAME=add_users  # Create new migration
# =============================================================================

# -----------------------------------------------------------------------------
# Variables - Common paths and settings
# -----------------------------------------------------------------------------

# Application metadata
APP_NAME          := $(shell head -n 1 go.mod | cut -d ' ' -f 2)
APP_VERSION       := $(shell git describe --tags --always 2>/dev/null || echo "dev")
BUILD_DATE        := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

# Directories
ROOT_DIR          := $(shell pwd)
BIN_DIR           := $(ROOT_DIR)/bin
CMD_DIR           := $(ROOT_DIR)/cmd/server
CONFIG_DIR        := $(ROOT_DIR)/configs
MIGRATIONS_DIR    := $(ROOT_DIR)/sqlc/migrations
QUERIES_DIR       := $(ROOT_DIR)/sqlc/queries
DOCKER_DIR        := $(ROOT_DIR)/deployments/docker
DOCS_DIR          := $(ROOT_DIR)/docs
TEST_DIR          := $(ROOT_DIR)/test
INTERNAL_DIR      := $(ROOT_DIR)/internal
SCRIPTS_DIR       := $(ROOT_DIR)/scripts

# Build outputs
BINARY_NAME       := server
BINARY_PATH       := $(BIN_DIR)/$(BINARY_NAME)
COVERAGE_FILE     := $(ROOT_DIR)/coverage.out

# Environment (default to local)
ENV               ?= local
ENV_FILE          := $(CONFIG_DIR)/$(ENV).yaml

# Go toolchain
GO                := go
GOCACHE           := $(shell go env GOCACHE)
GOMODCACHE        := $(shell go env GOMODCACHE)

# External tools
SQLC              := sqlc
GOLANGCI_LINT     := golangci-lint
DOCKER_COMPOSE    := docker-compose
SWAGGER           := swag

# Docker settings
DOCKER_IMAGE_NAME := $(APP_NAME)
DOCKER_TAG        := $(APP_VERSION)
DOCKER_REGISTRY   ?=
DOCKERFILE        := $(DOCKER_DIR)/Dockerfile

# Colors for terminal output (compatible with most terminals)
COLOR_RED         := \033[31m
COLOR_GREEN       := \033[32m
COLOR_YELLOW      := \033[33m
COLOR_BLUE        := \033[34m
COLOR_CYAN        := \033[36m
COLOR_WHITE       := \033[37m
COLOR_RESET       := \033[0m

# Go build flags
LDFLAGS           := -s -w -X main.version=$(APP_VERSION) -X main.buildDate=$(BUILD_DATE)
GOFLAGS           := -ldflags="$(LDFLAGS)"

# Test flags
TEST_FLAGS        := -v -race -coverprofile=$(COVERAGE_FILE) -covermode=atomic
UNIT_TEST_FLAGS   := -v -short -coverprofile=$(COVERAGE_FILE) -covermode=atomic
INTEG_TEST_FLAGS  := -v -tags=integration -coverprofile=$(COVERAGE_FILE) -covermode=atomic

# -----------------------------------------------------------------------------
# Phony targets - Always run these
# -----------------------------------------------------------------------------
.PHONY: all help build build-all run clean dev seed-db deps tidy
.PHONY: test test-all test-unit test-integration test-coverage test-race
.PHONY: lint lint-fix fmt vet
.PHONY: sqlc-gen sqlc-verify
.PHONY: migrate-up migrate-down migrate-create
.PHONY: docker-build docker-up docker-down docker-logs
.PHONY: swagger-gen swagger-validate generate
.PHONY: set-env env-dev env-local env-prod env-uat

# =============================================================================
# Core Targets
# =============================================================================

# Display help information
help:
	@printf '%s\n' "$(COLOR_CYAN)========================================$(COLOR_RESET)"
	@printf '%s\n' "$(COLOR_CYAN)   $(APP_NAME) - Build & Dev Commands   $(COLOR_RESET)"
	@printf '%s\n' "$(COLOR_CYAN)========================================$(COLOR_RESET)"
	@printf '\n'
	@printf '$(COLOR_YELLOW)Core Commands$(COLOR_RESET)\n'
	@printf '  $(COLOR_GREEN)make help$(COLOR_RESET)                Show this help message\n'
	@printf '  $(COLOR_GREEN)make build$(COLOR_RESET)               Build the application binary\n'
	@printf '  $(COLOR_GREEN)make run$(COLOR_RESET)                 Run the application locally\n'
	@printf '  $(COLOR_GREEN)make clean$(COLOR_RESET)               Remove build artifacts\n'
	@printf '\n'
	@printf '$(COLOR_YELLOW)Development Commands$(COLOR_RESET)\n'
	@printf '  $(COLOR_GREEN)make dev$(COLOR_RESET)                 Run in development mode\n'
	@printf '  $(COLOR_GREEN)make seed-db$(COLOR_RESET)             Seed the database\n'
	@printf '  $(COLOR_GREEN)make deps$(COLOR_RESET)                Download dependencies\n'
	@printf '  $(COLOR_GREEN)make tidy$(COLOR_RESET)                Clean up dependencies\n'
	@printf '\n'
	@printf '$(COLOR_YELLOW)Testing Commands$(COLOR_RESET)\n'
	@printf '  $(COLOR_GREEN)make test$(COLOR_RESET)                Run all tests\n'
	@printf '  $(COLOR_GREEN)make test-unit$(COLOR_RESET)           Run unit tests only\n'
	@printf '  $(COLOR_GREEN)make test-integration$(COLOR_RESET)    Run integration tests only\n'
	@printf '  $(COLOR_GREEN)make test-coverage$(COLOR_RESET)       Generate coverage report\n'
	@printf '  $(COLOR_GREEN)make test-race$(COLOR_RESET)           Run tests with race detector\n'
	@printf '\n'
	@printf '$(COLOR_YELLOW)Code Quality Commands$(COLOR_RESET)\n'
	@printf '  $(COLOR_GREEN)make lint$(COLOR_RESET)                Run golangci-lint\n'
	@printf '  $(COLOR_GREEN)make lint-fix$(COLOR_RESET)            Run golangci-lint with auto-fix\n'
	@printf '  $(COLOR_GREEN)make fmt$(COLOR_RESET)                 Format Go code\n'
	@printf '  $(COLOR_GREEN)make vet$(COLOR_RESET)                 Run go vet\n'
	@printf '\n'
	@printf '$(COLOR_YELLOW)SQLC Commands$(COLOR_RESET)\n'
	@printf '  $(COLOR_GREEN)make sqlc-gen$(COLOR_RESET)            Generate SQLC code\n'
	@printf '  $(COLOR_GREEN)make sqlc-verify$(COLOR_RESET)         Verify SQLC configuration\n'
	@printf '\n'
	@printf '$(COLOR_YELLOW)Database Migration Commands$(COLOR_RESET)\n'
	@printf '  $(COLOR_GREEN)make migrate-up$(COLOR_RESET)          Run database migrations up\n'
	@printf '  $(COLOR_GREEN)make migrate-down$(COLOR_RESET)        Rollback database migrations\n'
	@printf '  $(COLOR_GREEN)make migrate-create NAME=xxx$(COLOR_RESET) Create new migration\n'
	@printf '\n'
	@printf '$(COLOR_YELLOW)Docker Commands$(COLOR_RESET)\n'
	@printf '  $(COLOR_GREEN)make docker-build$(COLOR_RESET)        Build Docker image\n'
	@printf '  $(COLOR_GREEN)make docker-up$(COLOR_RESET)           Start Docker containers\n'
	@printf '  $(COLOR_GREEN)make docker-down$(COLOR_RESET)         Stop Docker containers\n'
	@printf '  $(COLOR_GREEN)make docker-logs$(COLOR_RESET)         View container logs\n'
	@printf '\n'
	@printf '$(COLOR_YELLOW)Swagger Commands$(COLOR_RESET)\n'
	@printf '  $(COLOR_GREEN)make swagger-gen$(COLOR_RESET)         Generate Swagger docs\n'
	@printf '  $(COLOR_GREEN)make swagger-validate$(COLOR_RESET)    Validate Swagger docs\n'
	@printf '\n'
	@printf '$(COLOR_YELLOW)Environment Commands$(COLOR_RESET)\n'
	@printf '  $(COLOR_GREEN)make set-env ENV=xxx$(COLOR_RESET)     Set environment (dev/local/prod/uat)\n'
	@printf '  $(COLOR_GREEN)make env-dev$(COLOR_RESET)             Set development environment\n'
	@printf '  $(COLOR_GREEN)make env-local$(COLOR_RESET)           Set local environment\n'
	@printf '  $(COLOR_GREEN)make env-prod$(COLOR_RESET)            Set production environment\n'
	@printf '  $(COLOR_GREEN)make env-uat$(COLOR_RESET)             Set UAT environment\n'
	@printf '\n'
	@printf '$(COLOR_CYAN)Note: Use $(COLOR_WHITE)ENV=value$(COLOR_CYAN) to override default environment$(COLOR_RESET)\n'
	@printf '$(COLOR_CYAN)Note: Use $(COLOR_WHITE)NAME=value$(COLOR_CYAN) for migration creation$(COLOR_RESET)\n'
	@printf '\n'

# Build the application binary
build: $(BINARY_PATH)
	@printf '$(COLOR_GREEN)[OK]$(COLOR_RESET) Binary built successfully: $(BINARY_PATH)\n'

# Build with all build tags
build-all:
	@printf '$(COLOR_YELLOW)[BUILD-ALL]$(COLOR_RESET) Building with all build tags...\n'
	$(GO) build $(GOFLAGS) -tags='$(ENV)' -o $(BINARY_PATH) $(CMD_DIR)/main.go
	@chmod +x $(BINARY_PATH)
	@printf '$(COLOR_GREEN)[OK]$(COLOR_RESET) Binary built with all tags: $(BINARY_PATH)\n'

$(BINARY_PATH): $(wildcard $(CMD_DIR)/*.go) $(wildcard $(INTERNAL_DIR)/**/*.go) go.mod
	@mkdir -p $(BIN_DIR)
	@printf '$(COLOR_YELLOW)[BUILD]$(COLOR_RESET) Building $(APP_NAME)...\n'
	$(GO) build $(GOFLAGS) -o $(BINARY_PATH) $(CMD_DIR)/main.go
	@chmod +x $(BINARY_PATH)

# Run the application locally
run: build
	@printf '$(COLOR_YELLOW)[RUN]$(COLOR_RESET) Running $(APP_NAME) with ENV=$(ENV)...\n'
	@ENV_FILE=$(ENV_FILE) $(BINARY_PATH)

# Clean build artifacts and temporary files
clean:
	@printf '$(COLOR_YELLOW)[CLEAN]$(COLOR_RESET) Cleaning build artifacts...\n'
	@rm -rf $(BIN_DIR)
	@rm -f $(COVERAGE_FILE)
	@rm -rf $(GOCACHE)
	@printf '$(COLOR_GREEN)[OK]$(COLOR_RESET) Clean completed\n'

# =============================================================================
# Development Targets
# =============================================================================

# Run the application in development mode
dev:
	@printf '$(COLOR_YELLOW)[DEV]$(COLOR_RESET) Starting development server...\n'
	@chmod +x $(SCRIPTS_DIR)/run-dev.sh
	@$(SCRIPTS_DIR)/run-dev.sh

# Seed the database
seed-db:
	@printf '$(COLOR_YELLOW)[SEED]$(COLOR_RESET) Seeding database...\n'
	@chmod +x $(SCRIPTS_DIR)/seed-db.sh
	@$(SCRIPTS_DIR)/seed-db.sh

# Download dependencies
deps:
	@printf '$(COLOR_YELLOW)[DEPS]$(COLOR_RESET) Downloading dependencies...\n'
	$(GO) mod download
	@printf '$(COLOR_GREEN)[OK]$(COLOR_RESET) Dependencies downloaded\n'

# Clean up dependencies
tidy:
	@printf '$(COLOR_YELLOW)[TIDY]$(COLOR_RESET) Tidying go.mod...\n'
	$(GO) mod tidy
	@printf '$(COLOR_GREEN)[OK]$(COLOR_RESET) Dependencies tidied\n'

# =============================================================================
# Testing Targets
# =============================================================================

# Run all tests
test: test-unit test-integration
	@printf '$(COLOR_GREEN)[OK]$(COLOR_RESET) All tests completed\n'

# Run all tests (alias for test)
test-all: test
	@printf '$(COLOR_GREEN)[OK]$(COLOR_RESET) Test-all completed\n'

# Run unit tests only
test-unit:
	@printf '$(COLOR_YELLOW)[TEST-UNIT]$(COLOR_RESET) Running unit tests...\n'
	@if [ -d ./test/unit ] && [ -n "$(find ./test/unit -name '*_test.go' 2>/dev/null)" ]; then \
		$(GO) test $(UNIT_TEST_FLAGS) ./test/unit/... ./internal/...; \
	else \
		$(GO) test $(UNIT_TEST_FLAGS) ./internal/...; \
	fi
	@if [ -f $(COVERAGE_FILE) ]; then \
		printf '$(COLOR_CYAN)[COVERAGE]$(COLOR_RESET) Unit test coverage: '; \
		go tool cover -func=$(COVERAGE_FILE) | tail -1; \
	fi

# Run integration tests only
test-integration:
	@printf '$(COLOR_YELLOW)[TEST-INTEG]$(COLOR_RESET) Running integration tests...\n'
	$(GO) test $(INTEG_TEST_FLAGS) ./test/integration/...
	@if [ -f $(COVERAGE_FILE) ]; then \
		printf '$(COLOR_CYAN)[COVERAGE]$(COLOR_RESET) Integration test coverage: '; \
		go tool cover -func=$(COVERAGE_FILE) | tail -1; \
	fi

# Generate test coverage report
test-coverage:
	@printf '$(COLOR_YELLOW)[COVERAGE]$(COLOR_RESET) Generating coverage report...\n'
	$(GO) test $(TEST_FLAGS) ./...
	@if [ -f $(COVERAGE_FILE) ]; then \
		go tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_FILE:.out=.html); \
		printf '$(COLOR_GREEN)[OK]$(COLOR_RESET) Coverage report: $(COVERAGE_FILE:.out=.html)\n'; \
	fi

# Run tests with race detector
test-race:
	@printf '$(COLOR_YELLOW)[RACE]$(COLOR_RESET) Running tests with race detector...\n'
	$(GO) test -race -v ./...
	@printf '$(COLOR_GREEN)[OK]$(COLOR_RESET) Race detector check completed\n'

# =============================================================================
# Code Quality Targets
# =============================================================================

# Run golangci-lint
lint:
	@printf '$(COLOR_YELLOW)[LINT]$(COLOR_RESET) Running golangci-lint...\n'
	$(GOLANGCI_LINT) run ./...

# Run golangci-lint with auto-fix
lint-fix:
	@printf '$(COLOR_YELLOW)[LINT-FIX]$(COLOR_RESET) Running golangci-lint with auto-fix...\n'
	$(GOLANGCI_LINT) run --fix ./...

# Format Go code
fmt:
	@printf '$(COLOR_YELLOW)[FMT]$(COLOR_RESET) Formatting Go code...\n'
	$(GO) fmt ./...
	@printf '$(COLOR_GREEN)[OK]$(COLOR_RESET) Code formatted\n'

# Run go vet
vet:
	@printf '$(COLOR_YELLOW)[VET]$(COLOR_RESET) Running go vet...\n'
	$(GO) vet ./...
	@printf '$(COLOR_GREEN)[OK]$(COLOR_RESET) Vet check completed\n'

# =============================================================================
# SQLC Targets
# =============================================================================

# Generate SQLC code from queries
sqlc-gen:
	@printf '$(COLOR_YELLOW)[SQLC]$(COLOR_RESET) Generating SQLC code...\n'
	$(SQLC) generate
	@printf '$(COLOR_GREEN)[OK]$(COLOR_RESET) SQLC code generated in $(INTERNAL_DIR)/sqlc/\n'

# Verify SQLC configuration
sqlc-verify:
	@printf '$(COLOR_YELLOW)[SQLC]$(COLOR_RESET) Verifying SQLC configuration...\n'
	$(SQLC) verify
	@printf '$(COLOR_GREEN)[OK]$(COLOR_RESET) SQLC configuration verified\n'

# Generate SQLC code and Swagger documentation
generate: sqlc-gen swagger-gen
	@printf '$(COLOR_GREEN)[OK]$(COLOR_RESET) All code generation completed\n'

# =============================================================================
# Database Migration Targets
# =============================================================================

# Run database migrations up
migrate-up:
	@printf '$(COLOR_YELLOW)[MIGRATE]$(COLOR_RESET) Running database migrations up...\n'
	@# Placeholder - implement based on your migration tool (e.g., migrate, golang-migrate)
	@if command -v migrate >/dev/null 2>&1; then \
		migrate -path $(MIGRATIONS_DIR) -database "$$DATABASE_URL" up; \
	elif command -v goose >/dev/null 2>&1; then \
		goose -dir $(MIGRATIONS_DIR) postgres "$$DATABASE_URL" up; \
	else \
		printf '$(COLOR_YELLOW)[WARN]$(COLOR_RESET) No migration tool found. Using sqlx migrations if available.\n'; \
		$(GO) run -tags '$(ENV)' ./cmd/migrate; \
	fi

# Rollback database migrations
migrate-down:
	@printf '$(COLOR_YELLOW)[MIGRATE]$(COLOR_RESET) Rolling back database migrations...\n'
	@# Placeholder - implement based on your migration tool
	@if command -v migrate >/dev/null 2>&1; then \
		migrate -path $(MIGRATIONS_DIR) -database "$$DATABASE_URL" down 1; \
	elif command -v goose >/dev/null 2>&1; then \
		goose -dir $(MIGRATIONS_DIR) postgres "$$DATABASE_URL" down 1; \
	else \
		printf '$(COLOR_YELLOW)[WARN]$(COLOR_RESET) No migration tool found.\n'; \
	fi

# Create a new migration
migrate-create:
	@if [ -z "$(NAME)" ]; then \
		printf '$(COLOR_RED)[ERROR]$(COLOR_RESET) Please provide NAME parameter: make migrate-create NAME=add_users\n'; \
		exit 1; \
	fi
	@printf '$(COLOR_YELLOW)[MIGRATE]$(COLOR_RESET) Creating migration: $(NAME)...\n'
	@TIMESTAMP=$$(date +%Y%m%d_%H%M%S); \
	mkdir -p $(MIGRATIONS_DIR); \
	echo "-- Migration: $(NAME)" > $(MIGRATIONS_DIR)/$${TIMESTAMP}_$(NAME).up.sql; \
	echo "-- Rollback: $(NAME)" > $(MIGRATIONS_DIR)/$${TIMESTAMP}_$(NAME).down.sql; \
	@printf '$(COLOR_GREEN)[OK]$(COLOR_RESET) Migration files created:\n'; \
	@ls -la $(MIGRATIONS_DIR)/$${TIMESTAMP}_$(NAME).*.sql

# =============================================================================
# Docker Targets
# =============================================================================

# Build Docker image
docker-build:
	@printf '$(COLOR_YELLOW)[DOCKER]$(COLOR_RESET) Building Docker image: $(DOCKER_REGISTRY)$(DOCKER_IMAGE_NAME):$(DOCKER_TAG)...\n'
	@docker build -t $(DOCKER_REGISTRY)$(DOCKER_IMAGE_NAME):$(DOCKER_TAG) -f $(DOCKERFILE) .
	@printf '$(COLOR_GREEN)[OK]$(COLOR_RESET) Docker image built: $(DOCKER_IMAGE_NAME):$(DOCKER_TAG)\n'

# Start Docker containers
docker-up:
	@printf '$(COLOR_YELLOW)[DOCKER]$(COLOR_RESET) Starting Docker containers...\n'
	@cd $(DOCKER_DIR) && $(DOCKER_COMPOSE) up -d
	@printf '$(COLOR_GREEN)[OK]$(COLOR_RESET) Docker containers started\n'

# Stop Docker containers
docker-down:
	@printf '$(COLOR_YELLOW)[DOCKER]$(COLOR_RESET) Stopping Docker containers...\n'
	@cd $(DOCKER_DIR) && $(DOCKER_COMPOSE) down
	@printf '$(COLOR_GREEN)[OK]$(COLOR_RESET) Docker containers stopped\n'

# View Docker container logs
docker-logs:
	@printf '$(COLOR_YELLOW)[DOCKER]$(COLOR_RESET) Showing container logs (Ctrl+C to exit)...\n'
	@cd $(DOCKER_DIR) && $(DOCKER_COMPOSE) logs -f

# =============================================================================
# Swagger Targets
# =============================================================================

# Generate Swagger documentation
swagger-gen:
	@printf '$(COLOR_YELLOW)[SWAGGER]$(COLOR_RESET) Generating Swagger documentation...\n'
	@$(SWAG) init --dir cmd/server -o $(DOCS_DIR) --parseDependency --parseInternal
	@printf '$(COLOR_GREEN)[OK]$(COLOR_RESET) Swagger docs generated in $(DOCS_DIR)/\n'

# Validate Swagger documentation
swagger-validate:
	@printf '$(COLOR_YELLOW)[SWAGGER]$(COLOR_RESET) Validating Swagger documentation...\n'
	@if command -v swagger >/dev/null 2>&1; then \
		swagger validate $(DOCS_DIR)/swagger.json; \
		swagger validate $(DOCS_DIR)/swagger.yaml; \
	else \
		printf '$(COLOR_YELLOW)[WARN]$(COLOR_RESET) swagger not installed. Skipping validation.\n'; \
	fi

# =============================================================================
# Environment Targets
# =============================================================================

# Set environment
set-env:
	@if [ -z "$(ENV)" ]; then \
		printf '$(COLOR_RED)[ERROR]$(COLOR_RESET) Please provide ENV parameter: make set-env ENV=dev\n'; \
		exit 1; \
	fi
	@printf '$(COLOR_YELLOW)[ENV]$(COLOR_RESET) Setting environment to: $(ENV)\n'
	@printf '$(COLOR_CYAN)Config file: $(COLOR_RESET)$(CONFIG_DIR)/$(ENV).yaml\n'
	@export ENV=$(ENV) && export ENV_FILE=$(CONFIG_DIR)/$(ENV).yaml && \
		printf '$(COLOR_GREEN)[OK]$(COLOR_RESET) Environment set to $(ENV)\n'

# Set development environment
env-dev:
	@make set-env ENV=dev

# Set local environment
env-local:
	@make set-env ENV=local

# Set production environment
env-prod:
	@make set-env ENV=prod

# Set UAT environment
env-uat:
	@make set-env ENV=uat

# =============================================================================
# Utility Targets
# =============================================================================

# Show project information
info:
	@printf '$(COLOR_CYAN)Project Information$(COLOR_RESET)\n'
	@printf '  Name:    $(APP_NAME)\n'
	@printf '  Version: $(APP_VERSION)\n'
	@printf '  Build:   $(BUILD_DATE)\n'
	@printf '\n'
	@printf '$(COLOR_CYAN)Paths$(COLOR_RESET)\n'
	@printf '  Binary:  $(BINARY_PATH)\n'
	@printf '  Config:  $(CONFIG_DIR)\n'
	@printf '  Docker:  $(DOCKER_DIR)\n'
	@printf '  Docs:    $(DOCS_DIR)\n'

# Default target
all: help
