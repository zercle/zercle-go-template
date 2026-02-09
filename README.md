# Zercle Go Template

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.25+-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Go Version">
  <img src="https://img.shields.io/badge/Echo-v4.15-00ADD8?style=for-the-badge" alt="Echo Version">
  <img src="https://img.shields.io/badge/PostgreSQL-14+-336791?style=for-the-badge&logo=postgresql&logoColor=white" alt="PostgreSQL">
  <img src="https://img.shields.io/badge/Docker-Ready-2496ED?style=for-the-badge&logo=docker&logoColor=white" alt="Docker">
</p>

<p align="center">
  <img src="https://img.shields.io/github/license/zercle/zercle-go-template?style=flat-square" alt="License">
  <img src="https://img.shields.io/github/actions/workflow/status/zercle/zercle-go-template/ci.yml?style=flat-square" alt="CI Status">
  <img src="https://img.shields.io/codecov/c/github/zercle/zercle-go-template?style=flat-square" alt="Coverage">
  <img src="https://img.shields.io/github/v/release/zercle/zercle-go-template?style=flat-square" alt="Release">
</p>

<p align="center">
  <b>Production-ready REST API template with clean architecture, JWT auth, and comprehensive testing.</b>
</p>

---

## Table of Contents

- [Features](#features)
- [Quick Start](#quick-start)
- [Installation](#installation)
- [Configuration](#configuration)
- [API Documentation](#api-documentation)
- [Testing](#testing)
- [Project Structure](#project-structure)
- [Architecture](#architecture)
- [Contributing](#contributing)
- [License](#license)

---

## Features

### Core Features

- **🔐 JWT Authentication** - Secure token-based auth with access/refresh tokens
- **👤 User Management** - Complete CRUD operations with pagination
- **🗄️ Type-Safe Database** - PostgreSQL with SQLC for compile-time query validation
- **📚 Auto Documentation** - Swagger/OpenAPI specs generated from code annotations
- **🧪 Comprehensive Testing** - Unit and integration tests with mocking
- **🐳 Docker Ready** - Multi-stage builds for optimized production images
- **📊 Structured Logging** - JSON logging with correlation IDs
- **⚡ High Performance** - Echo framework with zero-allocation routing

### Developer Experience

- **40+ Makefile Commands** - Build, test, lint, migrate, and more
- **Hot Reload** - Air integration for rapid development
- **Pre-commit Hooks** - Automated code quality checks
- **Mock Generation** - Auto-generate mocks for testing
- **Database Migrations** - Version-controlled schema changes

---

## Quick Start

### Prerequisites

- [Go 1.21+](https://golang.org/dl/)
- [Docker](https://docs.docker.com/get-docker/) (for PostgreSQL)
- [Make](https://www.gnu.org/software/make/)

### 5-Minute Setup

```bash
# 1. Clone the repository
git clone https://github.com/zercle/zercle-go-template.git
cd zercle-go-template

# 2. Install dependencies and tools
make setup

# 3. Start PostgreSQL
docker run -d --name postgres \
  -e POSTGRES_PASSWORD=postgres \
  -p 5432:5432 postgres:14-alpine

# 4. Run migrations
export DB_USER=postgres DB_PASSWORD=postgres DB_HOST=localhost \
       DB_PORT=5432 DB_NAME=zercle_template DB_SSLMODE=disable
make migrate

# 5. Start the server
make run
```

The API is now running at `http://localhost:8080`

- API Base URL: `http://localhost:8080/api/v1`
- Swagger UI: `http://localhost:8080/swagger/index.html`
- Health Check: `http://localhost:8080/health`

---

## Installation

### Step 1: Clone and Setup

```bash
git clone https://github.com/zercle/zercle-go-template.git my-api
cd my-api

# Replace module name (optional)
find . -type f -name "*.go" -exec sed -i '' 's/zercle-go-template/my-api/g' {} +
go mod edit -module my-api
```

### Step 2: Install Development Tools

```bash
# Install all required tools
make install-tools

# This installs:
# - golangci-lint (linting)
# - swag (Swagger generation)
# - mockgen (Mock generation)
# - sqlc (SQL code generation)
# - migrate (Database migrations)
```

### Step 3: Configure Environment

```bash
# Copy configuration file
cp configs/config.yaml configs/config.local.yaml

# Edit with your settings
vim configs/config.local.yaml
```

### Step 4: Setup Database

```bash
# Start PostgreSQL
docker run -d \
  --name postgres \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=zercle_template \
  -p 5432:5432 \
  postgres:14-alpine

# Run migrations
make migrate

# Generate SQLC code
make sqlc
```

---

## Configuration

### Configuration Hierarchy

1. **Environment variables** (highest priority)
2. **Configuration file** (`configs/config.yaml`)
3. **Default values** (lowest priority)

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `APP_APP_NAME` | Application name | `zercle-go-template` |
| `APP_APP_ENVIRONMENT` | Environment | `development` |
| `APP_SERVER_HOST` | Server bind address | `0.0.0.0` |
| `APP_SERVER_PORT` | Server port | `8080` |
| `APP_DATABASE_HOST` | Database host | `localhost` |
| `APP_DATABASE_PORT` | Database port | `5432` |
| `APP_DATABASE_DATABASE` | Database name | `zercle_template` |
| `APP_DATABASE_USERNAME` | Database user | `postgres` |
| `APP_DATABASE_PASSWORD` | Database password | *(empty)* |
| `APP_DATABASE_SSL_MODE` | SSL mode | `disable` |
| `APP_LOG_LEVEL` | Log level | `info` |
| `APP_LOG_FORMAT` | Log format | `json` |

### Example: Production Configuration

```yaml
# configs/config.production.yaml
app:
  environment: "production"

server:
  read_timeout: "60s"
  write_timeout: "60s"

log:
  level: "warn"
  format: "json"

database:
  ssl_mode: "require"
```

Run with:
```bash
APP_APP_ENVIRONMENT=production go run ./cmd/api
```

---

## API Documentation

### Swagger UI

Interactive API documentation is available at:
```
http://localhost:8080/swagger/index.html
```

### API Endpoints

#### Health
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check |

#### Authentication
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/auth/login` | User login |

#### Users
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/users` | Create user |
| GET | `/api/v1/users` | List users (paginated) |
| GET | `/api/v1/users/:id` | Get user by ID |
| PUT | `/api/v1/users/:id` | Update user |
| DELETE | `/api/v1/users/:id` | Delete user |
| PUT | `/api/v1/users/:id/password` | Update password |

### Example Requests

**Create User:**
```bash
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "securepassword123",
    "name": "John Doe"
  }'
```

**Login:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "securepassword123"
  }'
```

**List Users:**
```bash
curl "http://localhost:8080/api/v1/users?page=1&limit=10" \
  -H "Authorization: Bearer <token>"
```

---

## Testing

### Running Tests

```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# View HTML coverage report
make test-coverage-html

# Run integration tests (requires test database)
make test-integration

# Run benchmarks
make benchmark
```

### Test Structure

```
internal/feature/user/
├── handler/
│   └── user_handler_test.go        # HTTP handler tests
├── repository/
│   └── sqlc_repository_test.go     # Repository tests
└── usecase/
    └── user_usecase_test.go        # Business logic tests
```

### Writing Tests

**Unit Test Example:**
```go
func TestUserUsecase_CreateUser(t *testing.T) {
    tests := []struct {
        name    string
        req     dto.CreateUserRequest
        wantErr bool
    }{
        {
            name: "success",
            req: dto.CreateUserRequest{
                Email:    "test@example.com",
                Password: "password123",
                Name:     "Test User",
            },
            wantErr: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

---

## Project Structure

```
zercle-go-template/
├── cmd/
│   └── api/
│       └── main.go                 # Application entry point
├── internal/
│   ├── config/                     # Configuration management
│   │   └── config.go
│   ├── container/                  # Dependency injection
│   │   └── container.go
│   ├── errors/                     # Custom error types
│   │   └── errors.go
│   ├── feature/                    # Feature modules
│   │   ├── auth/                   # Authentication feature
│   │   │   ├── domain/
│   │   │   ├── middleware/
│   │   │   └── usecase/
│   │   └── user/                   # User management feature
│   │       ├── domain/
│   │       ├── dto/
│   │       ├── handler/
│   │       ├── repository/
│   │       └── usecase/
│   ├── infrastructure/             # External dependencies
│   │   └── db/
│   │       ├── migrations/         # Database migrations
│   │       ├── queries/            # SQLC queries
│   │       └── sqlc/               # Generated code
│   ├── logger/                     # Logging utilities
│   └── middleware/                 # HTTP middleware
├── api/
│   └── docs/                       # Swagger documentation
├── configs/
│   └── config.yaml                 # Configuration file
├── .agents/rules/memory-bank/      # Project documentation
├── plans/                          # Architecture plans
├── Makefile                        # Build automation
├── Dockerfile                      # Container build
├── docker-compose.test.yml         # Test environment
├── sqlc.yaml                       # SQLC configuration
└── go.mod                          # Go module definition
```

---

## Architecture

### Clean Architecture

This template implements **Clean Architecture** with clear separation of concerns:

```
┌─────────────────────────────────────────┐
│  Presentation Layer (Handler)           │
│  - HTTP request/response handling       │
│  - Input validation                     │
│  - Swagger documentation                │
├─────────────────────────────────────────┤
│  Business Layer (Usecase)               │
│  - Business logic                       │
│  - Orchestration                        │
│  - Domain rules enforcement             │
├─────────────────────────────────────────┤
│  Data Layer (Repository)                │
│  - Data access abstraction              │
│  - SQLC implementation                  │
│  - In-memory implementation (tests)     │
├─────────────────────────────────────────┤
│  Domain Layer                           │
│  - Entities                             │
│  - Value objects                        │
│  - Domain errors                        │
└─────────────────────────────────────────┘
```

### Key Design Patterns

- **Repository Pattern** - Abstract data access
- **Dependency Injection** - Loose coupling via container
- **DTO Pattern** - Separate API contracts from domain
- **Middleware Chain** - Cross-cutting concerns

### Request Flow

```
HTTP Request → Router → Middleware → Handler → Usecase → Repository → Database
                                              ↓
HTTP Response ← JSON ← Handler ← Usecase ← Domain Objects
```

---

## Development Commands

### Essential Commands

```bash
# Development
make run              # Run the application
make dev              # Run with hot reload (requires Air)
make build            # Build binary
make clean            # Clean build artifacts

# Testing
make test             # Run all tests
make test-coverage    # Generate coverage report
make benchmark        # Run benchmarks

# Code Quality
make lint             # Run linter
make fmt              # Format code
make check            # Run all checks (fmt, vet, lint, test)

# Database
make migrate          # Run migrations
make migrate-create   # Create new migration
make sqlc             # Generate SQLC code

# Documentation
make swagger          # Generate Swagger docs

# Docker
make docker-build     # Build Docker image
make docker-run       # Run Docker container
```

### Full List

```bash
make help             # Show all available commands
```

---

## Contributing

We welcome contributions! Please follow these guidelines:

### Getting Started

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Run quality checks (`make check`)
5. Commit your changes (`git commit -m 'Add amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

### Development Guidelines

- Follow [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Write tests for new features
- Update documentation for API changes
- Ensure all checks pass before submitting PR

### Commit Message Format

```
type(scope): subject

body (optional)

footer (optional)
```

Types: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`

Example:
```
feat(user): add email verification

- Add email verification token generation
- Send verification email on registration
- Add verify endpoint
```

---

## Deployment

### Docker

```bash
# Build production image
make docker-build

# Run container
make docker-run

# Or manually
docker build -t my-api .
docker run -p 8080:8080 \
  -e APP_DATABASE_HOST=db.example.com \
  -e APP_DATABASE_PASSWORD=secret \
  my-api
```

### Environment-Specific Configurations

Create separate config files for each environment:

```
configs/
├── config.yaml              # Default
├── config.development.yaml  # Development overrides
├── config.staging.yaml      # Staging overrides
└── config.production.yaml   # Production overrides
```

### Health Checks

The application includes a health check endpoint:

```bash
curl http://localhost:8080/health
```

Response:
```json
{
  "success": true,
  "data": {
    "status": "healthy",
    "timestamp": "2026-02-08T18:30:00Z"
  }
}
```

---

## Memory Bank

This project uses a comprehensive documentation system in `.agents/rules/memory-bank/`:

| Document | Description |
|----------|-------------|
| [brief.md](.agents/rules/memory-bank/brief.md) | Project overview and requirements |
| [product.md](.agents/rules/memory-bank/product.md) | Product documentation and roadmap |
| [architecture.md](.agents/rules/memory-bank/architecture.md) | System architecture and design patterns |
| [tech.md](.agents/rules/memory-bank/tech.md) | Technology stack and setup instructions |
| [context.md](.agents/rules/memory-bank/context.md) | Current work focus and decisions |
| [tasks.md](.agents/rules/memory-bank/tasks.md) | Development workflows and guides |

---

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## Acknowledgments

- [Echo Framework](https://echo.labstack.com/) - High performance web framework
- [SQLC](https://sqlc.dev/) - Type-safe SQL generator
- [Zerolog](https://github.com/rs/zerolog) - Zero-allocation JSON logger
- [Viper](https://github.com/spf13/viper) - Configuration management

---

<p align="center">
  Built with ❤️ by <a href="https://github.com/zercle">Zercle</a>
</p>
