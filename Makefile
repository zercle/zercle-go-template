.PHONY: build build-server build-client test test-integration test-all clean migrate-up migrate-down docker-build docker-up docker-down generate mocks sqlc lint fmt help

# Build
build: build-server build-client

build-server:
	go build -o bin/server ./cmd/server

build-client:
	go build -o bin/client ./cmd/client

# Testing
test:
	go test -race -coverprofile=coverage.out ./...

test-integration:
	go test -v -race -tags=integration ./test/integration/...

test-all: test test-integration

# Database
migrate-up:
	@echo "Running migrations up..."
	@read -p "Enter database URL: " DB_URL; \
	migrate -path internal/infrastructure/db/migrations -database "$$DB_URL" up

migrate-down:
	@echo "Running migrations down..."
	@read -p "Enter database URL: " DB_URL; \
	migrate -path internal/infrastructure/db/migrations -database "$$DB_URL" down

migrate-create:
	@echo "Creating new migration..."
	@read -p "Enter migration name: " NAME; \
	migrate create -ext sql -dir internal/infrastructure/db/migrations $$NAME

# Code Generation
generate:
	go generate ./...

mocks:
	go install go.uber.org/mock/mockgen@latest
	go generate ./...

sqlc:
	sqlc generate

# Docker
docker-build:
	docker-compose -f deployments/docker/compose.yaml build

docker-up:
	docker-compose -f deployments/docker/compose.yaml up -d

docker-down:
	docker-compose -f deployments/docker/compose.yaml down

docker-logs:
	docker-compose -f deployments/docker/compose.yaml logs -f

# Development
lint:
	golangci-lint run --timeout=5m

fmt:
	gofmt -s -w .
	goimports -l -w .

help:
	@echo "Available targets:"
	@echo "  build           - Build server and client"
	@echo "  build-server    - Build the gRPC server"
	@echo "  build-client    - Build the HTTP client"
	@echo "  test            - Run unit tests"
	@echo "  test-integration - Run integration tests"
	@echo "  test-all        - Run all tests"
	@echo "  migrate-up      - Run database migrations"
	@echo "  migrate-down    - Rollback database migrations"
	@echo "  migrate-create  - Create new migration"
	@echo "  generate        - Run go generate"
	@echo "  mocks           - Generate mocks"
	@echo "  sqlc            - Generate sqlc code"
	@echo "  docker-build    - Build Docker images"
	@echo "  docker-up       - Start Docker containers"
	@echo "  docker-down     - Stop Docker containers"
	@echo "  lint            - Run linters"
	@echo "  fmt             - Format code"
