# Technology Stack

**Last Updated:** 2026-02-10

## Core Technologies

### Language & Runtime
- **Go:** 1.25.7
- **Module:** zercle-go-template

### Web Framework
- **Echo v4:** 4.15.0
  - High-performance HTTP router
  - Middleware support
  - Built-in request/response handling
  - Graceful shutdown

### Database
- **PostgreSQL:** Primary data store
- **pgx v5:** 5.8.0
  - Pure Go PostgreSQL driver
  - Connection pooling (pgxpool)
  - Context support
  - Performance optimized
- **sqlc:** Type-safe SQL query generation
  - Generates Go code from SQL
  - Compile-time query validation
  - Reduces boilerplate

### Authentication
- **JWT:** golang-jwt/jwt/v5 v5.3.1
  - Access tokens (15 min default)
  - Refresh tokens (7 days default)
  - HMAC-SHA256 signing
- **Argon2id:** golang.org/x/crypto v0.48.0
  - OWASP-recommended password hashing
  - Configurable parameters
  - Constant-time comparison

### Configuration
- **Viper:** 1.21.0
  - Multi-source configuration
  - Environment variables
  - .env file support
  - YAML configuration
  - Type-safe unmarshaling

### Logging
- **Zerolog:** 1.34.0
  - Structured logging
  - Zero-allocation JSON logging
  - Context-aware logging
  - Multiple output formats

## Development Tools

### Code Generation
- **Mockgen:** go.uber.org/mock v0.6.0
  - Interface mocking
  - Used with go:generate directives
  - Generates mocks in `mocks/` directories

### API Documentation
- **Swagger:** github.com/swaggo/swag v1.16.6
  - OpenAPI 3.0 specification
  - Auto-generated docs from annotations
  - Available at `/swagger/*`
- **Echo-Swagger:** github.com/swaggo/echo-swagger v1.4.1
  - Echo integration for Swagger UI

### Validation
- **go-playground/validator:** v10.30.1
  - Struct validation
  - Custom validators
  - Detailed error messages

### Testing
- **Testify:** github.com/stretchr/testify v1.11.1
  - Assertions
  - Test suites
  - Mocking utilities
- **Go Testing:** Built-in testing framework
  - Unit tests
  - Integration tests
  - Benchmark tests

### Linting & Code Quality
- **golangci-lint:** Comprehensive linting
  - Configured in `.golangci.yml`
  - Multiple linters enabled
  - CI/CD integration
- **Pre-commit hooks:** .pre-commit-config.yaml
  - Automated checks before commit
  - Formatting validation
  - Linting enforcement

## Dependencies Summary

### Core Dependencies
```
github.com/labstack/echo/v4 v4.15.0
github.com/jackc/pgx/v5 v5.8.0
github.com/golang-jwt/jwt/v5 v5.3.1
github.com/spf13/viper v1.21.0
github.com/rs/zerolog v1.34.0
github.com/google/uuid v1.6.0
golang.org/x/crypto v0.48.0
```

### Development Dependencies
```
github.com/stretchr/testify v1.11.1
github.com/swaggo/swag v1.16.6
github.com/swaggo/echo-swagger v1.4.1
go.uber.org/mock v0.6.0
github.com/go-playground/validator/v10 v10.30.1
```

## Configuration Management

### Configuration Sources (Priority Order)
1. Runtime environment variables (highest)
2. .env file
3. YAML config file (`configs/config.yaml`)
4. Default values (lowest)

### Environment Variables Prefix
- Prefix: `APP_`
- Example: `APP_SERVER_PORT=8080`

### Configuration Structure
```yaml
app:
  name: string
  version: string
  environment: string

server:
  host: string
  port: int
  read_timeout: duration
  write_timeout: duration
  shutdown_timeout: duration

database:
  host: string
  port: int
  database: string
  username: string
  password: string
  ssl_mode: string

log:
  level: string (debug|info|warn|error)
  format: string (json|console)

jwt:
  secret: string
  access_token_ttl: duration
  refresh_token_ttl: duration

security:
  argon2_memory: int (KB)
  argon2_iterations: int
  argon2_parallelism: int
  argon2_salt_length: int
  argon2_key_length: int
```

## Deployment Configuration

### Docker
- **Dockerfile:** Multi-stage build
- **Docker Compose:** Test environment setup
- **.dockerignore:** Exclude unnecessary files

### Database Migrations
- **Location:** `internal/infrastructure/db/migrations/`
- **Format:** SQL up/down migrations
- **Naming:** `001_initial_schema.up.sql`

### Connection Pool Settings
- Max connections: 25
- Min connections: 5
- Max connection lifetime: 1 hour
- Max connection idle time: 30 minutes
- Health check period: 5 minutes

## Security Configuration

### Argon2id Defaults
- **Production:** Memory=64MB, Iterations=3, Parallelism=4, Salt=16B, Key=32B
- **Development:** Memory=16MB, Iterations=3, Parallelism=4, Salt=16B, Key=32B

### JWT Configuration
- Signing method: HMAC-SHA256
- Access token TTL: 15 minutes (configurable)
- Refresh token TTL: 7 days (configurable)
- Issuer: zercle-go-template

## Testing Strategy

### Test Types
1. **Unit Tests:** Isolated component testing
2. **Integration Tests:** Database interaction testing
3. **Benchmark Tests:** Performance measurement
4. **Mock Tests:** Interface contract verification

### Test Coverage Goal
- 80%+ for critical paths
- All business logic covered
- All handlers covered

### Test Naming Convention
- Unit: `*_test.go`
- Integration: `*_integration_test.go`
- Benchmark: `*_benchmark_test.go`
- Mocks: `mocks/*.go`

## Code Standards

### Go Conventions
- Follow Effective Go guidelines
- Use gofmt for formatting
- Use goimports for import management
- Follow SOLID principles
- Keep functions under 50 lines

### Naming Conventions
- Interfaces: Simple, descriptive (e.g., `UserRepository`)
- Implementations: Specific (e.g., `SqlcUserRepository`)
- Packages: Lowercase, single word when possible
- Exports: PascalCase
- Privates: camelCase

### Error Handling
- Wrap errors with context
- Use typed errors where appropriate
- Log all errors with context
- Provide user-friendly messages

## Performance Considerations

### Optimization Strategies
- Connection pooling for database
- Lazy initialization where appropriate
- Zero-copy patterns where possible
- Efficient data structures
- Proper indexing in database

### Monitoring Points
- Request latency
- Database query performance
- Connection pool statistics
- Error rates
- Token generation performance
