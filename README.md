# Zercle Go Template

[![Go Version](https://img.shields.io/badge/go-1.24+-00ADD8?style=for-the-badge&logo=go&logoColor=white)](https://golang.org/doc/devel/release.html)
[![Tests](https://img.shields.io/badge/tests-passing-brightgreen?style=for-the-badge&logo=testinglibrary&logoColor=white)]()
[![License](https://img.shields.io/badge/license-MIT-blue?style=for-the-badge)](LICENSE)
[![Docker](https://img.shields.io/badge/docker-ready-2496ED?style=for-the-badge&logo=docker&logoColor=white)]()

> 🚀 Production-ready Go backend template implementing Clean Architecture with user management and JWT authentication

[📖 Documentation](#table-of-contents) • [🚀 Quick Start](#quick-start) • [📚 API Docs](#api-documentation)

---

## 📋 Table of Contents

- [Overview](#overview)
- [Features](#features)
- [Quick Start](#quick-start)
- [Project Structure](#project-structure)
- [Architecture](#architecture)
- [API Documentation](#api-documentation)
- [Configuration](#configuration)
- [Testing](#testing)
- [Development](#development)
- [Deployment](#deployment)
- [Contributing](#contributing)
- [License](#license)

---

## 🎯 Overview

**Zercle Go Template** is a production-ready Go backend template designed for rapid REST API development. It provides a solid foundation for building scalable, maintainable, and well-tested microservices following industry best practices.

### Who Should Use This?

- **Teams** starting new Go projects who want a proven architecture
- **Developers** learning Clean Architecture and domain-driven design
- **Startups** needing to ship production APIs quickly
- **Engineers** who value type safety, testability, and maintainability

### Key Benefits

- ✅ **Clean Architecture** - Clear separation of concerns with domain-driven design
- ✅ **Production-Ready** - Includes logging, error handling, middleware, and more
- ✅ **Type-Safe** - SQL code generation with sqlc ensures compile-time query validation
- ✅ **Well-Tested** - Comprehensive unit and integration test coverage
- ✅ **Developer-Friendly** - Hot reload, make commands, and clear documentation

---

## ✨ Features

### 🏗️ Clean Architecture Implementation
- Layered architecture: Handler → Usecase → Repository
- Domain-driven design with clear boundaries
- Dependency injection for testability
- Interface-driven development

### 🔐 JWT Authentication
- Access and refresh token support
- Configurable token TTL
- Secure password hashing with bcrypt
- Auth middleware for protected routes

### 👤 User Management
- Complete CRUD operations
- Password management with validation
- Pagination support for listing users
- Input validation with struct tags

### 🗄️ Type-Safe Database
- PostgreSQL with sqlc code generation
- Compile-time SQL query validation
- Migration support
- Connection pooling

### 🛠️ Production-Ready Tooling
- Structured logging with Zerolog
- Configuration management with Viper
- Swagger/OpenAPI documentation
- Request ID tracking
- Graceful shutdown handling

### 🧪 Comprehensive Testing
- Unit tests with mocks
- Integration tests with real database
- Race condition detection
- Coverage reporting

### 🐳 Docker Support
- Multi-stage Dockerfile
- Docker Compose for local development
- Docker Compose for testing

---

## 🚀 Quick Start

### Prerequisites

- **Go** 1.24+ ([Download](https://golang.org/dl/))
- **PostgreSQL** 14+ ([Download](https://www.postgresql.org/download/))
- **Docker** (optional, for containerized development)
- **Make** (optional, for convenience commands)

### Installation

1. **Clone the repository**

```bash
git clone https://github.com/yourusername/zercle-go-template.git
cd zercle-go-template
```

2. **Install dependencies**

```bash
go mod download
```

3. **Set up environment variables**

```bash
cp .env.example .env
# Edit .env with your configuration
```

4. **Set up the database**

```bash
# Create PostgreSQL database
createdb zercle_template

# Run migrations
go run ./cmd/migrate
```

### Running the Application

**Using Make (recommended):**

```bash
make run
```

**Using Go directly:**

```bash
go run ./cmd/api
```

**Using Docker:**

```bash
docker-compose up -d
```

The API will be available at `http://localhost:8080`

### Environment Configuration

Create a `.env` file from the example:

```bash
cp .env.example .env
```

Key configuration options:

| Variable | Description | Default |
|----------|-------------|---------|
| `APP_APP_ENVIRONMENT` | Environment mode | `development` |
| `APP_SERVER_PORT` | HTTP server port | `8080` |
| `APP_DATABASE_HOST` | PostgreSQL host | `localhost` |
| `APP_DATABASE_PORT` | PostgreSQL port | `5432` |
| `APP_DATABASE_DATABASE` | Database name | `zercle_template` |
| `APP_DATABASE_USERNAME` | Database username | `postgres` |
| `APP_DATABASE_PASSWORD` | Database password | `postgres` |
| `APP_JWT_SECRET` | JWT signing secret | *(required)* |
| `APP_JWT_ACCESS_TOKEN_TTL` | Access token lifetime | `15m` |
| `APP_JWT_REFRESH_TOKEN_TTL` | Refresh token lifetime | `168h` |
| `APP_LOG_LEVEL` | Log level | `info` |

---

## 📁 Project Structure

```
zercle-go-template/
├── api/
│   └── docs/                  # Swagger/OpenAPI documentation
├── cmd/
│   └── api/                   # Application entry point
├── configs/
│   └── config.yaml            # Default configuration
├── internal/
│   ├── config/                # Configuration management
│   ├── container/             # Dependency injection container
│   ├── errors/                # Custom error types
│   ├── feature/               # Feature modules
│   │   ├── auth/              # Authentication feature
│   │   │   ├── domain/        # JWT domain models
│   │   │   ├── middleware/    # Auth middleware
│   │   │   └── usecase/       # JWT usecases
│   │   └── user/              # User feature
│   │       ├── domain/        # User domain models
│   │       ├── dto/           # Data transfer objects
│   │       ├── handler/       # HTTP handlers
│   │       ├── repository/    # Data access layer
│   │       └── usecase/       # Business logic
│   ├── infrastructure/
│   │   └── db/                # Database layer
│   │       ├── migrations/    # SQL migrations
│   │       ├── queries/       # SQL queries
│   │       └── sqlc/          # Generated code
│   ├── logger/                # Logging utilities
│   └── middleware/            # HTTP middleware
├── plans/                     # Architecture plans
├── docker-compose.yml         # Docker services
├── Dockerfile                 # Application image
├── Makefile                   # Build automation
├── sqlc.yaml                  # SQLC configuration
└── go.mod                     # Go module definition
```

### Key Directories Explained

| Directory | Purpose |
|-----------|---------|
| `internal/feature/` | Domain features following Clean Architecture |
| `internal/feature/*/domain/` | Domain entities and business rules |
| `internal/feature/*/dto/` | Request/response data structures |
| `internal/feature/*/handler/` | HTTP request handlers |
| `internal/feature/*/usecase/` | Business logic and orchestration |
| `internal/feature/*/repository/` | Data persistence abstraction |
| `internal/infrastructure/db/` | Database implementation details |

---

## 🏛️ Architecture

### Clean Architecture Layers

This project follows **Clean Architecture** principles with clear dependency direction:

```
┌─────────────────────────────────────────────────────────────┐
│                    External Layer                           │
│         (HTTP Handlers, Database, External APIs)            │
└───────────────────────────┬─────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                   Usecase Layer                             │
│            (Business Logic, Application Services)           │
└───────────────────────────┬─────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                    Domain Layer                             │
│         (Entities, Domain Services, Repository Interfaces)  │
└─────────────────────────────────────────────────────────────┘
```

### Data Flow

```
HTTP Request → Handler → DTO → Usecase → Repository → Database
     │                                                    │
     │            Response ← DTO ← Domain ← Data          │
     └────────────────────────────────────────────────────┘
```

### Key Principles

1. **Dependency Inversion** - Dependencies point inward toward the domain
2. **Interface Segregation** - Small, focused interfaces (e.g., [`UserRepository`](internal/feature/user/repository/user_repository.go))
3. **Single Responsibility** - Each layer has one reason to change
4. **Testability** - All layers can be tested in isolation with mocks

---

## 📚 API Documentation

### Swagger UI

Interactive API documentation is available at:

```
http://localhost:8080/swagger/index.html
```

### API Endpoints

#### Authentication

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| `POST` | `/api/v1/auth/login` | Login with credentials | No |
| `POST` | `/api/v1/auth/refresh` | Refresh access token | No |
| `POST` | `/api/v1/auth/logout` | Logout and invalidate token | Yes |

#### Users

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| `GET` | `/api/v1/users` | List all users (paginated) | Yes |
| `POST` | `/api/v1/users` | Create new user | No |
| `GET` | `/api/v1/users/:id` | Get user by ID | Yes |
| `PUT` | `/api/v1/users/:id` | Update user | Yes |
| `DELETE` | `/api/v1/users/:id` | Delete user | Yes |
| `PUT` | `/api/v1/users/:id/password` | Update password | Yes |

### Example Requests

**Login:**

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123"
  }'
```

**Create User:**

```bash
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com",
    "password": "securepassword"
  }'
```

**List Users (Authenticated):**

```bash
curl -X GET "http://localhost:8080/api/v1/users?page=1&limit=10" \
  -H "Authorization: Bearer <access_token>"
```

### Authentication Flow

```
┌──────────┐         Login          ┌──────────┐
│  Client  │ ─────────────────────→ │   API    │
└──────────┘                        └──────────┘
     │                                   │
     │    Access Token + Refresh Token   │
     │ ←─────────────────────────────────│
     │                                   │
     │ ───── API Call (Access Token) ─→ │
     │                                   │
     │ ←────── Protected Resource ──────│
     │                                   │
     │ ─── Token Expired ─────────────→ │
     │                                   │
     │ ───── Refresh Token Request ───→ │
     │                                   │
     │    New Access + Refresh Tokens    │
     │ ←─────────────────────────────────│
```

---

## ⚙️ Configuration

### Environment Variables

All configuration is managed through environment variables with the `APP_` prefix:

#### Application

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `APP_APP_NAME` | string | `zercle-go-template` | Application name |
| `APP_APP_VERSION` | string | `1.0.0` | Application version |
| `APP_APP_ENVIRONMENT` | string | `development` | Environment mode |

#### Server

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `APP_SERVER_HOST` | string | `0.0.0.0` | Server bind address |
| `APP_SERVER_PORT` | int | `8080` | Server port |
| `APP_SERVER_READ_TIMEOUT` | duration | `30s` | HTTP read timeout |
| `APP_SERVER_WRITE_TIMEOUT` | duration | `30s` | HTTP write timeout |
| `APP_SERVER_SHUTDOWN_TIMEOUT` | duration | `10s` | Graceful shutdown timeout |

#### Database

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `APP_DATABASE_HOST` | string | `localhost` | PostgreSQL host |
| `APP_DATABASE_PORT` | int | `5432` | PostgreSQL port |
| `APP_DATABASE_DATABASE` | string | - | Database name |
| `APP_DATABASE_USERNAME` | string | - | Database username |
| `APP_DATABASE_PASSWORD` | string | - | Database password |
| `APP_DATABASE_SSL_MODE` | string | `disable` | SSL mode |

#### JWT

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `APP_JWT_SECRET` | string | - | JWT signing secret |
| `APP_JWT_ACCESS_TOKEN_TTL` | duration | `15m` | Access token TTL |
| `APP_JWT_REFRESH_TOKEN_TTL` | duration | `168h` | Refresh token TTL |

#### Logging

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `APP_LOG_LEVEL` | string | `info` | Log level (debug/info/warn/error) |
| `APP_LOG_FORMAT` | string | `json` | Log format (json/console) |

### Configuration Precedence

Configuration is loaded in the following order (later overrides earlier):

1. Default values
2. Configuration file (`configs/config.yaml`)
3. Environment variables
4. Command-line flags

### Example `.env` Setup

```bash
# Application
APP_APP_NAME=my-service
APP_APP_ENVIRONMENT=development

# Server
APP_SERVER_PORT=8080

# Database
APP_DATABASE_HOST=localhost
APP_DATABASE_PORT=5432
APP_DATABASE_DATABASE=mydb
APP_DATABASE_USERNAME=postgres
APP_DATABASE_PASSWORD=secretpassword

# JWT (generate with: openssl rand -base64 32)
APP_JWT_SECRET=your-256-bit-secret-here
APP_JWT_ACCESS_TOKEN_TTL=15m
APP_JWT_REFRESH_TOKEN_TTL=168h

# Logging
APP_LOG_LEVEL=debug
APP_LOG_FORMAT=console
```

---

## 🧪 Testing

### Running Tests

**All tests:**

```bash
make test
```

**Unit tests only:**

```bash
make test-unit
```

**Integration tests (requires database):**

```bash
make test-integration
```

**With coverage report:**

```bash
make test-coverage
```

### Test Types

| Type | Description | Location |
|------|-------------|----------|
| **Unit** | Tests individual components with mocks | `*_test.go` (alongside source) |
| **Integration** | Tests database interactions | `*_integration_test.go` |
| **Handler** | Tests HTTP handlers | `handler/*_test.go` |

### Coverage Report

Generate and view coverage:

```bash
make test-coverage
open coverage.html
```

The project maintains high test coverage for critical paths:
- Domain logic: 90%+
- Usecases: 85%+
- Handlers: 80%+

---

## 💻 Development

### Available Make Commands

| Command | Description |
|---------|-------------|
| `make all` | Run all checks and build |
| `make build` | Build the application binary |
| `make run` | Run the application locally |
| `make test` | Run all tests with race detection |
| `make test-coverage` | Generate coverage report |
| `make lint` | Run linters (golangci-lint) |
| `make fmt` | Format code |
| `make vet` | Run go vet |
| `make deps` | Download dependencies |
| `make clean` | Clean build artifacts |
| `make docker-build` | Build Docker image |
| `make docker-up` | Start Docker services |
| `make generate` | Run all code generation |
| `make generate-sqlc` | Generate SQLC code |
| `make generate-mocks` | Generate mock implementations |
| `make migrate-up` | Run database migrations |
| `make migrate-down` | Rollback migrations |

### Adding New Features

1. **Define domain models** in `internal/feature/<feature>/domain/`
2. **Create DTOs** in `internal/feature/<feature>/dto/`
3. **Implement repository interface** in `internal/feature/<feature>/repository/`
4. **Write usecase logic** in `internal/feature/<feature>/usecase/`
5. **Create HTTP handler** in `internal/feature/<feature>/handler/`
6. **Add routes** in the feature's setup function
7. **Write tests** for each layer

### Code Generation

**SQLC (type-safe SQL):**

Edit SQL queries in `internal/infrastructure/db/queries/`, then run:

```bash
make generate-sqlc
```

**Mocks (for testing):**

```bash
make generate-mocks
```

---

## 🐳 Deployment

### Docker Build

**Build image:**

```bash
make docker-build
```

**Or manually:**

```bash
docker build -t zercle-go-template:latest .
```

### Running with Docker Compose

**Development:**

```bash
docker-compose up -d
```

This starts:
- Go application on port 8080
- PostgreSQL on port 5432

**Testing:**

```bash
docker-compose -f docker-compose.test.yml up --abort-on-container-exit
```

### Production Considerations

1. **Environment Variables:**
   - Use strong JWT secrets (256-bit)
   - Enable SSL for database connections
   - Set appropriate log levels

2. **Security:**
   - Run as non-root user in container
   - Use secrets management (e.g., Docker secrets, Kubernetes secrets)
   - Enable HTTPS/TLS termination

3. **Performance:**
   - Tune database connection pool
   - Set appropriate timeouts
   - Enable request rate limiting

4. **Monitoring:**
   - Use structured logging (JSON format)
   - Set up log aggregation
   - Configure health checks

### Environment-Specific Configs

Use `APP_APP_ENVIRONMENT` to control behavior:

| Environment | Characteristics |
|-------------|-----------------|
| `development` | Debug logging, detailed errors, CORS enabled |
| `staging` | Info logging, production-like setup |
| `production` | JSON logging, minimal error details, security headers |

---

## 🤝 Contributing

We welcome contributions! Here's how to get started:

### Getting Started

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/my-feature`
3. Make your changes
4. Run tests: `make test`
5. Run linters: `make lint`
6. Commit with clear messages
7. Push and create a pull request

### Code Standards

- Follow [Effective Go](https://go.dev/doc/effective_go) guidelines
- Use `gofmt` for formatting
- Pass `golangci-lint` checks
- Write tests for new code
- Maintain backward compatibility

### Pull Request Process

1. Ensure all tests pass
2. Update documentation if needed
3. Add changelog entry for significant changes
4. Request review from maintainers
5. Address review feedback
6. Squash commits if requested

### Code of Conduct

- Be respectful and constructive
- Focus on the code, not the person
- Help newcomers learn and grow

---

## 📄 License

This project is licensed under the **MIT License** - see below for details:

```
MIT License

Copyright (c) 2024 Zercle

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```

---

## 🙏 Acknowledgments

- [Echo](https://echo.labstack.com/) - High performance Go web framework
- [sqlc](https://sqlc.dev/) - Type-safe SQL generator
- [Zerolog](https://github.com/rs/zerolog) - Zero allocation JSON logger
- [Viper](https://github.com/spf13/viper) - Go configuration solution

---

<div align="center">

**Made with ❤️ by the Zercle Team**

[⬆ Back to Top](#zercle-go-template)

</div>
