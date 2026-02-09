# Technology Documentation: Zercle Go Template

**Last Updated:** 2026-02-08  
**Go Version:** 1.25.7  
**Status:** Production-Ready

---

## Technology Stack

### Core Technologies

| Category | Technology | Version | Purpose |
|----------|------------|---------|---------|
| **Language** | Go | 1.25.7 | Primary programming language |
| **Web Framework** | Echo | v4.15.0 | HTTP router and middleware |
| **Database** | PostgreSQL | 14+ | Primary data store |
| **SQL Generator** | SQLC | v1.28.0 | Type-safe SQL code generation |
| **Driver** | pgx | v5.8.0 | PostgreSQL driver with pooling |

### Authentication & Security

| Technology | Version | Purpose |
|------------|---------|---------|
| JWT | v5.3.1 | Token-based authentication |
| bcrypt | v0.47.0 | Password hashing |
| validator | v10.30.1 | Request validation |

### Configuration & Logging

| Technology | Version | Purpose |
|------------|---------|---------|
| Viper | v1.21.0 | Configuration management |
| Zerolog | v1.34.0 | Structured logging |

### Documentation

| Technology | Version | Purpose |
|------------|---------|---------|
| Swaggo | v1.16.6 | Swagger annotation parser |
| Echo Swagger | v1.4.1 | Swagger UI middleware |

### Testing

| Technology | Version | Purpose |
|------------|---------|---------|
| Testify | v1.11.1 | Test assertions and suites |
| Mock (gomock) | v0.6.0 | Mock generation for tests |

### Utilities

| Technology | Version | Purpose |
|------------|---------|---------|
| UUID | v1.6.0 | Unique identifier generation |

---

## Frameworks and Libraries

### Echo Framework (labstack/echo)

**Why Echo?**
- High performance (zero dynamic memory allocation in routing)
- Extensible middleware framework
- Excellent documentation and community
- Built-in validation support
- Graceful shutdown support

**Key Features Used**:
```go
// Router with parameter support
e.GET("/users/:id", handler.GetUser)

// Middleware chain
e.Use(middleware.Recover())
e.Use(middleware.RequestID())

// Request binding and validation
c.Bind(&req)
c.Validate(&req)

// Group routing
api := e.Group("/api/v1")
```

### SQLC (sqlc-dev/sqlc)

**Why SQLC?**
- Write SQL, get type-safe Go code
- Compile-time query checking
- No runtime reflection overhead
- IDE support for SQL files

**Configuration** ([`sqlc.yaml`](../../../../sqlc.yaml)):
```yaml
version: "2"
sql:
  - schema: "internal/infrastructure/db/migrations"
    queries: "internal/infrastructure/db/queries"
    engine: "postgresql"
    gen:
      go:
        package: "sqlc"
        out: "internal/infrastructure/db/sqlc"
        sql_package: "pgx/v5"
```

**Example**:
```sql
-- queries/users.sql
-- name: CreateUser :one
INSERT INTO users (id, email, password_hash, name, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;
```

Generates:
```go
func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (User, error)
```

### pgx (jackc/pgx)

**Why pgx over lib/pq?**
- Active maintenance
- Connection pooling built-in
- Better performance
- Support for advanced PostgreSQL features
- Compatible with database/sql

**Connection Pooling**:
```go
config, _ := pgxpool.ParseConfig(connString)
config.MaxConns = 100
config.MinConns = 10
config.MaxConnLifetime = time.Hour
pool, _ := pgxpool.NewWithConfig(ctx, config)
```

### Viper (spf13/viper)

**Configuration Hierarchy**:
1. Environment variables (highest priority)
2. Configuration file
3. Default values (lowest priority)

**Environment Variable Mapping**:
```yaml
# config.yaml
database:
  host: "localhost"
  port: 5432
```

Override with:
```bash
export APP_DATABASE_HOST=prod-db.example.com
export APP_DATABASE_PORT=5432
```

### Zerolog (rs/zerolog)

**Structured Logging**:
```go
log.Info().
    Str("method", "POST").
    Str("path", "/api/v1/users").
    Int("status", 201).
    Dur("latency", time.Since(start)).
    Msg("request completed")
```

Output:
```json
{"level":"info","method":"POST","path":"/api/v1/users","status":201,"latency":45,"time":"2026-02-08T18:30:00Z"}
```

---

## Development Tools

### Required Tools

| Tool | Purpose | Installation |
|------|---------|--------------|
| **Go** | Language runtime | https://golang.org/dl/ |
| **Docker** | Containerization | https://docs.docker.com/get-docker/ |
| **Make** | Build automation | `brew install make` (macOS) |
| **golangci-lint** | Linting | `brew install golangci-lint` |
| **pre-commit** | Git hooks | `pip install pre-commit` |

### Optional Tools

| Tool | Purpose | Installation |
|------|---------|--------------|
| **Air** | Hot reload | `go install github.com/air-verse/air@latest` |
| **swag** | Swagger generation | `go install github.com/swaggo/swag/cmd/swag@latest` |
| **mockgen** | Mock generation | `go install go.uber.org/mock/mockgen@latest` |
| **sqlc** | SQL code gen | `go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest` |
| **migrate** | DB migrations | `go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest` |
| **gosec** | Security scanner | `go install github.com/securego/gosec/v2/cmd/gosec@latest` |

### Installation Script

```bash
# Install all development tools
make install-tools

# Or manually:
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install github.com/swaggo/swag/cmd/swag@latest
go install github.com/securego/gosec/v2/cmd/gosec@latest
go install go.uber.org/mock/mockgen@latest
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

---

## Setup Instructions

### Quick Start

```bash
# 1. Clone the repository
git clone https://github.com/zercle/zercle-go-template.git
cd zercle-go-template

# 2. Install dependencies
make deps

# 3. Set up configuration
cp configs/config.yaml configs/config.local.yaml
# Edit configs/config.local.yaml with your settings

# 4. Start PostgreSQL (using Docker)
docker run -d \
  --name postgres \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=zercle_template \
  -p 5432:5432 \
  postgres:14-alpine

# 5. Run migrations
make migrate DB_USER=postgres DB_PASSWORD=postgres DB_HOST=localhost DB_PORT=5432 DB_NAME=zercle_template DB_SSLMODE=disable

# 6. Run the application
make run

# 7. Open Swagger UI
open http://localhost:8080/swagger/index.html
```

### Configuration

**Environment Variables**:

| Variable | Description | Default |
|----------|-------------|---------|
| `APP_APP_NAME` | Application name | zercle-go-template |
| `APP_APP_VERSION` | Application version | 1.0.0 |
| `APP_APP_ENVIRONMENT` | Environment (development/staging/production) | development |
| `APP_SERVER_HOST` | Server bind address | 0.0.0.0 |
| `APP_SERVER_PORT` | Server port | 8080 |
| `APP_DATABASE_HOST` | Database host | localhost |
| `APP_DATABASE_PORT` | Database port | 5432 |
| `APP_DATABASE_DATABASE` | Database name | zercle_template |
| `APP_DATABASE_USERNAME` | Database user | postgres |
| `APP_DATABASE_PASSWORD` | Database password | (empty) |
| `APP_DATABASE_SSL_MODE` | SSL mode (disable/require/verify-ca/verify-full) | disable |
| `APP_LOG_LEVEL` | Log level (debug/info/warn/error) | info |
| `APP_LOG_FORMAT` | Log format (json/console) | json |

---

## Testing Approach

### Test Pyramid

```
       /\
      /  \
     / E2E \          (Future: API integration tests)
    /--------\
   /          \
  / Integration \    (Repository tests with test DB)
 /--------------\
/                \
/     Unit         \  (Usecase tests with mocks)
/--------------------\
```

### Running Tests

```bash
# Run all tests
make test

# Run only unit tests
make test-unit

# Run integration tests (requires test database)
make test-integration

# Generate coverage report
make test-coverage

# View coverage in browser
make test-coverage-html

# Run benchmarks
make benchmark
```

### Test Database

Integration tests use Docker Compose to spin up a temporary PostgreSQL instance:

```bash
# docker-compose.test.yml
version: '3.8'
services:
  postgres_test:
    image: postgres:14-alpine
    environment:
      POSTGRES_USER: test
      POSTGRES_PASSWORD: test
      POSTGRES_DB: test_db
    ports:
      - "5433:5432"
```

### Mock Generation

```bash
# Generate all mocks
make mock

# Clean and regenerate
make mock-clean mock

# Verify mocks are up to date
make mock-verify
```

---

## Security Practices

### Authentication

- **JWT Tokens**: HS256 signed, configurable secret
- **Access Token TTL**: 15 minutes
- **Refresh Token TTL**: 7 days
- **Password Hashing**: bcrypt with cost 12

### Input Validation

```go
type CreateUserRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=8,max=72"`
    Name     string `json:"name" validate:"required,min=2,max=100"`
}
```

### SQL Injection Prevention

SQLC generates parameterized queries automatically:

```go
// Safe - uses parameterization
const createUser = `-- name: CreateUser :one
INSERT INTO users (email, password_hash)
VALUES ($1, $2)
RETURNING id, email`
```

### Security Scanning

```bash
# Run security scanner
make security

# Or directly
gosec ./...
```

### Security Headers (Future)

- `X-Content-Type-Options: nosniff`
- `X-Frame-Options: DENY`
- `X-XSS-Protection: 1; mode=block`
- `Strict-Transport-Security: max-age=31536000`

---

## Deployment Information

### Docker Build

```bash
# Build production image
make docker-build

# Build with specific tag
make docker-build DOCKER_TAG=v1.0.0

# Run container
make docker-run
```

### Multi-Stage Dockerfile

```dockerfile
# Stage 1: Builder
golang:1.25-bookworm
- Compile with optimizations
- Static binary output

# Stage 2: Runtime
gcr.io/distroless/static:nonroot
- Minimal attack surface
- Non-root user (UID 65532)
- ~20MB final image
```

### Environment-Specific Configuration

```
configs/
├── config.yaml              # Default configuration
├── config.development.yaml  # Development overrides
├── config.staging.yaml      # Staging overrides
└── config.production.yaml   # Production overrides
```

### Health Checks

```bash
# Liveness probe
GET /health

Response:
{
  "success": true,
  "data": {
    "status": "healthy",
    "timestamp": "2026-02-08T18:30:00Z"
  }
}
```

### Graceful Shutdown

```go
// Wait for interrupt signal
quit := make(chan os.Signal, 1)
signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
<-quit

// Graceful shutdown with timeout
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
e.Shutdown(ctx)
```

---

## Performance Characteristics

### Benchmarks

| Metric | Target | Notes |
|--------|--------|-------|
| Cold Start | < 100ms | Application startup time |
| Request Latency (P99) | < 50ms | Simple CRUD operations |
| Memory Usage | < 50MB | Baseline at idle |
| Build Time | < 2 min | Full CI pipeline |
| Docker Image | < 20MB | Compressed size |
| Test Suite | < 30s | Full test run |

### Optimization Strategies

1. **Connection Pooling**: pgx handles this automatically
2. **Prepared Statements**: SQLC generates these
3. **JSON Pooling**: Echo uses `sync.Pool` for buffers
4. **Zero-Allocation Logging**: Zerolog design
5. **Static Binary**: CGO disabled for faster startup

---

## Troubleshooting

### Common Issues

**Port Already in Use**:
```bash
lsof -ti:8080 | xargs kill -9
```

**Database Connection Failed**:
```bash
# Check PostgreSQL is running
docker ps | grep postgres

# Test connection
psql postgres://postgres:postgres@localhost:5432/zercle_template
```

**Linting Errors**:
```bash
# Auto-fix issues
make fmt

# Run linter with details
golangci-lint run --verbose
```

**Test Failures**:
```bash
# Run with verbose output
go test -v ./...

# Run specific test
go test -v -run TestCreateUser ./internal/feature/user/usecase/
```

---

**Related Documents**:
- [brief.md](brief.md) - Project overview
- [architecture.md](architecture.md) - System architecture
- [tasks.md](tasks.md) - Development workflows
