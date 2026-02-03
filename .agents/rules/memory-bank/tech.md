# Zercle Go Template - Technology Stack

## Core Technologies

### Language & Runtime

| Technology | Version | Purpose |
|------------|---------|---------|
| Go | 1.24+ | Primary language |
| Module System | Go Modules | Dependency management |

**Go Features Used:**
- Generics (where appropriate)
- Context for cancellation
- Interfaces for abstraction
- Struct embedding for composition

### Web Framework

**Echo v4** ([`labstack/echo`](https://github.com/labstack/echo))

```go
e := echo.New()
e.Use(middleware.Recover())
e.Use(middleware.Logger())
e.GET("/users/:id", userHandler.GetUser)
```

**Features used:**
- Routing with path parameters
- Middleware chain
- Request binding and validation
- Custom error handling
- Static file serving (Swagger UI)

### Database

**PostgreSQL 13+** with **pgx/v5**

| Component | Package | Purpose |
|-----------|---------|---------|
| Driver | `github.com/jackc/pgx/v5/pgxpool` | Connection pooling |
| SQL Builder | `sqlc` | Type-safe query generation |
| Migrations | `golang-migrate/migrate` | Schema versioning |

**Why pgx over lib/pq:**
- Better performance
- Native Go (no CGO)
- Connection pool management
- Support for advanced PostgreSQL features

### Authentication

**JWT** with `github.com/golang-jwt/jwt/v5`

```go
// Token structure
type Claims struct {
    UserID string `json:"user_id"`
    jwt.RegisteredClaims
}
```

**Token Strategy:**
- Access tokens: 15 minutes, in-memory only
- Refresh tokens: 7 days, stored in httpOnly cookies

### Configuration

**Viper** (`github.com/spf13/viper`)

Configuration precedence (highest to lowest):
1. Runtime environment variables (`APP_*` prefix)
2. `.env` file
3. YAML config file (`configs/config.yaml`)
4. Default values

```yaml
# Example structure
app:
  name: "zercle-api"
  port: 8080
  env: "development"

database:
  host: "localhost"
  port: 5432
  name: "zercle"

jwt:
  secret: "your-secret-key"
  expiry: 900  # seconds
```

### Logging

**Zerolog** (`github.com/rs/zerolog`)

Structured JSON logging with levels:
- `debug`: Development details
- `info`: General operations
- `warn`: Recoverable issues
- `error`: Failures requiring attention
- `fatal`: Application cannot continue

```go
logger.Info().
    Str("user_id", userID).
    Str("method", c.Request().Method).
    Str("path", c.Path()).
    Msg("user created")
```

### Validation

**go-playground/validator/v10**

```go
type CreateUserRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=8,max=64"`
    Name     string `json:"name" validate:"required,max=100"`
}
```

Built-in validators: `required`, `email`, `min`, `max`, `uuid`, `url`, `datetime`

### Documentation

**Swagger/OpenAPI** with `swaggo/swag`

- Annotations in handler code
- Auto-generated `api/docs/` files
- Interactive UI at `/swagger/index.html`

```go
// @Summary     Create user
// @Description Create a new user account
// @Tags        users
// @Accept      json
// @Produce     json
// @Param       request body dto.CreateUserRequest true "User data"
// @Success     201 {object} dto.UserResponse
// @Failure     400 {object} dto.ErrorResponse
// @Router      /users [post]
```

## Development Tools

### Linting

**golangci-lint** ([`.golangci.yml`](.golangci.yml))

Enabled linters:
- `errcheck`: Unchecked errors
- `gosimple`: Code simplification suggestions
- `govet`: Go vet analysis
- `ineffassign`: Ineffective assignments
- `staticcheck`: Advanced static analysis
- `unused`: Dead code detection

Run: `make lint` or `golangci-lint run`

### Testing

| Tool | Purpose |
|------|---------|
| `testing` (stdlib) | Unit test framework |
| `testify` | Assertions and test suites |
| `go.uber.org/mock` | Mock generation |
| `testcontainers-go` | Integration test DB |

**Test Commands:**
```bash
make test              # All tests
make test-unit         # Unit tests only
make test-integration  # Integration tests only
make test-coverage     # With coverage report
```

### Pre-commit Hooks

**pre-commit** ([`.pre-commit-config.yaml`](.pre-commit-config.yaml))

Runs on every commit:
1. `trailing-whitespace`
2. `end-of-file-fixer`
3. `check-yaml`
4. `golangci-lint`
5. `go-fmt`

### Build Tools

**Makefile** targets:

```makefile
build          # Build binary to bin/
dev            # Run with hot reload (air)
test           # Run all tests
lint           # Run golangci-lint
sqlc           # Generate SQLC code
swag           # Generate Swagger docs
docker-up      # Start with Docker Compose
docker-down    # Stop Docker Compose
migrate-up     # Run migrations
migrate-down   # Rollback migrations
```

## Security Practices

### Password Hashing

**bcrypt** (`golang.org/x/crypto/bcrypt`)

```go
hash, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
// Cost: 10 (adjustable based on hardware)
```

### JWT Security

- Algorithm: HS256 (HMAC-SHA256) or RS256 (RSA)
- Secret: 32+ byte random string
- Expiration: Short-lived access tokens (15 min)
- Validation: `iat`, `exp`, `sub` claims

### Input Validation

- Handler layer: Request struct validation
- Domain layer: Business rule validation
- Repository: Parameterized queries (SQL injection safe)

### Secrets Management

- Development: `.env` file (gitignored)
- Production: Environment variables or secrets manager
- Never commit secrets to repository

## Deployment

### Docker

**Multi-stage Dockerfile:**

1. **Builder stage**: Compile with Go toolchain
2. **Final stage**: Minimal image (distroless or alpine)

```dockerfile
FROM golang:1.24-alpine AS builder
# ... build steps ...

FROM gcr.io/distroless/static:nonroot
COPY --from=builder /app/bin/api /api
ENTRYPOINT ["/api"]
```

### Docker Compose

**Services:**
- `api`: Go application
- `postgres`: PostgreSQL database
- (Future) `redis`: Caching layer
- (Future) `prometheus`: Metrics collection

### Health Checks

```yaml
healthcheck:
  test: ["CMD", "/api", "health"]
  interval: 30s
  timeout: 10s
  retries: 3
```

## Environment Configuration

### Required Environment Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `APP_ENV` | Environment name | `production` |
| `APP_PORT` | HTTP server port | `8080` |
| `DB_HOST` | Database host | `postgres` |
| `DB_PORT` | Database port | `5432` |
| `DB_USER` | Database user | `zercle` |
| `DB_PASSWORD` | Database password | `secret` |
| `DB_NAME` | Database name | `zercle` |
| `JWT_SECRET` | JWT signing key | `random-string-32-chars` |
| `JWT_EXPIRY` | Token expiry (seconds) | `900` |

### Configuration Files

| File | Purpose |
|------|---------|
| `configs/config.yaml` | Default configuration values |
| `.env.example` | Template for environment variables |
| `.env` | Local environment (gitignored) |

## Monitoring & Observability (Future)

| Tool | Purpose |
|------|---------|
| Prometheus | Metrics collection |
| Grafana | Metrics visualization |
| Jaeger/Zipkin | Distributed tracing |
| ELK/Loki | Log aggregation |

## Technology Constraints

### Compatibility

- **Go**: 1.24 minimum (uses modern features)
- **PostgreSQL**: 13+ (for JSONB, CTE support)
- **Docker**: 20.10+ (for BuildKit features)

### Resource Requirements

| Component | Minimum | Recommended |
|-----------|---------|-------------|
| CPU | 0.5 cores | 1+ cores |
| Memory | 256 MB | 512 MB |
| Disk | 100 MB | 1 GB |

## Dependency Update Strategy

1. **Security patches**: Immediate update
2. **Minor versions**: Monthly review
3. **Major versions**: Quarterly evaluation with testing
4. **Go version**: Follow official release schedule (6 months)

## Technology References

- [Echo Documentation](https://echo.labstack.com/)
- [sqlc Documentation](https://docs.sqlc.dev/)
- [pgx Documentation](https://github.com/jackc/pgx)
- [Viper Documentation](https://github.com/spf13/viper)
- [Zerolog Documentation](https://github.com/rs/zerolog)
