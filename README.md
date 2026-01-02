# Zercle Go Template

A production-ready Go web application template using the Echo framework, designed with Domain-Driven Design (DDD) principles and type-safe database operations using sqlc.

## 📋 Table of Contents

- [Features](#-features)
- [Project Structure](#-project-structure)
- [Prerequisites](#-prerequisites)
- [Installation](#-installation)
- [Configuration](#-configuration)
- [Running the Application](#-running-the-application)
- [Database Migrations](#-database-migrations)
- [Testing](#-testing)
- [API Documentation](#-api-documentation)
- [Development Workflow](#-development-workflow)
- [Deployment](#-deployment)
- [Contributing](#-contributing)
- [License](#-license)

## ✨ Features

- **Echo Framework** - High performance, minimalist Go web framework
- **Domain-Driven Design** - Clean architecture with clear separation of concerns
- **sqlc** - Type-safe SQL queries with compile-time validation
- **PostgreSQL** - Robust relational database with migrations
- **Docker Support** - Containerized development and production deployment
- **Multiple Environments** - dev, local, prod, uat configurations
- **Comprehensive Testing** - Unit, integration, and mock tests
- **Swagger API Docs** - Auto-generated API documentation
- **Code Quality** - golangci-lint integration
- **Dependency Injection** - Clean component wiring

## 🏗️ Project Structure

```
zercle-go-template/
├── cmd/
│   └── server/                 # Application entry point
│       └── main.go
├── configs/                     # Environment configurations
│   ├── dev.yaml
│   ├── local.yaml
│   ├── prod.yaml
│   └── uat.yaml
├── deployments/
│   └── docker/                  # Docker configuration
│       ├── Dockerfile
│       └── docker-compose.yml
├── docs/                        # Swagger documentation
│   ├── docs.go
│   ├── swagger.json
│   └── swagger.yaml
├── internal/
│   ├── app/                     # Application setup
│   │   └── app.go
│   ├── domain/                  # Domain layer (DDD)
│   │   ├── task/                # Task domain module
│   │   │   ├── entity/
│   │   │   ├── handler/
│   │   │   ├── mock/
│   │   │   ├── repository/
│   │   │   ├── request/
│   │   │   ├── response/
│   │   │   ├── usecase/
│   │   │   └── interface.go
│   │   └── user/                # User domain module
│   │       ├── entity/
│   │       ├── handler/
│   │       ├── mock/
│   │       ├── repository/
│   │       ├── request/
│   │       ├── response/
│   │       ├── usecase/
│   │       └── interface.go
│   └── infrastructure/          # Infrastructure layer
│       ├── config/
│       ├── db/
│       ├── http/
│       ├── logger/
│       ├── password/
│       └── sqlc/
├── scripts/                      # Utility scripts
│   ├── run-dev.sh
│   └── seed-db.sh
├── sqlc/
│   ├── migrations/              # Database migrations
│   └── queries/                 # SQL queries for sqlc
├── test/                        # Test suite
│   ├── integration/
│   ├── mock/
│   └── unit/
├── .env.example
├── .golangci.yml
├── Makefile
└── sqlc.yaml
```

### Layer Explanation

| Layer | Purpose |
|-------|---------|
| **cmd** | Application entry points and CLI commands |
| **domain** | Business logic, entities, use cases, and domain interfaces |
| **infrastructure** | External concerns: database, HTTP, logging, configuration |
| **test** | Unit, integration, and mock test implementations |

## 🔧 Prerequisites

- **Go** 1.21 or higher
- **Docker** and **Docker Compose**
- **PostgreSQL** 15+ (for local development without Docker)
- **golangci-lint** (optional, for code quality checks)
- **sqlc** (optional, for code generation)

## 📦 Installation

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd zercle-go-template
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Generate sqlc code**
   ```bash
   make generate
   ```

4. **Configure environment**
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

## ⚙️ Configuration

### Environment Variables

Copy the example environment file and modify according to your needs:

```bash
cp .env.example .env
```

### Configuration Files

The application uses YAML configuration files for different environments:

| File | Environment | Description |
|------|-------------|-------------|
| `configs/dev.yaml` | Development | Development settings with verbose logging |
| `configs/local.yaml` | Local | Local machine settings |
| `configs/uat.yaml` | UAT | User acceptance testing environment |
| `configs/prod.yaml` | Production | Production-ready settings |

### Key Configuration Options

```yaml
# Example config structure
app:
  host: "0.0.0.0"
  port: 8080
  env: "development"

database:
  host: "localhost"
  port: 5432
  username: "postgres"
  password: "password"
  name: "app_db"
  sslmode: "disable"
  max_open_conns: 25
  max_idle_conns: 5
  conn_max_lifetime: 5m

logging:
  level: "debug"
  format: "json"
```

## 🚀 Running the Application

### Local Development

Using the provided script:
```bash
./scripts/run-dev.sh
```

Or with Makefile:
```bash
make run
```

### Using Docker Compose

```bash
# Start all services including database
docker-compose -f deployments/docker/docker-compose.yml up -d

# View logs
docker-compose -f deployments/docker/docker-compose.yml logs -f

# Stop services
docker-compose -f deployments/docker/docker-compose.yml down
```

### Environment-Specific Runs

```bash
# Run with local configuration
make run-local

# Run with development configuration
make run-dev

# Run with UAT configuration
make run-uat

# Run with production configuration
make run-prod
```

## 🗃️ Database Migrations

### Running Migrations

Using Docker:
```bash
docker-compose -f deployments/docker/docker-compose.yml exec app go run ./cmd/server/main.go migrate up
```

Using Makefile:
```bash
make migrate-up
make migrate-down
```

### Migration Commands

| Command | Description |
|---------|-------------|
| `make migrate-up` | Apply all pending migrations |
| `make migrate-down` | Rollback all migrations |
| `make migrate-create NAME=my_migration` | Create a new migration file |

### Seed Database

```bash
./scripts/seed-db.sh
```

Or:
```bash
make seed
```

## 🧪 Testing

### Run All Tests

```bash
make test
```

### Test Categories

```bash
# Unit tests
make test-unit

# Integration tests
make test-integration

# All tests with coverage
make test-coverage
```

### Test Configuration

- **Unit tests** - Located in `test/unit/` and domain-specific directories
- **Integration tests** - Located in `test/integration/`
- **Mock tests** - Located in `test/mock/` with sqlmock support

### Coverage Report

```bash
make test-coverage
# Open coverage.html for detailed report
```

## 📖 API Documentation

### Swagger UI

Once the application is running, access Swagger documentation at:

```
http://localhost:8080/swagger/index.html
```

### Regenerate Swagger Docs

```bash
make swagger
```

### Endpoints

The API provides RESTful endpoints for User and Task management:

#### User Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | /api/v1/users | List all users |
| GET | /api/v1/users/:id | Get user by ID |
| POST | /api/v1/users | Create new user |
| PUT | /api/v1/users/:id | Update user |
| DELETE | /api/v1/users/:id | Delete user |

#### Task Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | /api/v1/tasks | List all tasks |
| GET | /api/v1/tasks/:id | Get task by ID |
| POST | /api/v1/tasks | Create new task |
| PUT | /api/v1/tasks/:id | Update task |
| DELETE | /api/v1/tasks/:id | Delete task |

## 🔨 Development Workflow

### Makefile Commands

| Command | Description |
|---------|-------------|
| `make run` | Run the application |
| `make build` | Build the application |
| `make test` | Run all tests |
| `make lint` | Run golangci-lint |
| `make format` | Format Go code |
| `make generate` | Generate sqlc code |
| `make swagger` | Generate Swagger docs |
| `make clean` | Clean build artifacts |
| `make docker-build` | Build Docker image |
| `make docker-run` | Run Docker container |

### Code Quality

```bash
# Run linter
make lint

# Format code
make format

# Check for issues
make vet
```

### Git Hooks

Configure pre-commit hooks for code quality checks (add to `.pre-commit-config.yaml`):

```yaml
repos:
  - repo: local
    hooks:
      - id: lint
        name: Run golangci-lint
        entry: make lint
        language: system
        pass_filenames: false
```

## 🚢 Deployment

### Docker Deployment

#### Build Image

```bash
make docker-build
```

#### Run with Docker Compose (Production)

```bash
docker-compose -f deployments/docker/docker-compose.yml up -d --build
```

#### Environment Variables for Production

```yaml
# Override in production environment
app:
  env: "production"
  host: "0.0.0.0"
  port: 8080

database:
  host: "postgres"
  port: 5432
  sslmode: "require"
```

### Kubernetes Deployment

1. Build and push Docker image:
   ```bash
   docker build -t your-registry/zercle-go-template:latest .
   docker push your-registry/zercle-go-template:latest
   ```

2. Deploy with kubectl:
   ```bash
   kubectl apply -f deployments/k8s/
   ```

### Health Checks

The application exposes health check endpoints:

- `/health` - Basic health check
- `/ready` - Readiness probe
- `/live` - Liveness probe

## 🤝 Contributing

### Getting Started

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Commit changes: `git commit -m 'Add amazing feature'`
4. Push to branch: `git push origin feature/amazing-feature`
5. Open a Pull Request

### Code Standards

- Follow Go formatting conventions (`go fmt`)
- Run linter before committing (`make lint`)
- Write tests for new functionality
- Update documentation as needed
- Use conventional commits

### Pull Request Process

1. Ensure all tests pass (`make test`)
2. Verify code quality checks pass (`make lint`)
3. Update README.md if needed
4. Request review from maintainers

## 📄 License

This project is licensed under the MIT License - see the [LICENSE.md](LICENSE.md) file for details.

## 🙏 Acknowledgments

- [Echo Framework](https://echo.labstack.com/) - Web framework
- [sqlc](https://sqlc.dev/) - Type-safe SQL
- [golangci-lint](https://golangci-lint.run/) - Code linter
- [Swaggo](https://github.com/swaggo/swag) - Swagger documentation
