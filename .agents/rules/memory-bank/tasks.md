# Task Workflows - Zercle Go Fiber Template

## Development Workflows

### 1. Initial Project Setup
**Purpose**: Set up development environment from scratch

**Steps**:
```bash
# 1. Clone and navigate
git clone <repo-url>
cd zercle-go-template

# 2. Initialize dependencies
make init

# 3. Install development tools
make install-tools

# 4. Setup environment
cp .env.example .env
# Edit .env with your values

# 5. Start infrastructure
make docker-up

# 6. Run migrations
make migrate-up

# 7. Generate code
make generate

# 8. Run tests
make test

# 9. Start development
make dev
```

**Expected Outcome**: Running application at http://localhost:3000 with Swagger at http://localhost:3000/swagger

---

### 2. Daily Development
**Purpose**: Regular development workflow

**Steps**:
```bash
# Before coding
make lint
make test

# During coding (frequently)
make fmt
make test

# After coding
make test-coverage
make lint
make generate  # if SQL or interfaces changed

# Test in dev mode
make dev
```

**Expected Outcome**: Code passes all checks and tests

---

### 3. Adding a New Feature
**Purpose**: Extend the application with new functionality

**Steps**:

#### A. Create Domain Entity
1. Define entity in `internal/core/domain/`
2. Add domain errors if needed
3. Write unit tests for domain logic

#### B. Define Repository Interface
1. Create interface in `internal/core/port/repository.go`
2. Add CRUD methods
3. Write SQL queries in `sql/queries/`
4. Run `make generate`

#### C. Implement Repository
1. Create implementation in `internal/adapter/storage/postgres/`
2. Use sqlc-generated code
3. Write unit tests with sqlmock

#### D. Define Service Interface & Implementation
1. Add interface in `internal/core/port/service.go`
2. Create service in `internal/core/service/`
3. Inject repository via DI
4. Write unit tests with mock repository

#### E. Create DTOs
1. Create request DTOs in `pkg/dto/`
2. Add validation tags
3. Create response DTOs

#### F. Implement HTTP Handler
1. Create handler in `internal/adapter/handler/http/`
2. Inject service via DI
3. Implement endpoints
4. Add Swagger annotations
5. Write integration tests

#### G. Register in DI Container
1. Update `internal/infrastructure/container/di.go`
2. Register repository and service
3. Register handler

#### H. Update Migrations
1. Create migration in `sql/migrations/`
2. Add schema changes
3. Test migration up/down

#### I. Final Checks
```bash
make generate  # Regenerate code
make test      # Run all tests
make lint      # Lint code
make test-coverage  # Check coverage
```

**Expected Outcome**: New feature fully integrated with tests and documentation

---

### 4. Database Migration Workflow
**Purpose**: Add or modify database schema

**Steps**:
```bash
# 1. Create new migration file
# Format: YYYYMMDDHHMMSS_description.sql
touch sql/migrations/$(date +%Y%m%d%H%M%S)_add_feature.sql

# 2. Write UP migration
# sql/migrations/YYYYMMDDHHMMSS_add_feature.sql:
-- +migrate Up
CREATE TABLE ...;

# 3. Write DOWN migration
-- +migrate Down
DROP TABLE ...;

# 4. Test migration up
make migrate-up

# 5. Test migration down
make migrate-down

# 6. Re-run up
make migrate-up

# 7. Update sqlc queries if needed
# Edit files in sql/queries/

# 8. Regenerate code
make generate

# 9. Verify with tests
make test
```

**Expected Outcome**: Schema changes applied safely with rollback capability

---

### 5. Running Tests
**Purpose**: Execute test suite

**Steps**:
```bash
# Unit tests only
go test -v ./internal/core/domain/...
go test -v ./internal/core/service/...
go test -v ./internal/adapter/storage/...

# All tests
make test

# With race detection
go test -race ./...

# With coverage
make test-coverage
# Open coverage.html for detailed report

# Specific package
go test -v ./internal/adapter/handler/http/

# Verbose output
go test -v -count=1 ./...

# Integration tests
go test -v ./test/integration/...
```

**Expected Outcome**: All tests pass with acceptable coverage

---

### 6. Code Generation
**Purpose**: Regenerate code from SQL and interfaces

**Steps**:
```bash
# Generate all (sqlc, mocks, swagger)
make generate

# Or individually:
sqlc generate                    # SQL to Go
go generate ./...                # Mocks
swag init -g cmd/server/main.go  # Swagger
```

**Triggers for Generation**:
- SQL query changes in `sql/queries/`
- Interface changes in `internal/core/port/`
- Handler changes (for Swagger)
- API endpoint changes

**Expected Outcome**: Type-safe code generated and up-to-date

---

### 7. Linting & Formatting
**Purpose**: Ensure code quality

**Steps**:
```bash
# Format code
make fmt
# or: gofmt -s -w .

# Run linter
make lint
# or: golangci-lint run ./...

# Fix auto-fixable issues
golangci-lint run --fix ./...
```

**Expected Outcome**: Code passes all linting rules and is formatted

---

### 8. Building & Running
**Purpose**: Compile and execute application

**Steps**:
```bash
# Build binary
make build
# Output: ./bin/service

# Run compiled binary
./bin/service

# Run in development mode
make dev
# Uses: go run ./cmd/server

# Build Docker image
make docker-build

# Run with Docker Compose
make docker-up
make docker-down
```

**Expected Outcome**: Application running and accessible

---

### 9. Production Deployment
**Purpose**: Deploy to production

**Steps**:
```bash
# 1. Ensure production config
# Edit configs/prod.yaml or use environment variables

# 2. Validate configuration
# Verify JWT_SECRET is not default
# Verify database credentials
# Verify CORS origins

# 3. Build binary
make build

# 4. Run tests
make test

# 5. Build Docker image
make docker-build

# 6. Tag for registry
docker tag zercle-go-template:latest your-registry/zercle-go-template:v1.0.0

# 7. Push to registry
docker push your-registry/zercle-go-template:v1.0.0

# 8. Deploy to infrastructure
# (Kubernetes, ECS, etc.)

# 9. Run migrations
# Execute migrate-up on production database

# 10. Verify deployment
curl http://your-domain/health
```

**Expected Outcome**: Application running in production environment

---

## Testing Workflows

### 10. Writing Unit Tests
**Purpose**: Add tests for new code

**Domain Layer**:
```go
// Test in internal/core/domain/*_test.go
func TestUser_Validate(t *testing.T) {
    tests := []struct {
        name    string
        user    User
        wantErr bool
    }{
        {"valid user", User{...}, false},
        {"invalid email", User{...}, true},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateUser(tt.user)
            if (err != nil) != tt.wantErr {
                t.Errorf("ValidateUser() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

**Service Layer**:
```go
// Test in internal/core/service/*_test.go
func TestUserService_Register(t *testing.T) {
    // Setup mocks
    mockRepo := &MockUserRepository{}
    service := NewUserService(mockRepo)

    // Test
    user, err := service.Register(input)

    // Assertions
    require.NoError(t, err)
    assert.Equal(t, expectedUser, user)
}
```

**Repository Layer**:
```go
// Test in internal/adapter/storage/postgres/*_test.go
func TestUserRepository_Create(t *testing.T) {
    db, mock, _ := sqlmock.New()
    defer db.Close()

    repo := NewUserRepository(db)

    // Test
    err := repo.Create(user)

    // Assertions
    assert.NoError(t, err)
}
```

**Handler Layer**:
```go
// Test in internal/adapter/handler/http/*_test.go
func TestUserHandler_Register(t *testing.T) {
    app := fiber.New()
    handler := NewUserHandler(service)
    app.Post("/auth/register", handler.Register)

    // Test
    req := httptest.NewRequest("POST", "/auth/register", body)
    resp, _ := app.Test(req)

    // Assertions
    assert.Equal(t, 201, resp.StatusCode)
}
```

---

### 11. Integration Testing
**Purpose**: Test full stack integration

**Setup**:
```go
// test/integration/auth_test.go
func TestUserRegistration(t *testing.T) {
    // Setup test database
    db := setupTestDB()
    defer db.Close()

    // Setup DI
    container := di.NewContainer(db)
    app := setupFiberApp(container)

    // Test
    req := httptest.NewRequest("POST", "/auth/register", jsonBody)
    resp, body := app.Test(req)

    // Assertions
    assert.Equal(t, 201, resp.StatusCode)
    assert.Contains(t, body, "token")
}
```

---

## Debugging Workflows

### 12. Debugging Failed Tests
**Steps**:
```bash
# Run with verbose output
go test -v ./package/...

# Run with race detector
go test -race ./package/...

# Run single test
go test -v -run TestSpecific ./...

# Run with debugger
dlv test ./package/...

# Check coverage for specific test
go test -coverprofile=coverage.out ./package/...
go tool cover -func=coverage.out
```

---

### 13. Database Debugging
**Steps**:
```bash
# Connect to database
psql -h localhost -U postgres -d zercle_db

# Check migrations
migrate -path sql/migrations -database "postgres://..." version

# Check table structure
\dt
\d users

# Check indexes
\di

# Run query with EXPLAIN
EXPLAIN ANALYZE SELECT * FROM users WHERE email = 'test@example.com';
```

---

### 14. Performance Debugging
**Steps**:
```bash
# CPU profiling
go test -cpuprofile=cpu.prof -count=1 ./...
go tool pprof cpu.prof

# Memory profiling
go test -memprofile=mem.prof -count=1 ./...
go tool pprof mem.prof

# Benchmark tests
go test -bench=. -benchmem ./...

# Race detection
go test -race -count=1 ./...
```

---

## Maintenance Workflows

### 15. Dependency Updates
**Steps**:
```bash
# Check for updates
go list -u -m all

# Update specific dependency
go get github.com/gofiber/fiber/v2@v2.53.0
go mod tidy

# Update all dependencies
go get -u ./...
go mod tidy

# Verify
make test
make lint

# Commit
git add go.mod go.sum
git commit -m "chore: update dependencies"
```

---

### 16. Security Updates
**Steps**:
```bash
# Check for vulnerabilities
go list -json -m all | nancy sleuth

# Update vulnerable dependencies
go get -u github.com/vulnerable/package
go mod tidy

# Run security scan
make docker-build
# Scan image with security scanner

# Verify
make test
```

---

### 17. Code Cleanup
**Steps**:
```bash
# Format code
make fmt

# Remove unused imports
go mod tidy

# Run linter
make lint

# Fix issues
golangci-lint run --fix ./...

# Verify
make test
make test-coverage

# Clean up
make clean
```

---

## Emergency Workflows

### 18. Database Rollback
**Steps**:
```bash
# Check current version
migrate -path sql/migrations -database "postgres://..." version

# Rollback one step
migrate -path sql/migrations -database "postgres://..." down 1

# Rollback to specific version
migrate -path sql/migrations -database "postgres://..." down -version=20240101000000

# Verify
make test
```

---

### 19. Hotfix Deployment
**Steps**:
```bash
# Create hotfix branch
git checkout -b fix/critical-bug

# Fix the issue
# ...

# Test
make test
make lint

# Deploy
make build
make docker-build
# Deploy to production

# Merge back
git checkout main
git merge fix/critical-bug
git push origin main
```

---

### 20. Performance Incident Response
**Steps**:
```bash
# Check health
curl http://localhost:3000/health

# Check logs
tail -f logs/app.log

# Check database
# Connect and check active connections
SELECT * FROM pg_stat_activity;

# Scale if needed
# Update docker-compose.yml or Kubernetes deployment

# Profile application
go test -cpuprofile=cpu.prof ./...
go tool pprof cpu.prof
```

---

## Common Commands Reference

| Task | Command |
|------|---------|
| Initialize | `make init` |
| Generate Code | `make generate` |
| Run Dev | `make dev` |
| Run Tests | `make test` |
| Coverage | `make test-coverage` |
| Lint | `make lint` |
| Format | `make fmt` |
| Build | `make build` |
| Docker Up | `make docker-up` |
| Docker Down | `make docker-down` |
| Migrate Up | `make migrate-up` |
| Migrate Down | `make migrate-down` |
| Install Tools | `make install-tools` |
| Clean | `make clean` |

## Task Automation Tips

1. **Pre-commit Hook**: Add to `.git/hooks/pre-commit`
   ```bash
   #!/bin/bash
   make fmt
   make lint
   make test
   ```

2. **Makefile Targets**: Add custom targets for repeated tasks

3. **Shell Aliases**: Add to `~/.bashrc` or `~/.zshrc`
   ```bash
   alias test-all="make test && make test-coverage"
   alias deploy-prod="make build && make docker-build"
   ```

4. **VS Code Tasks**: Configure in `.vscode/tasks.json`

5. **Docker Compose Override**: Use `docker-compose.override.yml` for dev
