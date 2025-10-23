# System Architecture

## Architectural Style
Clean Architecture with Domain-Driven Design (DDD) principles.

## Layer Structure

### 1. Domain Layer (`internal/domain/`)
**Purpose:** Core business logic and entities, independent of infrastructure.

**Components per Domain:**
- `entity/` - Business entities with domain logic
- `interface.go` - Domain interfaces (Repository, Service, Handler)
- `repository/` - Repository implementations
- `usecase/` - Business logic and orchestration
- `handler/` - HTTP request handlers
- `request/` - DTOs for incoming requests
- `response/` - DTOs for outgoing responses
- `mock/` - Mock implementations for testing

**Current Domains:**
- `user/` - User authentication and profile management
- `task/` - Task management (example domain)

### 2. Infrastructure Layer (`internal/infrastructure/`)
**Purpose:** External concerns and technical implementations.

**Sub-packages:**
- `config/` - Configuration management with Viper
- `db/` - Database abstraction and factory
- `http/client/` - HTTP client (Resty wrapper)
- `logger/` - Structured logging (zerolog wrapper)
- `password/` - Password hashing (Argon2id)
- `sqlc/db/` - SQLC-generated database code

### 3. Application Layer (`internal/app/`)
**Purpose:** Application orchestration and dependency injection.

**Key Components:**
- `app.go` - Main application structure
- Dependency wiring
- Middleware setup
- Route registration
- Server lifecycle management

### 4. Entry Point (`cmd/server/`)
**Purpose:** Application bootstrap.

**Components:**
- `main.go` - Application initialization and startup

## Component Boundaries

### Domain Boundaries
- Each domain is self-contained with its own interfaces
- Domains communicate through well-defined interfaces
- No direct dependencies between domains
- Shared infrastructure through interfaces only

### Infrastructure Boundaries
- Infrastructure implements domain interfaces
- Domain layer depends on abstractions, not implementations
- Database access abstracted through repository pattern
- External services abstracted through client interfaces

## Data Flow

### Request Flow
```
Client → Handler → UseCase → Repository → Database
         ↓         ↓          ↓
    Request   Business    Data Access
    DTO       Logic       Layer
         ↓         ↓          ↓
    Response  Entity    SQLC Query
    DTO       Mapping    Generation
```

### Authentication Flow
1. User submits credentials to `/api/v1/auth/login`
2. Handler validates request DTO
3. UseCase retrieves user by email
4. Password verified using Argon2id
5. JWT token generated with user ID and email
6. Token returned in response

### Protected Route Flow
1. Request includes JWT in Authorization header
2. JWT middleware validates token
3. User ID extracted from token context
4. Handler processes request with user context
5. UseCase enforces ownership rules

## Module Interactions

### Dependency Injection
- Application layer creates all dependencies
- Dependencies passed through constructors
- Interfaces used for all external dependencies
- Mock implementations for testing

### Database Access Patterns
- SQLC generates type-safe queries
- Repository pattern abstracts database
- Transactions managed at repository level
- Connection pooling handled by pgx/v5

### Error Handling Strategy
- Domain-specific errors in usecase layer
- Repository errors wrapped with context
- HTTP status codes mapped in handler layer
- Structured error responses to clients

## Integration Points

### External Services
- PostgreSQL database (primary data store)
- Future: Redis (caching)
- Future: Message queues (async processing)

### API Integration
- RESTful API endpoints
- Swagger documentation at `/swagger/*`
- Health checks at `/health` and `/readiness`

## Scalability Considerations

### Horizontal Scaling
- Stateless application design
- JWT tokens (no session storage)
- Database connection pooling
- Rate limiting per client

### Vertical Scaling
- Configurable database connections
- Efficient query generation
- Connection pooling optimization
- Memory-efficient data structures

## Deployment Architecture

### Container Strategy
- Single container for application
- Docker Compose for local development
- Environment-specific configurations
- Health checks for orchestration

### Configuration Management
- YAML configuration files per environment
- Environment variable overrides
- Viper for configuration loading
- Type-safe configuration structs
