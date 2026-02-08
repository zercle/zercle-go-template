# Migration Plan: zercle-go-template Refactoring

## Executive Summary

This document outlines a comprehensive refactoring plan for the `zercle-go-template` project. The migration involves renaming the `service` layer to `usecase`, enhancing mock generation, integrating sqlc for type-safe SQL queries, improving configuration loading hierarchy, setting up git hooks, and enhancing the CI/CD pipeline.

**Key Changes:**
1. Rename `service/` → `usecase/` across all features
2. Standardize mock generation with mockery
3. Integrate sqlc for type-safe database queries
4. Enhance config loading: YAML → .env → runtime env vars (with override)
5. Setup pre-commit git hooks
6. Enhance CI/CD workflow with integration tests

---

## 1. Service Layer Rename (service → usecase)

### 1.1 Current State

The project currently uses `service` package naming for business logic:

```
internal/feature/
├── auth/
│   └── service/
│       └── jwt_service.go
└── user/
    └── service/
        └── user_service.go
```

### 1.2 File Renaming Mapping

| Current Path | New Path |
|--------------|----------|
| `internal/feature/auth/service/` | `internal/feature/auth/usecase/` |
| `internal/feature/auth/service/jwt_service.go` | `internal/feature/auth/usecase/jwt_usecase.go` |
| `internal/feature/user/service/` | `internal/feature/user/usecase/` |
| `internal/feature/user/service/user_service.go` | `internal/feature/user/usecase/user_usecase.go` |

### 1.3 Interface and Type Changes

| Current Name | New Name |
|--------------|----------|
| `service.UserService` | `usecase.UserUsecase` |
| `service.JWTService` | `usecase.JWTUsecase` |
| `service.userService` | `usecase.userUsecase` |
| `service.jwtService` | `usecase.jwtUsecase` |
| `service.NewUserService` | `usecase.NewUserUsecase` |
| `service.NewJWTService` | `usecase.NewJWTUsecase` |

### 1.4 Import Path Updates

Files that import service packages need updates:

| File | Current Import | New Import |
|------|----------------|------------|
| `internal/feature/user/handler/user_handler.go` | `zercle-go-template/internal/feature/user/service` | `zercle-go-template/internal/feature/user/usecase` |
| `internal/feature/user/handler/user_handler.go` | `zercle-go-template/internal/feature/auth/service` | `zercle-go-template/internal/feature/auth/usecase` |
| `internal/container/container.go` | `zercle-go-template/internal/feature/user/service` | `zercle-go-template/internal/feature/user/usecase` |
| `internal/container/container.go` | `zercle-go-template/internal/feature/auth/service` | `zercle-go-template/internal/feature/auth/usecase` |
| `internal/feature/user/user.go` (documentation) | `internal/feature/user/service` | `internal/feature/user/usecase` |
| `internal/feature/auth/auth.go` (documentation) | `internal/feature/auth/service` | `internal/feature/auth/usecase` |

### 1.5 Container Type Updates

In `internal/container/container.go`:

```go
// Current
type Container struct {
    Config      *config.Config
    Logger      logger.Logger
    UserRepo    userrepository.UserRepository
    UserService userservice.UserService
    JWTService  authservice.JWTService
}

// New
type Container struct {
    Config      *config.Config
    Logger      logger.Logger
    UserRepo    userrepository.UserRepository
    UserUsecase userusecase.UserUsecase
    JWTUsecase  authusecase.JWTUsecase
}
```

### 1.6 Handler Field Updates

In `internal/feature/user/handler/user_handler.go`:

```go
// Current
type UserHandler struct {
    userService service.UserService
    jwtService  authservice.JWTService
    logger      logger.Logger
}

// New
type UserHandler struct {
    userUsecase usecase.UserUsecase
    jwtUsecase  authusecase.JWTUsecase
    logger      logger.Logger
}
```

---

## 2. Mock Generation Enhancement

### 2.1 Current State

The project already has `//go:generate mockgen` directives:
- `internal/feature/user/service/user_service.go`
- `internal/feature/auth/service/jwt_service.go`
- `internal/feature/user/repository/user_repository.go`

### 2.2 Recommended Tool: mockery

**Why mockery over mockgen:**
- Better interface discovery
- Simpler configuration
- Generates mocks in a consistent location
- Better support for complex interfaces

### 2.3 Installation

```bash
go install github.com/vektra/mockery/v2@latest
```

### 2.4 Mock Generation Strategy

#### 2.4.1 Interfaces to Mock

| Feature | Interface | Location |
|---------|-----------|----------|
| User | `UserRepository` | `internal/feature/user/repository/user_repository.go` |
| User | `UserUsecase` | `internal/feature/user/usecase/user_usecase.go` |
| Auth | `JWTUsecase` | `internal/feature/auth/usecase/jwt_usecase.go` |
| Logger | `Logger` | `internal/logger/logger.go` |

#### 2.4.2 Mock File Organization

```
internal/
├── feature/
│   ├── user/
│   │   ├── repository/
│   │   │   ├── user_repository.go
│   │   │   └── mocks/
│   │   │       └── UserRepository.go
│   │   └── usecase/
│   │       ├── user_usecase.go
│   │       └── mocks/
│   │           └── UserUsecase.go
│   └── auth/
│       └── usecase/
│           ├── jwt_usecase.go
│           └── mocks/
│               └── JWTUsecase.go
└── logger/
    ├── logger.go
    └── mocks/
        └── Logger.go
```

#### 2.4.3 mockery Configuration

Create `.mockery.yaml`:

```yaml
version: 2
with-expecter: true
mockname: "{{.InterfaceName}}"
outpkg: mocks
dir: "{{.InterfaceDir}}/mocks"
filename: "{{.InterfaceName}}.go"
packages:
  zercle-go-template/internal/feature/user/repository:
    interfaces:
      UserRepository:
        config:
          dir: "internal/feature/user/repository/mocks"
  zercle-go-template/internal/feature/user/usecase:
    interfaces:
      UserUsecase:
        config:
          dir: "internal/feature/user/usecase/mocks"
  zercle-go-template/internal/feature/auth/usecase:
    interfaces:
      JWTUsecase:
        config:
          dir: "internal/feature/auth/usecase/mocks"
  zercle-go-template/internal/logger:
    interfaces:
      Logger:
        config:
          dir: "internal/logger/mocks"
```

#### 2.4.4 Generate Mocks Command

```bash
# Generate all mocks
mockry --all

# Generate mocks for specific package
mockry --dir=internal/feature/user/repository
```

#### 2.4.5 Update go:generate Directives

Replace existing `//go:generate` directives with:

```go
// In repository files
//go:generate mockery --name=UserRepository --output=./mocks --outpkg=mocks

// In usecase files
//go:generate mockery --name=UserUsecase --output=./mocks --outpkg=mocks
```

---

## 3. Sqlc Integration

### 3.1 Current State

The project uses in-memory repository (`memory_repository.go`) for development. No database adapter is implemented.

### 3.2 Sqlc Overview

Sqlc generates type-safe Go code from SQL queries. It provides:
- Type-safe query execution
- Compile-time query validation
- Automatic struct generation
- Support for PostgreSQL, MySQL, SQLite

### 3.3 Installation

```bash
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
```

### 3.4 Directory Structure

```
internal/
└── feature/
    └── user/
        ├── repository/
        │   ├── user_repository.go          # Interface
        │   ├── memory_repository.go        # In-memory impl
        │   ├── sqlc_repository.go         # NEW: Sqlc impl
        │   └── mocks/
        │       └── UserRepository.go
        └── sqlc/                           # NEW: Sqlc directory
            ├── schema.sql                  # Database schema
            ├── queries.sql                 # SQL queries
            └── db.go                       # Generated code (by sqlc)
```

### 3.5 Sqlc Configuration

Create `sqlc.yaml` at project root:

```yaml
version: "2"
sql:
  - engine: "postgresql"
    queries: "internal/feature/user/sqlc/queries.sql"
    schema: "internal/feature/user/sqlc/schema.sql"
    gen:
      go:
        package: "db"
        out: "internal/feature/user/sqlc"
        sql_package: "database/sql"
        emit_json_tags: true
        emit_prepared_queries: false
        emit_interface: false
        emit_exact_table_names: false
```

### 3.6 Database Schema

Create `internal/feature/user/sqlc/schema.sql`:

```sql
-- Schema: users
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Index for email lookups
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

-- Trigger for updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
```

### 3.7 SQL Queries

Create `internal/feature/user/sqlc/queries.sql`:

```sql
-- name: CreateUser :one
INSERT INTO users (id, email, name, password_hash)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1 LIMIT 1;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1 LIMIT 1;

-- name: ListUsers :many
SELECT * FROM users
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountUsers :one
SELECT COUNT(*) FROM users;

-- name: UpdateUser :one
UPDATE users
SET email = $2, name = $3
WHERE id = $1
RETURNING *;

-- name: UpdatePassword :exec
UPDATE users
SET password_hash = $2
WHERE id = $1;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;

-- name: EmailExists :one
SELECT EXISTS(
    SELECT 1 FROM users WHERE email = $1
);
```

### 3.8 Sqlc Repository Implementation

Create `internal/feature/user/repository/sqlc_repository.go`:

```go
package repository

import (
	"context"
	"database/sql"

	"zercle-go-template/internal/feature/user/domain"
	"zercle-go-template/internal/feature/user/sqlc"
)

// sqlcUserRepository implements UserRepository using sqlc.
type sqlcUserRepository struct {
	db *sql.DB
	q  *sqlc.Queries
}

// NewSqlcUserRepository creates a new sqlc-based user repository.
func NewSqlcUserRepository(db *sql.DB) UserRepository {
	return &sqlcUserRepository{
		db: db,
		q:  sqlc.New(db),
	}
}

// Implement all UserRepository interface methods...
```

### 3.9 Generated Code Structure

After running `sqlc generate`, the following files will be created:

```
internal/feature/user/sqlc/
├── db.go              # Generated: DBTX interface, models
├── models.go          # Generated: User struct
├── user.sql.go        # Generated: Query functions
└── query.sql.go       # Generated: Query interface
```

### 3.10 Dependency Updates

Add to `go.mod`:

```bash
go get github.com/lib/pq  # PostgreSQL driver
```

### 3.11 Integration with Config

Update `internal/config/config.go` to include sqlc DSN:

```go
type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Database string `mapstructure:"database"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	SSLMode  string `mapstructure:"ssl_mode"`
	DSN      string `mapstructure:"dsn"`  // NEW: Full connection string
}
```

---

## 4. Enhanced Configuration Loading

### 4.1 Current State

The config system uses Viper with priority:
1. Default values
2. Config file (`configs/config.yaml`)
3. Environment variables (APP_* prefix)

### 4.2 Enhanced Priority Order

**New Hierarchy (highest to lowest priority):**

1. **Runtime Environment Variables** - Override everything
2. **.env File** - Environment-specific overrides
3. **YAML Config File** - Base configuration
4. **Default Values** - Fallback

### 4.3 Implementation Approach

#### 4.3.1 Add godotenv for .env Support

```bash
go get github.com/joho/godotenv
```

#### 4.3.2 Updated Config Loading Logic

In `internal/config/config.go`:

```go
// Load loads configuration from all sources.
// Priority order: defaults < config file < .env file < runtime env vars
func Load() (*Config, error) {
	v := viper.New()

	// Step 1: Set default values
	setDefaults(v)

	// Step 2: Load YAML config file
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("./configs")

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// Step 3: Load .env file (if exists)
	// Note: .env values will be loaded into env vars before viper reads them
	if err := godotenv.Load(); err != nil {
		// .env file not found is not fatal
		// Continue with other config sources
	}

	// Step 4: Configure environment variable handling
	// Runtime env vars override everything
	v.SetEnvPrefix("APP")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Unmarshal config into struct
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}
```

### 4.4 Configuration Precedence Examples

| Config Key | YAML Value | .env Value | Runtime Env | Final Value |
|------------|------------|------------|-------------|-------------|
| `server.port` | `8080` | `3000` | `9000` | `9000` |
| `app.name` | `myapp` | `myapp-dev` | (not set) | `myapp-dev` |
| `log.level` | `debug` | (not set) | `info` | `info` |
| `jwt.secret` | (not set) | `dev-secret` | `prod-secret` | `prod-secret` |

### 4.5 Environment Variable Naming

| Config Path | Environment Variable |
|-------------|----------------------|
| `app.name` | `APP_APP_NAME` |
| `server.port` | `APP_SERVER_PORT` |
| `database.host` | `APP_DATABASE_HOST` |
| `jwt.secret` | `APP_JWT_SECRET` |

### 4.6 .env File Updates

Update `.env.example` to reflect all config options:

```bash
# Application
APP_NAME=zercle-go-template
APP_VERSION=1.0.0
APP_ENVIRONMENT=development

# Server
APP_SERVER_HOST=0.0.0.0
APP_SERVER_PORT=8080

# Database (PostgreSQL)
APP_DATABASE_HOST=localhost
APP_DATABASE_PORT=5432
APP_DATABASE_NAME=zercle_db
APP_DATABASE_USERNAME=postgres
APP_DATABASE_PASSWORD=changeme
APP_DATABASE_SSL_MODE=disable

# JWT
APP_JWT_SECRET=your-jwt-secret-key-here
APP_JWT_ACCESS_TOKEN_TTL=1h
APP_JWT_REFRESH_TOKEN_TTL=24h

# Logging
APP_LOG_LEVEL=debug
APP_LOG_FORMAT=console
```

---

## 5. Git Hooks Setup

### 5.1 Current State

No git hooks are currently configured.

### 5.2 Recommended Tool: pre-commit

**Why pre-commit:**
- Language-agnostic framework
- Easy to configure
- Supports multiple hook types
- Can be installed via homebrew or pip

### 5.3 Installation

```bash
# macOS
brew install pre-commit

# Or via pip
pip install pre-commit
```

### 5.4 Pre-commit Configuration

Create `.pre-commit-config.yaml`:

```yaml
# Pre-commit hooks configuration
# See: https://pre-commit.com/

repos:
  # Go hooks using pre-commit-go
  - repo: https://github.com/pre-commit/mirrors-prettier
    rev: v3.1.0
    hooks:
      - id: prettier
        types: [yaml, markdown]
        exclude: ^(go\.sum|\.git/)

  # Go specific hooks
  - repo: https://github.com/golangci/golangci-lint
    rev: v1.59.1
    hooks:
      - id: golangci-lint
        args: [--timeout=5m, --config=.golangci.yml]

  # Go fmt check
  - repo: local
    hooks:
      - id: go-fmt
        name: go fmt
        entry: gofmt
        language: system
        args: [-l, -w]
        files: \.go$

  # Go imports
  - repo: https://github.com/dnephin/pre-commit-golang
    rev: v0.5.1
    hooks:
      - id: goimports
        args: [-local, zercle-go-template]

  # Go tests
  - repo: local
    hooks:
      - id: go-test
        name: go test
        entry: go test
        language: system
        args: [-v, ./...]
        pass_filenames: false

  # Go mod tidy
  - repo: local
    hooks:
      - id: go-mod-tidy
        name: go mod tidy
        entry: go mod tidy
        language: system
        pass_filenames: false

  # Check for merge conflict markers
  - repo: https://github.com/pre-commit/pre-commit-hooks
    rev: v4.6.0
    hooks:
      - id: check-merge-conflict
      - id: check-yaml
      - id: end-of-file-fixer
      - id: trailing-whitespace
```

### 5.5 Installing Hooks

```bash
# Install hooks in .git/hooks/
pre-commit install

# Install pre-commit hook only
pre-commit install --hook-type pre-commit

# Install pre-push hook (optional)
pre-commit install --hook-type pre-push
```

### 5.6 Manual Hook Testing

```bash
# Run all hooks on all files
pre-commit run --all-files

# Run specific hook
pre-commit run go-fmt --all-files
```

### 5.7 Alternative: husky (for Node.js projects)

If the project has Node.js tooling, husky can be used:

```bash
npm install --save-dev husky
npx husky install
npx husky add .husky/pre-commit "go test ./... && golangci-lint run && gofmt -w ."
```

---

## 6. CI/CD Pipeline Enhancement

### 6.1 Current State

The project has `.github/workflows/ci.yml` with:
- Lint job (golangci-lint)
- Security scan (gosec)
- Format check
- Test job
- Build job
- Docker job

### 6.2 Enhanced Workflow Structure

Create `.github/workflows/ci-enhanced.yml`:

```yaml
name: CI Enhanced

on:
  push:
    branches: [main, develop]
    tags: ['v*']
  pull_request:
    branches: [main, develop]

concurrency:
  group: ${{ github.workflow }}-${{ github.ref }}
  cancel-in-progress: true

permissions:
  contents: read
  security-events: write
  pull-requests: read

env:
  GO_VERSION: "1.24"
  POSTGRES_VERSION: "16"

jobs:
  # Job 1: Lint
  lint:
    name: Lint
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}
          cache: true

      - name: Run golangci-lint
        uses: golangci/golangci-lint-action@v6
        with:
          version: latest
          args: --timeout=5m --out-format=github-actions

  # Job 2: Format Check
  format:
    name: Check Format
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}
          cache: true

      - name: Check gofmt
        run: |
          if [ -n "$(gofmt -l .)" ]; then
            echo "The following files need formatting:"
            gofmt -l .
            exit 1
          fi

      - name: Check goimports
        run: |
          go install golang.org/x/tools/cmd/goimports@latest
          if [ -n "$(goimports -l .)" ]; then
            echo "The following files need import formatting:"
            goimports -l .
            exit 1
          fi

  # Job 3: Unit Tests
  test-unit:
    name: Unit Tests
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}
          cache: true

      - name: Run unit tests
        run: |
          go test -v -race -coverprofile=coverage.out -covermode=atomic ./...

      - name: Upload coverage to Codecov
        uses: codecov/codecov-action@v4
        with:
          file: ./coverage.out
          flags: unittests
          name: codecov-umbrella

  # Job 4: Integration Tests
  test-integration:
    name: Integration Tests
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:${{ env.POSTGRES_VERSION }}
        env:
          POSTGRES_USER: test
          POSTGRES_PASSWORD: test
          POSTGRES_DB: test_db
        ports:
          - 5432:5432
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}
          cache: true

      - name: Install sqlc
        run: |
          go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

      - name: Generate sqlc code
        run: |
          sqlc generate

      - name: Run integration tests
        run: |
          go test -v -tags=integration ./...
        env:
          DATABASE_URL: postgres://test:test@localhost:5432/test_db?sslmode=disable

  # Job 5: Build
  build:
    name: Build
    runs-on: ubuntu-latest
    needs: [lint, format, test-unit]
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}
          cache: true

      - name: Build binary
        run: |
          go build -v -o bin/app ./cmd/api

      - name: Upload artifact
        uses: actions/upload-artifact@v4
        with:
          name: app-binary
          path: bin/app

  # Job 6: Security
  security:
    name: Security Scan
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}
          cache: true

      - name: Run Gosec
        uses: securego/gosec@master
        with:
          args: "-no-fail -fmt sarif -out results.sarif ./..."

      - name: Upload SARIF
        uses: github/codeql-action/upload-sarif@v3
        if: always()
        with:
          sarif_file: results.sarif
```

### 6.3 Workflow Changes Summary

| Change | Description |
|--------|-------------|
| **Separate Unit/Integration Tests** | Split tests into two jobs for better isolation |
| **Add Integration Test Service** | PostgreSQL container for integration tests |
| **Add sqlc Generation Step** | Generate type-safe SQL code before tests |
| **Add goimports Check** | Ensure proper import formatting |
| **Coverage Upload** | Upload coverage to Codecov |
| **Artifact Upload** | Store built binary for deployment |

---

## 7. Step-by-Step Migration Sequence

### Phase 1: Preparation (No Breaking Changes)

1. **Install New Dependencies**
   ```bash
   go install github.com/vektra/mockery/v2@latest
   go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
   go get github.com/joho/godotenv
   go get github.com/lib/pq
   ```

2. **Create Configuration Files**
   - Create `.mockery.yaml`
   - Create `sqlc.yaml`
   - Create `.pre-commit-config.yaml`

3. **Create Sqlc Directory Structure**
   - Create `internal/feature/user/sqlc/`
   - Create `schema.sql`
   - Create `queries.sql`

4. **Update Makefile**
   - Add `mocks` target
   - Add `sqlc-generate` target
   - Add `hooks-install` target

### Phase 2: Service Layer Rename (Breaking Changes)

5. **Rename Directories**
   ```bash
   mv internal/feature/auth/service internal/feature/auth/usecase
   mv internal/feature/user/service internal/feature/user/usecase
   ```

6. **Rename Files**
   ```bash
   mv internal/feature/auth/usecase/jwt_service.go internal/feature/auth/usecase/jwt_usecase.go
   mv internal/feature/user/usecase/user_service.go internal/feature/user/usecase/user_usecase.go
   ```

7. **Update Package Declarations**
   - Change `package service` to `package usecase` in both files

8. **Update Type Names**
   - `UserService` → `UserUsecase`
   - `JWTService` → `JWTUsecase`
   - `userService` → `userUsecase`
   - `jwtService` → `jwtUsecase`
   - `NewUserService` → `NewUserUsecase`
   - `NewJWTService` → `NewJWTUsecase`

9. **Update Import Paths**
   - Update all files importing `.../service` to `.../usecase`
   - Update `internal/container/container.go`
   - Update `internal/feature/user/handler/user_handler.go`
   - Update `internal/feature/user/user.go`
   - Update `internal/feature/auth/auth.go`

10. **Update Container Types**
    - Update `Container` struct fields
    - Update initialization code

11. **Update Handler Types**
    - Update `UserHandler` struct fields
    - Update all method implementations

### Phase 3: Mock Generation

12. **Generate Mocks with Mockery**
    ```bash
    mockery --all
    ```

13. **Update Test Files**
    - Update import paths for mocks
    - Update mock type references

### Phase 4: Sqlc Integration

14. **Generate Sqlc Code**
    ```bash
    sqlc generate
    ```

15. **Implement Sqlc Repository**
    - Create `sqlc_repository.go`
    - Implement `UserRepository` interface

16. **Update Container**
    - Add database connection initialization
    - Add sqlc repository option

17. **Update Config**
    - Add DSN configuration option
    - Update validation

### Phase 5: Enhanced Config Loading

18. **Update Config Loading**
    - Add godotenv loading
    - Update priority order

19. **Update .env.example**
    - Ensure all config options are documented

### Phase 6: Git Hooks

20. **Install Pre-commit**
    ```bash
    pre-commit install
    ```

21. **Test Hooks**
    ```bash
    pre-commit run --all-files
    ```

### Phase 7: CI/CD Enhancement

22. **Create Enhanced Workflow**
    - Create `.github/workflows/ci-enhanced.yml`

23. **Test Workflow**
    - Push to test branch
    - Verify all jobs pass

### Phase 8: Documentation

24. **Update Documentation**
    - Update `README.md`
    - Update `AGENTS.md` (if applicable)
    - Update Memory Bank files

25. **Update Architecture Documentation**
    - Update `.agents/rules/memory-bank/architecture.md`
    - Update `.agents/rules/memory-bank/tech.md`

---

## 8. Risk Assessment

### 8.1 High Risk Items

| Risk | Impact | Mitigation |
|------|--------|------------|
| **Service layer rename breaks imports** | High | Comprehensive search and replace; run tests after each change |
| **Container changes break initialization** | High | Update container first, then handlers; test initialization |
| **Sqlc generated code conflicts** | Medium | Generate in separate branch; review generated code |
| **Config priority changes cause unexpected behavior** | Medium | Document priority clearly; add logging of config source |

### 8.2 Medium Risk Items

| Risk | Impact | Mitigation |
|------|--------|------------|
| **Mock generation fails** | Medium | Use mockery with config file; test generation locally |
| **Git hooks block commits** | Low | Hooks can be bypassed with `--no-verify` |
| **CI/CD workflow changes** | Medium | Test on feature branch; keep old workflow as backup |

### 8.3 Low Risk Items

| Risk | Impact | Mitigation |
|------|--------|------------|
| **Documentation updates** | Low | Non-breaking; can be done after code changes |
| **Makefile additions** | Low | Additive changes only |

---

## 9. Rollback Strategy

### 9.1 Git Branch Strategy

```bash
# Create feature branch
git checkout -b feature/migration-refactor

# Commit changes incrementally
git add .
git commit -m "feat: rename service layer to usecase"

# If rollback is needed
git revert HEAD
# Or reset to previous commit
git reset --hard HEAD~1
```

### 9.2 Feature Flags

Consider adding feature flags for gradual rollout:

```go
// In config
type Config struct {
    Features struct {
        UseSqlcRepository bool `mapstructure:"use_sqlc_repository"`
    }
}
```

### 9.3 Database Migration Rollback

For sqlc changes, maintain migration scripts:

```sql
-- migrations/001_create_users_table.up.sql
CREATE TABLE users (...);

-- migrations/001_create_users_table.down.sql
DROP TABLE users;
```

### 9.4 Config Rollback

Keep old config loading as fallback:

```go
func Load() (*Config, error) {
    // Try enhanced loading first
    cfg, err := loadEnhanced()
    if err != nil {
        // Fallback to old loading
        return loadLegacy()
    }
    return cfg, nil
}
```

---

## 10. Testing Strategy

### 10.1 Unit Tests

- Ensure all existing tests pass after each change
- Add tests for new sqlc repository
- Add tests for enhanced config loading

### 10.2 Integration Tests

- Test sqlc repository with real PostgreSQL
- Test config loading with different sources
- Test git hooks with various scenarios

### 10.3 Manual Testing

- Test application startup with new config
- Test authentication flow
- Test user CRUD operations

---

## 11. Post-Migration Checklist

- [ ] All tests pass (unit + integration)
- [ ] Linting passes (golangci-lint)
- [ ] Code formatted (gofmt, goimports)
- [ ] Git hooks installed and working
- [ ] CI/CD pipeline passes
- [ ] Documentation updated
- [ ] Memory Bank updated
- [ ] No deprecation warnings
- [ ] Performance benchmarks pass
- [ ] Security scan passes

---

## 12. References

- [mockery Documentation](https://vektra.github.io/mockery/)
- [sqlc Documentation](https://docs.sqlc.dev/)
- [pre-commit Documentation](https://pre-commit.com/)
- [Viper Documentation](https://github.com/spf13/viper)
- [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
