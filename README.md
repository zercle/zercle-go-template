# Zercle Go Template

![Go Version](https://img.shields.io/badge/Go-1.26%2B-blue)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Build Status](https://img.shields.io/badge/build-passing-brightgreen)](https://github.com/zercle/zercle-go-template/actions)

A production-ready Go repository template following **feature-based clean architecture** with Go 2026 best practices.

## Features

- **Feature-Based Structure**: Code organized by business capabilities for better maintainability
- **Clean Architecture**: Clear separation of concerns across domain, application, infrastructure, and transport layers
- **Type-Safe SQL**: sqlc integration for compile-time SQL verification
- **Zero Import Cycles**: Architecture designed to prevent circular dependencies
- **Testability**: Built-in testing infrastructure with mocks and integration tests
- **Dependency Injection**: Clean DI container for managing dependencies

## Quick Start

### Prerequisites

- Go 1.26+
- PostgreSQL 16+
- Docker (for development environment)

### Setup

1. **Clone the repository**:
   ```bash
   git clone https://github.com/zercle/zercle-go-template.git
   cd zercle-go-template
   ```

2. **Install dependencies**:
   ```bash
   go mod download
   ```

3. **Start development environment**:
   ```bash
   make docker-up
   ```

4. **Run migrations**:
   ```bash
   make migrate-up
   ```

5. **Start the server**:
   ```bash
   make run
   ```

The server will start at `http://localhost:8080`.

### Makefile Commands

| Command | Description |
|---------|-------------|
| `make build` | Build the binary |
| `make run` | Build and run the server |
| `make dev` | Run the server in development mode |
| `make test` | Run unit tests |
| `make test-integration` | Run integration tests |
| `make test-coverage` | Generate coverage report |
| `make lint` | Run linter |
| `make fmt` | Format code |
| `make sqlc-generate` | Generate sqlc code |
| `make migrate-up` | Apply migrations |
| `make migrate-down` | Rollback migrations |
| `make migrate-create name=<name>` | Create a new migration |
| `make docker-up` | Start development containers |
| `make docker-down` | Stop development containers |

## Project Structure

```
zercle-go-template/
├── cmd/                          # Application entrypoints
│   └── server/                   # HTTP API server
│       └── main.go
│
├── internal/                     # Private application code
│   ├── app/                      # Application composition root
│   │   ├── container.go          # Application struct & initialization
│   │   ├── database.go           # Database connection setup
│   │   ├── repositories.go       # Repository initialization
│   │   ├── services.go           # Service initialization
│   │   └── handlers.go           # Handler initialization
│   │
│   ├── user/                     # User feature (combined domain + application)
│   │   ├── entity.go            # User entity, UserID, UserStatus
│   │   ├── entity_test.go        # Entity tests
│   │   ├── errors.go             # Sentinel errors
│   │   ├── valueobjects.go       # Email, Password value objects
│   │   ├── repository.go         # UserRepository interface, ListParams, ListResult
│   │   ├── mocks/                # Mock implementations
│   │   │   └── repository.go     # Mock repository
│   │   ├── dto.go                # CreateRequest, UpdateRequest, Response, ListQuery
│   │   ├── mapper.go             # Domain-DTO converters
│   │   ├── password.go           # Password hashing utilities
│   │   ├── service.go            # Service struct definition
│   │   ├── service_*.go          # Service methods (create, get, list, update, delete)
│   │   └── service_test.go       # Service tests
│   │
│   ├── infrastructure/           # Infrastructure layer
│   │   ├── database/            # Database implementation
│   │   │   ├── postgres.go      # PostgreSQL connection
│   │   │   ├── migrate/         # Database migrations
│   │   │   └── sqlc/            # sqlc generated code
│   │   ├── repository/         # Repository implementations
│   │   │   └── user/            # User repository implementation
│   │   ├── cache/               # Cache implementation
│   │   └── logging/             # Logging infrastructure
│   │
│   └── transport/               # Presentation layer
│       └── http/                 # HTTP transport
│           ├── router/          # Route definitions
│           ├── middleware/       # HTTP middleware
│           ├── response/         # Response utilities
│           └── handler/         # HTTP handlers
│               └── user/        # User feature handlers
│
├── api/                         # API specifications
│   └── openapi/
│       └── openapi.yaml        # OpenAPI specification
│
├── test/                        # Test infrastructure
│   ├── helpers.go
│   ├── mocks/                   # Mock implementations
│   └── integration/            # Integration tests
│
├── docs/                        # Documentation
│   ├── ARCHITECTURE.md         # Architecture details
│   ├── DEVELOPMENT.md          # Development guide
│   ├── API.md                  # API documentation
│   └── CONTRIBUTING.md        # Contribution guidelines
│
├── Makefile                    # Build & development commands
├── sqlc.yaml                   # sqlc configuration
├── docker-compose.yml          # Development environment
└── go.mod                      # Go module definition
```

## Adding a New Feature

See the [Development Guide](docs/DEVELOPMENT.md) for detailed instructions on:

1. Creating domain entities and interfaces
2. Defining repository interfaces
3. Creating SQL queries with sqlc
4. Implementing repositories
5. Creating application services
6. Creating HTTP handlers
7. Registering routes
8. Writing tests

## API Documentation

The API specification is available in OpenAPI format:
- [OpenAPI YAML](api/openapi/openapi.yaml)
- [API Documentation](docs/API.md)

### Base Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check |
| POST | `/api/v1/users` | Create a new user |
| GET | `/api/v1/users` | List users |
| GET | `/api/v1/users/{id}` | Get user by ID |
| PUT | `/api/v1/users/{id}` | Update user |
| DELETE | `/api/v1/users/{id}` | Delete user |

## Testing

### Unit Tests
Run unit tests:
```bash
go test -v ./...
```

Or use Makefile:
```bash
make test
```

### Integration Tests
Integration tests require Docker:
```bash
go test -v ./test/integration/...
```

Or use Makefile:
```bash
make test-integration
```

### Coverage
Run tests with coverage:
```bash
go test -cover ./...
```

Or generate a detailed coverage report:
```bash
make test-coverage
```

### Current Coverage Metrics

| Package | Coverage |
|---------|----------|
| Task Usecase | 73.8% |
| User Usecase | 68.7% |
| HTTP Response | 50.0% |

### Test Structure

```
test/
├── helpers.go              # Test utilities
├── mocks/                  # Generated mocks
│   └── user_repository.go # User repository mock
└── integration/           # Integration tests
    ├── setup.go           # Test setup with testcontainers
    └── user_test.go       # User integration tests
```

## Architecture

The template follows feature-based clean architecture with three main layers:

1. **Feature Layer** (`internal/{feature}/`): Combined domain and application code with entities, value objects, repository interfaces, DTOs, services, mappers, and password utilities
2. **Infrastructure Layer** (`internal/infrastructure/`): Database, cache, and external service implementations
3. **Transport Layer** (`internal/transport/`): HTTP handlers and middleware

See [ARCHITECTURE.md](docs/ARCHITECTURE.md) for detailed architecture documentation.

## Contributing

Please read [CONTRIBUTING.md](docs/CONTRIBUTING.md) for details on our development workflow, code style, and pull request process.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
