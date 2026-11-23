// Echo security headers middleware, built on echo's Secure middleware.
package middleware

import (
	"github.com/labstack/echo/v5"
	echomw "github.com/labstack/echo/v5/middleware"

	"github.com/zercle/zercle-go-template/internal/infrastructure/config"
)

// Secure returns echo's Secure middleware configured with the template's
// security headers. HSTSMaxAge comes from cfg.HTTP.HSTSMaxAge; a zero value
// (the default) disables the Strict-Transport-Security header, and a nil cfg
// is treated the same way.
func Secure(cfg *config.Config) echo.MiddlewareFunc {
	var hstsMaxAge int
	if cfg != nil {
		hstsMaxAge = int(cfg.HTTP.HSTSMaxAge.Seconds())
	}

	return echomw.SecureWithConfig(echomw.SecureConfig{
		XSSProtection:         "1; mode=block",
		ContentTypeNosniff:    "nosniff",
		XFrameOptions:         "SAMEORIGIN",
		HSTSMaxAge:            hstsMaxAge,
		HSTSExcludeSubdomains: false,
		HSTSPreloadEnabled:    false,
		ContentSecurityPolicy: "",
	})
}
