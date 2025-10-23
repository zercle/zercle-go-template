package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/zercle/zercle-go-template/internal/middleware"
)

func TestRateLimitMiddleware(t *testing.T) {
	// Limit 2 requests per 1 second
	limiter := middleware.NewRateLimiter(2, time.Second)
	app := fiber.New()
	app.Use(middleware.RateLimitMiddleware(limiter))
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	// 1st request - OK
	req1 := httptest.NewRequest("GET", "/", nil)
	resp1, err := app.Test(req1)
	assert.NoError(t, err)
	if resp1 != nil {
		defer resp1.Body.Close()
		assert.Equal(t, http.StatusOK, resp1.StatusCode)
	}

	// 2nd request - OK
	req2 := httptest.NewRequest("GET", "/", nil)
	resp2, err := app.Test(req2)
	assert.NoError(t, err)
	if resp2 != nil {
		defer resp2.Body.Close()
		assert.Equal(t, http.StatusOK, resp2.StatusCode)
	}

	// 3rd request - Too Many Requests
	req3 := httptest.NewRequest("GET", "/", nil)
	resp3, err := app.Test(req3)
	assert.NoError(t, err)
	if resp3 != nil {
		defer resp3.Body.Close()
		assert.Equal(t, http.StatusTooManyRequests, resp3.StatusCode)
	}
}
