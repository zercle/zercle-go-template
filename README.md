# Zercle Go Fiber Template

A production-ready Go microservice template using the **Fiber** framework with **Clean Architecture**, **Domain-Driven Design**, and **PostgreSQL** (UUIDv7).

## Features

- **High Performance**: Fiber framework for ultra-fast HTTP handling.
- **Clean Architecture**: Strict layered architecture (Handler → Service → Repository → Domain).
- **PostgreSQL + UUIDv7**: Native support for time-sorted UUIDs (`uuidv7`) for scalable primary keys.
- **Type-Safe Database**: `sqlc` for generating type-safe Go code from SQL.
- **Dependency Injection**: `samber/do/v2` for robust DI container management.
- **Authentication**: JWT-based middleware with secure token management.
- **JSend Responses**: Standardized JSON API response format `{ status, data, message }`.
- **Configuration**: `viper` for environment-based configuration.
- **Structured Logging**: `slog` for context-aware structured logging.
- **Observability**: Ready for metrics and tracing integration.
- **API Documentation**: Auto-generated Swagger/OpenAPI 2.0 via `swaggo`.

## Project Structure

```bash
zercle-go-template/
├── cmd/server/               # Application entry point
├── internal/                 # Private application code
│   ├── adapter/              # Interface adapters (HTTP, DB)
│   │   ├── handler/          # HTTP handlers (Fiber)
│   │   └── storage/          # Database repositories (Postgres)
│   ├── core/                 # Core business logic
│   │   ├── domain/           # Domain entities & errors
│   │   ├── port/             # Input/Output interfaces
│   │   └── service/          # Business services
│   ├── infrastructure/       # Infrastructure wiring
│   │   ├── config/           # Config loading
│   │   ├── container/        # DI containers
│   │   ├── server/           # Fiber server setup
│   │   └── sqlc/             # Generated database code
│   └── middleware/           # HTTP middleware (Auth, CORS, Logger)
├── pkg/                      # Shared public code
│   ├── dto/                  # Data Transfer Objects
│   └── utils/                # Utilities (Response, Logger)
├── sql/                      # SQL assets
│   ├── migrations/           # Database migrations
│   └── queries/              # sqlc query definitions
├── test/                     # Testing
│   ├── integration/          # Integration tests
│   └── mocks/                # Generated/Manual mocks
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
make test
# or
go test -v ./...
```

### Database
This template uses **PostgreSQL**.
- **Migrations**: `sql/migrations` (golang-migrate)
- **Queries**: `sql/queries` (sqlc)

## API Endpoints

### System
- `GET /health` - System health check (Readiness)
- `GET /health/live` - Container liveness check

### Auth
- `POST /auth/register` - Register new user
- `POST /auth/login` - Login and get JWT

### Users
- `GET /users/me` - Get current user profile

### Posts
- `POST /posts` - Create post
- `GET /posts` - List all posts
- `GET /posts/:id` - Get post by ID

## License
MIT
