# Zercle Go Template - Context & Implementation Notes

## Architectural Decisions

### Technology Choices

#### Go Echo Framework
**Decision**: Use Echo v4 instead of Gin, Fiber, or standard library
**Rationale**:
- Excellent performance (competitive with Gin)
- Minimalist but extensible
- Built-in support for middleware, routing, and WebSocket
- Active community and maintenance
- Clean API design

**Trade-offs**:
- Slightly more opinionated than standard library
- Requires learning Echo-specific patterns

#### SQLC vs ORM (GORM/sqlx)
**Decision**: Use SQLC for database access
**Rationale**:
- Type-safe queries generated at compile time
- No runtime reflection overhead
- SQL queries are explicit and reviewable
- Better performance than ORMs
- Catches SQL errors at compile time

**Trade-offs**:
- More verbose than ORMs for complex queries
- Requires writing SQL queries manually
- No automatic migration generation

#### PostgreSQL vs MySQL
**Decision**: Use PostgreSQL 18+
**Rationale**:
- Advanced features (JSONB, arrays, custom types)
- Better transaction handling
- Superior concurrency model
- Excellent for complex queries
- Native support for UUIDs

**Trade-offs**:
- Slightly higher resource usage
- Not as widely available on cheap hosting

#### JWT Authentication
**Decision**: JWT tokens with HS256 signing
**Rationale**:
- Stateless authentication (no server-side session storage)
- Easy to scale horizontally
- Standard and widely adopted
- Built-in expiration handling
- Simple implementation

**Trade-offs**:
- Token revocation requires additional infrastructure
- Larger token size than session IDs
- Requires secure secret management

#### JSend Response Format
**Decision**: Standardize on JSend for all API responses
**Rationale**:
- Clear distinction between success, fail, and error states
- Widely adopted standard
- Predictable structure for clients
- Easy to implement middleware

**Trade-offs**:
- Slightly more verbose than simple JSON responses
- Requires consistent implementation across all handlers

#### Zerolog vs Zap/Logrus
**Decision**: Use zerolog for structured logging
**Rationale**:
- Zero-allocation logging (high performance)
- Simple API
- Built-in structured logging support
- JSON output for production
- Console output for development

**Trade-offs**:
- Less feature-rich than Zap
- Smaller community than Logrus

#### Viper Configuration
**Decision**: Use Viper for configuration management
**Rationale**:
- Support for multiple formats (YAML, JSON, ENV)
- Environment variable override
- Live reload support (not used in production)
- Standard for Go configuration

**Trade-offs**:
- Additional dependency
- Overkill for simple use cases

### Project Structure Decisions

#### Domain-Driven Design
**Decision**: Organize code by domain (user, service, booking, payment)
**Rationale**:
- Clear business boundary separation
- Easy to locate feature code
- Supports future microservices extraction
- Aligns with clean architecture principles

**Trade-offs**:
- More directories than feature-based organization
- May feel complex for small projects

#### Clean Architecture Layers
**Decision**: Enforce strict layer separation (Handler → UseCase → Repository)
**Rationale**:
- Testability: Each layer can be tested independently
- Maintainability: Changes are localized to layers
- Flexibility: Easy to swap implementations
- Clarity: Clear responsibility boundaries

**Trade-offs**:
- More boilerplate code
- Indirection can make simple operations verbose
- Steeper learning curve for new developers

#### Interface-Based Design
**Decision**: Define interfaces for repositories and use cases
**Rationale**:
- Enables mocking for testing
- Supports dependency injection
- Makes implementations interchangeable
- Follows Go best practices

**Trade-offs**:
- More files to maintain
- Interface drift from implementations

### Database Design Decisions

#### UUID Primary Keys
**Decision**: Use UUIDs instead of auto-increment integers
**Rationale**:
- No sequential ID exposure (security)
- Easy distributed generation
- No coordination needed for IDs
- Works well with API endpoints

**Trade-offs**:
- Larger storage size (16 bytes vs 4/8 bytes)
- Slightly slower indexing
- Less human-readable

#### Migration Files
**Decision**: Use timestamped migration files
**Rationale**:
- Clear ordering of migrations
- Standard practice with golang-migrate
- Easy to track schema evolution
- Supports rollback files

**Trade-offs**:
- Requires careful naming conventions
- Merge conflicts need manual resolution

#### SQLC Query Organization
**Decision**: Separate query files by table
**Rationale**:
- Easy to locate queries
- Clear ownership of queries
- Supports code generation
- Aligns with domain structure

**Trade-offs**:
- Some queries span multiple tables (join queries)
- May require additional query files

## Domain Rules

### User Domain
- **Email uniqueness**: Enforced at database level with UNIQUE constraint
- **Password requirements**: Minimum 8 characters, hashed with bcrypt (cost 10)
- **JWT expiration**: Configurable, default 3600 seconds (1 hour)
- **Token claims**: user_id, exp, iat
- **User ownership**: Users can only access their own resources

### Service Domain
- **Service ownership**: Only service creators can update/delete
- **Price validation**: Must be positive number
- **Duration validation**: Must be positive integer (minutes)
- **Search behavior**: Case-insensitive partial match on name and description

### Booking Domain
- **Booking status workflow**: pending → confirmed → cancelled
- **Date validation**: Booking date must be in the future
- **Overlap prevention**: Cannot double-book same service at same time
- **Cancellation rules**: Only pending bookings can be cancelled
- **Status transitions**: Validated in use case layer

### Payment Domain
- **Payment workflow**: pending → completed/refunded/failed
- **Amount validation**: Cannot exceed booking total
- **Refund rules**: Only completed payments can be refunded
- **Booking payment**: Multiple payments allowed per booking
- **Confirmation**: Payment confirmation updates booking status to confirmed

## File Purpose Summaries

### Core Application Files
- **cmd/server/main.go**: Application entry point, initializes all layers and starts HTTP server
- **go.mod**: Module definition and dependency management
- **go.sum**: Dependency checksums for reproducible builds
- **Makefile**: Build automation with common development tasks
- **Dockerfile**: Multi-stage container image for production deployment
- **docker-compose.yml**: Local development orchestration (API + PostgreSQL)
- **docker-compose.test.yml**: Test database orchestration

### Configuration Files
- **configs/local.yaml**: Local development configuration (default)
- **configs/dev.yaml**: Development environment configuration
- **configs/uat.yaml**: User acceptance testing configuration
- **configs/prod.yaml**: Production environment configuration
- **.env.example**: Example environment variables template
- **infrastructure/config/config.go**: Configuration struct and loading logic

### Infrastructure Layer
- **infrastructure/db/postgres.go**: Database connection and connection pooling
- **infrastructure/logger/logger.go**: Structured logger wrapper (zerolog)
- **infrastructure/http/client/resty.go**: HTTP client for external API calls
- **infrastructure/sqlc/db/**: SQLC generated code (do not edit manually)

### Domain Files (Pattern: domain/{name}/)

#### User Domain
- **handler/handler.go**: HTTP handlers for auth and user profile endpoints
- **usecase/usecase.go**: Business logic for authentication and user management
- **repository/repository.go**: Data access layer for user operations
- **model/user.go**: Internal user domain model
- **request/user.go**: HTTP request DTOs for user operations
- **response/user.go**: HTTP response DTOs for user operations
- **interface.go**: Shared interfaces for user domain

#### Service Domain
- **handler/handler.go**: HTTP handlers for service CRUD operations
- **usecase/usecase.go**: Business logic for service management
- **repository/repository.go**: Data access layer for service operations
- **model/service.go**: Internal service domain model
- **request/service.go**: HTTP request DTOs for service operations
- **response/service.go**: HTTP response DTOs for service operations
- **interface.go**: Shared interfaces for service domain

#### Booking Domain
- **handler/handler.go**: HTTP handlers for booking operations
- **usecase/usecase.go**: Business logic for booking workflow
- **repository/repository.go**: Data access layer for booking operations
- **model/booking.go**: Internal booking domain model
- **request/booking.go**: HTTP request DTOs for booking operations
- **response/booking.go**: HTTP response DTOs for booking operations
- **interface.go**: Shared interfaces for booking domain

#### Payment Domain
- **handler/handler.go**: HTTP handlers for payment operations
- **usecase/usecase.go**: Business logic for payment processing
- **repository/repository.go**: Data access layer for payment operations
- **model/payment.go**: Internal payment domain model
- **request/payment.go**: HTTP request DTOs for payment operations
- **response/payment.go**: HTTP response DTOs for payment operations
- **interface.go**: Shared interfaces for payment domain

### Database Files
- **sql/migration/**: Database schema migration files (up/down)
- **sql/query/**: SQLC query definitions organized by table
- **infrastructure/sqlc/db/**: Generated Go code from SQLC queries

### Test Files
- **test/integration/**: Integration tests with test server
- **test/unit/**: Unit tests for individual components
- **domain/*/*_test.go**: Unit tests for handlers, use cases, repositories

### Package Files (pkg/)
- **pkg/middleware/**: Echo middleware (request ID, logging, CORS, rate limiting, JWT)
- **pkg/response/**: JSend response builders and helpers
- **pkg/health/**: Health check handlers

### Documentation Files
- **docs/**: Swagger/OpenAPI generated documentation
- **README.md**: Project overview and getting started guide
- **PODMAN.md**: Podman setup instructions
- **requirements.md**: Project requirements and specifications
- **CLAUDE.md**: AI assistant instructions and memory bank rules

## Implementation Notes

### Dependency Injection Pattern
Each domain uses constructor functions for initialization:
```go
func Initialize(queries *sqlc.Queries, log *logger.Logger) Repository {
    return &repositoryImpl{queries, log}
}
```

This pattern:
- Makes dependencies explicit
- Enables easy mocking for tests
- Supports compile-time dependency checking
- Follows Go idioms

### Error Handling Strategy
1. **Repository layer**: Returns raw database errors wrapped with context
2. **UseCase layer**: Maps repository errors to domain errors, adds business logic validation
3. **Handler layer**: Maps domain errors to HTTP status codes and JSend responses

Example:
```go
// Repository
return fmt.Errorf("failed to get user: %w", err)

// UseCase
if errors.Is(err, sql.ErrNoRows) {
    return ErrUserNotFound
}
return fmt.Errorf("repository error: %w", err)

// Handler
if errors.Is(err, ErrUserNotFound) {
    return c.JSON(http.StatusNotFound, response.Error("User not found"))
}
```

### Request Validation Strategy
1. **Handler layer**: Validates request structure using go-playground/validator
2. **UseCase layer**: Validates business rules and constraints
3. **Repository layer**: Relies on database constraints for data integrity

### Logging Strategy
- **Request logging**: Middleware logs all incoming requests with timing
- **Error logging**: All errors logged with contextual fields (request_id, user_id)
- **Debug logging**: Enabled in local/dev environments
- **Production logging**: JSON format for log aggregation

### Testing Strategy
1. **Unit tests**: Test individual components in isolation using mocks
2. **Integration tests**: Test full HTTP request/response cycle
3. **Test structure**: Table-driven tests for multiple scenarios
4. **Coverage target**: >80% overall, >90% for critical business logic

### Configuration Loading
1. Read `SERVER_ENV` environment variable (defaults to "local")
2. Load corresponding YAML file from `configs/{env}.yaml`
3. Override with environment variables if present
4. Validate configuration structure
5. Fail fast on invalid configuration

### Graceful Shutdown
1. Listen for SIGINT/SIGTERM signals
2. Stop accepting new requests
3. Wait up to 30 seconds for in-flight requests to complete
4. Close database connections
5. Flush logger
6. Exit

### API Versioning
- Current version: `/api/v1/`
- Future versions: `/api/v2/`, `/api/v3/`
- No breaking changes within a version
- Deprecate old versions before removing

## Known Issues & Limitations

### Current Limitations
1. **No file upload support**: Service images require manual URLs
2. **No caching layer**: All queries hit database directly
3. **No async processing**: Background jobs require external queue
4. **No pagination**: List endpoints return all results (performance concern)
5. **No rate limiting per user**: Global rate limiting only

### Future Improvements
1. Add Redis caching layer for frequently accessed data
2. Implement pagination for all list endpoints
3. Add user-specific rate limiting
4. Implement file upload service for service images
5. Add background job processing for notifications
6. Implement audit logging for sensitive operations

## Dependencies Summary

### Core Dependencies
- **github.com/labstack/echo/v4**: Web framework
- **github.com/jackc/pgx/v5**: PostgreSQL driver
- **github.com/rs/zerolog**: Structured logging
- **github.com/spf13/viper**: Configuration management
- **github.com/swaggo/swag**: Swagger documentation generation

### Development Dependencies
- **github.com/stretchr/testify**: Testing assertions and mocks
- **go.uber.org/mock**: Mock generation
- **github.com/swaggo/echo-swagger**: Swagger UI for Echo

### Build Dependencies
- **github.com/go-playground/validator/v10**: Request validation
- **github.com/golang-jwt/jwt/v5**: JWT token generation and validation
- **github.com/google/uuid**: UUID generation
- **golang.org/x/crypto**: Password hashing (bcrypt)

## Environment Configuration

### Local Development
```bash
SERVER_ENV=local
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=zercle_db
JWT_SECRET=dev-jwt-secret-key-change-in-production
```

### Development Environment
```bash
SERVER_ENV=dev
# Database and other credentials from infrastructure
```

### Production Environment
```bash
SERVER_ENV=prod
# All secrets from secure secret management
# No defaults in production
```

## Git Workflow

### Branch Strategy
- **main**: Production-ready code
- **develop**: Development branch for next release
- **feature/***: Feature branches
- **bugfix/***: Bug fix branches
- **hotfix/***: Emergency production fixes

### Commit Message Format
```
<type>: <description>

[optional body]

[optional footer]
```

Types: feat, fix, docs, style, refactor, test, chore

Example:
```
feat: add user authentication

Implement JWT-based authentication with login and registration endpoints.

Closes #123
```

## Performance Considerations

### Database Optimization
- Use indexes on frequently queried columns
- Use connection pooling (MaxOpenConns: 25, MaxIdleConns: 5)
- Set connection lifetime (5 minutes)
- Use prepared statements (SQLC handles this)

### API Optimization
- Implement pagination for list endpoints (future)
- Add caching layer for frequently accessed data (future)
- Use compression for large responses (future)
- Optimize JSON serialization (future)

### Memory Management
- Preallocate slices with capacity
- Use bytes.Buffer for string building
- Reuse objects where possible
- Limit concurrent goroutines

## Security Considerations

### Authentication & Authorization
- Never store passwords in plain text (use bcrypt)
- Use HTTPS in production
- Validate JWT signatures on every request
- Implement token expiration
- Use secure secret management

### Input Validation
- Validate all input at handler layer
- Sanitize data before database operations
- Use parameterized queries (SQLC prevents SQL injection)
- Limit file upload sizes (when implemented)

### Data Protection
- Never expose sensitive data in JSON responses
- Hash passwords before storing
- Use environment variables for secrets
- Implement CORS properly
- Add rate limiting to prevent abuse

## Monitoring & Observability

### Health Checks
- `/health`: Liveness probe (always returns 200 if running)
- `/readiness`: Readiness probe (checks database connectivity)

### Logging
- Structured JSON logs in production
- Include request_id for tracing
- Log all errors with context
- Use appropriate log levels

### Metrics (Future)
- Request count and latency
- Database query performance
- Error rates by endpoint
- Active connections

## Deployment Considerations

### Container Orchestration
- Support for Docker and Podman
- docker-compose for local development
- Kubernetes ready (health checks, graceful shutdown)
- Horizontal scaling support (stateless handlers)

### Configuration Management
- Environment-specific YAML files
- Environment variable override support
- No secrets in code or configuration files
- Validation on startup

### Database Migrations
- Run migrations on deployment
- Support for rollbacks
- Version-controlled migration files
- Test migrations before production

## Integration Points

### External Services (Future)
- **Payment Gateway**: Stripe/PayPal integration
- **Email Service**: SendGrid/Mailgun for notifications
- **SMS Service**: Twilio for SMS notifications
- **Storage Service**: S3/GCS for file uploads

### API Integrations (Future)
- **Calendar Sync**: Google Calendar/Outlook integration
- **Video Conferencing**: Zoom/Teams integration
- **Analytics**: Google Analytics/Mixpanel