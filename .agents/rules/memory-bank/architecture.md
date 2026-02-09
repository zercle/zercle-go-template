# Architecture Documentation: Zercle Go Template

**Last Updated:** 2026-02-08  
**Status:** Production-Ready  
**Pattern:** Clean Architecture / Domain-Driven Design

---

## System Architecture Overview

### High-Level Design

The Zercle Go Template implements **Clean Architecture** (also known as Onion Architecture or Ports and Adapters), which emphasizes:

1. **Independence of Frameworks**: The core business logic doesn't depend on Echo, PostgreSQL, or any external tool
2. **Testability**: Business rules can be tested without UI, database, or external services
3. **Independence of UI**: The API can be replaced without changing business rules
4. **Independence of Database**: PostgreSQL can be swapped without affecting use cases
5. **Independence of External Services**: External dependencies are abstracted behind interfaces

### Layered Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        PRESENTATION LAYER                        │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────────┐  │
│  │   Handler   │  │ Middleware  │  │    Swagger/OpenAPI      │  │
│  │  (Echo HTTP)│  │(Auth/Logging│  │      Documentation      │  │
│  └──────┬──────┘  └─────────────┘  └─────────────────────────┘  │
└─────────┼────────────────────────────────────────────────────────┘
          │ DTOs / Requests / Responses
          ▼
┌─────────────────────────────────────────────────────────────────┐
│                         BUSINESS LAYER                           │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │                      USE CASE LAYER                      │    │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────────┐  │    │
│  │  │User Usecase │  │ AuthUsecase │  │ Future Usecases │  │    │
│  │  │(Create User)│  │(JWT Tokens) │  │  (Order, etc.)  │  │    │
│  │  └─────────────┘  └─────────────┘  └─────────────────┘  │    │
│  └─────────────────────────────────────────────────────────┘    │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │                      DOMAIN LAYER                        │    │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────────┐  │    │
│  │  │ User Entity │  │ JWT Claims  │  │ Domain Errors   │  │    │
│  │  │ (Business   │  │ (Token      │  │ (Typed Errors)  │  │    │
│  │  │   Rules)    │  │  Structure) │  │                 │  │    │
│  │  └─────────────┘  └─────────────┘  └─────────────────┘  │    │
│  └─────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────┘
          │ Domain Objects
          ▼
┌─────────────────────────────────────────────────────────────────┐
│                        DATA LAYER                                │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │                   REPOSITORY LAYER                       │    │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────────┐  │    │
│  │  │   SQLC      │  │   Memory    │  │  Interface      │  │    │
│  │  │Repository   │  │Repository   │  │  (Abstraction)  │  │    │
│  │  │(PostgreSQL) │  │  (Testing)  │  │                 │  │    │
│  │  └─────────────┘  └─────────────┘  └─────────────────┘  │    │
│  └─────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────┘
          ▼
┌─────────────────────────────────────────────────────────────────┐
│                     INFRASTRUCTURE LAYER                         │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐              │
│  │ PostgreSQL  │  │   Logger    │  │   Config    │              │
│  │  Database   │  │  (Zerolog)  │  │   (Viper)   │              │
│  └─────────────┘  └─────────────┘  └─────────────┘              │
└─────────────────────────────────────────────────────────────────┘
```

---

## Component Diagrams

### Request Flow

```
HTTP Request
     │
     ▼
┌──────────────────────────────────────────────────────────────┐
│  Echo Router                                                  │
│  - Middleware (Recovery, RequestID, Logging)                  │
└──────────────────────────────────────────────────────────────┘
     │
     ▼
┌──────────────────────────────────────────────────────────────┐
│  Handler (internal/feature/*/handler)                         │
│  - Parse request (Bind JSON)                                  │
│  - Validate input (validator)                                 │
│  - Call use case                                              │
│  - Format response                                            │
└──────────────────────────────────────────────────────────────┘
     │
     ▼
┌──────────────────────────────────────────────────────────────┐
│  Use Case (internal/feature/*/usecase)                        │
│  - Business logic                                             │
│  - Orchestrate repositories                                   │
│  - Domain rule enforcement                                    │
└──────────────────────────────────────────────────────────────┘
     │
     ▼
┌──────────────────────────────────────────────────────────────┐
│  Repository (internal/feature/*/repository)                   │
│  - Data access abstraction                                    │
│  - SQLC or In-memory implementation                           │
└──────────────────────────────────────────────────────────────┘
     │
     ▼
┌──────────────────────────────────────────────────────────────┐
│  Database (PostgreSQL)                                        │
│  - Migrations managed                                         │
│  - Connection pooling (pgx)                                   │
└──────────────────────────────────────────────────────────────┘
```

### Dependency Injection Flow

```
┌──────────────────────────────────────────────────────────────┐
│                    Container (internal/container)             │
│                                                               │
│   ┌─────────────┐    ┌─────────────┐    ┌─────────────┐      │
│   │   Config    │───▶│  Database   │───▶│ Repository  │      │
│   │   (Viper)   │    │   (pgx)     │    │   (SQLC)    │      │
│   └─────────────┘    └─────────────┘    └──────┬──────┘      │
│                                                │              │
│   ┌─────────────┐    ┌─────────────┐           │              │
│   │   Logger    │───▶│  Use Case   │◀──────────┘              │
│   │  (Zerolog)  │    │             │                          │
│   └─────────────┘    └──────┬──────┘                          │
│                             │                                 │
│   ┌─────────────┐           │                                 │
│   │   Handler   │◀──────────┘                                 │
│   │   (Echo)    │                                             │
│   └─────────────┘                                             │
│                                                               │
│   Lifecycle: Config → Infrastructure → Repos → Usecases → Handlers│
└──────────────────────────────────────────────────────────────┘
```

---

## Data Flow Descriptions

### User Creation Flow

```
1. Client sends POST /api/v1/users
   Body: { "email": "user@example.com", "password": "secure123", "name": "John" }

2. Handler Layer (user_handler.go)
   - Bind JSON to CreateUserRequest DTO
   - Validate struct (email format, password length, name presence)
   - Call userUsecase.CreateUser(ctx, req)

3. Use Case Layer (user_usecase.go)
   - Check if email already exists (repository.GetByEmail)
   - Hash password using bcrypt (cost 12)
   - Create domain.User entity
   - Save to repository (repository.Create)
   - Return created user

4. Repository Layer (sqlc_repository.go)
   - Map domain.User to SQLC params
   - Execute SQLC generated query
   - Map result back to domain.User
   - Handle unique constraint violations

5. Response Flow
   - Repository → Use Case → Handler
   - Handler converts domain.User to UserResponse DTO
   - Returns 201 Created with user data
```

### Authentication Flow

```
1. Client sends POST /api/v1/auth/login
   Body: { "email": "user@example.com", "password": "secure123" }

2. Handler Layer
   - Validate request body
   - Call userUsecase.Authenticate(ctx, email, password)

3. Use Case Layer
   - Fetch user by email (repository.GetByEmail)
   - Compare password with bcrypt
   - Return user on success, error on failure

4. JWT Generation (jwt_usecase.go)
   - Generate access token (15 min expiry)
   - Generate refresh token (7 day expiry)
   - Sign with HS256 algorithm

5. Response
   - Returns { user, access_token, refresh_token }
   - Client stores tokens for subsequent requests

6. Authenticated Request
   - Client sends Authorization: Bearer <token>
   - Auth middleware validates JWT signature and expiry
   - Extracts user ID from claims
   - Request proceeds to handler
```

---

## Design Patterns Used

### 1. Repository Pattern

**Purpose**: Abstract data access to enable testing and database swapping

```go
// Interface defines the contract
type UserRepository interface {
    Create(ctx context.Context, user *domain.User) error
    GetByID(ctx context.Context, id string) (*domain.User, error)
    // ...
}

// SQLC implementation
type sqlcRepository struct{ db *sqlc.Queries }

// In-memory implementation for testing
type memoryRepository struct{ users map[string]*domain.User }
```

### 2. Use Case Pattern (Application Service)

**Purpose**: Encapsulate business logic, coordinate repositories

```go
type UserUsecase interface {
    CreateUser(ctx context.Context, req dto.CreateUserRequest) (*domain.User, error)
    Authenticate(ctx context.Context, email, password string) (*domain.User, error)
}
```

### 3. Dependency Injection

**Purpose**: Loose coupling, testability

```go
// Container wires all dependencies
type Container struct {
    Config      *config.Config
    DB          *pgxpool.Pool
    UserRepo    repository.UserRepository
    UserUsecase usecase.UserUsecase
    // ...
}
```

### 4. Middleware Pattern

**Purpose**: Cross-cutting concerns (auth, logging, recovery)

```go
// Chain of responsibility
e.Use(middleware.Recover())      // Panic recovery
e.Use(middleware.RequestID())    // Correlation IDs
e.Use(loggingMiddleware)         // Request logging
```

### 5. DTO (Data Transfer Object)

**Purpose**: Decouple internal models from API contracts

```go
// Request/Response objects separate from domain
type CreateUserRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=8"`
}
```

### 6. Factory Pattern

**Purpose**: Create objects with dependencies

```go
func NewUserHandler(uc UserUsecase, jwt JWTUsecase, log Logger) *UserHandler {
    return &UserHandler{userUsecase: uc, jwtUsecase: jwt, logger: log}
}
```

---

## Technical Decisions & Rationale

### 1. SQLC Over ORM

**Decision**: Use SQLC for type-safe SQL instead of GORM/Ent

**Rationale**:
- Compile-time query validation
- No runtime reflection overhead
- Plain SQL—no query DSL to learn
- Easy to optimize queries
- Smaller binary size

**Trade-off**: Manual query writing vs. automatic migration generation

### 2. Clean Architecture Over Simple MVC

**Decision**: Strict layer separation with dependency inversion

**Rationale**:
- Business logic is framework-agnostic
- Easy to test without HTTP or database
- Can swap Echo for another framework
- Clear boundaries prevent spaghetti code

**Trade-off**: More boilerplate, steeper learning curve

### 3. UUID Over Auto-Increment IDs

**Decision**: Use UUID v4 for all entity IDs

**Rationale**:
- Distributed system friendly
- No ID enumeration attacks
- Safe to generate client-side if needed
- No coordination required

**Trade-off**: Slightly more storage, not sortable by creation time

### 4. Bcrypt Over Argon2/scrypt

**Decision**: Use bcrypt for password hashing

**Rationale**:
- Battle-tested, widely supported
- Go standard library support (golang.org/x/crypto)
- Adjustable cost factor
- Resistant to GPU/ASIC attacks

**Trade-off**: Memory-hard algorithms (Argon2) are more future-proof

### 5. Zerolog Over Standard Library

**Decision**: Use Zerolog for structured logging

**Rationale**:
- Zero-allocation JSON logging
- Structured logs for log aggregation
- Contextual logging support
- Multiple output formats (JSON/console)

**Trade-off**: External dependency vs. standard library

### 6. Feature-Based Organization

**Decision**: Organize by feature, not by layer

```
internal/feature/user/        # Everything user-related
  ├── domain/                 # User entity
  ├── dto/                    # User DTOs
  ├── handler/                # User handlers
  ├── repository/             # User repositories
  └── usecase/                # User use cases
```

**Rationale**:
- Cohesion: related files are together
- Easier navigation
- Clear feature boundaries
- Supports microservice extraction

**Trade-off**: Some duplication of layer patterns across features

---

## Scalability Considerations

### Horizontal Scaling

```
┌──────────────────────────────────────────────────────────────┐
│                         Load Balancer                         │
│                     (NGINX / Cloud LB)                        │
└──────────────────────────────────────────────────────────────┘
     │              │              │              │
     ▼              ▼              ▼              ▼
┌─────────┐   ┌─────────┐   ┌─────────┐   ┌─────────┐
│ API Pod │   │ API Pod │   │ API Pod │   │ API Pod │
│  (Go)   │   │  (Go)   │   │  (Go)   │   │  (Go)   │
└────┬────┘   └────┬────┘   └────┬────┘   └────┬────┘
     │              │              │              │
     └──────────────┴──────────────┴──────────────┘
                         │
                         ▼
              ┌─────────────────────┐
              │   PostgreSQL        │
              │   (Primary-Replica) │
              └─────────────────────┘
```

**Statelessness**:
- JWT tokens contain all auth state
- No server-side sessions
- Any pod can handle any request

**Database Scaling**:
- Read replicas for GET operations
- Connection pooling (pgx default: max 100)
- Prepared statement caching

### Caching Strategy (Future)

```
┌─────────┐     ┌─────────────┐     ┌─────────────┐
│  Client │────▶│  API Server │────▶│    Redis    │
└─────────┘     └─────────────┘     │   (Cache)   │
                                    └──────┬──────┘
                                           │
                                    ┌──────┴──────┐
                                    │ PostgreSQL  │
                                    │  (Source)   │
                                    └─────────────┘
```

**Cache-Aside Pattern**:
- Check cache first
- On miss: query DB, populate cache
- Write-through for updates

### Performance Optimizations

| Area | Current | Future |
|------|---------|--------|
| Database | Connection pooling | Read replicas, query optimization |
| JSON | Standard encoding | EasyJSON or similar for hot paths |
| Logging | Async zerolog | Sampling for high throughput |
| Middleware | Standard Echo | Custom optimized versions |

---

## Security Architecture

### Defense in Depth

```
┌─────────────────────────────────────────────────────────────┐
│ Layer 1: Network                                             │
│ - HTTPS/TLS termination                                      │
│ - WAF (Web Application Firewall)                             │
│ - DDoS protection                                            │
└─────────────────────────────────────────────────────────────┘
     │
     ▼
┌─────────────────────────────────────────────────────────────┐
│ Layer 2: Application                                         │
│ - Input validation (validator)                               │
│ - Authentication (JWT)                                       │
│ - CORS configuration                                         │
│ - Rate limiting (future)                                     │
└─────────────────────────────────────────────────────────────┘
     │
     ▼
┌─────────────────────────────────────────────────────────────┐
│ Layer 3: Data                                                │
│ - SQL injection prevention (SQLC)                            │
│ - Password hashing (bcrypt)                                  │
│ - Sensitive data encryption at rest                          │
└─────────────────────────────────────────────────────────────┘
```

### Authentication Flow Security

- Tokens signed with HMAC-SHA256 (configurable secret)
- Short-lived access tokens (15 minutes)
- Refresh token rotation (future)
- Secure password storage (bcrypt cost 12)

---

**Related Documents**:
- [brief.md](brief.md) - Project overview
- [tech.md](tech.md) - Technology stack details
- [tasks.md](tasks.md) - Development workflows
