# Current Work Context

**Last Updated:** 2026-02-21

## Recent Refactoring (2026-02)

### Go 1.26 Upgrade (2026-02)
- **Decision:** Upgrade Go version from 1.25.7 to 1.26
- **Impact:** Updated go.mod, Dockerfile, Makefile GO_VERSION variable
- **Files Changed:** go.mod, Makefile, Dockerfile

### Echo v5 Migration (2026-02)
- **Decision:** Migrate from Echo v4 to Echo v5
- **Impact:** Updated imports and some API changes
- **Files Changed:** cmd/api/main.go, all handlers
- **Dependencies Updated:** github.com/labstack/echo/v5 v5.0.4

### JWT Token Caching (2026-02)
- **Decision:** Add JWT token validation caching
- **Implementation:**
  - Configurable cache enabled by default
  - Cache TTL configurable (default 5 minutes)
  - Reduces authentication latency
- **Config Options:**
  - `jwt.cache_enabled` (default: true)
  - `jwt.cache_ttl` (default: 5m)

### Service → Usecase Rename
- **Rationale:** Align with Clean Architecture terminology
- **Impact:** All `service/` directories renamed to `usecase/`
- **Files Changed:**
  - `internal/feature/user/service/` → `internal/feature/user/usecase/`
  - `internal/feature/auth/service/` → `internal/feature/auth/usecase/`
  - Import statements updated across codebase
  - Documentation updated

### SQLC Integration
- **Purpose:** Type-safe SQL query generation
- **Implementation:**
  - Added `internal/infrastructure/db/sqlc/` directory
  - Created `internal/infrastructure/db/queries/users.sql`
  - Configured `sqlc.yaml` for code generation
  - Generated `sqlc/db.go`, `sqlc/models.go`, `sqlc/users.sql.go`
- **Benefits:**
  - Compile-time query validation
  - Reduced boilerplate code
  - Type-safe database operations

## Current State

### Completed Features
1. **User Management**
   - Create, read, update, delete users
   - Email uniqueness validation
   - Password update with verification
   - Paginated listing

2. **Authentication**
   - JWT token generation (access + refresh)
   - Token validation middleware
   - Password hashing with Argon2id
   - Login endpoint

3. **Infrastructure**
   - Configuration management (multi-source)
   - Structured logging with Zerolog
   - Database connection pooling
   - Dependency injection container

4. **Testing**
   - Unit tests for all components
   - Integration tests for repositories
   - Benchmark tests for performance
   - Mock generation for interfaces

### Codebase Structure
```
zercle-go-template/
├── cmd/api/                    # Application entry point
├── internal/
│   ├── config/                 # Configuration management
│   ├── container/              # Dependency injection
│   ├── errors/                 # Custom error types
│   ├── feature/
│   │   ├── auth/              # Authentication feature
│   │   │   ├── domain/        # JWT domain entities
│   │   │   ├── middleware/    # Auth middleware
│   │   │   └── usecase/       # JWT business logic
│   │   └── user/              # User management feature
│   │       ├── domain/        # User domain entities
│   │       ├── dto/           # Data transfer objects
│   │       ├── handler/       # HTTP handlers
│   │       ├── repository/    # Data access
│   │       └── usecase/       # Business logic
│   ├── infrastructure/
│   │   └── db/               # Database layer
│   │       ├── migrations/    # SQL migrations
│   │       ├── queries/       # SQL query definitions
│   │       └── sqlc/          # Generated SQLC code
│   ├── logger/                # Logging infrastructure
│   └── middleware/            # Cross-cutting middleware
├── api/docs/                  # Swagger documentation
├── configs/                   # Configuration files
└── .agents/rules/memory-bank/ # Memory bank (this directory)
```

## Known Patterns and Conventions

### Feature Organization
Each feature follows this structure:
```
feature/<name>/
├── domain/          # Pure domain logic, no dependencies
├── dto/             # Request/response structures
├── handler/         # HTTP handlers (presentation)
├── usecase/         # Business logic (orchestration)
└── repository/      # Data access interfaces
```

### Dependency Injection
- Use `container.New(cfg, opts...)` pattern
- Functional options for configuration
- Auto-selection based on database availability
- Graceful fallback to in-memory repository

### Error Handling
- Domain errors in `domain/` packages
- App errors in `internal/errors/`
- HTTP status code mapping
- Structured error logging
- User-friendly error messages

### Logging
- Use Zerolog for structured logging
- Context-aware logging with `logger.WithContext(ctx)`
- Field-based logging with `logger.String()`, `logger.Int()`, etc.
- Log levels: debug, info, warn, error, fatal

### Testing
- Mock generation with `//go:generate mockgen`
- Table-driven tests for multiple scenarios
- Integration tests with real database
- Benchmark tests for performance measurement
- Test naming: `*_test.go`, `*_integration_test.go`, `*_benchmark_test.go`

### Repository Pattern
- Interface defined in `repository/`
- Implementations: `MemoryUserRepository`, `SqlcUserRepository`
- Context support for all operations
- Error handling with typed errors

### Configuration
- Multi-source loading (env > .env > YAML > defaults)
- Environment prefix: `APP_`
- Type-safe configuration structs
- Validation on load

## Dependencies

### External Dependencies
- Echo v4 (web framework)
- pgx v5 (PostgreSQL driver)
- golang-jwt/jwt (JWT library)
- Viper (configuration)
- Zerolog (logging)
- sqlc (SQL code generation)

### Internal Dependencies
- `internal/config` - Configuration management
- `internal/container` - DI container
- `internal/errors` - Error types
- `internal/logger` - Logging interface
- `internal/infrastructure/db` - Database layer

## Next Steps

### Immediate Priorities
1. **Refresh Token Endpoint**
   - Implement refresh token validation
   - Generate new access tokens
   - Update JWT usecase

2. **Token Revocation**
   - Implement token blacklist
   - Add logout endpoint
   - Handle token expiration gracefully

3. **Rate Limiting**
   - Add rate limiting middleware
   - Configure per-IP and per-user limits
   - Integrate with Redis (optional)

### Medium-term Enhancements
1. **RBAC Implementation**
   - Define role/permission models
   - Add role-based middleware
   - Implement permission checking

2. **Email Verification**
   - Add email verification flow
   - Generate verification tokens
   - Send verification emails

3. **Password Reset**
   - Implement reset token generation
   - Add password reset endpoint
   - Send reset emails

### Long-term Considerations
1. **Caching Layer**
   - Redis integration
   - Cache frequently accessed data
   - Implement cache invalidation

2. **Event-Driven Architecture**
   - Add event bus
   - Implement event publishing
   - Add event consumers

3. **GraphQL Support**
   - Add GraphQL server
   - Define schema
   - Implement resolvers

## Known Issues

None currently documented.

## Decisions Log

### 2026-02: Service → Usecase Rename
- **Decision:** Rename all `service/` directories to `usecase/`
- **Rationale:** Align with Clean Architecture terminology and industry standards
- **Impact:** Breaking change for external consumers (if any)
- **Status:** Completed

### 2026-02: SQLC Integration
- **Decision:** Use sqlc for type-safe SQL queries
- **Rationale:** Reduce boilerplate, improve type safety, catch errors at compile time
- **Alternatives Considered:** GORM, sqlx, raw SQL
- **Status:** Completed

### Initial: Argon2id for Password Hashing
- **Decision:** Use Argon2id for password hashing
- **Rationale:** OWASP recommended, memory-hard, resistant to GPU attacks
- **Alternatives Considered:** bcrypt, scrypt, PBKDF2
- **Status:** Implemented with configurable parameters

## Code Quality Standards

### Linting
- golangci-lint with comprehensive rules
- Pre-commit hooks enforce standards
- CI/CD integration for PR validation

### Testing
- 80%+ coverage for critical paths
- All business logic tested
- Integration tests for database operations
- Benchmark tests for performance-critical code

### Documentation
- Swagger/OpenAPI for API docs
- Inline comments for complex logic
- README for project overview
- Memory bank for architectural decisions
