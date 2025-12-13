# Architecture - Zercle Go Fiber Template

## Architectural Pattern
**Clean Architecture** with **Domain-Driven Design (DDD)** and **Hexagonal Architecture** principles.

## Modular Architecture
**Build Tags System**: Conditional compilation with build tags for modular deployments
- **Tags**: `health`, `user`, `post`, `all`
- **Purpose**: Build minimal binaries with only required handlers
- **Benefit**: Reduced binary size, faster deployment, selective functionality
- **Documentation**: See BUILD_TAGS.md for detailed usage

**Route Modularization**: Routes split into separate files for better organization
- `routes_base.go` - Base route configuration
- `routes_health.go` - Health check routes
- `routes_user.go` - User management routes
- `routes_post.go` - Post management routes

**DI Hooks System**: Modular dependency injection registration
- `di.go` - Main DI container
- `di_health.go` - Health-specific DI setup
- `di_post.go` - Post module DI setup
- `di_user.go` - User module DI setup
- `di_hooks_*.go` - Conditional hooks for modular registration

## Layered Architecture

### 1. Handler Layer (HTTP Adapters)
**Location**: `internal/adapter/handler/http/`
**Responsibility**: HTTP request/response handling
**Dependencies**: Depends on Service layer (inward)

**Components**:
- `health_handler.go` - Health check endpoints (readiness and liveness)
- `user_handler.go` - User authentication & profile
- `post_handler.go` - Post CRUD operations
- `response/` - JSend response formatting utilities

**Rules**:
- No business logic in handlers
- Only HTTP-specific concerns (parsing, validation, response formatting)
- Depends on Service interfaces (ports)

### 2. Service Layer (Core Business Logic)
**Location**: `internal/core/service/`
**Responsibility**: Orchestrate business workflows
**Dependencies**: Depends on Repository layer (inward)

**Components**:
- `user_service.go` - User registration, login, profile
- `post_service.go` - Post creation, listing, retrieval
- `health_service.go` - Health checks with database connectivity verification

**Rules**:
- Contains business rules and logic
- Coordinates between multiple repositories
- Depends on repository interfaces (ports)
- No HTTP or database knowledge

### 3. Repository Layer (Data Access)
**Location**: `internal/adapter/storage/postgres/`
**Responsibility**: Data persistence and retrieval
**Dependencies**: Depends on Domain and Infrastructure

**Components**:
- `user_repo.go` - User data access
- `post_repo.go` - Post data access
- `health_repo.go` - Database health check and connectivity verification
- `*_test.go` - Repository tests

**Rules**:
- Implements repository interfaces (ports)
- Contains SQL queries and database operations
- Uses sqlc-generated code
- No business logic, only data access

### 4. Domain Layer (Core Business)
**Location**: `internal/core/domain/`
**Responsibility**: Business entities and rules
**Dependencies**: None (innermost layer)

**Components**:
- `user.go` - User entity definition
- `post.go` - Post entity definition
- `health.go` - Health status entity definition
- `errors/` - Domain-specific error types

**Rules**:
- Pure business logic, no external dependencies
- Defines entities, value objects, and domain events
- Framework-agnostic
- Contains business rules and invariants

## Dependency Flow
```
Handler → Service → Repository → Domain
          ↓
    Infrastructure (sqlc, config, etc.)
```

## Ports & Adapters

### Ports (Interfaces)
**Location**: `internal/core/port/`
**Purpose**: Define contracts for external dependencies

**Interfaces**:
- `service.go` - Service port definitions
- `repository.go` - Repository port definitions
- `service_health.go` - Health service port
- `repository_health.go` - Health repository port

**Rules**:
- Interfaces define what is needed, not how
- Inward-pointing dependencies
- Mockable for testing

### Adapters (Implementations)
**Purpose**: Implement ports with external systems

**HTTP Adapter**:
- Fiber handlers implementing service ports

**Storage Adapter**:
- PostgreSQL repositories implementing repository ports

## Infrastructure Layer
**Location**: `internal/infrastructure/`
**Responsibility**: Framework and external system integration

**Components**:
- `config/` - Configuration management
- `container/` - Dependency injection setup with hooks
- `server/` - Fiber server configuration
- `sqlc/` - Generated database code

## Dependency Injection
**Framework**: samber/do/v2
**Location**: `internal/infrastructure/container/di.go`

**Wiring**:
- Register concrete implementations
- Inject dependencies into handlers
- Manage lifecycle (singleton, transient)
- Modular registration via hooks

**Benefits**:
- Decouples code from concrete dependencies
- Enables easy testing with mocks
- Centralized dependency management
- Modular and conditional wiring

## Data Flow Examples

### User Registration Flow
```
1. HTTP POST /auth/register
   ↓
2. UserHandler.Register()
   ↓
3. UserService.Register()
   ↓
4. UserRepository.Create()
   ↓
5. PostgreSQL (via sqlc)
   ↓
6. Response flows back through layers
```

### Get Post Flow
```
1. HTTP GET /posts/:id
   ↓
2. PostHandler.GetByID()
   ↓
3. PostService.GetByID()
   ↓
4. PostRepository.FindByID()
   ↓
5. PostgreSQL (via sqlc)
   ↓
6. Response flows back through layers
```

## Error Handling Strategy
- Domain errors in `internal/core/domain/errors/`
- Repository errors wrapped with context
- Service errors propagated with context
- Handler converts to HTTP responses
- JSend format for all responses

## Configuration Management
**Pattern**: Configuration struct with validation
**Location**: `internal/infrastructure/config/config.go`

**Loading Order**:
1. Defaults (hardcoded)
2. YAML config file
3. Environment variables

**Validation**:
- Struct tag validation (validator/v10)
- Custom business rules (e.g., production JWT secret check)

## Database Architecture
**ORM**: None (raw SQL with sqlc)
**Migration Tool**: golang-migrate
**Query Generator**: sqlc
**Connection**: lib/pq (native PostgreSQL driver)

**Benefits**:
- Type-safe queries
- No ORM complexity
- Explicit SQL control
- Better performance

## API Design
**Format**: RESTful JSON
**Response Format**: JSend {status, data, message}
**Documentation**: Swagger/OpenAPI 2.0

**Validation**:
- Request validation middleware
- DTO validation with custom tags
- Error aggregation

## Security Architecture
**Authentication**: JWT tokens
**Middleware**:
- Auth middleware (token validation)
- CORS middleware
- Rate limiting
- Request ID
- Logging

**Password Security**:
- bcrypt hashing via golang.org/x/crypto
- Salt rounds configured

## Testing Architecture
**Strategy**: Test pyramid

**Unit Tests**:
- Domain entities (no dependencies)
- Services (with mock repositories)
- Repositories (with sqlmock)

**Integration Tests**:
- Full HTTP stack
- Real database connection
- `test/integration/` directory

**Mocking**:
- Generated mocks via mockery
- Manual mocks for special cases
- sqlmock for database testing

## Build & Deployment
**Binary**: Compiled to `./bin/service`
**Build Tags**: Conditional compilation for modular builds
**Docker**: Multi-stage Dockerfile with non-root user
**Compose**: Development stack with PostgreSQL

**Entry Point**: `cmd/server/main.go`
**Server**: Fiber HTTP server
**Port**: Configurable (default 3000)

## Observability
**Logging**:
- Structured logging (zerolog/slog)
- Request ID correlation
- Contextual log fields
- Multiple levels (debug, info, warn, error)

**Health Checks**:
- `/health` - Readiness probe
- `/health/live` - Liveness probe

**Metrics**: Ready for OpenTelemetry integration

## Module Boundaries
**Strict Separation**:
- No circular dependencies
- Inner layers cannot depend on outer layers
- Interfaces define contracts
- Dependencies point inward
- Build tags control module inclusion

**Communication**:
- Layer-to-layer via interfaces
- No skip-layer dependencies
- Explicit dependency injection
- Modular registration via hooks

## Code Organization Rules
- One file per domain entity
- Clear package naming
- Small, focused functions
- Explicit error handling
- No global state
- Composition over inheritance
- Modular route organization

## Scalability Considerations
- Stateless service design
- Database connection pooling
- JWT for horizontal scaling
- UUIDv7 for distributed IDs
- Rate limiting for protection
- Build tags for selective deployment
- Prepared for caching layer

## Future Extensibility
**Pattern**: Open/Closed Principle
- Add new endpoints without modifying existing
- Add new services following patterns
- Swap implementations via DI
- New databases via repository pattern
- New protocols via adapter pattern
- New modules via build tags
