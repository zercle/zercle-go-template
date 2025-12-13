# Context - Zercle Go Fiber Template

## Project Identity
**Name**: Zercle Go Fiber Template
**Purpose**: Production-ready microservice template using Clean Architecture, DDD, and PostgreSQL
**Language**: Go 1.25.0
**Primary Framework**: Fiber v2.52.9 (high-performance HTTP framework)

## Recent Updates & Evolution
**Build Tags System**: Conditional compilation with build tags (`health`, `user`, `post`, `all`) for modular deployments
**Route Modularization**: Routes split into separate files (routes_base.go, routes_health.go, routes_user.go, routes_post.go)
**DI Hooks System**: Modular DI registration with hooks (di_health.go, di_hooks_*.go, di_post.go, di_user.go)
**Memory Bank**: Comprehensive documentation system for long-term project knowledge
**Feature Completion**: All features (Health, User, Post) fully implemented with Clean Architecture
**Test Coverage**: Complete test suite with 16 test files covering all layers (handler, service, repository)
**Project Cleanup**: Removed all .gitkeep placeholder files, final documentation updates

## Key Domain Concepts

### Entities
- **User**: Core user entity with UUIDv7 ID, name, email, password, timestamps
- **Post**: Post entity with UUIDv7 ID, title, content, user association, timestamps
- **HealthStatus**: Health check status entity with timestamp, database status, and error information

### Value Objects & Types
- **UUIDv7**: Time-sorted UUIDs for scalable primary keys and distributed systems
- **Time**: Standard Go time.Time for all timestamps
- **JWT Claims**: UserID in tokens for authentication context

### Domain Errors
Located in `internal/core/domain/errors/errors.go`
Custom error types with separation of domain errors from infrastructure errors
Wrapped errors with context at each layer

### Core Services
- **UserService**: User business logic (registration, authentication, profile management)
- **PostService**: Post business logic (create, list, retrieve by ID)
- **HealthService**: Health check logic (readiness, liveness, database connectivity)

### Repositories
- **UserRepository**: PostgreSQL-based user data access with sqlc-generated code
- **PostRepository**: PostgreSQL-based post data access with sqlc-generated code
- **HealthRepository**: Database health check and connectivity verification
- Support for soft deletes and hard deletes

### API Endpoints
- **Health**: `GET /health` (readiness), `GET /health/live` (liveness)
- **Auth**: `POST /auth/register`, `POST /auth/login`
- **Users**: `GET /users/me` (protected)
- **Posts**: `POST /posts`, `GET /posts`, `GET /posts/:id`

## Technology Stack

### Core
- **Fiber v2.52.9**: HTTP web framework (Express.js-inspired for Go)
- **PostgreSQL 18+**: Primary database with UUIDv7 support
- **Go 1.25.0**: Language version

### Database Tools
- **sqlc v1.26+**: Type-safe SQL query compiler with prepared statements
- **golang-migrate**: Database migration tool with versioned files
- **UUIDv7**: Time-ordered UUID generation for distributed systems

### Infrastructure
- **samber/do/v2**: Dependency Injection container with centralized wiring
- **viper**: Configuration management (YAML + env vars with validation)
- **slog**: Structured logging (Go 1.21+)
- **zerolog**: High-performance structured logging with contextual fields
- **JWT v5**: Authentication tokens
- **bcrypt**: Password hashing via golang.org/x/crypto

### Validation & Security
- **go-playground/validator/v10**: Request validation with custom tags
- **CORS**: Configurable cross-origin policies
- **Rate Limiting**: Protection against DoS and abuse

### API & Documentation
- **Swagger/OpenAPI 2.0**: Auto-generated API docs via swaggo/swag
- **JSend Format**: Standardized JSON responses {status, data, message}

### Testing & Quality
- **testify**: Testing framework with assertions
- **go-sqlmock**: SQL mocking for tests
- **Mockery**: Mock generation via go.uber.org/mock/mockgen
- **golangci-lint**: Go linter with strict rules
- **race detection**: Always enabled in test runs

### DevOps
- **Docker & Docker Compose**: Containerization with multi-stage builds
- **GitHub Actions**: CI/CD pipeline with caching, test, lint, and Docker image build
- **GitHub Container Registry**: Registry integration
- **Non-root Containers**: Security best practice

## Configuration Layers (Precedence)
1. Hardcoded defaults
2. YAML config file (`configs/{env}.yaml`: dev, uat, prod)
3. Environment variables (highest priority)

## Key Environment Variables
- `SERVER_HOST`, `SERVER_PORT`, `SERVER_ENV`
- `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_DRIVER`
- `JWT_SECRET`, `JWT_EXPIRATION`
- `LOG_LEVEL`, `LOG_FORMAT`
- `CORS_ALLOWED_ORIGINS`
- `RATE_LIMIT_REQUESTS`, `RATE_LIMIT_WINDOW`

## Development Workflow
1. `make init` - Install dependencies
2. `make docker-up` - Start PostgreSQL
3. `make migrate-up` - Run migrations
4. `make generate` - Regenerate code (sqlc, mocks, swagger)
5. `make dev` - Run in development mode with hot reload
6. `make test` - Run tests with race detection
7. `make test-coverage` - Generate coverage report (minimum 80%)
8. `make lint` - Run static analysis
9. `make build` - Build binary with tags

## Build Tags System
**Tags**: `health`, `user`, `post`, `all`
**Purpose**: Conditional compilation for modular deployments
**Usage**: `go build -tags "health,user" ./cmd/server`
**Benefit**: Build minimal binaries with only required handlers
**Documentation**: See BUILD_TAGS.md for details

## Architecture Principles
- **Clean Architecture**: Strict layered separation with inward dependencies
- **DDD**: Domain-driven design with bounded contexts
- **Hexagonal/Ports & Adapters**: Dependency inversion for testability
- **Dependency Inversion**: Depend on abstractions, not concretions
- **SOLID Principles**: Applied throughout codebase
- **No ORM**: Direct SQL with sqlc for type safety and performance

## Critical File Locations
- **Entry Point**: `cmd/server/main.go`
- **Config**: `internal/infrastructure/config/config.go`
- **DI Container**: `internal/infrastructure/container/di.go`
- **DI Hooks**: `internal/infrastructure/container/di_*.go`
- **Server Setup**: `internal/infrastructure/server/server.go`
- **Routes**: `cmd/server/routes_*.go`
- **Domain Entities**: `internal/core/domain/`
- **Services**: `internal/core/service/`
- **Handlers**: `internal/adapter/handler/http/`
- **Repositories**: `internal/adapter/storage/postgres/`
- **DTOs**: `pkg/dto/`
- **Ports (Interfaces)**: `internal/core/port/`
- **SQL Queries**: `sql/queries/`
- **Migrations**: `sql/migrations/`
- **Build Tags Doc**: `BUILD_TAGS.md`

## Build Outputs
- **Binary**: `./bin/service`
- **Server**: http://localhost:3000
- **Swagger**: http://localhost:3000/swagger/index.html

## Container Stack
- PostgreSQL 16-alpine database (via docker-compose)
- Application service with health checks
- Bridge network for service communication
- Volumes for persistent database data and code mount

## Generated Code
- **sqlc**: Generates `internal/infrastructure/sqlc/*`
- **Mocks**: Generated in `test/mocks/` via mockery
- **Swagger**: Auto-generated in `docs/`

## Testing Strategy
**Test Pyramid**:
- **Unit Tests**: Domain (no dependencies), Service (mock repositories), Repository (sqlmock)
- **Handler Tests**: HTTP handling with httptest
- **Integration Tests**: Full HTTP stack with real database in `test/integration/`
- **Mocking**: Generated mocks for all interfaces
- **Coverage**: Minimum 80% with race detection
- **Commands**: `make test`, `make test-coverage`

**Test Files Created (16 total)**:
- **Health Feature**: health_handler_test.go, health_service_test.go, health_repo_test.go
- **User Feature**: user_handler_test.go, user_service_test.go, user_repo_test.go
- **Post Feature**: post_handler_test.go, post_service_test.go, post_repo_test.go
- **Existing**: auth_test.go, response_test.go, validator_test.go, integration/auth_test.go
- **Mock Files**: repository_mock.go, service_mock.go

## Security Features
- JWT-based stateless authentication
- bcrypt password hashing (configurable cost)
- Request validation middleware
- CORS configuration (explicit origins, not wildcard)
- Rate limiting middleware
- Non-root Docker containers
- Input validation at handler and middleware levels

## Logging & Observability
- Structured logging with zerolog and slog
- Request ID tracking for correlation
- Multiple log levels (debug, info, warn, error)
- JSON format (production), text format (development)
- Contextual log fields (request ID, user ID, HTTP method, path, status, duration)

## Migration Strategy
- Versioned migrations in `sql/migrations/` with timestamped files
- Tool: golang-migrate
- Up/down migration support
- Test migrations on copy before production

## Code Generation Triggers
- Database schema changes → `make generate`
- Interface changes → `make generate`
- API changes → Swagger auto-regeneration
- Command: `go generate ./...`

## Performance Optimizations
- **Database**: Configurable connection pooling, prepared statements (sqlc)
- **HTTP**: Configurable timeouts, keep-alive, compression (gzip/brotli)
- **Memory**: Object pooling for frequent allocations, minimal heap usage
- **Logging**: Fast structured logging with minimal allocations
- **Fiber**: High-performance HTTP handling

## Recent Changes (from git status)
- **Build Tags System**: Added conditional compilation with build tags for modular builds
- **Route Modularization**: Split routes into separate files for better organization
- **DI Hooks**: Modular DI registration with conditional hooks
- **Memory Bank**: Added comprehensive documentation system for project knowledge
- **Makefile Updates**: Enhanced with new targets and build tag support
- **Feature Refactoring**: Complete refactoring to feature-based architecture
- **Health Feature**: Fully implemented with Clean Architecture layers
- **Test Restoration**: Created comprehensive test suite (16 test files)
- **Cleanup**: Removed all .gitkeep placeholder files
- **Port Interfaces**: Added health-specific port interfaces
- **DI Wiring**: Complete dependency injection setup for all features

## Project Evolution
1. **Initial commit (0185abd)**: Bootstrap structure
2. **Second commit (da9257f)**: Bootstrap application with Fiber, core handlers, middleware
3. **Third commit (70f3044)**: Dev configuration, Memory Bank documentation, GEMINI documentation
4. **Fourth commit (e4f5bb8)**: Added dev configuration, Memory Bank documentation, BUILD_TAGS.md
5. **Feature Refactor**: Complete feature-based architecture with modular builds
6. **Health Implementation**: Full Clean Architecture implementation for Health feature
7. **Test Suite**: Comprehensive testing with 16 test files across all layers
8. **Final State**: Production-ready template with complete test coverage

## Final State (Current)
**Status**: ✅ Complete and Production-Ready
**Architecture**: Feature-based Clean Architecture with modular builds
**Features**: Health (complete), User (complete), Post (complete)
**Test Coverage**: Comprehensive test suite with 16 test files
**Build System**: Modular build tags system for selective deployment
**Documentation**: Memory Bank, README, and inline documentation complete
**Dependencies**: All dependencies managed and locked
**Quality**: Linting, testing, and validation configured
