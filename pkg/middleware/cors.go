package middleware

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/zercle/zercle-go-template/infrastructure/config"
)

// CORS creates an Echo middleware that handles Cross-Origin Resource Sharing (CORS) headers.
// It allows configured origins and standard HTTP methods with credentials support.
// The middleware caches preflight requests for 24 hours.
func CORS(cfg *config.CORSConfig) echo.MiddlewareFunc {
	return middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     cfg.AllowedOrigins,
		AllowMethods:     []string{echo.GET, echo.POST, echo.PUT, echo.PATCH, echo.DELETE, echo.OPTIONS},
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
		ExposeHeaders:    []string{echo.HeaderContentLength, "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           86400, // 24 hours
	})
}
