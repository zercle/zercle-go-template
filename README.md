# Zercle Go Template

A production-ready RESTful API template built with Go Echo framework, featuring clean architecture, JWT authentication, and PostgreSQL database. This template provides a solid foundation for building Go microservices or REST APIs with best practices already implemented.

## Features

- **Clean Architecture** - Domain-driven design with clear separation of concerns
- **Type-safe Database Operations** - SQLC for compile-time safe SQL queries
- **JWT Authentication** - Stateless authentication with configurable expiration
- **Password Security** - Argon2id hashing for secure password storage
- **Comprehensive Testing** - Unit, integration, and mock testing infrastructure
- **API Documentation** - Swagger/OpenAPI documentation out of the box
- **Structured Logging** - Zerolog for zero-allocation JSON logging
- **Docker Support** - Containerized deployment with Docker Compose
- **Rate Limiting** - Configurable request rate limiting
- **CORS Support** - Configurable cross-origin resource sharing
- **Health Checks** - Application and readiness endpoints

## Tech Stack

- **Language**: Go 1.24.0+
- **Web Framework**: Echo v4
- **Database**: PostgreSQL 12+ with pgx/v5 driver
- **ORM/Query Builder**: SQLC (type-safe SQL generation)
- **Authentication**: JWT (golang-jwt/jwt/v5)
- **Password Hashing**: Argon2id (golang.org/x/crypto)
- **Configuration**: Viper
- **Logging**: Zerolog
- **Validation**: go-playground/validator/v10
- **Documentation**: Swaggo (Swagger)
- **Testing**: testify, go.uber.org/mock, testcontainers

## Project Structure

```
.
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/
│   ├── app/
│   │   └── app.go               # Application orchestration & DI
│   ├── domain/
│   │   ├── user/                # User domain (example)
│   │   │   ├── entity/          # Business entities
│   │   │   ├── handler/         # HTTP handlers
│   │   │   ├── repository/      # Data access layer
│   │   │   ├── usecase/         # Business logic
│   │   │   ├── request/         # Request DTOs
│   │   │   ├── response/        # Response DTOs
│   │   │   ├── mock/            # Mock implementations
│   │   │   └── interface.go     # Domain interfaces
│   │   └── task/                # Task domain (example)
│   └── infrastructure/
│       ├── config/              # Configuration management
│       ├── db/                  # Database abstraction
│       ├── http/                # HTTP client
│       ├── logger/              # Structured logging
│       ├── password/            # Password hashing
│       └── sqlc/db/             # SQLC-generated code
├── sqlc/
│   ├── migrations/              # Database migrations
│   └── queries/                 # SQL query files
├── configs/
│   ├── local.yaml               # Local development config
│   ├── dev.yaml                 # Development config
│   ├── uat.yaml                 # UAT config
│   └── prod.yaml                # Production config
├── test/
│   ├── integration/             # Integration tests
│   └── mock/                    # Mock utilities
├── scripts/
│   ├── run-dev.sh               # Development runner
│   └── seed-db.sh               # Database seeding
├── deployments/
│   └── docker/
│       ├── Dockerfile           # Docker image
│       └── docker-compose.yml   # Docker Compose setup
├── docs/                        # Swagger documentation
├── .env.example                 # Environment variables template
├── go.mod                       # Go module definition
├── go.sum                       # Go dependencies
├── Makefile                     # Common operations
├── sqlc.yaml                    # SQLC configuration
└── .golangci.yml                # Linting configuration
```

## Architecture

This template follows **Clean Architecture** with **Domain-Driven Design (DDD)** principles:

### Layers

1. **Domain Layer** (`internal/domain/`) - Core business logic and entities, independent of infrastructure
2. **Infrastructure Layer** (`internal/infrastructure/`) - External concerns and technical implementations
3. **Application Layer** (`internal/app/`) - Application orchestration and dependency injection
4. **Entry Point** (`cmd/server/`) - Application bootstrap

### Data Flow

```
Client → Handler → UseCase → Repository → Database
         ↓         ↓          ↓
     Request   Business    Data Access
     DTO       Logic       Layer
```

## Getting Started

### Prerequisites

- Go 1.24.0 or higher
- PostgreSQL 12+
- Docker (optional, for containerized deployment)

### Installation

1. Clone the repository:
```bash
git clone https://github.com/zercle/zercle-go-template.git
cd zercle-go-template
```

2. Copy environment variables:
```bash
cp .env.example .env
```

3. Install dependencies:
```bash
go mod download
```

4. Configure database connection in `.env` or `configs/local.yaml`

5. Run database migrations:
```bash
# Using migration tool (to be added)
migrate -path sqlc/migrations -database "postgres://user:pass@localhost:5432/dbname?sslmode=disable" up
```

6. Generate SQLC code:
```bash
sqlc generate
```

7. Generate Swagger documentation:
```bash
swag init -g cmd/server/main.go
```

### Running the Application

#### Development

```bash
# Set environment
export SERVER_ENV=local

# Run the application
go run cmd/server/main.go
```

Or use the provided script:
```bash
./scripts/run-dev.sh
```

#### Production

```bash
# Build the binary
go build -o bin/server cmd/server/main.go

# Run the binary
./bin/server
```

#### Docker

```bash
# Build Docker image
docker build -t zercle-go-template .

# Run container
docker run -p 3000:3000 \
  -e SERVER_ENV=prod \
  -e DATABASE_URL=postgres://user:pass@host:5432/dbname \
  zercle-go-template
```

#### Docker Compose

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop services
docker-compose down
```

## Configuration

Configuration is managed through YAML files in the `configs/` directory:

- `local.yaml` - Local development
- `dev.yaml` - Development environment
- `uat.yaml` - User acceptance testing
- `prod.yaml` - Production

Environment variables can override configuration values. Set `SERVER_ENV` to select the configuration file.

### Key Configuration Sections

- **Database**: Connection string, pool settings
- **JWT**: Secret key, expiration time
- **Server**: Port, timeout settings
- **Logging**: Level, format, output
- **CORS**: Allowed origins, methods, headers
- **Rate Limiting**: Requests per window, window duration

## API Documentation

Once the application is running, access the Swagger documentation at:

```
http://localhost:3000/swagger/index.html
```

### Available Endpoints

#### Health Checks
- `GET /health` - Application health check
- `GET /readiness` - Readiness probe

#### Authentication
- `POST /api/v1/auth/register` - User registration
- `POST /api/v1/auth/login` - User login

#### User Management
- `GET /api/v1/users` - List users (paginated)
- `GET /api/v1/users/:id` - Get user profile
- `PUT /api/v1/users/:id` - Update user profile (protected)
- `DELETE /api/v1/users/:id` - Delete account (protected)

#### Task Management (Example Domain)
- `POST /api/v1/tasks` - Create task (protected)
- `GET /api/v1/tasks` - List tasks (protected, paginated)
- `GET /api/v1/tasks/:id` - Get task (protected)
- `PUT /api/v1/tasks/:id` - Update task (protected)
- `DELETE /api/v1/tasks/:id` - Delete task (protected)

## Testing

### Run All Tests

```bash
go test ./...
```

### Run Tests with Coverage

```bash
go test -cover ./...
```

### Generate Coverage Report

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Run Integration Tests

```bash
go test -tags=integration ./test/integration/
```

### Run Specific Test

```bash
go test -v -run TestLogin ./internal/domain/user/usecase/
```

## Development Guidelines

### Adding a New Domain

1. Create domain structure under `internal/domain/<domain>/`
2. Define entity in `entity/` directory
3. Create interfaces in `interface.go`
4. Implement repository, usecase, and handler
5. Add request/response DTOs
6. Write tests
7. Wire dependencies in `internal/app/app.go`
8. Register routes
9. Update Swagger documentation

### Code Style

- Follow Go standard formatting (`go fmt`)
- Use `golangci-lint` for linting
- Write godoc comments for exported functions
- Keep functions under 50 lines when possible
- Follow SOLID principles

### Testing Standards

- Write unit tests for business logic
- Use table-driven tests for multiple scenarios
- Mock external dependencies
- Aim for >80% coverage on critical paths
- Test error paths, not just happy paths

### Database Migrations

1. Create migration files in `sqlc/migrations/`
2. Format: `YYYYMMDD_NNN_description`
3. Write both up and down migrations
4. Apply migrations in order
5. Regenerate SQLC code: `sqlc generate`

## Common Commands

### Linting

```bash
# Run linter
golangci-lint run

# Fix issues automatically
golangci-lint run --fix
```

### Formatting

```bash
# Format code
go fmt ./...

# Check for issues
go vet ./...
```

### Dependencies

```bash
# Tidy dependencies
go mod tidy

# Update dependencies
go get -u ./...

# Verify dependencies
go mod verify
```

### SQLC

```bash
# Generate SQLC code
sqlc generate

# Validate SQLC configuration
sqlc validate
```

### Documentation

```bash
# Generate Swagger docs
swag init -g cmd/server/main.go
```

## Environment Variables

Key environment variables (see `.env.example`):

- `SERVER_ENV` - Environment (local, dev, uat, prod)
- `DATABASE_URL` - PostgreSQL connection string
- `JWT_SECRET` - JWT signing secret
- `JWT_EXPIRATION` - Token expiration time
- `SERVER_PORT` - Server port (default: 3000)

## Security

- Passwords hashed with Argon2id
- JWT tokens for stateless authentication
- Input validation on all endpoints
- CORS configuration per environment
- Rate limiting to prevent abuse
- SQL injection prevention via SQLC

## Performance

- Database connection pooling
- Efficient query generation via SQLC
- Structured logging with minimal overhead
- Graceful shutdown handling
- Configurable timeouts

## Deployment

### Production Checklist

- [ ] Set strong JWT secret
- [ ] Configure production database
- [ ] Enable HTTPS/TLS
- [ ] Set appropriate CORS origins
- [ ] Configure rate limiting
- [ ] Set log level to INFO or WARN
- [ ] Enable health checks
- [ ] Configure monitoring and alerting
- [ ] Run database migrations
- [ ] Test all endpoints

### Docker Deployment

The provided Dockerfile uses a multi-stage build for optimization:

- Builder stage: Compiles the Go binary
- Runtime stage: Alpine-based minimal image
- Non-root user for security
- Health checks configured

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Pull Request Guidelines

- Follow existing code style
- Add tests for new features
- Update documentation
- Ensure all tests pass
- Run linter and fix issues

## License

This project is licensed under the MIT License - see the [LICENSE.md](LICENSE.md) file for details.

## Support

For issues, questions, or contributions, please visit the GitHub repository.

## Roadmap

Future enhancements planned:

- [ ] Redis caching layer
- [ ] Message queue integration (RabbitMQ/Kafka)
- [ ] Metrics collection (Prometheus)
- [ ] Distributed tracing (OpenTelemetry)
- [ ] API versioning strategy
- [ ] GraphQL support option
- [ ] Additional example domains

## Acknowledgments

Built with best practices and modern Go development tools. Special thanks to the open-source community for the excellent libraries and frameworks used in this project.
