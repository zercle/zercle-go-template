# Operational Workflows

## Test-Driven Development (TDD)

### TDD Cycle
1. **Red:** Write a failing test for the desired behavior
2. **Green:** Write minimal code to make the test pass
3. **Refactor:** Improve code while keeping tests green

### When to Write Tests
- **Before implementing:** New features or business logic
- **Before fixing bugs:** Reproduce the bug with a test
- **After refactoring:** Ensure behavior unchanged
- **Critical paths:** Authentication, authorization, data persistence

### Test Organization

**Unit Tests:**
- Location: Same package as implementation (`*_test.go`)
- Scope: Single function or method
- Dependencies: Mock external dependencies
- Examples: `internal/domain/user/usecase/usecase_test.go`

**Integration Tests:**
- Location: `test/integration/`
- Scope: End-to-end API flows
- Dependencies: Real database (testcontainers)
- Examples: `test/integration/api_test.go`

**Mock Tests:**
- Location: `test/mock/`
- Scope: Database interactions
- Dependencies: SQL mocks
- Examples: `test/mock/sqlmock_test.go`

### Test Structure Template

```go
func Test<FunctionName>_<Scenario>_<ExpectedResult>(t *testing.T) {
    // Arrange
    ctx := context.Background()
    mockRepo := NewMockUserRepository(ctrl)
    useCase := NewUserUseCase(mockRepo, cfg, argon2Cfg, log)
    
    // Setup expectations
    mockRepo.EXPECT().GetByEmail(ctx, email).Return(nil, repository.ErrUserNotFound)
    
    // Act
    result, err := useCase.Register(ctx, request)
    
    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, result)
    assert.NotEmpty(t, result.Token)
}
```

### Table-Driven Tests

```go
func TestValidateEmail(t *testing.T) {
    tests := []struct {
        name    string
        email   string
        wantErr bool
    }{
        {"valid email", "user@example.com", false},
        {"invalid format", "invalid", true},
        {"empty", "", true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateEmail(tt.email)
            if (err != nil) != tt.wantErr {
                t.Errorf("ValidateEmail() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### Running Tests

**All tests:**
```bash
go test ./...
```

**Specific package:**
```bash
go test ./internal/domain/user/usecase/
```

**With coverage:**
```bash
go test -cover ./...
```

**Coverage report:**
```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

**Integration tests:**
```bash
go test -tags=integration ./test/integration/
```

### Test Coverage Goals
- **Critical business logic:** >90%
- **Domain use cases:** >80%
- **Handlers:** >70%
- **Infrastructure:** >60%
- **Overall:** >70%

## Refactoring Procedures

### When to Refactor
- Code duplication detected
- Complex functions (>50 lines)
- God objects with too many responsibilities
- Poor naming or unclear intent
- Performance bottlenecks identified
- Adding new features becomes difficult

### Refactoring Checklist
- [ ] Ensure tests exist and pass
- [ ] Identify the smell/problem
- [ ] Plan the refactoring approach
- [ ] Make small, incremental changes
- [ ] Run tests after each change
- [ ] Verify behavior unchanged
- [ ] Update documentation if needed
- [ ] Commit with clear message

### Common Refactorings

**Extract Method:**
- Move code to a new function
- Give it a descriptive name
- Replace original code with function call

**Extract Interface:**
- Identify common behavior
- Create interface with methods
- Implement interface in concrete types
- Update dependencies to use interface

**Replace Magic Numbers:**
- Identify constants in code
- Create named constants
- Replace numbers with constants
- Add documentation

**Simplify Conditional:**
- Use guard clauses
- Replace nested if-else with switch
- Extract complex conditions to named functions

**Remove Dead Code:**
- Identify unused code
- Remove or comment out
- Run tests to verify
- Commit removal

### Refactoring Example

**Before:**
```go
func (h *UserHandler) Register(c echo.Context) error {
    var req request.RegisterUser
    if err := c.Bind(&req); err != nil {
        return c.JSON(http.StatusBadRequest, ErrorResponse{Message: err.Error()})
    }
    if err := h.validator.Struct(req); err != nil {
        return c.JSON(http.StatusBadRequest, ErrorResponse{Message: err.Error()})
    }
    // ... more code
}
```

**After:**
```go
func (h *UserHandler) Register(c echo.Context) error {
    req, err := h.bindAndValidateRequest(c)
    if err != nil {
        return h.errorResponse(c, http.StatusBadRequest, err)
    }
    // ... more code
}

func (h *UserHandler) bindAndValidateRequest(c echo.Context) (*request.RegisterUser, error) {
    var req request.RegisterUser
    if err := c.Bind(&req); err != nil {
        return nil, err
    }
    if err := h.validator.Struct(req); err != nil {
        return nil, err
    }
    return &req, nil
}
```

## Code Review Checklist

### General Review
- [ ] Code follows project coding standards
- [ ] Naming is clear and descriptive
- [ ] Functions are small and focused
- [ ] No code duplication
- [ ] Comments explain "why", not "what"
- [ ] No commented-out code left behind
- [ ] Proper error handling throughout
- [ ] Logging at appropriate levels

### Architecture Review
- [ ] Follows clean architecture principles
- [ ] Dependencies point inward
- [ ] Domain logic isolated from infrastructure
- [ ] Interfaces used for external dependencies
- [ ] No circular dependencies
- [ ] Proper separation of concerns

### Security Review
- [ ] Input validation on all user inputs
- [ ] SQL injection prevention (SQLC handles this)
- [ ] Authentication/authorization enforced
- [ ] Sensitive data not logged
- [ ] Secrets not hardcoded
- [ ] CORS properly configured
- [ ] Rate limiting applied

### Performance Review
- [ ] No N+1 query problems
- [ ] Database queries optimized
- [ ] Connection pooling configured
- [ ] No unnecessary allocations
- [ ] Efficient data structures used
- [ ] Caching considered where appropriate

### Testing Review
- [ ] Tests added for new functionality
- [ ] Tests cover edge cases
- [ ] Tests are readable and maintainable
- [ ] Mocks used appropriately
- [ ] Test coverage adequate
- [ ] Integration tests included for API changes

### Documentation Review
- [ ] Godoc comments on exported functions
- [ ] API documentation updated (Swagger)
- [ ] README updated if needed
- [ ] Architecture docs updated if major change
- [ ] Migration files documented

### Specific Domain Reviews

**User Domain:**
- [ ] Password hashing with Argon2id
- [ ] Email uniqueness enforced
- [ ] JWT token properly generated
- [ ] User ownership verified

**Task Domain:**
- [ ] Task ownership verified
- [ ] Status values validated
- [ ] Priority values validated
- [ ] Due date handling correct

**Database:**
- [ ] Migration files created
- [ ] SQLC queries updated
- [ ] Indexes added if needed
- [ ] Foreign keys defined

## Debugging Protocols

### Debugging Workflow

1. **Reproduce the Issue**
   - Get exact steps to reproduce
   - Identify affected environment
   - Gather error messages and logs
   - Note request/response data

2. **Gather Information**
   - Check application logs
   - Review database state
   - Examine request/response
   - Check configuration values

3. **Formulate Hypothesis**
   - Based on symptoms
   - Consider recent changes
   - Review related code
   - Check known issues

4. **Test Hypothesis**
   - Add logging to verify
   - Write reproduction test
   - Use debugger if needed
   - Isolate the problem

5. **Implement Fix**
   - Write minimal fix
   - Add tests for fix
   - Verify fix works
   - Check for side effects

### Debugging Tools

**Logging:**
```go
log.Debug("Processing request", "user_id", userID, "task_id", taskID)
log.Error("Failed to update task", "error", err, "task_id", taskID)
```

**Structured Logging:**
- Include request ID in all logs
- Use consistent field names
- Log at appropriate levels
- Include context for errors

**Error Inspection:**
```go
if err != nil {
    log.Error("Operation failed", 
        "error", err,
        "operation", "createUser",
        "email", req.Email)
    // Use errors.Is() and errors.As() for error checking
}
```

**Database Debugging:**
```bash
# Connect to database
psql -h localhost -U postgres -d postgres

# Check recent queries
SELECT * FROM pg_stat_statements ORDER BY total_time DESC LIMIT 10;

# Check connection pool
SELECT * FROM pg_stat_activity;
```

**HTTP Debugging:**
```bash
# Check API endpoint
curl -X GET http://localhost:3000/health

# With authentication
curl -X GET http://localhost:3000/api/v1/users \
  -H "Authorization: Bearer <token>"

# Verbose output
curl -v http://localhost:3000/api/v1/tasks
```

### Common Issues & Solutions

**Database Connection Issues:**
- Check database is running
- Verify connection string
- Check connection pool settings
- Review firewall rules

**Authentication Failures:**
- Verify JWT secret matches
- Check token expiration
- Validate token format
- Review middleware configuration

**Performance Issues:**
- Check database query performance
- Review connection pool settings
- Profile with pprof
- Check for N+1 queries

**Test Failures:**
- Run tests with verbose output
- Check test data setup
- Verify mock expectations
- Review test isolation

### Adding Debug Logging

**Before Production:**
```go
func (uc *userUseCase) Login(ctx context.Context, req request.LoginUser) (*userResponse.LoginResponse, error) {
    log.Debug("Login attempt", "email", req.Email)
    
    userModel, err := uc.repo.GetByEmail(ctx, req.Email)
    if err != nil {
        log.Error("User not found", "email", req.Email, "error", err)
        return nil, ErrInvalidCredentials
    }
    
    // ... rest of code
}
```

**Remove Before Production:**
- Remove debug-level logs
- Keep error and warn logs
- Ensure no sensitive data in logs

### Performance Debugging

**Enable Profiling:**
```go
import _ "net/http/pprof"

// Add to routes
e.GET("/debug/pprof/*", echo.WrapHandler(http.DefaultServeMux))
```

**Profile CPU:**
```bash
go tool pprof http://localhost:3000/debug/pprof/profile
```

**Profile Memory:**
```bash
go tool pprof http://localhost:3000/debug/pprof/heap
```

**Profile Goroutines:**
```bash
go tool pprof http://localhost:3000/debug/pprof/goroutine
```

### Integration Testing Debugging

**Run Single Test:**
```bash
go test -v -run TestLogin ./test/integration/
```

**Keep Database Running:**
```bash
# Add this to test
testcontainers.CleanupContainer(t, container)
```

**View Test Database:**
```bash
# Get container ID
docker ps

# Connect to test database
docker exec -it <container_id> psql -U postgres -d postgres
```

## Adding a New Domain

### Step-by-Step Process

1. **Create Domain Structure**
   ```
   internal/domain/<domain>/
     entity/
     handler/
     repository/
     usecase/
     request/
     response/
     mock/
     interface.go
   ```

2. **Define Entity**
   - Create entity in `entity/<domain>.go`
   - Add UUID primary key
   - Add timestamps (created_at, updated_at)
   - Add business logic methods

3. **Create Interface**
   - Define Repository, Service, Handler interfaces
   - Follow existing patterns
   - Use domain-specific types

4. **Implement Repository**
   - Create SQL queries in `sqlc/queries/<domain>.sql`
   - Run `sqlc generate` to create types
   - Implement repository interface
   - Handle errors appropriately

5. **Implement UseCase**
   - Create business logic
   - Define domain-specific errors
   - Implement validation rules
   - Add logging

6. **Implement Handler**
   - Create HTTP handlers
   - Map request/response DTOs
   - Handle errors
   - Register routes

7. **Add Tests**
   - Unit tests for usecase
   - Integration tests for API
   - Mock tests for repository

8. **Update Application**
   - Wire dependencies in `app.go`
   - Register routes
   - Update Swagger documentation

9. **Update Documentation**
   - Add to architecture.md
   - Update context.md
   - Add API examples

## Database Migration Workflow

### Creating a Migration

1. **Create Migration File**
   ```bash
   # Format: YYYYMMDD_NNN_description
   touch sqlc/migrations/20260101_003_add_orders_table.up.sql
   touch sqlc/migrations/20260101_003_add_orders_table.down.sql
   ```

2. **Write Up Migration**
   ```sql
   CREATE TABLE orders (
       id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
       user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
       total DECIMAL(10,2) NOT NULL,
       status VARCHAR(50) NOT NULL DEFAULT 'pending',
       created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
       updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
   );
   
   CREATE INDEX idx_orders_user_id ON orders(user_id);
   ```

3. **Write Down Migration**
   ```sql
   DROP INDEX IF EXISTS idx_orders_user_id;
   DROP TABLE IF EXISTS orders;
   ```

4. **Apply Migration**
   ```bash
   # Using migration tool (add to project)
   migrate -path sqlc/migrations -database "postgres://..." up
   ```

5. **Regenerate SQLC**
   ```bash
   sqlc generate
   ```

### Migration Best Practices
- Always write both up and down migrations
- Use transactions for complex changes
- Add indexes for foreign keys
- Consider data migration for schema changes
- Test migrations on development first
- Never modify existing migrations

## Running the Application

### Development
```bash
# Set environment
export SERVER_ENV=local

# Run with hot reload (add air to project)
air

# Or standard run
go run cmd/server/main.go
```

### Production
```bash
# Build
go build -o bin/server cmd/server/main.go

# Run
./bin/server
```

### Docker
```bash
# Build image
docker build -t zercle-go-template .

# Run container
docker run -p 3000:3000 \
  -e SERVER_ENV=prod \
  -e DATABASE_URL=... \
  zercle-go-template
```

### Docker Compose
```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop services
docker-compose down
```

## Common Commands

### Linting
```bash
# Run linter
golangci-lint run

# Fix issues
golangci-lint run --fix
```

### Formatting
```bash
# Format code
go fmt ./...

# Check formatting
go vet ./...
```

### Dependencies
```bash
# Tidy dependencies
go mod tidy

# Update dependencies
go get -u ./...

# Verify dependencies
go mod verify
```

### Documentation
```bash
# Generate Swagger docs
swag init -g cmd/server/main.go

# View Swagger UI
# Navigate to http://localhost:3000/swagger/index.html
```

### SQLC
```bash
# Generate SQLC code
sqlc generate

# Validate SQLC configuration
sqlc validate
```

## Environment Setup

### Prerequisites
- Go 1.24.0+
- PostgreSQL 12+
- Docker (optional, for containerized deployment)

### Local Development
1. Clone repository
2. Copy `.env.example` to `.env`
3. Configure database connection
4. Run migrations
5. Start application

### Database Setup
```bash
# Start PostgreSQL with Docker
docker run --name postgres \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=postgres \
  -p 5432:5432 \
  -d postgres:15

# Run migrations
# (Add migration tool to project)
```

### Seed Data
```bash
# Run seed script
./scripts/seed-db.sh
```
