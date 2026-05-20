.PHONY: build build-server test test-integration test-e2e test-all clean \
        migrate-up migrate-down migrate-create \
        docker-build docker-up docker-down docker-logs \
        generate mocks sqlc lint fmt verify cover-html help

# Build
build: build-server

build-server:
	CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/server ./cmd/server

# Testing
test:
	go test -race -count=1 -coverprofile=coverage.out ./...

test-integration:
	go test -v -race -count=1 -tags=integration ./test/integration/...

test-e2e:
	go test -v -count=1 -tags=e2e ./test/e2e/...

test-all: test test-integration

cover-html:
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Database
migrate-up:
	@echo "Running migrations..."
	@go run ./cmd/migrate

migrate-down:
	@echo "Rolling back migrations..."
	@go run ./cmd/migrate -down

migrate-create:
	@echo "Creating new migration..."
	@read -p "Enter migration name: " NAME; \
	migrate create -ext sql -dir migrations $$NAME

# Code Generation
generate:
	go generate ./...

mocks:
	@which mockgen > /dev/null || go install go.uber.org/mock/mockgen@latest
	go generate ./...

sqlc:
	@which sqlc > /dev/null || go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	sqlc generate

# Docker
docker-build:
	docker build -f Containerfile -t zercle-go-template:latest .

docker-up:
	docker compose up -d

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f

# Development
lint:
	golangci-lint run --timeout=5m

fmt:
	gofumpt -w .
	goimports -w .

verify:
	go mod verify
	go mod tidy
	@if [ -n "$$(git status --porcelain go.mod go.sum)" ]; then \
		echo "go.mod or go.sum is not tidy"; \
		git diff go.mod go.sum; \
		exit 1; \
	fi
	go vet ./...

run:
	go run ./cmd/server

help:
	@echo "Available targets:"
	@echo "  build             - Build the server binary"
	@echo "  build-server       - Build the gRPC/HTTP server"
	@echo "  test               - Run unit tests with race detection"
	@echo "  test-integration   - Run integration tests"
	@echo "  test-e2e           - Run end-to-end tests"
	@echo "  test-all           - Run all tests"
	@echo "  cover-html         - Generate HTML coverage report"
	@echo "  migrate-up         - Run database migrations"
	@echo "  migrate-down       - Rollback database migrations"
	@echo "  migrate-create     - Create new migration"
	@echo "  generate           - Run go generate (mocks)"
	@echo "  mocks              - Regenerate mock files"
	@echo "  sqlc               - Regenerate sqlc code"
	@echo "  docker-build       - Build Docker image"
	@echo "  docker-up          - Start services via docker compose"
	@echo "  docker-down        - Stop services"
	@echo "  lint               - Run golangci-lint"
	@echo "  fmt                - Format code with gofumpt + goimports"
	@echo "  verify             - Verify go.mod, go vet"
	@echo "  run                - Run the server"