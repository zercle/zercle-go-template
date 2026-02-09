# Task Workflows: Zercle Go Template

**Last Updated:** 2026-02-08  
**Status:** Production-Ready  
**Audience:** Developers using this template

---

## Common Development Workflows

### Daily Development Loop

```bash
# 1. Start the day - ensure everything is working
make check                    # Run all quality checks

# 2. Start development server
make dev                      # Hot reload with Air
# OR
make run                      # Standard run

# 3. Make changes, then verify
make test                     # Run tests
make lint                     # Check code quality

# 4. Before committing
make check                    # Full verification
make fmt                      # Format code
```

### Setting Up a New Development Environment

```bash
# 1. Clone and setup
git clone https://github.com/zercle/zercle-go-template.git
cd zercle-go-template

# 2. Install Go dependencies
make deps

# 3. Install development tools
make install-tools

# 4. Set up pre-commit hooks
make hooks-install

# 5. Configure application
cp configs/config.yaml configs/config.local.yaml
# Edit configs/config.local.yaml with your database settings

# 6. Start PostgreSQL
docker run -d \
  --name postgres \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=zercle_template \
  -p 5432:5432 \
  postgres:14-alpine

# 7. Run database migrations
export DB_USER=postgres
export DB_PASSWORD=postgres
export DB_HOST=localhost
export DB_PORT=5432
export DB_NAME=zercle_template
export DB_SSLMODE=disable
make migrate

# 8. Generate SQLC code
make sqlc

# 9. Run the application
make run

# 10. Verify installation
curl http://localhost:8080/health
```

---

## How to Add New Features

### Adding a New Feature Module

This template uses **feature-based organization**. Follow this pattern:

```
internal/feature/<featurename>/
├── domain/
│   └── <feature>.go          # Domain entities and business rules
├── dto/
│   └── <feature>.go          # Request/response DTOs
├── handler/
│   ├── <feature>_handler.go  # HTTP handlers
│   └── <feature>_handler_test.go
├── repository/
│   ├── <feature>_repository.go      # Interface
│   ├── sqlc_repository.go           # SQLC implementation
│   ├── memory_repository.go         # In-memory implementation
│   └── mocks/                       # Generated mocks
└── usecase/
    ├── <feature>_usecase.go         # Business logic
    ├── <feature>_usecase_test.go
    └── mocks/                       # Generated mocks
```

### Step-by-Step Feature Implementation

#### 1. Define the Domain Model

```go
// internal/feature/order/domain/order.go
package domain

import (
    "time"
    "github.com/google/uuid"
)

type Order struct {
    ID        uuid.UUID
    UserID    uuid.UUID
    Status    OrderStatus
    Total     float64
    CreatedAt time.Time
    UpdatedAt time.Time
}

type OrderStatus string

const (
    OrderStatusPending   OrderStatus = "pending"
    OrderStatusPaid      OrderStatus = "paid"
    OrderStatusShipped   OrderStatus = "shipped"
    OrderStatusDelivered OrderStatus = "delivered"
    OrderStatusCancelled OrderStatus = "cancelled"
)

func (o *Order) Validate() error {
    if o.UserID == uuid.Nil {
        return errors.New("user ID is required")
    }
    if o.Total < 0 {
        return errors.New("total cannot be negative")
    }
    return nil
}
```

#### 2. Create DTOs

```go
// internal/feature/order/dto/order.go
package dto

import "zercle-go-template/internal/feature/order/domain"

type CreateOrderRequest struct {
    UserID uuid.UUID       `json:"user_id" validate:"required"`
    Items  []OrderItemDTO  `json:"items" validate:"required,min=1"`
}

type OrderResponse struct {
    ID        string      `json:"id"`
    UserID    string      `json:"user_id"`
    Status    string      `json:"status"`
    Total     float64     `json:"total"`
    CreatedAt time.Time   `json:"created_at"`
}

type OrderItemDTO struct {
    ProductID uuid.UUID `json:"product_id" validate:"required"`
    Quantity  int       `json:"quantity" validate:"required,min=1"`
    Price     float64   `json:"price" validate:"required,gt=0"`
}

func ToOrderResponse(o *domain.Order) OrderResponse {
    return OrderResponse{
        ID:        o.ID.String(),
        UserID:    o.UserID.String(),
        Status:    string(o.Status),
        Total:     o.Total,
        CreatedAt: o.CreatedAt,
    }
}
```

#### 3. Define Repository Interface

```go
// internal/feature/order/repository/order_repository.go
package repository

import (
    "context"
    "zercle-go-template/internal/feature/order/domain"
)

//go:generate mockgen -source=order_repository.go -destination=mocks/order_repository.go -package=mocks

type OrderRepository interface {
    Create(ctx context.Context, order *domain.Order) error
    GetByID(ctx context.Context, id string) (*domain.Order, error)
    GetByUserID(ctx context.Context, userID string, page, limit int) ([]*domain.Order, int, error)
    Update(ctx context.Context, order *domain.Order) error
    Delete(ctx context.Context, id string) error
}
```

Generate mocks:
```bash
cd internal/feature/order/repository
go generate ./...
# OR
make mock
```

#### 4. Implement Use Case

```go
// internal/feature/order/usecase/order_usecase.go
package usecase

import (
    "context"
    "zercle-go-template/internal/feature/order/domain"
    "zercle-go-template/internal/feature/order/dto"
    "zercle-go-template/internal/feature/order/repository"
)

//go:generate mockgen -source=order_usecase.go -destination=mocks/order_usecase.go -package=mocks

type OrderUsecase interface {
    CreateOrder(ctx context.Context, req dto.CreateOrderRequest) (*domain.Order, error)
    GetOrder(ctx context.Context, id string) (*domain.Order, error)
    ListUserOrders(ctx context.Context, userID string, page, limit int) ([]*domain.Order, int, error)
}

type orderUsecase struct {
    orderRepo repository.OrderRepository
}

func NewOrderUsecase(orderRepo repository.OrderRepository) OrderUsecase {
    return &orderUsecase{orderRepo: orderRepo}
}

func (uc *orderUsecase) CreateOrder(ctx context.Context, req dto.CreateOrderRequest) (*domain.Order, error) {
    // Calculate total
    var total float64
    for _, item := range req.Items {
        total += item.Price * float64(item.Quantity)
    }

    order := &domain.Order{
        ID:        uuid.New(),
        UserID:    req.UserID,
        Status:    domain.OrderStatusPending,
        Total:     total,
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }

    if err := order.Validate(); err != nil {
        return nil, err
    }

    if err := uc.orderRepo.Create(ctx, order); err != nil {
        return nil, err
    }

    return order, nil
}
```

#### 5. Create Handler

```go
// internal/feature/order/handler/order_handler.go
package handler

import (
    "net/http"
    "github.com/labstack/echo/v4"
    "zercle-go-template/internal/feature/order/dto"
    "zercle-go-template/internal/feature/order/usecase"
)

type OrderHandler struct {
    orderUsecase usecase.OrderUsecase
}

func NewOrderHandler(uc usecase.OrderUsecase) *OrderHandler {
    return &OrderHandler{orderUsecase: uc}
}

func (h *OrderHandler) RegisterRoutes(router *echo.Group) {
    orders := router.Group("/orders")
    orders.POST("", h.CreateOrder)
    orders.GET("/:id", h.GetOrder)
    orders.GET("/user/:user_id", h.ListUserOrders)
}

// CreateOrder handles POST /orders
// @Summary Create order
// @Description Create a new order
// @Tags orders
// @Accept json
// @Produce json
// @Param request body dto.CreateOrderRequest true "Order data"
// @Success 201 {object} Response{data=dto.OrderResponse}
// @Router /orders [post]
func (h *OrderHandler) CreateOrder(c echo.Context) error {
    var req dto.CreateOrderRequest
    if err := c.Bind(&req); err != nil {
        return c.JSON(http.StatusBadRequest, errorResponse(err))
    }

    order, err := h.orderUsecase.CreateOrder(c.Request().Context(), req)
    if err != nil {
        return handleError(c, err)
    }

    return c.JSON(http.StatusCreated, successResponse(dto.ToOrderResponse(order)))
}
```

#### 6. Wire in Container

```go
// internal/container/container.go

type Container struct {
    // ... existing fields ...
    
    // Order feature
    OrderRepo    orderRepo.OrderRepository
    OrderUsecase orderUsecase.OrderUsecase
}

func New(cfg *config.Config) (*Container, error) {
    // ... existing initialization ...
    
    // Initialize order feature
    c.OrderRepo = orderRepo.NewSQLCRepository(c.DB)
    c.OrderUsecase = orderUsecase.NewOrderUsecase(c.OrderRepo)
    
    return c, nil
}
```

#### 7. Register Routes

```go
// cmd/api/main.go

func setupRouter(e *echo.Echo, container *container.Container, log logger.Logger) {
    // ... existing routes ...
    
    // Order routes
    orderHandler := orderHandler.NewOrderHandler(container.OrderUsecase)
    orderHandler.RegisterRoutes(api)
}
```

#### 8. Add Database Migration

```bash
# Create migration
make migrate-create name=create_orders_table
```

Edit the generated files:

```sql
-- internal/infrastructure/db/migrations/002_create_orders_table.up.sql
CREATE TABLE orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    total DECIMAL(10, 2) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_orders_user_id ON orders(user_id);
CREATE INDEX idx_orders_status ON orders(status);
```

```sql
-- internal/infrastructure/db/migrations/002_create_orders_table.down.sql
DROP INDEX IF EXISTS idx_orders_status;
DROP INDEX IF EXISTS idx_orders_user_id;
DROP TABLE IF EXISTS orders;
```

Run migration:
```bash
make migrate
```

#### 9. Add SQLC Queries

```sql
-- internal/infrastructure/db/queries/orders.sql
-- name: CreateOrder :one
INSERT INTO orders (id, user_id, status, total, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetOrderByID :one
SELECT * FROM orders WHERE id = $1;

-- name: GetOrdersByUserID :many
SELECT * FROM orders 
WHERE user_id = $1 
ORDER BY created_at DESC 
LIMIT $2 OFFSET $3;

-- name: CountOrdersByUserID :one
SELECT COUNT(*) FROM orders WHERE user_id = $1;
```

Generate code:
```bash
make sqlc
```

#### 10. Write Tests

```go
// internal/feature/order/usecase/order_usecase_test.go
package usecase

import (
    "context"
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "zercle-go-template/internal/feature/order/domain"
    "zercle-go-template/internal/feature/order/dto"
    "zercle-go-template/internal/feature/order/repository/mocks"
)

func TestOrderUsecase_CreateOrder(t *testing.T) {
    tests := []struct {
        name    string
        req     dto.CreateOrderRequest
        mock    func(*mocks.MockOrderRepository)
        wantErr bool
    }{
        {
            name: "success",
            req: dto.CreateOrderRequest{
                UserID: uuid.New(),
                Items: []dto.OrderItemDTO{
                    {ProductID: uuid.New(), Quantity: 2, Price: 10.00},
                },
            },
            mock: func(m *mocks.MockOrderRepository) {
                m.On("Create", mock.Anything, mock.AnythingOfType("*domain.Order")).
                    Return(nil)
            },
            wantErr: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockRepo := new(mocks.MockOrderRepository)
            tt.mock(mockRepo)

            uc := NewOrderUsecase(mockRepo)
            order, err := uc.CreateOrder(context.Background(), tt.req)

            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                assert.NotNil(t, order)
                assert.Equal(t, 20.00, order.Total) // 2 * 10.00
            }

            mockRepo.AssertExpectations(t)
        })
    }
}
```

---

## Testing Procedures

### Test-Driven Development Workflow

```
1. Write test (fails)
2. Implement minimal code (passes)
3. Refactor
4. Repeat
```

### Running Tests

```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# View HTML coverage report
make test-coverage-html
open coverage.html

# Run specific test
make test TEST_FLAGS="-v -run TestCreateUser"

# Run integration tests only
make test-integration

# Run benchmarks
make benchmark
```

### Writing Good Tests

#### Unit Test Pattern

```go
func TestUsecase_Method(t *testing.T) {
    type args struct {
        // Input parameters
    }
    
    tests := []struct {
        name    string
        args    args
        mock    func(*mocks.MockRepository)
        want    *domain.Entity
        wantErr error
    }{
        {
            name: "success case",
            args: args{/* ... */},
            mock: func(m *mocks.MockRepository) {
                m.On("Method", mock.Anything, /* ... */).
                    Return(/* ... */)
            },
            want:    &domain.Entity{/* ... */},
            wantErr: nil,
        },
        {
            name: "error case",
            args: args{/* ... */},
            mock: func(m *mocks.MockRepository) {
                m.On("Method", mock.Anything, /* ... */).
                    Return(nil, errors.New("database error"))
            },
            want:    nil,
            wantErr: errors.New("database error"),
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup mocks
            mockRepo := new(mocks.MockRepository)
            tt.mock(mockRepo)
            
            // Execute
            uc := NewUsecase(mockRepo)
            got, err := uc.Method(context.Background(), tt.args/* ... */)
            
            // Assert
            if tt.wantErr != nil {
                assert.Error(t, err)
                assert.Equal(t, tt.wantErr.Error(), err.Error())
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tt.want, got)
            }
            
            mockRepo.AssertExpectations(t)
        })
    }
}
```

#### Integration Test Pattern

```go
//go:build integration

package repository

import (
    "context"
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "zercle-go-template/internal/infrastructure/db"
)

func TestSQLCRepository_CreateUser_Integration(t *testing.T) {
    // Setup test database
    pool, cleanup := db.SetupTestDB(t)
    defer cleanup()
    
    repo := NewSQLCRepository(pool)
    
    ctx := context.Background()
    user := &domain.User{
        ID:       uuid.New(),
        Email:    "test@example.com",
        Password: "hashedpassword",
        Name:     "Test User",
    }
    
    // Execute
    err := repo.Create(ctx, user)
    
    // Assert
    require.NoError(t, err)
    
    // Verify by fetching
    fetched, err := repo.GetByID(ctx, user.ID.String())
    require.NoError(t, err)
    assert.Equal(t, user.Email, fetched.Email)
}
```

---

## Database Migration Process

### Creating a Migration

```bash
# Create new migration files
make migrate-create name=add_user_profile_fields

# This creates:
# internal/infrastructure/db/migrations/003_add_user_profile_fields.up.sql
# internal/infrastructure/db/migrations/003_add_user_profile_fields.down.sql
```

### Writing Migrations

**Best Practices**:
- Always provide both `up` and `down` migrations
- Make migrations idempotent where possible
- Avoid destructive operations in production (use separate migrations)
- Test migrations on a copy of production data

```sql
-- 003_add_user_profile_fields.up.sql
-- Add new columns to users table

-- Check if column exists before adding (idempotent)
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name='users' AND column_name='avatar_url') THEN
        ALTER TABLE users ADD COLUMN avatar_url VARCHAR(500);
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name='users' AND column_name='bio') THEN
        ALTER TABLE users ADD COLUMN bio TEXT;
    END IF;
END $$;
```

```sql
-- 003_add_user_profile_fields.down.sql
-- Revert changes

ALTER TABLE users DROP COLUMN IF EXISTS avatar_url;
ALTER TABLE users DROP COLUMN IF EXISTS bio;
```

### Running Migrations

```bash
# Set environment variables
export DB_USER=postgres
export DB_PASSWORD=postgres
export DB_HOST=localhost
export DB_PORT=5432
export DB_NAME=zercle_template
export DB_SSLMODE=disable

# Run all pending migrations
make migrate

# Rollback one migration
make migrate-down

# Reset all migrations (down then up)
make migrate-reset

# Check current version
migrate -path internal/infrastructure/db/migrations \
  -database "postgres://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME?sslmode=$DB_SSLMODE" \
  version
```

### Migration Checklist

Before committing migrations:

- [ ] Migration is idempotent (can run multiple times safely)
- [ ] Down migration is provided and tested
- [ ] Migration works on empty database
- [ ] Migration works on existing database with data
- [ ] Indexes added for new foreign keys
- [ ] No breaking changes to existing columns (renames, type changes)

---

## Code Review Checklist

### Before Submitting PR

```bash
# Run full check suite
make check

# Verify mocks are up to date
make mock-verify

# Verify SQLC is up to date
make sqlc-verify

# Check test coverage
make test-coverage
```

### Review Criteria

#### Code Quality
- [ ] Code follows Go conventions (gofmt, golint)
- [ ] No linting errors (`make lint`)
- [ ] All tests pass (`make test`)
- [ ] Test coverage maintained or improved
- [ ] No security issues (`make security`)

#### Architecture
- [ ] Clean architecture principles followed
- [ ] Domain logic in use cases, not handlers
- [ ] Repository interfaces properly abstracted
- [ ] Dependencies injected via container
- [ ] No circular dependencies

#### Documentation
- [ ] Swagger annotations added for handlers
- [ ] Complex logic has comments
- [ ] README updated if needed
- [ ] Memory Bank updated for significant changes

#### Testing
- [ ] Unit tests for use cases
- [ ] Integration tests for repositories
- [ ] Table-driven test patterns used
- [ ] Mocks generated and up to date
- [ ] Edge cases covered

#### Database
- [ ] Migrations provided (up and down)
- [ ] SQLC queries added
- [ ] Indexes added for performance
- [ ] No N+1 query problems

---

## Troubleshooting Common Issues

### Build Issues

```bash
# Clean and rebuild
make clean
make deps
make build

# Check Go version
go version  # Should be 1.21+
```

### Database Issues

```bash
# Reset database
docker rm -f postgres
docker run -d --name postgres -e POSTGRES_PASSWORD=postgres -p 5432:5432 postgres:14-alpine
make migrate

# Check connection
psql postgres://postgres:postgres@localhost:5432/zercle_template -c "\dt"
```

### Test Issues

```bash
# Run tests with verbose output
go test -v ./...

# Run specific failing test
go test -v -run TestName ./path/to/package

# Clean test cache
go clean -testcache
```

### Linting Issues

```bash
# Auto-fix formatting
make fmt

# Run linter with details
golangci-lint run --verbose

# Skip pre-commit hooks (emergency only)
git commit --no-verify -m "message"
```

---

**Related Documents**:
- [brief.md](brief.md) - Project overview
- [product.md](product.md) - Product documentation
- [architecture.md](architecture.md) - System architecture
- [tech.md](tech.md) - Technology stack
- [context.md](context.md) - Current context
