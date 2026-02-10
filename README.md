# 🚀 Zercle Go Template

<p align="center">
  <a href="https://go.dev"><img src="https://img.shields.io/badge/Go-1.25.7-00ADD8?style=for-the-badge&logo=go" alt="Go Version"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/License-MIT-green.svg?style=for-the-badge" alt="License: MIT"></a>
  <a href="https://github.com/zercle/zercle-go-template/actions"><img src="https://img.shields.io/badge/CI%2FCD-GitHub%20Actions-blue?style=for-the-badge&logo=githubactions" alt="CI/CD"></a>
</p>

<p align="center">
  <strong>Production-ready Go web application template with Clean Architecture, JWT authentication, and comprehensive tooling.</strong>
</p>

<p align="center">
  <a href="#-features">Features</a> •
  <a href="#-tech-stack">Tech Stack</a> •
  <a href="#-quick-start">Quick Start</a> •
  <a href="#-project-structure">Structure</a> •
  <a href="#-api-documentation">API Docs</a> •
  <a href="#-testing">Testing</a> •
  <a href="#-deployment">Deployment</a>
</p>

---

## ✨ Features

- 🏗️ **Clean Architecture** - Feature-based organization with clear separation of concerns
- 🔐 **JWT Authentication** - Secure authentication with Argon2id password hashing
- 🗄️ **PostgreSQL** - Type-safe database queries with sqlc
- 🌐 **Echo Framework** - High-performance, minimalist web framework
- 📝 **Swagger/OpenAPI** - Auto-generated API documentation
- 🧪 **Comprehensive Testing** - Unit, integration, and benchmark tests
- 🐳 **Docker Support** - Multi-stage builds for production-ready containers
- ⚡ **GitHub Actions** - Automated CI/CD pipeline
- 🪝 **Pre-commit Hooks** - Automated code quality checks
- 📊 **Structured Logging** - High-performance logging with Zerolog
- ⚙️ **Configuration Management** - Environment-based config with Viper
- 🔒 **Security Best Practices** - Input validation, secure headers, and more

---

## 🛠️ Tech Stack

| Category | Technology |
|----------|------------|
| **Language** | [Go 1.25.7](https://go.dev) |
| **Web Framework** | [Echo v4](https://echo.labstack.com/) |
| **Database** | [PostgreSQL](https://www.postgresql.org/) + [pgx v5](https://github.com/jackc/pgx) |
| **SQL Codegen** | [sqlc](https://sqlc.dev/) |
| **Authentication** | [golang-jwt](https://github.com/golang-jwt/jwt) + [Argon2id](https://pkg.go.dev/golang.org/x/crypto/argon2) |
| **Configuration** | [Viper](https://github.com/spf13/viper) |
| **Logging** | [Zerolog](https://github.com/rs/zerolog) |
| **Validation** | [validator](https://github.com/go-playground/validator) |
| **Documentation** | [Swagger](https://swagger.io/) |
| **Testing** | [Testify](https://github.com/stretchr/testify) + [Mockgen](https://github.com/uber-go/mock) |
| **Linting** | [golangci-lint](https://golangci-lint.run/) |
| **Containerization** | [Docker](https://www.docker.com/) + [Distroless](https://github.com/GoogleContainerTools/distroless) |
| **CI/CD** | [GitHub Actions](https://github.com/features/actions) |

---

## 🏛️ Architecture

This project follows **Clean Architecture** principles with a feature-based organization:

```
┌─────────────────────────────────────────────────────────────┐
│                        Presentation Layer                    │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │
│  │   Handler   │  │ Middleware  │  │   Swagger Docs      │  │
│  └──────┬──────┘  └─────────────┘  └─────────────────────┘  │
└─────────┼────────────────────────────────────────────────────┘
          │
          ▼
┌─────────────────────────────────────────────────────────────┐
│                        Application Layer                     │
│  ┌─────────────────────────────────────────────────────┐    │
│  │                    Use Cases                         │    │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────────────────┐ │    │
│  │  │   User   │ │   Auth   │ │   [Other Features]   │ │    │
│  │  └──────────┘ └──────────┘ └──────────────────────┘ │    │
│  └─────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────┘
          │
          ▼
┌─────────────────────────────────────────────────────────────┐
│                        Domain Layer                          │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │
│  │   Domain    │  │     DTO     │  │     Interfaces      │  │
│  │   Models    │  │             │  │                     │  │
│  └─────────────┘  └─────────────┘  └─────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
          │
          ▼
┌─────────────────────────────────────────────────────────────┐
│                     Infrastructure Layer                     │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │
│  │ Repository  │  │   Database  │  │   External APIs     │  │
│  │  (sqlc)     │  │  (PostgreSQL)│  │                     │  │
│  └─────────────┘  └─────────────┘  └─────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

### Key Principles

- **Dependency Inversion**: Dependencies point inward toward the domain
- **Single Responsibility**: Each layer has a clear, focused responsibility
- **Testability**: Business logic is decoupled from frameworks and infrastructure
- **Feature-Based**: Related code is colocated by feature, not by layer type

---

## 🚀 Quick Start

### Prerequisites

- Go 1.25.7 or later
- PostgreSQL 14+ (or Docker for containerized database)
- Make (optional, for using Makefile commands)

### Installation

1. **Clone the repository**

   ```bash
   git clone https://github.com/zercle/zercle-go-template.git
   cd zercle-go-template
   ```

2. **Copy environment file**

   ```bash
   cp .env.example .env
   # Edit .env with your database credentials and JWT secret
   ```

3. **Install dependencies**

   ```bash
   go mod download
   ```

4. **Run database migrations**

   ```bash
   # Install migrate tool if not already installed
   go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

   # Run migrations
   make migrate-up
   ```

5. **Run the application**

   ```bash
   make run
   ```

   The API will be available at `http://localhost:8080`

### Using Docker

```bash
# Build and run with Docker Compose
docker-compose up -d

# Or build the image manually
make docker-build
make docker-run
```

---

## 📁 Project Structure

```
zercle-go-template/
├── 📂 cmd/
│   └── 📂 api/                 # Application entry point
│       └── 📄 main.go          # Main application
│
├── 📂 internal/                # Private application code
│   ├── 📂 config/              # Configuration management (Viper)
│   │   ├── 📄 config.go
│   │   └── 📄 config_test.go
│   │
│   ├── 📂 container/           # Dependency injection container
│   │   ├── 📄 container.go
│   │   └── 📄 container_test.go
│   │
│   ├── 📂 errors/              # Custom error types
│   │   ├── 📄 errors.go
│   │   └── 📄 errors_test.go
│   │
│   ├── 📂 feature/             # Feature modules
│   │   ├── 📂 auth/            # Authentication feature
│   │   │   ├── 📂 domain/      # JWT domain logic
│   │   │   ├── 📂 middleware/  # Auth middleware
│   │   │   └── 📂 usecase/     # JWT use cases
│   │   │
│   │   └── 📂 user/            # User management feature
│   │       ├── 📂 domain/      # User domain models
│   │       ├── 📂 dto/         # Data transfer objects
│   │       ├── 📂 handler/     # HTTP handlers
│   │       ├── 📂 repository/  # Data access layer
│   │       └── 📂 usecase/     # Business logic
│   │
│   ├── 📂 infrastructure/
│   │   └── 📂 db/              # Database layer
│   │       ├── 📂 migrations/  # SQL migrations
│   │       ├── 📂 queries/     # sqlc query files
│   │       └── 📂 sqlc/        # Generated Go code
│   │
│   ├── 📂 logger/              # Structured logging (Zerolog)
│   │   ├── 📄 logger.go
│   │   └── 📄 logger_test.go
│   │
│   └── 📂 middleware/          # HTTP middleware
│       ├── 📄 logging.go
│       └── 📄 recovery.go
│
├── 📂 api/
│   └── 📂 docs/                # Swagger documentation
│       ├── 📄 docs.go
│       ├── 📄 swagger.json
│       └── 📄 swagger.yaml
│
├── 📂 configs/                 # Configuration files
│   └── 📄 config.yaml
│
├── 📄 .env.example             # Environment variables template
├── 📄 .golangci.yml            # Linter configuration
├── 📄 .pre-commit-config.yaml  # Pre-commit hooks
├── 📄 Dockerfile               # Multi-stage Docker build
├── 📄 docker-compose.test.yml  # Test environment
├── 📄 Makefile                 # Build automation
├── 📄 go.mod                   # Go module definition
├── 📄 go.sum                   # Go module checksums
└── 📄 sqlc.yaml                # sqlc configuration
```

---

## 📚 API Documentation

### Swagger UI

Once the application is running, access the interactive API documentation at:

```
http://localhost:8080/swagger/index.html
```

### Generating Swagger Docs

After modifying handlers or adding new endpoints, regenerate the documentation:

```bash
make swagger
```

This requires the `swag` CLI tool:

```bash
go install github.com/swaggo/swag/cmd/swag@latest
```

### API Endpoints

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| `POST` | `/api/v1/auth/login` | User login | No |
| `POST` | `/api/v1/auth/refresh` | Refresh access token | No |
| `GET` | `/api/v1/users` | List all users | Yes |
| `GET` | `/api/v1/users/:id` | Get user by ID | Yes |
| `POST` | `/api/v1/users` | Create new user | Yes |
| `PUT` | `/api/v1/users/:id` | Update user | Yes |
| `DELETE` | `/api/v1/users/:id` | Delete user | Yes |

---

## 🧪 Testing

### Running Tests

```bash
# Run all tests
make test

# Run unit tests only
make test-unit

# Run integration tests (requires database)
make test-integration

# Run tests with coverage
make test-coverage

# Generate HTML coverage report
make test-coverage-html

# Run benchmarks
make benchmark

# Run short tests (skip integration)
make test-short
```

### Test Structure

| Test Type | Pattern | Location |
|-----------|---------|----------|
| Unit Tests | `*_test.go` | Next to source files |
| Integration Tests | `*_integration_test.go` | Repository and handler packages |
| Benchmark Tests | `*_benchmark_test.go` | Domain and handler packages |
| Mocks | `mocks/` | Generated with mockgen |

### Writing Tests

```go
// Example unit test
func TestUserUsecase_CreateUser(t *testing.T) {
    // Arrange
    mockRepo := mocks.NewMockUserRepository(gomock.NewController(t))
    uc := usecase.NewUserUsecase(mockRepo)

    // Act
    user, err := uc.CreateUser(context.Background(), dto.CreateUserRequest{
        Email:    "test@example.com",
        Password: "password123",
    })

    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, user)
}
```

---

## ⚙️ Configuration

### Environment Variables

All configuration can be set via environment variables with the `APP_` prefix:

| Variable | Description | Default |
|----------|-------------|---------|
| `APP_APP_NAME` | Application name | `zercle-go-template` |
| `APP_APP_ENVIRONMENT` | Environment (development/staging/production) | `development` |
| `APP_SERVER_HOST` | Server host address | `0.0.0.0` |
| `APP_SERVER_PORT` | Server port | `8080` |
| `APP_DATABASE_HOST` | Database host | `localhost` |
| `APP_DATABASE_PORT` | Database port | `5432` |
| `APP_DATABASE_DATABASE` | Database name | `zercle_template` |
| `APP_DATABASE_USERNAME` | Database username | `postgres` |
| `APP_DATABASE_PASSWORD` | Database password | `postgres` |
| `APP_DATABASE_SSL_MODE` | SSL mode (disable/require/verify-ca/verify-full) | `disable` |
| `APP_JWT_SECRET` | JWT signing secret | *(required)* |
| `APP_JWT_ACCESS_TOKEN_TTL` | Access token TTL | `15m` |
| `APP_JWT_REFRESH_TOKEN_TTL` | Refresh token TTL | `168h` |
| `APP_LOG_LEVEL` | Log level (debug/info/warn/error) | `info` |
| `APP_LOG_FORMAT` | Log format (json/console) | `json` |

### Configuration File

Alternatively, use the YAML configuration file at [`configs/config.yaml`](configs/config.yaml):

```yaml
app:
  name: zercle-go-template
  version: 1.0.0
  environment: development

server:
  host: 0.0.0.0
  port: 8080
  read_timeout: 30s
  write_timeout: 30s
```

Environment variables take precedence over the configuration file.

---

## 💻 Development

### Prerequisites

```bash
# Install development tools
make install-tools

# Install pre-commit hooks
make hooks-install
```

### Development Workflow

```bash
# 1. Start the database
docker-compose -f docker-compose.test.yml up -d

# 2. Run migrations
make migrate-up

# 3. Run the application with hot reload (requires air)
make dev

# Or without hot reload
make run
```

### Code Quality

```bash
# Format code
make fmt

# Run linter
make lint

# Run security scan
make security

# Run all checks
make check
```

### Pre-commit Hooks

The project uses [pre-commit](https://pre-commit.com/) to ensure code quality:

```bash
# Install hooks
make hooks-install

# Run hooks manually
make hooks-run

# Update hook versions
make hooks-update
```

Hooks include:
- Trailing whitespace removal
- YAML/JSON validation
- Go formatting (`gofmt`)
- Go linting (`golangci-lint`)
- Dependency checking

### Generating Code

```bash
# Generate mocks
make mock

# Generate sqlc code
make sqlc

# Generate Swagger docs
make swagger
```

---

## 🐳 Deployment

### Docker

```bash
# Build production image
make docker-build

# Run container
make docker-run

# Push to registry
DOCKER_REGISTRY=your-registry.com make docker-push
```

### Kubernetes

Example deployment:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: zercle-go-template
spec:
  replicas: 3
  selector:
    matchLabels:
      app: zercle-go-template
  template:
    metadata:
      labels:
        app: zercle-go-template
    spec:
      containers:
        - name: api
          image: zercle-go-template:latest
          ports:
            - containerPort: 8080
          env:
            - name: APP_ENVIRONMENT
              value: "production"
            - name: APP_JWT_SECRET
              valueFrom:
                secretKeyRef:
                  name: jwt-secret
                  key: secret
```

### Environment-Specific Builds

```bash
# Development
APP_ENVIRONMENT=development make build

# Production (optimized)
APP_ENVIRONMENT=production make build

# Cross-compile for Linux
make build-linux
```

---

## 🤝 Contributing

We welcome contributions! Please follow these guidelines:

1. **Fork the repository** and create your branch from `main`
2. **Run tests** and ensure they pass: `make test`
3. **Run linting** and fix any issues: `make lint`
4. **Update documentation** if needed
5. **Submit a pull request** with a clear description

### Commit Message Convention

We follow [Conventional Commits](https://www.conventionalcommits.org/):

```
feat: add user authentication
fix: resolve database connection leak
docs: update API documentation
test: add integration tests for user handler
refactor: simplify error handling
```

### Code Style

- Follow [Effective Go](https://go.dev/doc/effective_go) guidelines
- Use `gofmt` for formatting
- Keep functions focused and under 50 lines when possible
- Add documentation comments for exported functions
- Write tests for new features

---

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## 🙏 Acknowledgments

- [Echo Framework](https://echo.labstack.com/) for the excellent web framework
- [sqlc](https://sqlc.dev/) for type-safe SQL
- [Zerolog](https://github.com/rs/zerolog) for blazing-fast logging
- [golangci-lint](https://golangci-lint.run/) for comprehensive linting

---

<p align="center">
  Built with ❤️ by <a href="https://github.com/zercle">Zercle</a>
</p>
