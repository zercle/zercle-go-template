# Zercle Go Template - Development Tasks & Workflows

## Common Development Workflows

### Starting Development

```bash
# 1. Clone and enter repository
cd zercle-go-template

# 2. Copy environment file
cp .env.example .env
# Edit .env with your values

# 3. Start development environment
make docker-up

# 4. Verify running
curl http://localhost:8080/health
# View Swagger: http://localhost:8080/swagger/index.html
```

### Daily Development Loop

```bash
# Pull latest changes
git pull origin main

# Run tests before making changes
make test

# Start development server
make dev

# Make changes to code...

# Run tests after changes
make test

# Run linter
make lint

# Commit (pre-commit hooks will run)
git add .
git commit -m "feat: description"
```

## Adding a New Feature

### Step-by-Step Guide

Follow this pattern to add a new feature (e.g., `order`):

#### 1. Create Domain Layer

```go
// internal/feature/order/domain/order.go
package domain

import "time"

type Order struct {
    ID        string
    UserID    string
    Total     float64
    Status    OrderStatus
    CreatedAt time.Time
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
    if o.UserID == "" {
        return errors.New("user_id is required")
    }
    if o.Total <= 0 {
        return errors.New("total must be positive")
    }
    return nil
}
```

#### 2. Create DTOs

```go
// internal/feature/order/dto/order.go
package dto

type CreateOrderRequest struct {
    UserID string  `json:"user_id" validate:"required,uuid"`
    Total  float64 `json:"total" validate:"required,gt=0"`
}

type OrderResponse struct {
    ID        string    `json:"id"`
    UserID    string    `json:"user_id"`
    Total     float64   `json:"total"`
    Status    string    `json:"status"`
    CreatedAt time.Time `json:"created_at"`
}
```

#### 3. Define Repository Interface

```go
// internal/feature/order/repository/order_repository.go
package repository

import (
    "context"
    "github.com/zercle/zercle-go-template/internal/feature/order/domain"
)

type OrderRepository interface {
    GetByID(ctx context.Context, id string) (*domain.Order, error)
    GetByUserID(ctx context.Context, userID string) ([]*domain.Order, error)
    Create(ctx context.Context, order *domain.Order) error
    Update(ctx context.Context, order *domain.Order) error
    Delete(ctx context.Context, id string) error
}
```

#### 4. Implement Repository (PostgreSQL)

```go
// internal/feature/order/repository/sqlc_repository.go
package repository

import "github.com/jackc/pgx/v5/pgxpool"

type SqlcOrderRepository struct {
    db *pgxpool.Pool
}

func NewSqlcOrderRepository(db *pgxpool.Pool) *SqlcOrderRepository {
    return &SqlcOrderRepository{db: db}
}

// Implement interface methods...
```

#### 5. Create Usecase

```go
// internal/feature/order/usecase/order_usecase.go
package usecase

type OrderUsecase struct {
    orderRepo OrderRepository
    userRepo  user.UserRepository  // Cross-feature dependency
    logger    logger.Logger
}

func NewOrderUsecase(
    orderRepo OrderRepository,
    userRepo user.UserRepository,
    logger logger.Logger,
) *OrderUsecase {
    return &OrderUsecase{
        orderRepo: orderRepo,
        userRepo:  userRepo,
        logger:    logger,
    }
}

func (u *OrderUsecase) CreateOrder(ctx context.Context, req dto.CreateOrderRequest) (*domain.Order, error) {
    // 1. Validate user exists
    _, err := u.userRepo.GetByID(ctx, req.UserID)
    if err != nil {
        return nil, err
    }
    
    // 2. Create order entity
    order := &domain.Order{
        ID:        uuid.New().String(),
        UserID:    req.UserID,
        Total:     req.Total,
        Status:    domain.OrderStatusPending,
        CreatedAt: time.Now(),
    }
    
    // 3. Validate
    if err := order.Validate(); err != nil {
        return nil, errors.ErrInvalidInput.WithMessage(err.Error())
    }
    
    // 4. Save
    if err := u.orderRepo.Create(ctx, order); err != nil {
        return nil, err
    }
    
    u.logger.Info().Str("order_id", order.ID).Msg("order created")
    return order, nil
}
```

#### 6. Create Handler

```go
// internal/feature/order/handler/order_handler.go
package handler

type OrderHandler struct {
    orderUsecase usecase.OrderUsecase
}

func NewOrderHandler(orderUsecase usecase.OrderUsecase) *OrderHandler {
    return &OrderHandler{orderUsecase: orderUsecase}
}

// @Summary     Create order
// @Description Create a new order
// @Tags        orders
// @Accept      json
// @Produce     json
// @Param       request body dto.CreateOrderRequest true "Order data"
// @Success     201 {object} dto.OrderResponse
// @Router      /orders [post]
func (h *OrderHandler) CreateOrder(c echo.Context) error {
    var req dto.CreateOrderRequest
    if err := c.Bind(&req); err != nil {
        return err
    }
    if err := c.Validate(&req); err != nil {
        return err
    }
    
    order, err := h.orderUsecase.CreateOrder(c.Request().Context(), req)
    if err != nil {
        return err
    }
    
    return c.JSON(http.StatusCreated, dto.ToOrderResponse(order))
}
```

#### 7. Add SQL Queries

```sql
-- internal/infrastructure/db/queries/orders.sql
-- name: GetOrderByID :one
SELECT * FROM orders WHERE id = $1;

-- name: GetOrdersByUserID :many
SELECT * FROM orders WHERE user_id = $1 ORDER BY created_at DESC;

-- name: CreateOrder :one
INSERT INTO orders (id, user_id, total, status, created_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: UpdateOrder :one
UPDATE orders
SET status = $2, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteOrder :exec
DELETE FROM orders WHERE id = $1;
```

#### 8. Create Migration

```sql
-- internal/infrastructure/db/migrations/002_add_orders.sql
CREATE TABLE orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    total DECIMAL(10, 2) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_orders_user_id ON orders(user_id);
CREATE INDEX idx_orders_status ON orders(status);
```

#### 9. Wire Dependencies

Update [`internal/container/container.go`](internal/container/container.go):

```go
// Add to Container struct
type Container struct {
    // ... existing fields ...
    orderRepo   order.OrderRepository
    orderUsecase order.OrderUsecase
}

// Add functional option
func WithPostgresOrderRepository(db *pgxpool.Pool) ContainerOption {
    return func(c *Container) error {
        c.orderRepo = orderRepo.NewSqlcOrderRepository(db)
        return nil
    }
}

// Add to Build()
func (c *Container) Build() error {
    // ... existing code ...
    c.orderUsecase = orderUC.NewOrderUsecase(c.orderRepo, c.userRepo, c.logger)
    return nil
}
```

#### 10. Register Routes

Update [`cmd/api/main.go`](cmd/api/main.go):

```go
// Add to route registration
orderHandler := container.GetOrderHandler()
orders := e.Group("/orders", authMiddleware)
{
    orders.POST("", orderHandler.CreateOrder)
    orders.GET("/:id", orderHandler.GetOrder)
    orders.GET("", orderHandler.ListOrders)
}
```

#### 11. Generate Code & Run Migrations

```bash
# Generate SQLC code
make sqlc

# Run migrations
make migrate-up

# Generate Swagger docs
make swag
```

#### 12. Write Tests

```bash
# Generate mocks
//go:generate mockgen -source=order_repository.go -destination=mocks/order_repository_mock.go

# Run tests
make test
```

## Testing Procedures

### Test Organization

```
feature/user/
├── handler/
│   ├── user_handler.go
│   └── user_handler_test.go          # Unit tests
│   └── user_handler_integration_test.go  # Integration tests
├── repository/
│   ├── sqlc_repository.go
│   └── sqlc_repository_integration_test.go
└── usecase/
    ├── user_usecase.go
    └── user_usecase_test.go
```

### Unit Test Pattern

```go
func TestUserUsecase_CreateUser(t *testing.T) {
    tests := []struct {
        name      string
        input     dto.CreateUserRequest
        mockSetup func(*mocks.MockUserRepository)
        wantErr   bool
        errCode   string
    }{
        {
            name:  "success",
            input: dto.CreateUserRequest{Email: "test@example.com", Password: "password123"},
            mockSetup: func(m *mocks.MockUserRepository) {
                m.EXPECT().GetByEmail(gomock.Any(), "test@example.com").Return(nil, errors.ErrNotFound)
                m.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
            },
            wantErr: false,
        },
        {
            name:  "email_already_exists",
            input: dto.CreateUserRequest{Email: "exists@example.com", Password: "password123"},
            mockSetup: func(m *mocks.MockUserRepository) {
                m.EXPECT().GetByEmail(gomock.Any(), "exists@example.com").Return(&domain.User{}, nil)
            },
            wantErr: true,
            errCode: "CONFLICT",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            ctrl := gomock.NewController(t)
            defer ctrl.Finish()
            
            mockRepo := mocks.NewMockUserRepository(ctrl)
            tt.mockSetup(mockRepo)
            
            uc := usecase.NewUserUsecase(mockRepo, logger.NewNop())
            _, err := uc.CreateUser(context.Background(), tt.input)
            
            if tt.wantErr {
                assert.Error(t, err)
                var appErr *errors.AppError
                if errors.As(err, &appErr) {
                    assert.Equal(t, tt.errCode, appErr.Code)
                }
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

### Integration Test Pattern

```go
func TestUserRepositoryIntegration(t *testing.T) {
    ctx := context.Background()
    
    // Setup test database
    db := setupTestDB(t)
    defer db.Close()
    
    repo := repository.NewSqlcUserRepository(db)
    
    t.Run("create_and_get_user", func(t *testing.T) {
        user := &domain.User{
            ID:    uuid.New().String(),
            Email: "test@example.com",
            Name:  "Test User",
        }
        
        // Create
        err := repo.Create(ctx, user)
        require.NoError(t, err)
        
        // Get
        found, err := repo.GetByID(ctx, user.ID)
        require.NoError(t, err)
        assert.Equal(t, user.Email, found.Email)
        assert.Equal(t, user.Name, found.Name)
    })
}
```

### Running Tests

```bash
# All tests
make test

# Unit tests only
go test ./... -short

# Integration tests only
go test ./... -run Integration

# Specific package
go test ./internal/feature/user/...

# With coverage
make test-coverage

# Verbose output
go test -v ./...
```

## Build and Deployment

### Local Build

```bash
# Build binary
make build
# Output: bin/api

# Run binary
./bin/api
```

### Docker Build

```bash
# Build image
docker build -t zercle-go-template:latest .

# Run container
docker run -p 8080:8080 --env-file .env zercle-go-template:latest
```

### Docker Compose

```bash
# Start all services
make docker-up
# or
docker-compose up -d

# View logs
docker-compose logs -f api

# Stop services
make docker-down
# or
docker-compose down

# Full reset (volumes)
docker-compose down -v
```

### Database Migrations

```bash
# Run migrations
make migrate-up

# Rollback last migration
make migrate-down

# Create new migration
migrate create -ext sql -dir internal/infrastructure/db/migrations -seq add_orders

# Check version
migrate -path internal/infrastructure/db/migrations -database $DB_URL version
```

## Code Review Checklist

### Before Creating PR

- [ ] Tests pass: `make test`
- [ ] Linter passes: `make lint`
- [ ] Code formatted: `go fmt ./...`
- [ ] Swagger updated: `make swag`
- [ ] SQLC generated: `make sqlc` (if queries changed)
- [ ] Migrations tested: `make migrate-up` works

### Architecture Review

- [ ] Domain layer has no external dependencies
- [ ] Interfaces defined by consumer (usecase defines repo interface)
- [ ] Errors use custom error types from `internal/errors`
- [ ] Context passed through all layers
- [ ] No business logic in handlers

### Code Quality

- [ ] Functions are focused and small (< 50 lines)
- [ ] Variable names are descriptive
- [ ] Exported functions have documentation comments
- [ ] No magic numbers/strings (use constants)
- [ ] Proper error wrapping with context

### Testing

- [ ] Unit tests for usecase layer
- [ ] Table-driven tests where appropriate
- [ ] Mock generation with `//go:generate mockgen`
- [ ] Integration tests for repository layer
- [ ] Edge cases covered

### Security

- [ ] Input validation at handler layer
- [ ] SQL queries use parameterized statements (via sqlc)
- [ ] Passwords hashed with bcrypt
- [ ] No secrets logged
- [ ] Authorization checks in usecase

### API Design

- [ ] RESTful endpoint naming
- [ ] Proper HTTP status codes
- [ ] Consistent response structure
- [ ] Swagger annotations complete
- [ ] Error responses follow standard format

## Troubleshooting

### Common Issues

**Tests failing with "connection refused"**
```bash
# Database not running
make docker-up
# or for unit tests only
go test ./... -short
```

**SQLC generation errors**
```bash
# Ensure schema is valid
make migrate-up
# Regenerate
make sqlc
```

**Linting errors**
```bash
# Auto-fix where possible
golangci-lint run --fix
```

**Port already in use**
```bash
# Find and kill process
lsof -ti:8080 | xargs kill -9
# or change port in .env
```

## Git Workflow

```bash
# Create feature branch
git checkout -b feature/order-management

# Regular commits
git add .
git commit -m "feat: add order domain model"

# Push branch
git push -u origin feature/order-management

# Create PR via GitHub/GitLab
# After review and approval...

# Merge to main
git checkout main
git pull origin main
```

### Commit Message Convention

```
feat:     New feature
fix:      Bug fix
docs:     Documentation changes
style:    Code style (formatting, no logic change)
refactor: Code refactoring
test:     Test changes
chore:    Build/config/tooling changes
```
