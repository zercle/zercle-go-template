// Echo CORS middleware configuration.
package middleware

import (
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"

	"github.com/zercle/zercle-go-template/internal/config"
)

// defaultCORSMethods mirrors the echo CORS middleware default when none are
// configured.
var defaultCORSMethods = []string{"GET", "HEAD", "PUT", "PATCH", "POST", "DELETE"}

// defaultCORSHeaders mirrors the echo CORS middleware default when none are
// configured.
var defaultCORSHeaders = []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization}

// CORS returns echo's built-in CORS middleware configured from cfg.HTTP.CORS*.
// When no origins are configured it defaults to allowing all origins.
func CORS(cfg *config.Config) echo.MiddlewareFunc {
	corsCfg := middleware.CORSConfig{
		AllowOrigins: cfg.HTTP.CORSAllowOrigins,
		AllowMethods: cfg.HTTP.CORSAllowMethods,
		AllowHeaders: cfg.HTTP.CORSAllowHeaders,
	}

	if len(corsCfg.AllowOrigins) == 0 {
		corsCfg.AllowOrigins = []string{"*"}
	}
	if len(corsCfg.AllowMethods) == 0 {
		corsCfg.AllowMethods = defaultCORSMethods
	}
	if len(corsCfg.AllowHeaders) == 0 {
		corsCfg.AllowHeaders = defaultCORSHeaders
	}

	return middleware.CORSWithConfig(corsCfg)
}
