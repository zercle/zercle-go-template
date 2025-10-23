# Context & Decisions

## Architectural Decisions

### Clean Architecture Choice
**Decision:** Adopted Clean Architecture with DDD principles
**Rationale:** Separates business logic from infrastructure, improves testability, enables independent evolution of layers
**Impact:** All new domains must follow the established layer structure

### SQLC for Database Access
**Decision:** Use SQLC instead of raw SQL or ORM
**Rationale:** Type-safe queries, compile-time safety, better performance than ORMs, explicit SQL control
**Impact:** All database queries must be SQLC-generated, placed in `sqlc/queries/`

### JWT Stateless Authentication
**Decision:** JWT tokens without server-side session storage
**Rationale:** Stateless design enables horizontal scaling, simpler architecture, no session management overhead
**Impact:** All protected routes require JWT middleware, tokens stored client-side

### Argon2id for Password Hashing
**Decision:** Argon2id algorithm for password hashing
**Rationale:** Memory-hard, resistant to GPU/ASIC attacks, recommended by security experts
**Impact:** All password operations must use the password.Hasher wrapper

### Echo Framework
**Decision:** Echo v4 as HTTP framework
**Rationale:** High performance, minimal boilerplate, excellent middleware support, active community
**Impact:** All HTTP handlers use Echo context and patterns

## Domain Rules

### User Domain
**Business Rules:**
- Email must be unique across all users
- Password must be hashed before storage
- Users can only update their own profiles
- Email cannot be changed after registration
- Minimum full name length: 2 characters

**Validation Rules:**
- Email format validated by validator/v10
- Password strength enforced by Argon2id parameters
- Phone number is optional
- Full name required for registration

**Ownership Rules:**
- Users can only access their own profile
- Admin endpoints (if added) can access all users
- User ID extracted from JWT token for authorization

### Task Domain
**Business Rules:**
- Tasks must have an owner (user_id)
- Users can only access their own tasks
- Task status must be one of: pending, in_progress, completed, cancelled
- Task priority must be one of: low, medium, high, urgent
- Completed tasks automatically set completed_at timestamp

**Validation Rules:**
- Title is required
- Description is optional
- Due date is optional
- Status defaults to "pending"
- Priority defaults to "medium" if not specified

**Ownership Rules:**
- All task operations verify user ownership
- Cannot access/modify tasks owned by other users
- Task list filtered by user_id

## File & Component Summaries

### Core Application Files

**cmd/server/main.go**
- Application entry point
- Loads environment-specific configuration
- Initializes logger and application
- Handles graceful shutdown

**internal/app/app.go**
- Main application structure
- Dependency injection container
- Middleware setup (RequestID, Logger, Recovery, CORS, RateLimit)
- Route registration
- Server lifecycle management

### Configuration

**internal/infrastructure/config/config.go**
- Configuration structs for all components
- Viper-based configuration loading
- Environment variable support
- Type-safe configuration access

**configs/*.yaml**
- Environment-specific configurations
- local, dev, uat, prod environments
- Database, JWT, logging, CORS, rate limit settings

### Database Layer

**internal/infrastructure/db/postgres.go**
- PostgreSQL database implementation
- Connection pooling configuration
- Health check implementation
- SQLC queries integration

**internal/infrastructure/db/factory.go**
- Database factory for creating connections
- Abstracts database type selection
- Currently supports PostgreSQL only

**internal/infrastructure/sqlc/db/**
- SQLC-generated code
- Type-safe database queries
- Models and querier interfaces
- Auto-generated from SQL files

### Domain: User

**internal/domain/user/entity/user.go**
- User entity definition
- UUID-based primary key
- Fields: id, email, password, full_name, phone, timestamps

**internal/domain/user/repository/repository.go**
- SQLC-based repository implementation
- CRUD operations for users
- Email uniqueness check
- Pagination support

**internal/domain/user/usecase/usecase.go**
- Business logic for user operations
- Register, Login, GetProfile, UpdateProfile, DeleteAccount, ListUsers
- Password hashing and verification
- JWT token generation
- Domain-specific error definitions

**internal/domain/user/handler/handler.go**
- HTTP handlers for user endpoints
- Request/response DTO mapping
- Error handling and HTTP status codes
- Route registration

### Domain: Task

**internal/domain/task/entity/task.go**
- Task entity definition
- UUID-based primary key
- Fields: id, user_id, title, description, status, priority, due_date, completed_at, timestamps

**internal/domain/task/repository/repository.go**
- pgx-based repository implementation
- CRUD operations for tasks
- User filtering for list operations
- Ownership verification

**internal/domain/task/usecase/usecase.go**
- Business logic for task operations
- CreateTask, GetTask, ListTasks, UpdateTask, DeleteTask
- Status and priority validation
- Ownership enforcement
- Domain-specific error definitions

**internal/domain/task/handler/handler.go**
- HTTP handlers for task endpoints
- Request/response DTO mapping
- Error handling and HTTP status codes
- Protected routes only

### Infrastructure Components

**internal/infrastructure/logger/logger.go**
- Zerolog-based structured logger
- Configurable log levels and format
- Request ID integration
- Context-aware logging

**internal/infrastructure/password/passworder.go**
- Argon2id password hashing wrapper
- Configurable parameters
- Hash and verify operations

**internal/infrastructure/http/client/resty.go**
- Resty HTTP client wrapper
- For making external HTTP requests
- Configurable timeouts and retries

**pkg/middleware/**
- Custom middleware implementations
- JWT authentication
- Request ID generation
- Structured logging
- CORS handling
- Rate limiting

**pkg/health/**
- Health check handler
- Database connectivity check
- Readiness probe

## Dependency Mapping

### Domain Dependencies
- **User Domain:** Depends on config, logger, password, middleware (JWT)
- **Task Domain:** Depends on logger only (uses pgx directly for DB)

### Infrastructure Dependencies
- **Database:** pgx/v5 driver
- **Config:** Viper
- **Logging:** Zerolog
- **Validation:** validator/v10
- **Auth:** golang-jwt/jwt/v5
- **Password:** golang.org/x/crypto

### External Dependencies
- **PostgreSQL:** Primary database
- **Testcontainers:** Integration testing
- **Swagger:** API documentation

## Key Implementation Details

### JWT Token Structure
- Contains user ID and email in claims
- Configurable expiration time
- Secret key from configuration
- Bearer token format in Authorization header

### Database Connection Pool
- Min connections: 5
- Max connections: 25
- Connection lifetime: 1 hour
- Idle timeout: 10 minutes
- Health check period: 1 minute

### Rate Limiting
- Configurable requests per time window
- Default: 100 requests per 60 seconds
- Applied at middleware level
- Per-client tracking

### CORS Configuration
- Allowed origins configurable per environment
- Local: localhost:3000, localhost:8080
- Methods: GET, POST, PUT, PATCH, DELETE, OPTIONS
- Headers: Authorization, Content-Type, X-Request-ID

### Error Handling Pattern
```go
// UseCase layer: Domain errors
if user == nil {
    return nil, ErrUserNotFound
}

// Repository layer: Wrap with context
if err != nil {
    return fmt.Errorf("failed to create user: %w", err)
}

// Handler layer: Map to HTTP status
if errors.Is(err, ErrUserNotFound) {
    return c.JSON(http.StatusNotFound, ErrorResponse{...})
}
```

### Request Validation
- Use validator/v10 struct tags
- Validate before business logic
- Return validation errors with field details
- Example: `validate:"required,email"`

### Pagination Pattern
```go
// Standard pagination parameters
limit, offset := getPaginationParams(c)

// Repository returns data + total count
users, total, err := repo.List(ctx, limit, offset)

// Response includes pagination metadata
return c.JSON(http.StatusOK, ListResponse{
    Users: users,
    Total: total,
    Limit: limit,
    Offset: offset,
})
```

## Testing Strategy

### Unit Tests
- Test usecase business logic
- Mock repository dependencies
- Test error paths and edge cases
- Located in same package as implementation

### Integration Tests
- Test API endpoints end-to-end
- Use testcontainers for real database
- Test authentication flow
- Located in test/integration/

### Mock Generation
- Use go.uber.org/mock
- Generate mocks from domain interfaces
- Located in domain/*/mock/ directories
- Regenerate when interfaces change

### Test Helpers
- test/mock/sqlmock.go - SQL mock utilities
- test/integration/test_helper.go - Integration test setup
- Common test fixtures and utilities

## Migration Strategy

### Database Migrations
- SQLC migration format
- Up and down migrations required
- Version naming: YYYYMMDD_NNN_description
- Apply migrations in order
- Rollback support with down migrations

### Schema Changes
- Add new migrations for schema changes
- Never modify existing migrations
- Use SQLC to regenerate queries after schema changes
- Test migrations in all environments

## Configuration Management

### Environment Hierarchy
1. Base config from YAML file
2. Environment variable overrides
3. Default values in struct tags

### Configuration Files
- `configs/local.yaml` - Local development
- `configs/dev.yaml` - Development environment
- `configs/uat.yaml` - User acceptance testing
- `configs/prod.yaml` - Production

### Environment Variables
- `SERVER_ENV` - Environment selector (default: local)
- Database credentials via env vars in production
- JWT secret via env vars in production
- Never commit secrets to repository

## Deployment Considerations

### Docker Deployment
- Multi-stage build for optimization
- Alpine-based final image
- Non-root user for security
- Health checks configured
- Port 3000 exposed

### Database Requirements
- PostgreSQL 12+ required
- Connection pool configuration important
- Migrations must be applied before startup
- Health check verifies connectivity

### Monitoring Points
- Health check endpoints
- Request/response logging
- Error logging with context
- Performance metrics (future)
- Database query performance (future)

## Known Constraints

### Current Limitations
- Only PostgreSQL supported (no MySQL, SQLite)
- No caching layer implemented
- No message queue integration
- No distributed tracing
- No metrics collection
- Single-region deployment only

### Technical Debt
- Task domain uses pgx directly instead of SQLC
- Mixed database access patterns (SQLC vs pgx)
- Consider standardizing on one approach

### Future Considerations
- Add Redis caching layer
- Implement message queue for async operations
- Add Prometheus metrics
- Implement distributed tracing with OpenTelemetry
- Add GraphQL support as alternative to REST
- Consider gRPC for internal service communication
