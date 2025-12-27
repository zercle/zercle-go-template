# Zercle Go Template - Technical Standards

## Language & Runtime

### Go Version
- **Minimum**: Go 1.25.0
- **Target**: Latest stable Go 1.25.x release
- **Module**: `github.com/zercle/zercle-go-template`

### Dependency Management
```bash
# Add new dependency
go get github.com/package/name

# Update all dependencies
go get -u ./...

# Tidy dependencies
go mod tidy

# Verify dependencies
go mod verify
```

## Code Organization

### Standard Go Project Layout
```
zercle-go-template/
├── cmd/                    # Main applications
│   └── server/            # API server entry point
├── configs/               # Configuration files
├── domain/                # Business domains (clean architecture)
├── infrastructure/        # External dependencies
├── pkg/                   # Public libraries
├── sql/                   # Database schemas and queries
├── test/                  # Additional test utilities
├── .github/               # CI/CD workflows
├── docs/                  # Generated documentation
├── Dockerfile             # Container image
├── docker-compose.yml     # Local development orchestration
├── Makefile              # Build automation
├── go.mod                # Module definition
└── go.sum                # Dependency checksums
```

### Package Naming Conventions
- Use lowercase, single words when possible
- Avoid underscores or mixed caps
- Follow Go package naming guidelines
- Examples: `handler`, `usecase`, `repository`, `model`

### File Naming Conventions
- **Go files**: lowercase with underscores: `handler.go`, `usecase_test.go`
- **Test files**: append `_test.go`: `handler_test.go`
- **Generated files**: append `.sql.go` for SQLC: `users.sql.go`
- **Mock files**: `interface.go` for generated mocks

## Coding Standards

### Naming Conventions

#### Packages
```go
package handler      // Good
package handlers     // Avoid plural form
package Handler      // Never use capital letters
```

#### Constants
```go
const (
    maxRetries    = 3
    defaultTimeout = 30 * time.Second
    jwtExpiration = 3600
)
```

#### Variables
```go
var (
    errUserNotFound = errors.New("user not found")
    defaultConfig   = &Config{Port: 3000}
)
```

#### Functions
```go
// Good: descriptive, action-oriented
func GetUserByID(id string) (*User, error)
func ValidateRequest(req *Request) error

// Bad: vague, abbreviations
func GetUser(id string) (*User, error)  // What ID?
func Validate(req *Request) error       // Validate what?
```

#### Interfaces
```go
// Prefix with capability or action
type UserRepository interface {
    Create(ctx context.Context, user *User) error
    GetByID(ctx context.Context, id string) (*User, error)
    Update(ctx context.Context, user *User) error
    Delete(ctx context.Context, id string) error
}

// Single-method interfaces often end in -er
type Reader interface {
    Read(p []byte) (n int, err error)
}
```

#### Structs
```go
// Exported types have godoc comments
type User struct {
    ID        string    `json:"id" db:"id"`
    Email     string    `json:"email" db:"email" validate:"required,email"`
    Password  string    `json:"-" db:"password"`              // Never expose in JSON
    CreatedAt time.Time `json:"created_at" db:"created_at"`
    UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Unexported fields in lowercase
type user struct {
    internalField string
}
```

### Function Design

#### Function Length
- **Ideal**: < 20 lines
- **Maximum**: < 50 lines (extract if longer)
- Break complex functions into smaller, named helpers

#### Parameters
- **Ideal**: 1-3 parameters
- **Maximum**: 4 parameters (use struct for more)
```go
// Good
func CreateUser(ctx context.Context, user *User) error

// Too many parameters - use struct instead
func CreateUser(ctx context.Context, name, email, password, phone, address string) error

// Better
type CreateUserData struct {
    Name     string
    Email    string
    Password string
    Phone    string
    Address  string
}
func CreateUser(ctx context.Context, data *CreateUserData) error
```

#### Return Values
- Return errors as last parameter
- Prefer concrete types over interfaces for return values
- Use context.Context as first parameter for exported functions

```go
func (r *repository) GetByID(ctx context.Context, id string) (*User, error) {
    // Implementation
}
```

### Error Handling

#### Error Creation
```go
// Use fmt.Errorf for wrapping with context
if err != nil {
    return fmt.Errorf("failed to get user: %w", err)
}

// Define domain errors
var (
    ErrUserNotFound  = errors.New("user not found")
    ErrInvalidEmail  = errors.New("invalid email format")
    ErrDuplicateUser = errors.New("user already exists")
)
```

#### Error Checking
- **Always** check errors
- Never ignore errors: `func(), _`
- Handle errors immediately or propagate up

```go
// Good
user, err := repo.GetByID(ctx, id)
if err != nil {
    if errors.Is(err, ErrUserNotFound) {
        return nil, ErrUserNotFound
    }
    return nil, fmt.Errorf("failed to get user: %w", err)
}

// Bad
user, _ := repo.GetByID(ctx, id)
```

#### Error Messages
- Include context: what operation failed
- Include identifying values: IDs, names
- Lowercase first letter (not starting sentence)
- No trailing punctuation

```go
// Good
"failed to create user with email %s"
"database connection failed after %d retries"

// Bad
"Failed to create user."
"Error"
"null"
```

### Comments & Documentation

#### Godoc Comments
```go
// User represents a user account in the system.
// It contains authentication information and profile data.
type User struct {
    // ID is the unique identifier for the user.
    ID string `json:"id" db:"id"`
    
    // Email is the user's email address and must be unique.
    Email string `json:"email" db:"email" validate:"required,email"`
}

// Create inserts a new user into the database.
// It returns the created user with the generated ID.
// Returns ErrDuplicateUser if email already exists.
func (r *repository) Create(ctx context.Context, user *User) error {
    // Implementation
}
```

#### Comment Style
- Exported identifiers **must** have godoc comments
- Use complete sentences
- Start with the identifier name
- No blank line between comment and declaration
- Use present tense: "Create inserts" not "Creates"

```go
// Good
// GetUserByID retrieves a user by their unique identifier.
func GetUserByID(id string) (*User, error) {
    // ...
}

// Bad
// get user by id
func GetUserByID(id string) (*User, error) {
    // ...
}
```

#### When NOT to Comment
- Don't comment obvious code
- Don't repeat what the code says
- Don't comment out code (delete it instead)

```go
// Bad
// Increment count
count++

// Good - no comment needed
count++
```

### Struct Tags

#### JSON Tags
```go
type User struct {
    ID       string `json:"id"`                    // Standard field
    Email    string `json:"email"`                 // Standard field
    Password string `json:"-"`                     // Never expose
    FullName string `json:"full_name,omitempty"`   // Snake case with omitempty
}
```

#### DB Tags
```go
type User struct {
    ID        string    `json:"id" db:"id"`
    Email     string    `json:"email" db:"email"`
    CreatedAt time.Time `json:"created_at" db:"created_at"`
}
```

#### Validation Tags
```go
type User struct {
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=8"`
    Age      int    `json:"age" validate:"gte=0,lte=150"`
}
```

## Design Patterns

### Repository Pattern
```go
// Define interface in domain layer
type UserRepository interface {
    Create(ctx context.Context, user *User) error
    GetByID(ctx context.Context, id string) (*User, error)
    List(ctx context.Context, limit, offset int) ([]*User, error)
    Update(ctx context.Context, user *User) error
    Delete(ctx context.Context, id string) error
}

// Implement in repository package
type repositoryImpl struct {
    queries *sqlc.Queries
    log     *logger.Logger
}

func (r *repositoryImpl) Create(ctx context.Context, user *User) error {
    params := sqlc.CreateUserParams{
        Email:    user.Email,
        Password: user.Password,
        // ...
    }
    err := r.queries.CreateUser(ctx, params)
    if err != nil {
        return fmt.Errorf("failed to create user: %w", err)
    }
    return nil
}
```

### Dependency Injection
```go
// Constructor function
func Initialize(queries *sqlc.Queries, log *logger.Logger) Repository {
    return &repositoryImpl{
        queries: queries,
        log:     log,
    }
}

// Usage in main.go
userRepo := userRepository.Initialize(database.Queries(), log)
userUseCase := userUsecase.Initialize(userRepo, cfg, log)
userHandler := userHandler.Initialize(userUseCase, log)
```

### Context Usage
```go
// Always accept context as first parameter
func (r *repository) GetByID(ctx context.Context, id string) (*User, error) {
    // Pass context to database calls
    user, err := r.queries.GetUser(ctx, id)
    if err != nil {
        return nil, err
    }
    return &user, nil
}

// Create context with timeout
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

user, err := repo.GetByID(ctx, userID)
```

## Testing Standards

### Test Structure
```go
func TestCreateUser(t *testing.T) {
    tests := []struct {
        name    string
        input   *User
        setup   func(*MockRepository)
        wantErr bool
        err     error
    }{
        {
            name: "success",
            input: &User{
                Email:    "test@example.com",
                Password: "hashedpassword",
            },
            setup: func(m *MockRepository) {
                m.CreateUserFunc = func(ctx context.Context, user *User) error {
                    return nil
                }
            },
            wantErr: false,
        },
        {
            name: "duplicate email",
            input: &User{
                Email:    "test@example.com",
                Password: "hashedpassword",
            },
            setup: func(m *MockRepository) {
                m.CreateUserFunc = func(ctx context.Context, user *User) error {
                    return ErrDuplicateUser
                }
            },
            wantErr: true,
            err:     ErrDuplicateUser,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockRepo := &MockRepository{}
            tt.setup(mockRepo)

            useCase := NewUseCase(mockRepo, &config.Config{}, &logger.Logger{})

            err := useCase.Create(context.Background(), tt.input)

            if (err != nil) != tt.wantErr {
                t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
            }

            if tt.wantErr && !errors.Is(err, tt.err) {
                t.Errorf("Create() error = %v, want %v", err, tt.err)
            }
        })
    }
}
```

### Test Coverage
- **Target**: >80% overall coverage
- **Critical code**: >90% coverage (use cases, handlers)
- **Simple CRUD**: >70% coverage (repositories)
- Run tests: `make test` or `go test ./...`
- Coverage report: `make test-coverage`

### Test Naming
```go
// Test<Function><Scenario>
func TestGetUserByID_Success(t *testing.T) {}
func TestGetUserByID_NotFound(t *testing.T) {}
func TestGetUserByID_DatabaseError(t *testing.T) {}
```

## Database Standards

### SQLC Usage
```sql
-- name: GetUser :one
SELECT * FROM users
WHERE id = $1;

-- name: ListUsers :many
SELECT * FROM users
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CreateUser :one
INSERT INTO users (email, password, full_name)
VALUES ($1, $2, $3)
RETURNING *;
```

### Migration Files
```sql
-- 20251226_initialize_schema.up.sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 20251226_initialize_schema.down.sql
DROP TABLE users;
```

### Query Naming
- Use PascalCase for SQLC query names
- Prefix with action: Get, List, Create, Update, Delete
- Use descriptive names: GetUserByEmail, ListActiveUsers

## Configuration Standards

### Configuration Structure
```go
type Config struct {
    Server    ServerConfig
    Database  DatabaseConfig
    JWT       JWTConfig
    Logging   LoggingConfig
    CORS      CORSConfig
    RateLimit RateLimitConfig
}
```

### Environment Variables
- Use uppercase with underscores: `DB_HOST`, `JWT_SECRET`
- Provide defaults for local development
- Document all environment variables in README

## Logging Standards

### Log Levels
```go
log.Debug("Detailed debugging information", "key", value)
log.Info("Normal operational information", "event", "user_created")
log.Warn("Warning messages", "retry", attempt)
log.Error("Error conditions", "error", err)
```

### Structured Logging
```go
// Always include contextual fields
log.Info("User logged in",
    "user_id", userID,
    "request_id", requestID,
    "ip", clientIP,
)

// Never use string formatting for structured data
log.Info("User logged in", fmt.Sprintf("user_id=%s", userID))  // Bad

// Instead
log.Info("User logged in", "user_id", userID)  // Good
```

## API Standards

### HTTP Methods
```go
// CRUD operations
GET    /api/v1/users         // List users
GET    /api/v1/users/:id     // Get user by ID
POST   /api/v1/users         // Create user
PUT    /api/v1/users/:id     // Update user
DELETE /api/v1/users/:id     // Delete user
```

### Status Codes
```go
// Success
200 OK              // Successful GET, PUT, DELETE
201 Created         // Successful POST
204 No Content      // Successful DELETE with no response body

// Client Errors
400 Bad Request     // Invalid request data
401 Unauthorized    // Authentication required
403 Forbidden       // Insufficient permissions
404 Not Found       // Resource not found
409 Conflict        // Duplicate resource
422 Unprocessable   // Validation failed
429 Too Many Requests // Rate limit exceeded

// Server Errors
500 Internal Server Error // Unexpected error
503 Service Unavailable   // Service down
```

### JSend Response Format
```go
// Success response
type JSendSuccess struct {
    Status string      `json:"status"`
    Data   interface{} `json:"data"`
}

// Fail response (validation errors)
type JSendFail struct {
    Status string      `json:"status"`
    Message string     `json:"message"`
    Errors []FieldError `json:"errors,omitempty"`
}

// Error response (system errors)
type JSendError struct {
    Status  string `json:"status"`
    Message string `json:"message"`
}
```

## Security Standards

### Password Hashing
```go
import "golang.org/x/crypto/bcrypt"

func HashPassword(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    return string(bytes), err
}

func CheckPassword(password, hash string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    return err == nil
}
```

### JWT Token Generation
```go
import "github.com/golang-jwt/jwt/v5"

func GenerateToken(userID string, secret string, expiration int) (string, error) {
    claims := jwt.MapClaims{
        "user_id": userID,
        "exp":     time.Now().Add(time.Second * time.Duration(expiration)).Unix(),
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(secret))
}
```

### Input Validation
```go
import "github.com/go-playground/validator/v10"

type CreateUserRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=8,max=128"`
    FullName string `json:"full_name" validate:"required,min=2,max=255"`
}

func ValidateRequest(req interface{}) error {
    validate := validator.New()
    return validate.Struct(req)
}
```

## Performance Standards

### Connection Pooling
```go
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(5)
db.SetConnMaxLifetime(5 * time.Minute)
```

### Context Timeouts
```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
```

### Buffer Management
```go
// Preallocate slices with capacity
users := make([]*User, 0, expectedCount)

// Use bytes.Buffer for string building
var buf bytes.Buffer
buf.WriteString("Hello ")
buf.WriteString("World")
```

## Linting & Formatting

### golangci-lint Configuration
Located in `.golangci.yml`:
```yaml
linters:
  enable:
    - gofmt
    - govet
    - staticcheck
    - unused
    - gosimple
    - ineffassign
    - misspell
```

### Run Linters
```bash
make lint
# or
golangci-lint run
```

### Format Code
```bash
make fmt
# or
go fmt ./...
```

## Build & Deployment Standards

### Makefile Targets
```makefile
make init          # Install dependencies
make generate      # Generate SQLC code and mocks
make build         # Build binary
make test          # Run tests
make test-coverage # Run tests with coverage
make lint          # Run linters
make fmt           # Format code
make clean         # Clean build artifacts
```

### Docker Standards
```dockerfile
# Multi-stage build
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o bin/service ./cmd/server

FROM alpine:latest
COPY --from=builder /app/bin/service /service
CMD ["/service"]
```

## Best Practices Summary

### DO
- Use descriptive names
- Handle errors explicitly
- Write table-driven tests
- Add godoc comments
- Use dependency injection
- Follow Go idioms
- Validate inputs
- Log with context
- Use interfaces for mocking
- Keep functions small

### DON'T
- Ignore errors
- Use abbreviations in names
- Write long functions (>50 lines)
- Expose sensitive data in JSON
- Use global variables
- Repeat code (DRY principle)
- Ignore test coverage
- Use panic for normal errors
- Expose internal details
- Over-engineer simple problems