# Technical Standards & Guidelines

## Language & Runtime
- **Go Version:** 1.24.0
- **Module:** github.com/zercle/zercle-go-template

## Core Dependencies

### Web Framework
- **Echo v4** - HTTP server framework
- **Labstack middleware** - Request ID, logger, recovery, CORS

### Database
- **pgx/v5** - PostgreSQL driver
- **SQLC** - Type-safe SQL query generation
- **Testcontainers** - Integration testing with real databases

### Authentication
- **golang-jwt/jwt/v5** - JWT token generation and validation
- **Argon2id** - Password hashing (golang.org/x/crypto)

### Configuration
- **Viper** - Configuration management
- **YAML** - Configuration file format

### Logging
- **Zerolog** - Structured, zero-allocation logging

### Validation
- **go-playground/validator/v10** - Request validation

### Documentation
- **Swaggo** - Swagger/OpenAPI documentation generation

### Testing
- **testify** - Assertions and mocking
- **go.uber.org/mock** - Mock generation
- **testcontainers** - Integration testing

## Coding Standards

### Naming Conventions
- **Files:** lowercase with underscores (e.g., `user_handler.go`)
- **Packages:** lowercase, single word (e.g., `handler`, `usecase`)
- **Interfaces:** Simple names describing capability (e.g., `UserRepository`)
- **Implementations:** Descriptive names (e.g., `userUseCase`, `UserHandler`)
- **Constants:** UPPER_SNAKE_CASE
- **Private variables:** camelCase
- **Public variables:** PascalCase

### Code Organization
- **Package structure:** One responsibility per package
- **File size:** Keep files focused and under 300 lines when possible
- **Function length:** Prefer functions under 50 lines
- **Exported functions:** Must have godoc comments
- **Error handling:** Always handle errors, never ignore

### Design Patterns

**Repository Pattern:**
- Abstract data access behind interfaces
- Domain entities mapped to database models
- Repository implementations in infrastructure layer

**Use Case Pattern:**
- Business logic encapsulated in use cases
- Coordinate between repositories and handlers
- Domain-specific error definitions

**Factory Pattern:**
- Database factory for creating connections
- Configuration-based instantiation

**Middleware Pattern:**
- Request/response processing pipeline
- Cross-cutting concerns (auth, logging, CORS)

### SOLID Principles

**Single Responsibility:**
- Each package has one clear purpose
- Functions do one thing well
- Classes/interfaces focused on single capability

**Open/Closed:**
- Interfaces for extensibility
- New features through new implementations
- Avoid modifying existing, stable code

**Liskov Substitution:**
- Interface contracts honored by implementations
- Mock implementations behave like real ones

**Interface Segregation:**
- Small, focused interfaces
- Clients depend only on needed methods

**Dependency Inversion:**
- Depend on abstractions (interfaces)
- High-level modules don't depend on low-level
- Inversion of Control through DI

## Testing Guidelines

### Test Structure
- **Unit tests:** Test individual functions/methods
- **Integration tests:** Test component interactions
- **Table-driven tests:** Multiple test cases in one function
- **Mock tests:** Use generated mocks for dependencies

### Test Organization
```
domain/
  user/
    handler/
      handler.go
      handler_test.go
    usecase/
      usecase.go
      usecase_test.go
test/
  integration/
    api_test.go
  mock/
    sqlmock_test.go
```

### Testing Best Practices
- Write tests for critical business logic
- Aim for >80% coverage on core paths
- Use table-driven tests for multiple scenarios
- Mock external dependencies (database, HTTP clients)
- Use testcontainers for real database integration tests
- Test error paths, not just happy paths

### Test Naming
- `Test<FunctionName>_<Scenario>_<ExpectedResult>`
- Example: `TestLogin_ValidCredentials_ReturnsToken`

## Security Standards

### Password Storage
- Always use Argon2id for password hashing
- Configurable memory, iterations, parallelism
- Never store plaintext passwords

### Authentication
- JWT tokens for stateless authentication
- Token expiration configurable
- Secret key must be environment-specific
- Validate tokens on protected routes

### Input Validation
- Validate all user inputs
- Use validator/v10 for request DTOs
- Sanitize database queries (SQLC prevents SQL injection)
- Validate file uploads (size, type)

### CORS Configuration
- Whitelist allowed origins per environment
- Configure allowed methods and headers
- Use secure defaults for production

### Rate Limiting
- Configurable requests per time window
- Apply to API endpoints
- Prevent abuse and DoS attacks

## Database Standards

### Migrations
- Use SQLC migration format
- Up and down migrations required
- Version with timestamp format: `YYYYMMDD_NNN_description`
- Place in `sqlc/migrations/` directory

### Queries
- Use SQLC for type-safe queries
- SQL files in `sqlc/queries/` directory
- Named queries for clarity
- Parameterized queries (SQLC handles this)

### Connection Pooling
- Configure min/max connections
- Set connection lifetime and idle timeout
- Health check period for stale connections
- Adjust based on application load

## Error Handling

### Error Types
- **Domain errors:** Business rule violations (e.g., `ErrUserNotFound`)
- **Repository errors:** Data access failures
- **Infrastructure errors:** External service failures
- **Validation errors:** Input validation failures

### Error Wrapping
- Wrap errors with context using `fmt.Errorf`
- Use `errors.Is()` and `errors.As()` for error checking
- Log errors with sufficient context
- Return appropriate HTTP status codes

### HTTP Status Codes
- 200 OK - Successful GET/PUT/PATCH
- 201 Created - Successful POST
- 400 Bad Request - Validation errors
- 401 Unauthorized - Missing/invalid JWT
- 404 Not Found - Resource not found
- 409 Conflict - Duplicate resources
- 500 Internal Server Error - Unexpected errors

## Logging Standards

### Log Levels
- **Debug:** Detailed diagnostic information
- **Info:** General informational messages
- **Warn:** Warning messages for potential issues
- **Error:** Error events that might still allow continued operation
- **Fatal:** Severe errors requiring immediate attention

### Log Format
- Structured JSON logging
- Include request ID for tracing
- Contextual fields (user_id, action, resource)
- Timestamps in ISO 8601 format

### What to Log
- Application startup/shutdown
- Request/response for API calls (with request ID)
- Errors with stack traces
- Business events (user registration, task creation)
- Performance metrics (slow queries, long-running operations)

## API Standards

### RESTful Design
- Use appropriate HTTP methods (GET, POST, PUT, PATCH, DELETE)
- Resource-based URLs (e.g., `/api/v1/users/:id`)
- Query parameters for filtering and pagination
- Consistent response format

### Response Format
```json
{
  "data": { ... },
  "error": null,
  "meta": { "total": 100, "page": 1 }
}
```

### Versioning
- URL-based versioning: `/api/v1/`
- Backward compatibility within major versions
- Deprecation notices for breaking changes

### Documentation
- Swagger/OpenAPI documentation
- Auto-generated from code annotations
- Example requests/responses
- Authentication requirements documented

## Deployment Guidelines

### Docker
- Multi-stage builds for optimization
- Alpine-based images for smaller size
- Non-root user for security
- Health checks defined in Dockerfile

### Configuration
- Environment-specific configs (local, dev, uat, prod)
- Sensitive data via environment variables
- Never commit secrets to repository

### Health Checks
- `/health` - Application health
- `/readiness` - Readiness for traffic
- Database connectivity check
- Dependency service checks

## Performance Guidelines

### Database
- Use connection pooling
- Optimize queries with proper indexes
- Batch operations when possible
- Use prepared statements (SQLC handles this)

### HTTP
- Enable compression for large responses
- Use appropriate cache headers
- Implement rate limiting
- Monitor response times

### Memory
- Reuse objects where possible
- Avoid allocations in hot paths
- Use value types for small structs
- Profile before optimizing

## Code Quality

### Linting
- Use golangci-lint
- Configure in `.golangci.yml`
- Run in CI/CD pipeline

### Code Review Checklist
- Follows coding standards
- Tests included and passing
- Error handling complete
- Documentation updated
- No security vulnerabilities
- Performance considered

### Documentation
- Godoc comments for exported functions
- README with setup instructions
- API documentation (Swagger)
- Architecture documentation (Memory Bank)
