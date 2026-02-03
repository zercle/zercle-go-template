// Package observability provides metrics, tracing, and health checking capabilities.
package observability

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// HealthChecker is the interface for health check components.
type HealthChecker interface {
	Name() string
	Check(ctx context.Context) error
}

// HealthAggregator aggregates multiple health checkers.
type HealthAggregator struct {
	checkers []HealthChecker
}

// NewHealthAggregator creates a new HealthAggregator.
func NewHealthAggregator() *HealthAggregator {
	return &HealthAggregator{
		checkers: make([]HealthChecker, 0),
	}
}

// Register adds a health checker to the aggregator.
func (h *HealthAggregator) Register(checker HealthChecker) {
	h.checkers = append(h.checkers, checker)
}

// Check runs all registered health checks and returns the aggregated result.
func (h *HealthAggregator) Check(ctx context.Context) *HealthResponse {
	response := &HealthResponse{
		Status:     "healthy",
		Components: make([]ComponentHealth, 0, len(h.checkers)),
	}

	for _, checker := range h.checkers {
		component := ComponentHealth{
			Name: checker.Name(),
		}

		checkCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		err := checker.Check(checkCtx)
		cancel()

		if err != nil {
			component.Status = "unhealthy"
			component.Error = err.Error()
			response.Status = "unhealthy"
		} else {
			component.Status = "healthy"
		}

		response.Components = append(response.Components, component)
	}

	return response
}

// HealthResponse represents the overall health check response.
type HealthResponse struct {
	Status     string            `json:"status"`
	Components []ComponentHealth `json:"components"`
}

// ComponentHealth represents the health of a single component.
type ComponentHealth struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

// DatabaseHealthChecker checks database connectivity.
type DatabaseHealthChecker struct {
	pool *pgxpool.Pool
}

// NewDatabaseHealthChecker creates a new DatabaseHealthChecker.
func NewDatabaseHealthChecker(pool *pgxpool.Pool) *DatabaseHealthChecker {
	return &DatabaseHealthChecker{pool: pool}
}

// Name returns "database".
func (d *DatabaseHealthChecker) Name() string {
	return "database"
}

// Check executes a simple query to verify database connectivity.
func (d *DatabaseHealthChecker) Check(ctx context.Context) error {
	return d.pool.QueryRow(ctx, "SELECT 1").Scan()
}

// CacheHealthChecker checks Redis/Valkey connectivity.
type CacheHealthChecker struct {
	client *redis.Client
}

// NewCacheHealthChecker creates a new CacheHealthChecker.
func NewCacheHealthChecker(client *redis.Client) *CacheHealthChecker {
	return &CacheHealthChecker{client: client}
}

// Name returns "cache".
func (c *CacheHealthChecker) Name() string {
	return "cache"
}

// Check executes a PING command to verify cache connectivity.
func (c *CacheHealthChecker) Check(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}
