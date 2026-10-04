// Echo CORS middleware configuration.
package middleware

import (
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"

	"github.com/zercle/zercle-go-template/internal/infrastructure/config"
)

// defaultCORSMethods mirrors the echo CORS middleware default when none are
// configured.
var defaultCORSMethods = []string{"GET", "HEAD", "PUT", "PATCH", "POST", "DELETE"}

// defaultCORSHeaders mirrors the echo CORS middleware default when none are
// configured.
var defaultCORSHeaders = []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization}

// defaultCORSExposeHeaders is the default list of response headers exposed to
// the browser when none are configured.
var defaultCORSExposeHeaders = []string{"Content-Length"}

// defaultCORSMaxAge is the default CORS preflight cache duration (seconds)
// when not configured.
const defaultCORSMaxAge = 86400

// CORS returns echo's built-in CORS middleware configured from cfg.HTTP.CORS*.
// A nil cfg or an empty list falls back to the package defaults: all origins,
// the standard methods/headers, Content-Length exposed, and a 24h preflight
// cache. Defaults are applied per field, so a partial config keeps them for the
// fields it does not set.
func CORS(cfg *config.Config) echo.MiddlewareFunc {
	var corsCfg middleware.CORSConfig
	if cfg != nil {
		corsCfg.AllowOrigins = cfg.HTTP.CORSAllowOrigins
		corsCfg.AllowMethods = cfg.HTTP.CORSAllowMethods
		corsCfg.AllowHeaders = cfg.HTTP.CORSAllowHeaders
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
	corsCfg.ExposeHeaders = defaultCORSExposeHeaders
	corsCfg.MaxAge = defaultCORSMaxAge

	return middleware.CORSWithConfig(corsCfg)
}
