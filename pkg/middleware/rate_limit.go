package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/zercle/zercle-go-template/infrastructure/config"
	"github.com/zercle/zercle-go-template/pkg/response"
)

// TokenBucket implements a token bucket rate limiting algorithm.
// Tokens are refilled at a fixed rate, and requests consume tokens.
type TokenBucket struct {
	lastRefill time.Time
	tokens     int
	maxTokens  int
	refillRate int
	window     time.Duration
	mu         sync.Mutex
}

// NewTokenBucket creates a new token bucket rate limiter with the specified capacity and refill rate.
// The bucket starts with full capacity.
func NewTokenBucket(maxTokens, refillRate int, window time.Duration) *TokenBucket {
	return &TokenBucket{
		tokens:     maxTokens,
		maxTokens:  maxTokens,
		refillRate: refillRate,
		window:     window,
		lastRefill: time.Now(),
	}
}

// Allow checks if a request can proceed under the rate limit.
// Returns true if a token is available, false if the bucket is empty.
// Automatically refills tokens based on elapsed time.
func (tb *TokenBucket) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(tb.lastRefill)
	if elapsed >= tb.window {
		tb.tokens = tb.maxTokens
		tb.lastRefill = now
	}

	if tb.tokens > 0 {
		tb.tokens--
		return true
	}
	return false
}

// RateLimiter manages rate limiting per IP address
type RateLimiter struct {
	buckets map[string]*TokenBucket
	cfg     *config.RateLimitConfig
	mu      sync.RWMutex
}

// NewRateLimiter creates a new rate limiter with the given configuration.
func NewRateLimiter(cfg *config.RateLimitConfig) *RateLimiter {
	return &RateLimiter{
		buckets: make(map[string]*TokenBucket),
		cfg:     cfg,
	}
}

// Allow checks if the specified IP address is allowed to make a request.
// Creates a new token bucket for the IP if it doesn't exist.
func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	bucket, exists := rl.buckets[ip]
	if !exists {
		bucket = NewTokenBucket(rl.cfg.Requests, rl.cfg.Requests, time.Duration(rl.cfg.Window)*time.Second)
		rl.buckets[ip] = bucket
	}
	rl.mu.Unlock()

	return bucket.Allow()
}

// RateLimit creates an Echo middleware that enforces per-IP rate limiting.
// Returns HTTP 429 (Too Many Requests) when the rate limit is exceeded.
func RateLimit(cfg *config.RateLimitConfig) echo.MiddlewareFunc {
	limiter := NewRateLimiter(cfg)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ip := c.RealIP()

			if !limiter.Allow(ip) {
				return response.Error(c, http.StatusTooManyRequests, "Rate limit exceeded")
			}

			return next(c)
		}
	}
}
