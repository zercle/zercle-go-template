# Zercle Go Fiber Template

A production-ready Go microservice template using the **Fiber** framework with **Clean Architecture**, **Domain-Driven Design**, **PostgreSQL** (UUIDv7), and **Modular Build System**.

## ✨ Features

### Core Architecture
- **Clean Architecture**: Strict layered architecture with dependency inversion
- **Feature-Based Organization**: Modular architecture with feature separation
- **Build Tags System**: Conditional compilation for modular deployments
- **Dependency Injection**: `samber/do/v2` with hooks for modular registration

### Database & Storage
- **PostgreSQL + UUIDv7**: Native support for time-sorted UUIDs for scalable primary keys
- **Type-Safe Database**: `sqlc` for generating type-safe Go code from SQL
- **Migrations**: Versioned database migrations with golang-migrate

### API & Web
- **High Performance**: Fiber framework for ultra-fast HTTP handling
- **Authentication**: JWT-based middleware with secure token management
- **JSend Responses**: Standardized JSON API response format `{ status, data, message }`
- **API Documentation**: Auto-generated Swagger/OpenAPI 2.0 via `swaggo`

### Infrastructure
- **Configuration**: `viper` for environment-based configuration
- **Structured Logging**: `slog` for context-aware structured logging
- **Health Checks**: Readiness and liveness probes with database connectivity
- **Docker Support**: Multi-stage builds with non-root containers

### Testing & Quality
- **Comprehensive Testing**: 16 test files covering all layers
- **Mock-Based Tests**: Unit tests with generated mocks
- **Integration Tests**: Full stack testing with real database
- **Code Quality**: golangci-lint integration with strict rules
- **Race Detection**: Always enabled in test runs

## 🏗️ Project Structure

```
zercle-go-template/
├── cmd/server/               # Application entry point
│   ├── main.go               # Server bootstrap
│   └── routes_*.go           # Modular route definitions
├── internal/                 # Private application code
│   ├── features/             # Feature-based architecture
│   │   ├── health/           # Health check feature
│   │   │   ├── domain/       # Health entities
│   │   │   ├── dto/          # Health DTOs
│   │   │   ├── handler/      # Health HTTP handlers
│   │   │   ├── repository/   # Health data access
│   │   │   └── service/      # Health business logic
│   │   ├── user/             # User management feature
│   │   │   ├── domain/       # User entities
│   │   │   ├── dto/          # User DTOs
│   │   │   ├── handler/      # User HTTP handlers
│   │   │   ├── repository/   # User data access
│   │   │   └── service/      # User business logic
│   │   └── post/             # Post management feature
│   │       ├── domain/       # Post entities
│   │       ├── dto/          # Post DTOs
│   │       ├── handler/      # Post HTTP handlers
│   │       ├── repository/   # Post data access
│   │       └── service/      # Post business logic
│   ├── core/                 # Core business logic
│   │   ├── domain/           # Shared domain entities
│   │   ├── port/             # Input/Output interfaces
│   │   │   ├── service_*.go  # Service ports
│   │   │   └── repository_*.go # Repository ports
│   │   └── service/          # Shared business services
│   ├── infrastructure/       # Infrastructure wiring
│   │   ├── config/           # Config loading
│   │   ├── container/        # DI containers with hooks
│   │   │   ├── di.go         # Main DI setup
│   │   │   ├── di_*.go       # Feature-specific DI
│   │   │   └── di_hooks_*.go # Conditional hooks
│   │   ├── server/           # Fiber server setup
│   │   └── sqlc/             # Generated database code
│   └── shared/               # Shared utilities
├── pkg/                      # Shared public code
├── sql/                      # SQL assets
│   ├── migrations/           # Database migrations
│   └── queries/              # sqlc query definitions
├── test/                     # Testing
│   ├── integration/          # Integration tests
│   └── mocks/                # Generated mocks
├── .agents/rules/memory-bank/# Project knowledge base
├── docs/                     # Swagger documentation
├── Dockerfile                # Multi-stage build
├── compose.yml               # Local dev stack
├── Makefile                  # Task runner
└── go.mod                    # Dependencies
```

## Quick Start

### Prerequisites
- Go 1.25+
- Docker & Docker Compose
- PostgreSQL 18+ (or via Docker)

### Installation

1. **Clone the repository**:
   ```bash
   git clone https://github.com/zercle/zercle-go-template.git
   cd zercle-go-template
   ```

2. **Setup Environment**:
   ```bash
   cp .env.example .env
   # Ensure DB_DRIVER=postgres
   ```

3. **Start Infrastructure**:
   ```bash
   make docker-up
   # Or manually start Postgres
   ```

4. **Run Migrations**:
   ```bash
   make migrate-up
   ```

5. **Run Application**:
   ```bash
   make run
   # Server: http://localhost:3000
   # Swagger: http://localhost:3000/swagger/index.html
   ```

## Development

### Code Generation
Regenerate SQLC code, Mocks, and Swagger docs:
```bash
make generate
# or individually:
# sqlc generate
# swag init -g cmd/server/main.go
```

### Testing
Run unit and integration tests:
```bash
# Run all tests with race detection
make test

# Generate coverage report
make test-coverage

# Run specific test
go test -v ./internal/features/health/service/

# Run with coverage
go test -v -cover ./...
```

### Build Tags System
Build modular binaries with selective features:
```bash
# Build with all features (default)
make build

# Build specific feature combinations
make build-health    # Health checks only (~35MB)
make build-user      # User management only (~36MB)
make build-post      # Post management only (~36MB)

# Custom build tags
go build -tags "health,user" ./cmd/server
go build -tags "post" ./cmd/server
```

**Available Tags**: `health`, `user`, `post`, `all`

### Database
This template uses **PostgreSQL**.
- **Migrations**: `sql/migrations` (golang-migrate)
- **Queries**: `sql/queries` (sqlc)

## 🚀 API Endpoints

### Health Checks
- `GET /health` - System health check (readiness probe)
- `GET /health/live` - Container liveness check with DB connectivity

### Authentication
- `POST /auth/register` - Register new user
  ```json
  {
    "email": "user@example.com",
    "password": "securepassword",
    "name": "John Doe"
  }
  ```
- `POST /auth/login` - Login and get JWT
  ```json
  {
    "email": "user@example.com",
    "password": "securepassword"
  }
  ```

### User Management
- `GET /users/me` - Get current user profile (requires JWT token)

### Posts
- `POST /posts` - Create post (requires JWT token)
  ```json
  {
    "title": "My Post",
    "content": "Post content here"
  }
  ```
- `GET /posts` - List all posts
- `GET /posts/:id` - Get post by ID

### Response Format
All endpoints return JSend format:
```json
{
  "status": "success",
  "data": { ... },
  "message": "optional message"
}
```

### Error Responses
```json
{
  "status": "error",
  "data": null,
  "message": "Error description"
}
```

## 🧪 Testing Coverage

This template includes comprehensive testing across all layers:

- **16 Test Files** covering:
  - Health feature (3 tests)
  - User feature (3 tests)
  - Post feature (3 tests)
  - Middleware, Integration, and Utility tests (7 tests)

### Test Structure
```
test/
├── integration/          # Integration tests with real DB
├── mocks/                # Generated mocks for interfaces
└── *_test.go            # Unit tests for each layer
```

### Test Commands
```bash
# Run all tests
make test

# Run with coverage and race detection
make test-coverage

# Run integration tests
go test -v ./test/integration/

# Generate mock files
make generate
```

## 📚 Documentation

- **Memory Bank**: Project knowledge in `.agents/rules/memory-bank/`
- **API Docs**: Auto-generated at `/swagger/index.html`
- **Architecture**: See `internal/features/` for feature-based structure
- **Build Tags**: See `BUILD_TAGS.md` for modular build documentation

## 🎯 Project Status

✅ **Production Ready**

- Complete feature implementation (Health, User, Post)
- Comprehensive test suite (16 test files)
- Modular build system with conditional compilation
- Docker support with multi-stage builds
- Full documentation and examples

## License
MIT
