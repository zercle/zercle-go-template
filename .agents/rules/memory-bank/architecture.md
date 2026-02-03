# Zercle Go Template - Architecture Documentation

## System Overview

The Zercle Go Template implements **Clean Architecture** (also known as Hexagonal or Ports & Adapters) with clear dependency rules. The architecture ensures:

- **Business logic independence**: Domain layer has zero external dependencies
- **Testability**: All layers can be tested in isolation with mocks
- **Flexibility**: Swap infrastructure (database, HTTP framework) without touching business logic
- **Maintainability**: Changes are localized to specific layers

## Layer Structure

```
┌─────────────────────────────────────────────────────────┐
│                    External World                       │
│  (HTTP Requests, Database, External APIs, Logger)      │
└─────────────────────────────────────────────────────────┘
                          ▲
                          │ Depends On
┌─────────────────────────────────────────────────────────┐
│  Infrastructure Layer                                    │
│  • HTTP Handlers (Echo)                                  │
│  • Database Repositories (PostgreSQL, In-Memory)         │
│  • Middleware (Logging, Recovery, Auth)                  │
│  • Configuration (Viper)                                 │
└─────────────────────────────────────────────────────────┘
                          ▲
                          │ Depends On
┌─────────────────────────────────────────────────────────┐
│  Use Case Layer                                          │
│  • Business Orchestration                                │
│  • Transaction Boundaries                                │
│  • Authorization Rules                                   │
└─────────────────────────────────────────────────────────┘
                          ▲
                          │ Depends On
┌─────────────────────────────────────────────────────────┐
│  Domain Layer                                            │
│  • Entities (User, Auth)                                 │
│  • Business Rules                                        │
│  • Value Objects                                         │
│  • Domain Errors                                         │
└─────────────────────────────────────────────────────────┘
```

## Dependency Rule

**Dependencies always point inward.**

- Domain layer knows nothing about use cases, infrastructure, or external frameworks
- Use cases know about domain but not infrastructure details
- Infrastructure depends on use cases and domain
- External frameworks (Echo, pgx, etc.) only exist in infrastructure

## Component Architecture

### Dependency Injection Container

The [`internal/container/container.go`](internal/container/container.go) is the composition root:

```
Container
├── Config          (*viper.Viper)
├── Logger          (zerolog.Logger)
├── DB              (*pgxpool.Pool)
├── UserRepository  (UserRepository interface)
└── UserUsecase     (UserUsecase interface)
```

Uses **functional options pattern** for flexible configuration:

```go
container.New(
    container.WithConfig(cfg),
    container.WithLogger(logger),
    container.WithPostgresRepository(db), // or WithMemoryRepository()
)
```

### Feature Module Structure

Each feature is self-contained:

```
internal/feature/user/
├── domain/
│   └── user.go              # User entity, validation rules
├── dto/
│   └── user.go              # CreateUserRequest, UpdateUserRequest
├── handler/
│   └── user_handler.go      # HTTP handlers, request binding
├── repository/
│   ├── user_repository.go   # Interface definition
│   ├── sqlc_repository.go   # PostgreSQL implementation
│   └── memory_repository.go # In-memory implementation
├── usecase/
│   └── user_usecase.go      # Business logic
└── user.go                  # Public exports
```

## Data Flow

### Request Flow (Create User)

```
HTTP Request
     ↓
[Echo Router] ──► [Auth Middleware] ──► [UserHandler.CreateUser()]
                                              ↓
                                      [Bind & Validate DTO]
                                              ↓
                                      [UserUsecase.CreateUser()]
                                              ↓
                                      [Business Logic & Validation]
                                              ↓
                                      [UserRepository.Create()]
                                              ↓
                                      [sqlc.Queries.CreateUser()]
                                              ↓
                                      [PostgreSQL]
                                              ↓
                                      [Response DTO]
                                              ↓
                                      [JSON Response]
```

### Authentication Flow

```
Login Request
     ↓
[AuthHandler.Login()]
     ↓
[Validate Credentials]
     ↓
[Generate JWT Pair]
     ├── Access Token (short-lived)
     └── Refresh Token (long-lived)
     ↓
[Return Tokens]

Protected Request
     ↓
[Auth Middleware]
     ↓
[Extract Bearer Token]
     ↓
[Validate JWT]
     ↓
[Set User Context]
     ↓
[Handler Execution]
```

## Design Patterns

### 1. Repository Pattern

Abstracts data access behind interfaces:

```go
// Interface in domain layer
type UserRepository interface {
    GetByID(ctx context.Context, id string) (*domain.User, error)
    GetByEmail(ctx context.Context, email string) (*domain.User, error)
    Create(ctx context.Context, user *domain.User) error
    Update(ctx context.Context, user *domain.User) error
    Delete(ctx context.Context, id string) error
}
```

**Implementations:**
- `SqlcUserRepository`: Production PostgreSQL via sqlc
- `MemoryUserRepository`: In-memory map for testing

### 2. Dependency Injection

Constructor injection throughout:

```go
// Usecase receives repository interface
type UserUsecase struct {
    userRepo UserRepository
    logger   logger.Logger
}

func NewUserUsecase(userRepo UserRepository, logger logger.Logger) *UserUsecase {
    return &UserUsecase{userRepo: userRepo, logger: logger}
}
```

### 3. Functional Options

Flexible container configuration:

```go
func WithPostgresRepository(db *pgxpool.Pool) ContainerOption {
    return func(c *Container) error {
        c.userRepo = repository.NewSqlcUserRepository(db)
        return nil
    }
}
```

### 4. Middleware Chain

Echo middleware for cross-cutting concerns:

```
Request
  ↓
[Recovery]      ← Catch panics, return 500
  ↓
[Logging]       ← Structured request logging
  ↓
[Auth]          ← JWT validation (protected routes)
  ↓
[Rate Limit]    ← Future: request throttling
  ↓
Handler
```

### 5. DTO Pattern

Separate domain entities from API contracts:

```go
// Domain entity - internal business logic
type User struct {
    ID           string
    Email        string
    PasswordHash string
    CreatedAt    time.Time
}

// DTO - API contract
type CreateUserRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=8"`
}
```

## Technical Decisions

### Why Clean Architecture?

| Alternative | Decision | Rationale |
|-------------|----------|-----------|
| MVC | Rejected | Business logic mixed with framework code |
| Layered (n-tier) | Rejected | Database drives design, hard to test |
| Clean Architecture | Selected | Framework agnostic, highly testable |

### Why sqlc over ORM?

| Aspect | sqlc | GORM |
|--------|------|------|
| Type Safety | Compile-time | Runtime reflection |
| Performance | Raw SQL speed | Reflection overhead |
| Complexity | Low | High (magic methods) |
| SQL Control | Full | Limited |

### Why Echo?

- **Performance**: Faster than Gin in benchmarks
- **Middleware**: Excellent middleware chain support
- **Validation**: Built-in request validation
- **Documentation**: Good Swagger integration

## Scalability Considerations

### Horizontal Scaling

Current state is stateless:
- JWT authentication (no server-side sessions)
- No in-memory state in handlers
- Database is external

**Ready for:**
- Multiple container instances behind load balancer
- Kubernetes deployment
- Auto-scaling based on CPU/memory

### Database Scaling

Current: Single PostgreSQL instance

**Future paths:**
1. **Read replicas**: Route read queries to replicas
2. **Connection pooling**: pgx already implements pooling
3. **Caching layer**: Redis for hot data
4. **Sharding**: By tenant or user ID ranges

### Caching Strategy (Future)

```
Request
  ↓
[Cache Check] ──► Hit ──► Return cached
  ↓ Miss
[Database]
  ↓
[Cache Write]
  ↓
[Response]
```

## Security Architecture

### Authentication

```
┌─────────────────────────────────────┐
│         Client Application          │
│  (Web, Mobile, Third-party)         │
└─────────────────────────────────────┘
              │
              ▼ Bearer Token
┌─────────────────────────────────────┐
│         API Gateway / LB            │
│         (SSL Termination)           │
└─────────────────────────────────────┘
              │
              ▼
┌─────────────────────────────────────┐
│         Auth Middleware             │
│  • Extract JWT from header          │
│  • Validate signature               │
│  • Check expiration                 │
│  • Set user context                 │
└─────────────────────────────────────┘
              │
              ▼
┌─────────────────────────────────────┐
│         Protected Handler           │
│  • Access user ID from context      │
│  • Enforce resource ownership       │
└─────────────────────────────────────┘
```

### Data Protection

- **Passwords**: bcrypt hashing (adaptive, slow)
- **Tokens**: JWT with HS256 or RS256
- **Transport**: TLS 1.3 required
- **Validation**: Input validation at handler layer

## Monitoring Points

For future observability:

| Layer | Metric | Implementation |
|-------|--------|----------------|
| HTTP | Request duration, rate, errors | Echo middleware |
| Business | Usecase execution time | Span in usecase |
| Database | Query duration, pool stats | pgx metrics |
| System | Memory, goroutines | runtime metrics |
