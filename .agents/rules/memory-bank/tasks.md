# Zercle Go Template - Workflows & Guides

## Common Workflows

### Adding a New Domain

Follow this pattern when adding a new business domain:

```bash
# 1. Create domain directory structure
mkdir -p domain/{name}/{handler,usecase,repository,model,request,response,mock}

# 2. Create interface.go
cat > domain/{name}/interface.go << 'EOF'
package {name}

// Repository defines the data access operations for {name}.
type Repository interface {
    Create(ctx context.Context, model *Model) error
    GetByID(ctx context.Context, id string) (*Model, error)
    List(ctx context.Context) ([]*Model, error)
    Update(ctx context.Context, model *Model) error
    Delete(ctx context.Context, id string) error
}

// UseCase defines the business logic operations for {name}.
type UseCase interface {
    Create(ctx context.Context, req *request.Create) (*Model, error)
    GetByID(ctx context.Context, id string) (*Model, error)
    List(ctx context.Context) ([]*Model, error)
    Update(ctx context.Context, id string, req *request.Update) (*Model, error)
    Delete(ctx context.Context, id string) error
}

// Handler defines the HTTP handlers for {name}.
type Handler interface {
    Create(c echo.Context) error
    GetByID(c echo.Context) error
    List(c echo.Context) error
    Update(c echo.Context) error
    Delete(c echo.Context) error
}
EOF

# 3. Implement each layer following existing domain patterns
# 4. Add tests for each layer
# 5. Register routes in cmd/server/main.go
# 6. Generate mocks: go generate ./domain/{name}/...
```

### Adding a New API Endpoint

```go
// 1. Define request/response DTOs
// domain/{name}/request/{operation}.go
type CreateSomethingRequest struct {
    Field1 string `json:"field1" validate:"required"`
    Field2 int    `json:"field2" validate:"gte=0"`
}

// domain/{name}/response/{operation}.go
type SomethingResponse struct {
    ID     string `json:"id"`
    Field1 string `json:"field1"`
    Field2 int    `json:"field2"`
}

// 2. Add use case method
// domain/{name}/usecase/usecase.go
func (uc *useCaseImpl) CreateSomething(ctx context.Context, req *request.CreateSomethingRequest) (*model.Something, error) {
    // Validate business rules
    // Call repository
    // Return result
}

// 3. Add handler method
// domain/{name}/handler/handler.go
func (h *handlerImpl) CreateSomething(c echo.Context) error {
    var req request.CreateSomethingRequest
    if err := c.Bind(&req); err != nil {
        return response.Error(c, http.StatusBadRequest, "Invalid request body")
    }
    
    if err := c.Validate(req); err != nil {
        return response.ValidationError(c, err)
    }
    
    model, err := h.useCase.CreateSomething(context.Background(), &req)
    if err != nil {
        return response.Error(c, http.StatusInternalServerError, err.Error())
    }
    
    return response.Success(c, http.StatusCreated, model)
}

// 4. Register route in cmd/server/main.go
api.POST("/something", handler.CreateSomething)
```

### Adding Database Queries

```sql
-- 1. Create SQL query file
-- sql/query/{table}/{operation}.sql

-- name: GetSomethingByID :one
SELECT * FROM something
WHERE id = $1;

-- name: ListSomething :many
SELECT * FROM something
WHERE deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CreateSomething :one
INSERT INTO something (field1, field2)
VALUES ($1, $2)
RETURNING *;
```

```bash
# 2. Generate SQLC code
make generate
# or
go generate ./...

# 3. Use generated code in repository
import "github.com/zercle/zercle-go-template/infrastructure/sqlc/db"

func (r *repositoryImpl) GetByID(ctx context.Context, id string) (*Model, error) {
    row, err := r.queries.GetSomethingByID(ctx, id)
    if err != nil {
        return nil, err
    }
    return mapToModel(row), nil
}
```

### Database Migrations

```bash
# 1. Create migration files
# Timestamp format: YYYYMMDD_description
cat > sql/migration/20251226_add_new_table.up.sql << 'EOF'
CREATE TABLE new_table (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    field1 VARCHAR(255) NOT NULL,
    field2 INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_new_table_field1 ON new_table(field1);
EOF

cat > sql/migration/20251226_add_new_table.down.sql << 'EOF'
DROP INDEX IF EXISTS idx_new_table_field1;
DROP TABLE IF EXISTS new_table;
EOF

# 2. Run migrations (manual or via migration tool)
# Using docker-compose
podman-compose -f docker-compose.yml exec db psql -U postgres -d zercle_db -f sql/migration/20251226_add_new_table.up.sql
```

### Testing Guide

#### Writing Unit Tests

```go
func TestUseCase_CreateSomething(t *testing.T) {
    tests := []struct {
        name    string
        input   *request.CreateSomethingRequest
        setup   func(*MockRepository)
        wantErr bool
        err     error
    }{
        {
            name: "success",
            input: &request.CreateSomethingRequest{
                Field1: "test",
                Field2: 100,
            },
            setup: func(m *MockRepository) {
                m.CreateFunc = func(ctx context.Context, model *Model) error {
                    return nil
                }
            },
            wantErr: false,
        },
        {
            name: "validation error",
            input: &request.CreateSomethingRequest{
                Field1: "", // Invalid: empty
                Field2: 100,
            },
            setup:   func(m *MockRepository) {},
            wantErr: true,
            err:     ErrInvalidField,
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

#### Writing Integration Tests

```go
func TestCreateUserIntegration(t *testing.T) {
    // Setup test server
    testServer := setupTestServer(t)
    defer testServer.Close()

    // Test data
    requestBody := map[string]interface{}{
        "email":    "test@example.com",
        "password": "password123",
        "full_name": "Test User",
    }

    // Make request
    resp := testServer.POST("/api/v1/auth/register").
        WithJSON(requestBody).
        Expect()

    // Assertions
    resp.Status(http.StatusCreated)
    resp.JSON().Path("$.status").String().Equal("success")
    resp.JSON().Path("$.data.email").String().Equal("test@example.com")
}
```

### Debugging Guide

#### Common Issues

**1. Database Connection Failed**
```bash
# Check database is running
podman ps | grep postgres

# Check database logs
podman logs zercle-db

# Test connection
psql -h localhost -U postgres -d zercle_db
```

**2. SQLC Generated Code Issues**
```bash
# Regenerate SQLC code
make generate

# Check SQL query syntax
# Ensure query files have correct comments:
-- name: QueryName :one or :many
```

**3. Tests Failing**
```bash
# Run specific test
go test ./domain/user/usecase -v -run TestCreateUser

# Run with race detector
go test -race ./...

# Run with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

**4. Import Errors**
```bash
# Tidy dependencies
go mod tidy

# Verify dependencies
go mod verify

# Update dependencies
go get -u ./...
```

#### Debug Logging

```go
// Add debug logging in use case
log.Debug("Processing request",
    "action", "create_user",
    "email", req.Email,
)

// Add error logging with context
log.Error("Failed to create user",
    "error", err,
    "email", req.Email,
    "request_id", requestID,
)
```

## Development Workflow

### Initial Setup

```bash
# 1. Clone repository
git clone https://github.com/zercle/zercle-go-template.git
cd zercle-go-template

# 2. Install dependencies
make init

# 3. Generate SQLC code and mocks
make generate

# 4. Set environment variables
export SERVER_ENV=local
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=password
export DB_NAME=zercle_db

# 5. Start database
make docker-up

# 6. Run migrations
# Run migration SQL files manually or via migration tool

# 7. Start application
make run
```

### Daily Development

```bash
# Start database
make docker-up

# Run application in development mode
make dev

# Run tests
make test

# Run linter
make lint

# Format code
make fmt

# Generate code after SQL changes
make generate
```

### Before Committing

```bash
# 1. Format code
make fmt

# 2. Run linter
make lint

# 3. Run all tests
make test

# 4. Run tests with coverage
make test-coverage

# 5. Check test coverage
go tool cover -func=coverage.out

# 6. Build application
make build
```

## Code Review Checklist

### Functionality
- [ ] Feature works as specified
- [ ] Edge cases are handled
- [ ] Error handling is comprehensive
- [ ] Business rules are enforced

### Code Quality
- [ ] Code follows project structure
- [ ] Functions are small and focused
- [ ] Names are descriptive and clear
- [ ] No code duplication
- [ ] Proper error wrapping

### Testing
- [ ] Unit tests cover all paths
- [ ] Tests use table-driven pattern
- [ ] Mocks are used appropriately
- [ ] Coverage >80% for new code
- [ ] Integration tests if applicable

### Documentation
- [ ] Godoc comments on exports
- [ ] Comments explain "why" not "what"
- [ ] README updated if needed
- [ ] Swagger annotations added

### Security
- [ ] Input validation is present
- [ ] Sensitive data not exposed
- [ ] SQL injection prevented (SQLC)
- [ ] Authentication/authorization enforced

### Performance
- [ ] No unnecessary database queries
- [ ] Efficient data structures used
- [ ] Context timeouts set
- [ ] Connection pooling configured

## Troubleshooting Guide

### Build Issues

**Error: cannot find package**
```bash
go mod download
go mod tidy
```

**Error: SQLC generated files missing**
```bash
make generate
```

### Runtime Issues

**Error: connection refused**
- Check database is running: `make docker-up`
- Check connection string in configs/local.yaml
- Verify database credentials

**Error: duplicate key value violates unique constraint**
- Check if record already exists
- Ensure proper error handling in repository
- Return appropriate error to client

### Test Issues

**Tests failing with connection error**
```bash
# Start test database
podman-compose -f docker-compose.test.yml up -d

# Run tests
make test
```

**Coverage below threshold**
- Add tests for uncovered paths
- Use `go tool cover -html=coverage.out` to visualize
- Focus on critical business logic

## Performance Optimization

### Database Queries

```go
// Bad: N+1 query problem
for _, booking := range bookings {
    user, _ := repo.GetUser(ctx, booking.UserID)
}

// Good: Single query with JOIN
bookingsWithUsers, _ := repo.GetBookingsWithUsers(ctx)
```

### Caching Strategy (Future)

```go
// Add Redis caching layer
func (uc *useCaseImpl) GetByID(ctx context.Context, id string) (*Model, error) {
    // Check cache first
    if cached, found := uc.cache.Get(id); found {
        return cached, nil
    }
    
    // Get from database
    model, err := uc.repo.GetByID(ctx, id)
    if err != nil {
        return nil, err
    }
    
    // Store in cache
    uc.cache.Set(id, model, 5*time.Minute)
    
    return model, nil
}
```

## Deployment Workflow

### Production Build

```bash
# 1. Set production environment
export SERVER_ENV=prod

# 2. Build optimized binary
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bin/service ./cmd/server

# 3. Build Docker image
make docker-build

# 4. Tag image
docker tag zercle-go-template:latest zercle-go-template:v1.0.0

# 5. Push to registry
docker push registry.example.com/zercle-go-template:v1.0.0
```

### Environment Variables (Production)

```bash
SERVER_ENV=prod
DB_HOST=prod-db.example.com
DB_PORT=5432
DB_USER=app_user
DB_PASSWORD=<secure-password>
DB_NAME=zercle_prod
JWT_SECRET=<secure-jwt-secret>
JWT_EXPIRATION=3600
LOG_LEVEL=info
```

### Health Checks

```bash
# Liveness probe
curl http://localhost:3000/health

# Readiness probe
curl http://localhost:3000/readiness

# Expected response
{"status":"healthy"}
```

## Refactoring Guidelines

### When to Refactor
- Code duplication detected
- Function or method too long (>50 lines)
- Complex conditional logic
- Poor performance identified
- Test coverage difficult to maintain

### Refactoring Steps
1. Write tests for existing behavior
2. Make small, incremental changes
3. Run tests after each change
4. Update documentation
5. Commit with "refactor:" prefix

### Example Refactoring

**Before: Long function with nested conditionals**
```go
func (h *handlerImpl) CreateUser(c echo.Context) error {
    var req request.CreateUser
    if err := c.Bind(&req); err != nil {
        return response.Error(c, http.StatusBadRequest, "Invalid request")
    }
    if req.Email == "" {
        return response.Error(c, http.StatusBadRequest, "Email required")
    }
    if req.Password == "" {
        return response.Error(c, http.StatusBadRequest, "Password required")
    }
    // ... more validation
    // ... business logic
    return response.Success(c, http.StatusCreated, user)
}
```

**After: Extracted validation and business logic**
```go
func (h *handlerImpl) CreateUser(c echo.Context) error {
    var req request.CreateUser
    if err := c.Bind(&req); err != nil {
        return response.Error(c, http.StatusBadRequest, "Invalid request")
    }
    
    if err := h.validateCreateUserRequest(&req); err != nil {
        return err
    }
    
    user, err := h.useCase.CreateUser(context.Background(), &req)
    if err != nil {
        return h.handleError(c, err)
    }
    
    return response.Success(c, http.StatusCreated, user)
}

func (h *handlerImpl) validateCreateUserRequest(req *request.CreateUser) error {
    if req.Email == "" {
        return response.Error(c, http.StatusBadRequest, "Email required")
    }
    if req.Password == "" {
        return response.Error(c, http.StatusBadRequest, "Password required")
    }
    return nil
}
```

## Security Best Practices

### Input Validation
```go
// Always validate input
type CreateUserRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=8"`
}

func (h *handlerImpl) CreateUser(c echo.Context) error {
    var req CreateUserRequest
    if err := c.Bind(&req); err != nil {
        return response.Error(c, http.StatusBadRequest, "Invalid request")
    }
    
    if err := c.Validate(&req); err != nil {
        return response.ValidationError(c, err)
    }
    
    // Continue with validated input
}
```

### SQL Injection Prevention
```go
// SQLC prevents SQL injection with parameterized queries
// Queries are defined in .sql files
func (r *repositoryImpl) GetByEmail(ctx context.Context, email string) (*User, error) {
    // SQLC generates type-safe parameterized query
    return r.queries.GetUserByEmail(ctx, email)
}
```

### Password Security
```go
import "golang.org/x/crypto/bcrypt"

// Always hash passwords before storing
func HashPassword(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    return string(bytes), err
}

// Never log passwords
log.Info("User created", "email", email) // Good
log.Info("User created", "password", password) // BAD - Never do this
```

## Monitoring & Logging

### Structured Logging
```go
// Good: Structured logging with context
log.Info("User logged in",
    "user_id", userID,
    "email", email,
    "ip", clientIP,
    "user_agent", userAgent,
    "request_id", requestID,
)

// Bad: Unstructured logging
log.Infof("User %s logged in from %s", userID, clientIP)
```

### Error Logging
```go
// Always log errors with context
log.Error("Failed to create user",
    "error", err,
    "email", req.Email,
    "request_id", requestID,
    "retry_attempt", attempt,
)
```

## API Design Patterns

### RESTful Conventions
```go
// Resource naming
GET    /api/v1/users           // List users
GET    /api/v1/users/:id       // Get specific user
POST   /api/v1/users           // Create user
PUT    /api/v1/users/:id       // Update user (full)
PATCH  /api/v1/users/:id       // Update user (partial)
DELETE /api/v1/users/:id       // Delete user

// Nested resources
GET    /api/v1/users/:id/bookings          // List user's bookings
POST   /api/v1/users/:id/bookings          // Create booking for user
GET    /api/v1/bookings/:id/payments       // List payments for booking
```

### Pagination (Future Implementation)
```go
// Request
GET /api/v1/users?page=1&limit=20&sort=created_at&order=desc

// Response
{
    "status": "success",
    "data": {
        "users": [...],
        "pagination": {
            "page": 1,
            "limit": 20,
            "total": 100,
            "total_pages": 5
        }
    }
}
```

## Quick Reference

### Makefile Commands
```bash
make init           # Install dependencies
make generate       # Generate SQLC code and mocks
make build          # Build application
make run            # Run compiled binary
make dev            # Run in development mode
make test           # Run tests
make test-coverage  # Run tests with coverage
make lint           # Run linter
make fmt            # Format code
make clean          # Clean build artifacts
make docker-build   # Build Docker image
make docker-up      # Start containers
make docker-down    # Stop containers
```

### Environment Variables
```bash
SERVER_ENV          # Environment: local/dev/uat/prod
DB_HOST            # Database host
DB_PORT            # Database port (default: 5432)
DB_USER            # Database user
DB_PASSWORD        # Database password
DB_NAME            # Database name
JWT_SECRET         # JWT signing secret
JWT_EXPIRATION     # JWT expiration in seconds
LOG_LEVEL          # Log level: debug/info/warn/error
```

### Common Ports
- API Server: 3000
- PostgreSQL: 5432
- Swagger UI: http://localhost:3000/swagger/index.html

### Directory Quick Access
```bash
cd domain/user           # User domain
cd domain/booking        # Booking domain
cd infrastructure/db     # Database layer
cd sql/migration         # Database migrations
cd test/integration      # Integration tests