// Echo request logging middleware, built on echo's RequestLogger middleware
// with zerolog output: one structured line per request.
package middleware

import (
	"github.com/labstack/echo/v5"
	echomw "github.com/labstack/echo/v5/middleware"
	"github.com/rs/zerolog"
)

// RequestLogger returns echo middleware that emits exactly one structured
// zerolog line per request with method, path, status, latency, and request id.
// The error from the handler chain keeps bubbling up so echo's centralized
// HTTPErrorHandler writes the response exactly once.
func RequestLogger(logger *zerolog.Logger) echo.MiddlewareFunc {
	return echomw.RequestLoggerWithConfig(echomw.RequestLoggerConfig{
		LogMethod:    true,
		LogURIPath:   true,
		LogStatus:    true,
		LogLatency:   true,
		LogRequestID: false,
		HandleError:  false,
		LogValuesFunc: func(c *echo.Context, v echomw.RequestLoggerValues) error {
			logger.Info().
				Str("request_id", RequestIDFromContext(c)).
				Str("method", v.Method).
				Str("path", v.URIPath).
				Int("status", v.Status).
				Dur("latency", v.Latency).
				Msg("http request")
			return nil
		},
	})
}
