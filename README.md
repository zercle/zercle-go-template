# zercle-go-template

Production-ready Go REST API template with clean architecture, online booking system, and comprehensive testing.

## Features

- **Architecture**: Clean Architecture with Handler → UseCase → Repository layers
- **Framework**: Echo - high-performance, minimalist web framework
- **Database**: PostgreSQL with SQLC for type-safe database access
- **Security**: JWT authentication, request validation, rate limiting, CORS
- **Observability**: Structured logging (zerolog), health checks, request tracing
- **Standards**: JSend response format, graceful shutdown, containerized deployment
- **Testing**: Table-driven tests, integration tests, 80%+ coverage target
- **Containers**: Podman-first with Docker compatibility

## Quick Start

### Prerequisites

- Go 1.25+
- PostgreSQL 18+
- Podman (recommended) or Docker & Docker Compose
  - See [PODMAN.md](./PODMAN.md) for Podman setup

### Installation

```bash
git clone https://github.com/zercle/zercle-go-template.git
cd zercle-go-template
make init       # Install dependencies
make generate   # Generate SQLC code and mocks
```

### Running

#### Podman/Docker (Recommended)

```bash
make docker-up    # Start API + PostgreSQL
make docker-down  # Stop services
podman logs -f zercle-api  # View logs
```

#### Local Development

```bash
export SERVER_ENV=local DB_HOST=localhost DB_PORT=5432 \
       DB_USER=postgres DB_PASSWORD=password DB_NAME=zercle_db
make run           # or: go run ./cmd/server
```

#### Build Binary

```bash
make build                    # Build for current platform
GOOS=linux GOARCH=amd64 make build
./bin/service
```

## API Examples

### Health Check

```bash
curl http://localhost:3000/health
```

### Authentication

```bash
# Register
curl -X POST http://localhost:3000/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123","full_name":"John Doe"}'

# Login
curl -X POST http://localhost:3000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123"}'
```

### Protected Routes

```bash
# Get profile
curl http://localhost:3000/api/v1/users/profile \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"

# Update profile
curl -X PUT http://localhost:3000/api/v1/users/profile \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"full_name":"John Updated"}'
```

## Project Structure

```
zercle-go-template/
├── cmd/server/              # Application entry point
├── configs/                 # Environment configs (local/dev/uat/prod)
├── domain/                  # Business logic layer
│   ├── user/               # User domain
│   │   ├── handler/        # HTTP handlers
│   │   ├── usecase/        # Business logic
│   │   ├── repository/     # Data access
│   │   ├── model/          # Domain models
│   │   ├── request/        # Request DTOs
│   │   └── response/       # Response DTOs
│   ├── service/            # Service domain
│   ├── booking/            # Booking domain
│   └── payment/            # Payment domain
├── infrastructure/          # External dependencies
│   ├── config/             # Configuration
│   ├── db/                 # Database connection
│   ├── logger/             # Logging
│   └── sqlc/db/            # SQLC generated code
├── pkg/                     # Shared packages
│   ├── middleware/         # HTTP middleware
│   ├── response/           # Response utilities
│   └── health/             # Health checks
├── sql/                     # Database
│   ├── migration/          # Migrations
│   └── query/              # SQLC queries
└── test/                    # Tests
    ├── unit/               # Unit tests
    └── integration/        # Integration tests
```

## Configuration

Config loaded from `configs/` based on `SERVER_ENV` environment variable.

| Variable | Description | Default |
|----------|-------------|---------|
| `SERVER_ENV` | Environment | `local` |
| `SERVER_PORT` | Server port | `3000` |
| `DB_HOST` | Database host | `localhost` |
| `DB_PORT` | Database port | `5432` |
| `DB_USER` | Database user | `postgres` |
| `DB_PASSWORD` | Database password | `password` |
| `DB_NAME` | Database name | `zercle_db` |
| `JWT_SECRET` | JWT signing secret | - |
| `JWT_EXPIRATION` | JWT expiration (seconds) | `3600` |
| `LOG_LEVEL` | Log level | `info` |

## Testing

```bash
make test            # Run all tests
make test-coverage   # Run with coverage
open coverage.html   # View report
```

### Test Structure

- **Unit Tests**: Table-driven tests for handlers, usecases, repositories
- **Integration Tests**: Full HTTP request/response cycle
- **Coverage Target**: Minimum 80%

## Makefile Commands

| Command | Description |
|---------|-------------|
| `make help` | Show available commands |
| `make init` | Initialize dependencies |
| `make generate` | Generate SQLC code and mocks |
| `make build` | Build application |
| `make dev` | Run in development mode |
| `make run` | Run compiled binary |
| `make test` | Run tests |
| `make test-coverage` | Run tests with coverage |
| `make lint` | Run linter |
| `make fmt` | Format code |
| `make clean` | Clean build artifacts |
| `make docker-build` | Build container image |
| `make docker-up` | Start containers |
| `make docker-down` | Stop containers |

## API Standards

### Response Format (JSend)

```json
{
  "status": "success",
  "data": { ... }
}
```

```json
{
  "status": "fail",
  "message": "Validation failed",
  "errors": [{ "field": "email", "message": "Must be a valid email address" }]
}
```

```json
{
  "status": "error",
  "message": "Internal server error"
}
```

### HTTP Status Codes

| Code | Description |
|------|-------------|
| 200 | Successful request |
| 201 | Resource created |
| 204 | Successful request, no response body |
| 400 | Invalid request data |
| 401 | Authentication required |
| 403 | Insufficient permissions |
| 404 | Resource not found |
| 429 | Rate limit exceeded |
| 500 | Server error |

### Request Headers

- `X-Request-ID` - Unique request identifier for tracing
- `Authorization` - Bearer token for authenticated requests

## Development Guidelines

### Clean Architecture

1. **Handlers**: HTTP request/response handling only
2. **UseCases**: Business logic and orchestration
3. **Repositories**: Data access abstraction, interface-based
4. **No direct database access from handlers**

### Coding Standards

- Table-driven tests for multiple scenarios
- Exported functions require godoc comments
- Descriptive names (no abbreviations)
- Explicit error handling
- Contextual error logging (request_id, user_id)

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push to branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

[MIT License](LICENSE)

## Acknowledgments

- [Echo](https://echo.labstack.com/) - High-performance web framework
- [SQLC](https://sqlc.dev/) - Type-safe SQL builder
- [zerolog](https://github.com/rs/zerolog) - Zero-allocation logging
- [go-playground/validator](https://github.com/go-playground/validator) - Struct validation
