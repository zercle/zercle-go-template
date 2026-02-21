# 🚀 Zercle Go Template

[![Go Version](https://img.shields.io/badge/Go-1.26-blue.svg)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](https://opensource.org/licenses/MIT)
[![Docker](https://img.shields.io/badge/Docker-ready-blue.svg)](https://www.docker.com/)
[![Swagger](https://img.shields.io/badge/Swagger-API%20Docs-green.svg)](https://swagger.io/)

A production-ready Go REST API template implementing **Clean Architecture** with JWT authentication, PostgreSQL database, and comprehensive testing infrastructure.

## 📋 Overview

This template provides a solid foundation for building scalable REST APIs in Go. It follows industry best practices including Clean Architecture, dependency injection, structured logging, and comprehensive testing patterns.

### 🎯 Key Highlights

- **Clean Architecture** — Separation of concerns with feature-based organization
- **JWT Authentication** — Secure token-based authentication with refresh tokens
- **PostgreSQL + SQLC** — Type-safe database operations with migrations
- **Structured Logging** — JSON logging with zerolog for production observability
- **API Documentation** — Auto-generated Swagger/OpenAPI documentation
- **Docker First** — Multi-stage builds with distroless runtime
- **Comprehensive Testing** — Unit, integration, benchmarks, and mocks

## ✨ Features

| Category | Features |
|----------|----------|
| **Architecture** | Clean Architecture, Dependency Injection, Feature-based organization |
| **Authentication** | JWT with access/refresh tokens, Password hashing (Argon2id) |
| **Database** | PostgreSQL, SQLC code generation, Database migrations |
| **API** | RESTful endpoints, Swagger/OpenAPI documentation, Middleware stack |
| **Logging** | Structured JSON logging, Request logging, Error tracking |
| **Testing** | Unit tests, Integration tests, Benchmarks, Mock generation |
| **DevOps** | Docker multi-stage builds, Docker Compose, Pre-commit hooks |
| **Code Quality** | golangci-lint, gofmt, go vet, Security scanning (gosec) |

## 📁 Project Structure

```
zercle-go-template/
├── api/                          # API documentation
│   └── docs/                    # Generated Swagger files
├── cmd/                         # Application entry points
│   └── api/                    # Main API server
│       └── main.go            # Application bootstrap
├── configs/                    # Configuration files
│   └── config.yaml            # Application configuration
├── internal/                   # Private application code
│   ├── config/                # Configuration loading
│   ├── container/             # Dependency injection container
│   ├── errors/                # Custom error types
│   ├── feature/               # Feature modules (Clean Architecture)
│   │   ├── auth/             # Authentication feature
│   │   │   ├── domain/       # Domain models & JWT
│   │   │   ├── middleware/    # Auth middleware
│   │   │   └── usecase/       # Auth business logic
│   │   └── user/             # User management feature
│   │       ├── domain/       # User domain models
│   │       ├── dto/          # Data transfer objects
│   │       ├── handler/      # HTTP handlers
│   │       ├── repository/   # Data access layer
│   │       └── usecase/      # User business logic
│   ├── infrastructure/        # External services
│   │   └── db/               # Database layer
│   │       ├── migrations/   # SQL migrations
│   │       ├── queries/      # SQL query files
│   │       └── sqlc/         # Generated SQLC code
│   ├── logger/               # Structured logging
│   └── middleware/           # HTTP middleware
├── docker-compose.test.yml   # Docker Compose for testing
├── Dockerfile                # Multi-stage Docker build
├── Makefile                  # Build automation
├── go.mod                    # Go module definition
├── go.sum                    # Go dependency checksums
└── sqlc.yaml                # SQLC configuration
```

### 🏗️ Architecture Layers

```
┌─────────────────────────────────────────────────────────────┐
│                    Presentation Layer                        │
│  (HTTP Handlers, Middleware, DTOs)                          │
├─────────────────────────────────────────────────────────────┤
│                    Application Layer                         │
│  (Use Cases, Business Logic)                                │
├─────────────────────────────────────────────────────────────┤
│                      Domain Layer                            │
│  (Entities, Domain Services, Interfaces)                    │
├─────────────────────────────────────────────────────────────┤
│                   Infrastructure Layer                       │
│  (Database, External Services, Repositories)               │
└─────────────────────────────────────────────────────────────┘
```

## 🔧 Prerequisites

| Tool | Version | Installation |
|------|---------|--------------|
| **Go** | 1.26+ | [Download](https://go.dev/dl/) |
| **Docker** | Latest | [Docker Desktop](https://www.docker.com/products/docker-desktop) |
| **Make** | Latest | Pre-installed on macOS/Linux |
| **PostgreSQL** | 14+ | Via Docker or [Installer](https://www.postgresql.org/download/) |

### Optional Development Tools

```bash
# Install all development tools at once
make install-tools

# Or install individually:
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest  # Linting
go install github.com/swaggo/swag/cmd/swag@latest                       # Swagger
go install github.com/securego/gosec/v2/cmd/gosec@latest               # Security
go install go.uber.org/mock/mockgen@latest                             # Mock generation
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest                   # SQL code generation
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest  # Migrations
go install github.com/air-verse/air@latest                            # Hot reload
```

## 🚀 Getting Started

### 1. Clone the Repository

```bash
git clone https://github.com/zercle/zercle-go-template.git
cd zercle-go-template
```

### 2. Set Up Environment

```bash
# Copy environment template
cp .env.example .env

# Edit configuration (optional - defaults work for development)
vim .env
```

### 3. Start Database with Docker

```bash
# Start PostgreSQL for development
docker run -d \
  --name postgres-dev \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=zercle_template \
  -p 5432:5432 \
  postgres:18-alpine
```

### 4. Run Database Migrations

```bash
# Set database credentials
export DB_HOST=localhost
export DB_PORT=5432
export DB_NAME=zercle_template
export DB_USER=postgres
export DB_PASSWORD=postgres
export DB_SSLMODE=disable

# Run migrations
make migrate
```

### 5. Run the Application

```bash
# Using Make
make run

# Or directly with Go
go run ./cmd/api
```

The API server will start at `http://localhost:8080`

### 6. Verify Installation

```bash
# Health check
curl http://localhost:8080/health

# Swagger documentation
open http://localhost:8080/swagger/index.html
```

## ⚙️ Configuration

### Configuration File (`configs/config.yaml`)

```yaml
app:
  name: "zercle-go-template"
  version: "1.0.0"
  environment: "development"

server:
  host: "0.0.0.0"
  port: 8080
  read_timeout: "30s"
  write_timeout: "30s"
  shutdown_timeout: "10s"

database:
  host: "localhost"
  port: 5432
  database: "zercle_template"
  username: "postgres"
  password: ""          # Set via APP_DATABASE_PASSWORD env var
  ssl_mode: "disable"

log:
  level: "info"
  format: "json"

security:
  argon2_memory: 0      # 0 = environment-based default
  argon2_iterations: 0
  argon2_parallelism: 0
```

### Environment Variables

Configuration can be overridden using environment variables. Prefix with `APP_`:

| Variable | Description | Default |
|----------|-------------|---------|
| `APP_APP_NAME` | Application name | `zercle-go-template` |
| `APP_APP_ENVIRONMENT` | Environment | `development` |
| `APP_SERVER_PORT` | HTTP server port | `8080` |
| `APP_DATABASE_HOST` | Database host | `localhost` |
| `APP_DATABASE_PASSWORD` | Database password | (empty) |
| `APP_JWT_SECRET` | JWT signing secret | (required) |
| `APP_LOG_LEVEL` | Logging level | `info` |
| `APP_LOG_FORMAT` | Log format (json/console) | `json` |

### JWT Configuration

```bash
# Generate a secure JWT secret
openssl rand -base64 32

# Set in environment
export APP_JWT_SECRET=your-generated-secret
export APP_JWT_ACCESS_TOKEN_TTL=15m
export APP_JWT_REFRESH_TOKEN_TTL=168h
```

## 📦 Available Commands

### Build & Run

| Command | Description |
|---------|-------------|
| `make build` | Build the application binary |
| `make build-linux` | Cross-compile for Linux AMD64 |
| `make run` | Run the application locally |
| `make dev` | Run with hot reload (requires air) |

### Testing

| Command | Description |
|---------|-------------|
| `make test` | Run all tests with race detection |
| `make test-unit` | Run unit tests only |
| `make test-integration` | Run integration tests (requires database) |
| `make test-coverage` | Generate coverage report |
| `make test-coverage-html` | Generate HTML coverage report |
| `make benchmark` | Run performance benchmarks |

### Code Quality

| Command | Description |
|---------|-------------|
| `make lint` | Run golangci-lint |
| `make fmt` | Format code with gofmt |
| `make vet` | Run go vet |
| `make fmt-check` | Check code formatting |
| `make security` | Run security scan with gosec |
| `make check` | Run all checks (fmt, vet, lint, test) |

### Database

| Command | Description |
|---------|-------------|
| `make migrate` | Run database migrations |
| `make migrate-down` | Rollback last migration |
| `make migrate-reset` | Reset all migrations |
| `make migrate-create name=<name>` | Create new migration |

### Code Generation

| Command | Description |
|---------|-------------|
| `make swagger` | Generate Swagger documentation |
| `make mock` | Generate mocks with mockgen |
| `make sqlc` | Generate SQLC Go code |
| `make install-tools` | Install all development tools |

### Docker

| Command | Description |
|---------|-------------|
| `make docker-build` | Build Docker image |
| `make docker-run` | Run Docker container |
| `make docker-push` | Push image to registry |
| `make docker-scan` | Scan image for vulnerabilities |

### Utilities

| Command | Description |
|---------|-------------|
| `make clean` | Clean build artifacts |
| `make clean-all` | Clean everything including dependencies |
| `make deps` | Download and verify dependencies |
| `make deps-update` | Update dependencies |
| `make hooks-install` | Install pre-commit hooks |
| `make help` | Show all available targets |

## 📚 API Documentation

### Swagger UI

Access the interactive API documentation at:

```
http://localhost:8080/swagger/index.html
```

### API Endpoints

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| `GET` | `/health` | Health check | ❌ |
| `GET` | `/swagger/*` | Swagger documentation | ❌ |
| `POST` | `/api/v1/users/register` | Register new user | ❌ |
| `POST` | `/api/v1/users/login` | User login | ❌ |
| `GET` | `/api/v1/users` | List all users | ✅ |
| `GET` | `/api/v1/users/:id` | Get user by ID | ✅ |
| `PUT` | `/api/v1/users/:id` | Update user | ✅ |
| `DELETE` | `/api/v1/users/:id` | Delete user | ✅ |
| `POST` | `/api/v1/users/refresh` | Refresh access token | ❌ |

### Authentication

This API uses JWT Bearer token authentication:

```bash
# Example request with authentication
curl -X GET http://localhost:8080/api/v1/users \
  -H "Authorization: Bearer <your-jwt-token>"
```

## 🧪 Testing

### Running Tests

```bash
# All tests with race detection
make test

# Unit tests only (no database required)
make test-unit

# Integration tests (requires PostgreSQL)
make test-integration

# Short tests (skips slow integration tests)
make test-short
```

### Test Coverage

```bash
# Generate coverage report
make test-coverage

# Generate HTML coverage report
make test-coverage-html
open coverage.html
```

### Benchmarks

```bash
# Run performance benchmarks
make benchmark
```

### Docker-Based Testing

```bash
# Start test infrastructure
docker-compose -f docker-compose.test.yml up --abort-on-container-exit

# Run integration tests only
docker-compose -f docker-compose.test.yml --profile test up test-runner
```

## 🗄️ Database Migrations

### Creating a New Migration

```bash
make migrate-create name=create_users_table
```

This creates two files:
- `internal/infrastructure/db/migrations/001_create_users_table.up.sql`
- `internal/infrastructure/db/migrations/001_create_users_table.down.sql`

### Writing Migrations

```sql
-- up.sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- down.sql
DROP TABLE users;
```

### Running Migrations

```bash
# Apply all pending migrations
make migrate

# Rollback one migration
make migrate-down

# Reset database (rollback all, then migrate up)
make migrate-reset
```

## 🐳 Deployment

### Docker Build

```bash
# Build the Docker image
make docker-build

# Tag for specific version
make DOCKER_TAG=v1.0.0 docker-build

# Build and run
make docker-run
```

### Docker Compose for Production

```yaml
# docker-compose.yml
version: "3.9"

services:
  app:
    image: zercle-go-template:latest
    ports:
      - "8080:8080"
    environment:
      - APP_APP_ENVIRONMENT=production
      - APP_DATABASE_HOST=postgres
      - APP_DATABASE_PASSWORD=${DB_PASSWORD}
      - APP_JWT_SECRET=${JWT_SECRET}
    depends_on:
      - postgres

  postgres:
    image: postgres:18-alpine
    environment:
      - POSTGRES_USER=app
      - POSTGRES_PASSWORD=${DB_PASSWORD}
      - POSTGRES_DB=production
    volumes:
      - postgres_data:/var/lib/postgresql/data

volumes:
  postgres_data:
```

### Environment-Specific Builds

```bash
# Production build
APP_ENVIRONMENT=production make build

# Build with custom version
make DOCKER_TAG=v1.0.0 docker-build
make docker-push
```

## 🔒 Security

### Password Hashing

The template uses **Argon2id** (OWASP recommended) for password hashing:

```bash
# Configuration (via environment or config.yaml)
APP_SECURITY_ARGON2_MEMORY=65536      # 64 MB
APP_SECURITY_ARGON2_ITERATIONS=3
APP_SECURITY_ARGON2_PARALLELISM=4
```

### Security Scanning

```bash
# Run gosec security scanner
make security
```

### Best Practices

- ✅ Always use HTTPS in production
- ✅ Store secrets in environment variables, not in config files
- ✅ Use strong JWT secrets (minimum 256-bit)
- ✅ Enable SSL for database connections in production
- ✅ Run as non-root user in Docker (already configured)
- ✅ Regular dependency updates (`make deps-update`)

## 🤝 Contributing

1. **Fork** the repository
2. Create a **feature branch** (`git checkout -b feature/amazing-feature`)
3. Commit your **changes** (`git commit -m 'Add amazing feature'`)
4. **Push** to the branch (`git push origin feature/amazing-feature`)
5. Open a **Pull Request**

### Code Style

This project follows standard Go conventions:
- Run `make check` before committing
- Ensure all tests pass
- Update Swagger docs when changing API endpoints
- Add unit tests for new features

### Pre-commit Hooks

```bash
# Install pre-commit hooks
make hooks-install

# Run hooks manually
make hooks-run
```

## 📄 License

This project is licensed under the **MIT License** - see the [LICENSE](LICENSE) file for details.

---

<div align="center">

**Built with ❤️ using Go + Echo + PostgreSQL**

[Documentation](#) · [Issues](#) · [Discussions](#)

</div>
