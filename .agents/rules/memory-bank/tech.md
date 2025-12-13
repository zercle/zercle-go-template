# Technical Standards - Zercle Go Fiber Template

## Go Coding Standards

### Version & Dependencies
- **Go Version**: 1.25.0 (specified in go.mod)
- **Strict Dependencies**: All dependencies locked in go.mod
- **Proxy**: GOPROXY=direct (configurable)
- **Module Management**: Use `make init` for tidy and download

### Code Formatting
- **Formatter**: gofmt (with -s flag)
- **Format Command**: `make fmt`
- **Imports**: Grouped and sorted
  1. Standard library
  2. Third-party packages
  3. Internal packages
- **Line Length**: No strict limit, but keep readable
- **Indentation**: Tabs (not spaces)

### Naming Conventions

#### General
- **Packages**: lowercase, single-word, descriptive
- **Files**: snake_case for multi-word, simple names
- **Variables**: camelCase, short but descriptive
- **Constants**: SCREAMING_SNAKE_CASE

#### Specific Patterns
- **Interfaces**: Add "er" suffix (Reader, Writer, Repository)
- **Errors**: Variable name starts with "Err" (ErrNotFound, ErrInvalidInput)
- **DTOs**: Suffix with DTO (CreateUserRequestDTO)
- **Handlers**: Suffix with Handler (UserHandler)
- **Services**: Suffix with Service (UserService)
- **Repositories**: Suffix with Repository (UserRepository)

#### Go-Specific
- **Receiver Names**: 1-2 letters (r, s, m)
- **Receiver Types**: Use pointers for mutation, values for read-only
- **Error Variables**: Package-level with Err prefix
- **Custom Types**: Use meaningful names (UserID instead of string)

## Architecture Patterns

### Clean Architecture Rules
1. **Layer Dependencies**: Must point inward only
   - Handler → Service → Repository → Domain
   - No dependencies between same-level packages
   - No dependencies skipping layers

2. **Dependency Inversion**:
   - Services depend on repository interfaces (ports)
   - Handlers depend on service interfaces
   - Concrete implementations injected via DI

3. **Dependency Injection**:
   - Use samber/do/v2
   - Register in `internal/infrastructure/container/di.go`
   - Singleton for services and repositories
   - Request-scoped for handlers

4. **Ports & Adapters**:
   - Ports in `internal/core/port/`
   - Adapters implement ports
   - Adapters in `internal/adapter/`

### Domain-Driven Design
- **Entities**: Structs with ID and business logic
- **Value Objects**: Types for specific domain concepts
- **Domain Errors**: Custom error types in `internal/core/domain/errors/`
- **Business Rules**: In domain layer, not services

### Repository Pattern
- **Interface**: Define in `internal/core/port/repository.go`
- **Implementation**: In `internal/adapter/storage/postgres/`
- **SQL**: Use sqlc for type-safe queries
- **Transactions**: Pass context, handle errors explicitly

## Error Handling Standards

### Error Types & Patterns
1. **Domain Errors**: Custom types in domain/errors package
2. **Repository Errors**: Wrap with context using fmt.Errorf("%w", err)
3. **Service Errors**: Propagate with business context
4. **Handler Errors**: Convert to appropriate HTTP status

### Error Wrapping
```go
// Good
return fmt.Errorf("create user: %w", err)

// Bad
return fmt.Errorf("create user: %s", err.Error())
```

### Error Context
- Include operation name
- Include entity type and ID when relevant
- Don't expose internal errors to clients
- Log internal errors with full context

### Error Response Format
```go
// JSend Format
type Response struct {
    Status  string      `json:"status"`
    Data    interface{} `json:"data,omitempty"`
    Message string      `json:"message,omitempty"`
}
```

## Testing Standards

### Test Structure
- **Naming**: `*_test.go`
- **Function Naming**: TestXxx(t *testing.T)
- **Table-Driven Tests**: For multiple scenarios
- **Assertions**: Use testify/require or testify/assert

### Test Organization
```
- Unit Tests: Test business logic (domain, service)
- Integration Tests: Full stack (test/integration/)
- Mock Tests: External dependencies (test/mocks/)
```

### Testing Rules
1. **Domain Tests**: No external dependencies
2. **Service Tests**: Use mock repositories
3. **Repository Tests**: Use sqlmock
4. **Handler Tests**: Use httptest
5. **Integration Tests**: Real database (test/integration/)

### Coverage Requirements
- **Minimum**: 80% coverage
- **Critical Paths**: 100% coverage
- **Command**: `make test-coverage`
- **Race Detection**: Always enabled

### Current Test Coverage (Updated)
**Total Test Files**: 16
**Coverage Status**: ✅ Complete and Comprehensive
**Test Distribution**:
- **Health Feature Tests** (3 files):
  - `health_handler_test.go` - HTTP handler tests with httptest
  - `health_service_test.go` - Service layer tests with mocks
  - `health_repo_test.go` - Repository tests with sqlmock
- **User Feature Tests** (3 files):
  - `user_handler_test.go` - Registration, login, profile tests
  - `user_service_test.go` - Business logic and JWT tests
  - `user_repo_test.go` - CRUD operations with sqlmock
- **Post Feature Tests** (3 files):
  - `post_handler_test.go` - CRUD handler tests
  - `post_service_test.go` - Post business logic tests
  - `post_repo_test.go` - Post data access tests
- **Existing Tests** (7 files):
  - `auth_test.go`, `response_test.go`, `validator_test.go`
  - `integration/auth_test.go`
  - `repository_mock.go`, `service_mock.go`

**Test Quality**:
- Table-driven tests for comprehensive coverage
- Mock-based unit tests for all layers
- Context cancellation and timeout handling
- Edge case testing (empty responses, errors, concurrent operations)
- JWT token validation and authentication flow
- Database connectivity and error scenarios

### Mock Generation
```bash
make generate  # Regenerates all mocks
```
- Interfaces in `internal/core/port/` mocked
- Tools: go.uber.org/mock/mockgen

## Code Generation

### sqlc
```bash
sqlc generate
```
- Generates type-safe Go from SQL
- Queries in `sql/queries/*.sql`
- Generated code in `internal/infrastructure/sqlc/`

### Swagger
```bash
swag init -g cmd/server/main.go -o docs
```
- Auto-generates OpenAPI 2.0 documentation
- Regenerate after API changes
- Endpoint: http://localhost:3000/swagger/index.html

### Mocks
```bash
go generate ./...
```
- Reads `//go:generate` directives
- Generates mocks for interfaces
- Updates test/mocks/

## Build Tags System

### Purpose
Conditional compilation for modular deployments
Build minimal binaries with only required handlers
Reduce binary size and deployment time

### Available Tags
- `health`: Health check endpoints only
- `user`: User management endpoints only
- `post`: Post management endpoints only
- `all`: All endpoints (default)

### Usage
```bash
# Build with specific tags
go build -tags "health,user" ./cmd/server

# Build minimal binary
go build -tags "health" ./cmd/server

# Build all features (default)
go build -tags "all" ./cmd/server
# or simply
go build ./cmd/server
```

### Makefile Integration
```bash
make build           # Build with all tags
make build-health    # Build health-only binary
make build-user      # Build user-only binary
make build-post      # Build post-only binary
```

### Best Practices
1. Always specify build tags explicitly
2. Document which tags include which functionality
3. Test builds with different tag combinations
4. Use `all` tag for development and testing
5. Use specific tags for production to minimize binary size

### Build Tags Status (Updated)
**Status**: ✅ Fully Implemented and Tested
**Available Tags**: `health`, `user`, `post`, `all`
**Build Verification**: ✅ All tag combinations compile successfully
**Binary Sizes** (verified):
- `health` only: ~35.3 MB
- `user` only: ~35.7 MB
- `post` only: ~35.6 MB
- `health+user`: ~35.7 MB
- `all`: ~35.8 MB

**DI Hooks Integration**: ✅ All features properly wired with conditional registration

### Route Modularization
Routes split into separate files:
- `routes_base.go`: Base configuration and middleware
- `routes_health.go`: Health check routes
- `routes_user.go`: User management routes
- `routes_post.go`: Post management routes

Each route file should be self-contained and only register its own routes.

### DI Hooks System
Modular dependency injection registration:
- `di.go`: Main DI container setup
- `di_health.go`: Health module DI registration
- `di_user.go`: User module DI registration
- `di_post.go`: Post module DI registration
- `di_hooks_*.go`: Conditional hooks for optional dependencies

DI hooks allow conditional registration of dependencies based on build tags.

## Configuration Management

### Config Structure
- **Location**: `internal/infrastructure/config/config.go`
- **Format**: Struct with validation tags
- **Loading**: Viper (YAML + env vars)
- **Validation**: validator/v10 + custom rules

### Config Rules
1. **Defaults**: Set hardcoded defaults
2. **YAML**: Environment-specific files in configs/
3. **Env Vars**: Override YAML (highest priority)
4. **Validation**: Must validate on load
5. **Production**: Extra checks (e.g., JWT secret)

### Environment Files
- `configs/dev.yaml` - Development
- `configs/uat.yaml` - User Acceptance Testing
- `configs/prod.yaml` - Production

## Database Standards

### SQL Conventions
- **Tool**: sqlc for type-safe queries
- **Location**: `sql/queries/`
- **Naming**: snake_case for columns
- **Transactions**: Explicit BEGIN/COMMIT
- **Parameters**: Use $1, $2, etc.

### Migration Standards
- **Tool**: golang-migrate
- **Location**: `sql/migrations/`
- **Naming**: timestamp_description.sql
- **Up/Down**: Both directions required
- **Testing**: Test on copy before production

### Connection Management
- **Pool Settings**: Configurable in Config struct
- **Max Open**: Number of max connections
- **Max Idle**: Number of idle connections
- **Lifetime**: Max connection lifetime
- **Context**: Always pass context to DB operations

## Security Standards

### Authentication
- **Method**: JWT (golang-jwt/jwt/v5)
- **Algorithm**: HS256 (configurable)
- **Secret**: Must be 32+ characters
- **Expiration**: Configurable duration
- **Storage**: Client-side (localStorage/sessionStorage)

### Password Security
- **Hashing**: bcrypt (golang.org/x/crypto/bcrypt)
- **Cost**: 12 rounds (configurable)
- **Never**: Store plaintext or reversible encrypted
- **Validation**: Use bcrypt.CompareHashAndPassword

### Validation
- **Framework**: go-playground/validator/v10
- **Tags**: Struct tags for validation rules
- **Custom**: Custom validation functions
- **Middleware**: Validation middleware in handlers

### CORS
- **Configuration**: Via Config struct
- **Origins**: Explicitly configured (not wildcard)
- **Credentials**: Configurable
- **Headers**: Explicitly listed

### Rate Limiting
- **Middleware**: Implemented in `internal/middleware/ratelimit.go`
- **Purpose**: Prevent abuse and DoS
- **Configuration**: Via middleware setup

## HTTP/API Standards

### Routing
- **Framework**: Fiber
- **Versioning**: URL path versioning (/v1/)
- **Resource Naming**: Plural nouns (/users, /posts)
- **HTTP Methods**: RESTful (GET, POST, PUT, DELETE)

### Request/Response
- **Format**: JSON
- **Response Standard**: JSend format
- **Content-Type**: application/json
- **Accept Header**: Respected for content negotiation

### Middleware Order
1. Request ID
2. Logging
3. CORS
4. Rate Limiting
5. Authentication
6. Recovery
7. Validation
8. Handler

### Status Codes
- 200 - OK
- 201 - Created
- 400 - Bad Request
- 401 - Unauthorized
- 403 - Forbidden
- 404 - Not Found
- 409 - Conflict
- 422 - Unprocessable Entity
- 500 - Internal Server Error

## Logging Standards

### Structured Logging
- **Library**: zerolog
- **Format**: JSON (production), text (development)
- **Fields**: Consistent field names
- **Context**: Include request ID, user ID when available

### Log Levels
- **Debug**: Detailed diagnostic info
- **Info**: General operational events
- **Warn**: Warning events
- **Error**: Error events

### Log Context
- Request ID
- User ID (when available)
- HTTP Method and Path
- Response Status
- Duration

## Performance Standards

### Database
- **Connection Pooling**: Configured limits
- **Prepared Statements**: sqlc handles this
- **Indexes**: Required on foreign keys and frequently queried columns
- **Query Optimization**: Use EXPLAIN ANALYZE

### HTTP
- **Timeouts**: Read/Write timeouts configured
- **Keep-Alive**: Configured in Fiber
- **Compression**: Enabled (gzip/brotli)

### Memory
- **Object Pooling**: For frequently allocated objects
- **Avoid Allocations**: In hot paths
- **Escape Analysis**: Ensure values don't escape to heap unnecessarily

## Linting & Static Analysis

### golangci-lint
- **Config**: `.golangci.yml`
- **Command**: `make lint`
- **Issues**: All warnings must be addressed
- **Rules**: Strict configuration

### Pre-commit Checks
1. `make fmt` - Format code
2. `make lint` - Lint code
3. `make test` - Run tests
4. `make generate` - Regenerate code if needed

## Documentation Standards

### Code Documentation
- **Public Functions**: Must have comments
- **Packages**: Package-level comment (// Package xyz...)
- **Exported Types**: Comments required
- **Examples**: Include usage examples

### API Documentation
- **Swagger**: Auto-generated from code
- **Comments**: Use swaggo annotations
- **Examples**: Include request/response examples
- **Updates**: Regenerate after API changes

### README
- **Setup Instructions**: Step-by-step
- **Configuration**: All environment variables
- **Examples**: Common commands
- **API**: Endpoint documentation

## Docker Standards

### Multi-stage Build
- **Base**: distroless or alpine
- **Build**: Go builder image
- **Final**: Minimal runtime image
- **User**: Non-root user

### Security
- **No Root**: Run as non-root user
- **Minimal Image**: Use distroless/alpine
- **Health Check**: In Dockerfile
- **Labels**: Standard labels

## Development Workflow

### Standard Commands
```bash
make init          # Initial setup
make generate      # Regenerate code
make dev           # Development mode
make test          # Run tests
make test-coverage # Coverage report
make lint          # Lint check
make build         # Build binary
make docker-up     # Start containers
make migrate-up    # Run migrations
```

### Pre-commit Checklist
1. Code formatted (`make fmt`)
2. Lint passes (`make lint`)
3. Tests pass (`make test`)
4. Coverage acceptable
5. Generated code up-to-date
6. No debug code committed

### Branching Strategy
- **main**: Production-ready code
- **feature/***: Feature branches
- **fix/***: Bug fix branches
- **chore/***: Maintenance branches

## Anti-Patterns to Avoid

1. **Dependency Violations**: Layer skipping
2. **Global State**: Package-level variables (except config)
3. **Naked Returns**: Always named returns
4. **Error Swallowing**: Ignoring errors
5. **Any Type**: Use specific types
6. **Panic**: Use errors instead
7. **Reflection**: Avoid unless necessary
8. **Mutex Global**: Pass dependencies explicitly
9. **Context Forget**: Always pass context
10. **SQL Injection**: Use parameterized queries only

## Code Review Checklist

### Architecture
- [ ] Follows clean architecture layers
- [ ] No dependency violations
- [ ] Dependency injection used
- [ ] Interfaces properly defined

### Code Quality
- [ ] Naming conventions followed
- [ ] Functions are small and focused
- [ ] Error handling explicit
- [ ] Context passed correctly

### Testing
- [ ] Unit tests included
- [ ] Edge cases covered
- [ ] Mocks properly used
- [ ] Integration tests for new features

### Security
- [ ] No hardcoded secrets
- [ ] Input validation present
- [ ] Authentication/authorization checked
- [ ] SQL injection prevented

### Performance
- [ ] Database queries optimized
- [ ] No unnecessary allocations
- [ ] Timeouts configured
- [ ] Proper connection pooling
