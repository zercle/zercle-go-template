# Development Guide

## Prerequisites

| Tool | Version | Installation |
|------|---------|--------------|
| Go | 1.26+ | [go.dev](https://go.dev/dl/) |
| Docker | Latest | [docker.com](https://www.docker.com/) |
| Docker Compose | Latest | Included with Docker |
| PostgreSQL | 18+ | Via Docker |
| Valkey | 9+ | Via Docker |

## Quick Start

```bash
# 1. Clone repository
git clone https://github.com/zercle/zercle-go-template.git
cd zercle-go-template

# 2. Copy environment file
cp .env.example .env

# 3. Start infrastructure (PostgreSQL + Valkey)
docker compose -f deployments/docker/docker-compose.yaml up -d

# 4. Run database migrations
make migrate-up

# 5. Build and run server
make build-server
./bin/server

# 6. In another terminal, build and run client
make build-client
./bin/client
```

## Project Setup

### 1. Environment Variables

Create `.env` file:

```bash
# Application
APP_ENVIRONMENT=development
APP_HOST=0.0.0.0
APP_PORT=8080

# Database
DB_HOST=localhost
DB_PORT=5432
DB_NAME=zercle_chat
DB_USER=postgres
DB_PASSWORD=postgres
DB_SSL_MODE=disable

# Valkey
VALKEY_HOST=localhost
VALKEY_PORT=6379
VALKEY_PASSWORD=
VALKEY_DB=0

# JWT
JWT_SECRET=your-256-bit-secret-key-here
JWT_EXPIRY=24h
REFRESH_EXPIRY=168h
```

### 2. Install Dependencies

```bash
go mod download

# Install development tools
go install go.uber.org/mock/mockgen@latest
go install github.com/sqlc-dev/sqlc@latest
go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
go install golang.org/x/tools/cmd/goimports@latest
```

### 3. Generate Code

```bash
# Generate mocks
go generate ./...

# Generate sqlc
sqlc generate
```

## Development Commands

### Makefile Targets

```bash
# Build
make build-server     # Build gRPC server
make build-client     # Build HTTP client
make build            # Build both

# Testing
make test             # Run unit tests
make test-integration # Run integration tests
make test-all         # Run all tests with coverage

# Database
make migrate-up       # Run migrations
make migrate-down     # Rollback migrations
make migrate-create  # Create new migration (NAME=create_users)

# Code Generation
make generate         # Run go generate
make mocks            # Generate mocks only
make sqlc             # Generate sqlc code

# Docker
make docker-build     # Build Docker images
make docker-up       # Start containers
make docker-down     # Stop containers
make docker-logs     # View logs

# Development
make watch           # Watch mode (requires air)
make lint            # Run linters
make fmt             # Format code
```

## Directory Structure

```
zercle-go-template/
├── cmd/
│   ├── server/              # gRPC server entry point
│   └── client/              # HTTP client entry point
├── internal/
│   ├── features/           # Feature modules
│   │   ├── auth/           # Authentication
│   │   ├── chat/           # Chat/Messaging
│   │   └── user/           # User management
│   ├── infrastructure/     # Infrastructure
│   │   ├── config/         # Configuration
│   │   ├── db/             # Database
│   │   └── messaging/      # Valkey
│   └── shared/             # Shared utilities
├── api/
│   ├── proto/              # gRPC definitions
│   └── openapi/            # REST API spec
├── configs/                # Config files
├── deployments/             # Docker & K8s
├── test/                   # Test utilities
├── Makefile
└── sqlc.yaml
```

## Testing

### Unit Tests

```bash
# Run all unit tests
go test -v ./...

# Run with coverage
go test -race -coverprofile=coverage.out ./...

# Watch mode (requires air)
go test -v -race ./...
```

### Integration Tests

```bash
# Requires running PostgreSQL and Valkey
go test -v -tags=integration ./test/integration/...
```

### Generate Mocks

```bash
# After defining interfaces, run:
go generate ./...
```

## Database Migrations

### Create Migration

```bash
migrate create -ext sql -dir internal/infrastructure/db/migrations create_users
```

### Run Migrations

```bash
# Up
migrate -path internal/infrastructure/db/migrations \
  -database "postgres://postgres:postgres@localhost:5432/zercle_chat?sslmode=disable" up

# Down
migrate -path internal/infrastructure/db/migrations \
  -database "postgres://postgres:postgres@localhost:5432/zercle_chat?sslmode=disable" down
```

### Using Make

```bash
# Set environment variables first
export DB_HOST=localhost
export DB_PORT=5432
export DB_NAME=zercle_chat
export DB_USER=postgres
export DB_PASSWORD=postgres

make migrate-up
```

## Code Generation

### sqlc

Configuration in `sqlc.yaml`:

```yaml
version: "2"
sql:
  - engine: "postgresql"
    queries: "internal/infrastructure/db/queries"
    schema: "internal/infrastructure/db/migrations"
    gen:
      go:
        package: "db"
        out: "internal/infrastructure/db/postgres"
```

### Protocol Buffers

```bash
# Install protoc plugins
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Generate code
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       api/proto/chat.proto
```

## Docker Development

### Local Development with Hot Reload

```yaml
# docker-compose.dev.yaml
version: '3.8'
services:
  app:
    build:
      context: .
      dockerfile: Dockerfile.dev
    volumes:
      - .:/app
    ports:
      - "8080:8080"
      - "50051:50051"
    environment:
      - DB_HOST=postgres
      - VALKEY_HOST=valkey
```

### Production Build

```bash
# Build multi-stage
docker build -f deployments/docker/Dockerfile.server -t zercle-server:latest .
docker build -f deployments/docker/Dockerfile.client -t zercle-client:latest .
```

## Troubleshooting

### Database Connection Failed

```bash
# Check if PostgreSQL is running
docker ps | grep postgres

# Check logs
docker logs postgres

# Test connection
docker exec -it postgres psql -U postgres -d zercle_chat
```

### Port Already in Use

```bash
# Find process using port
lsof -i :8080

# Kill process
kill -9 <PID>
```

### Migration Errors

```bash
# Force clean migration table (development only)
DELETE FROM schema_migrations;
```

## API Testing

### cURL Examples

```bash
# Register
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"john","email":"john@example.com","password":"password123"}'

# Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"john@example.com","password":"password123"}'

# Get rooms (with token)
curl -H "Authorization: Bearer <token>" http://localhost:8080/api/v1/rooms

# SSE stream
curl -N -H "Authorization: Bearer <token>" http://localhost:8080/api/v1/events/stream
```

### grpcurl Examples

```bash
# List services
grpcurl localhost:50051 list

# Call auth service
grpcurl -d '{"email":"john@example.com","password":"password123"}' \
  localhost:50051 chat.AuthService/Login
```

## Best Practices

1. **Always run tests before committing**: `make test`
2. **Format code**: `make fmt`
3. **Generate mocks after interface changes**: `go generate ./...`
4. **Use environment variables for secrets**: Never commit to git
5. **Keep dependencies updated**: `go get -u ./...`
6. **Write tests for new features**: Test-first approach
