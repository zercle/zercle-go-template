package middleware

import (
	"fmt"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

// RateLimitEntry tracks request counts for rate limiting
type RateLimitEntry struct {
	count     int
	resetTime time.Time
}

// RateLimiter implements token bucket algorithm for rate limiting
type RateLimiter struct {
	mu       sync.RWMutex
	requests map[string]*RateLimitEntry
	limit    int
	window   time.Duration
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		requests: make(map[string]*RateLimitEntry),
		limit:    limit,
		window:   window,
	}

	// Cleanup old entries periodically
	go rl.cleanup()

	return rl
}

// Allow checks if a request is allowed for the given IP
func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	entry, exists := rl.requests[ip]

	if !exists || now.After(entry.resetTime) {
		// New window
		rl.requests[ip] = &RateLimitEntry{
			count:     1,
			resetTime: now.Add(rl.window),
		}
		return true
	}

	if entry.count < rl.limit {
		entry.count++
		return true
	}

	return false
}

// cleanup removes expired entries
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(rl.window)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for ip, entry := range rl.requests {
			if now.After(entry.resetTime) {
				delete(rl.requests, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// RateLimitMiddleware creates a rate limiting middleware
func RateLimitMiddleware(limiter *RateLimiter) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ip := c.IP()

		if !limiter.Allow(ip) {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"success": false,
				"code":    "RATE_LIMIT_EXCEEDED",
				"message": fmt.Sprintf("Rate limit exceeded. Max requests per %v", limiter.window),
			})
		}

		return c.Next()
	}
}
