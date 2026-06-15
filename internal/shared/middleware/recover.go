// Echo panic recovery middleware.
package middleware

import (
	"github.com/labstack/echo/v5"
	"github.com/rs/zerolog"

	sharederrors "github.com/zercle/zercle-go-template/internal/shared/errors"
)

// Recover returns echo middleware that recovers from panics, logs the failure
// with the request id, and returns a structured 500 response.
func Recover(logger *zerolog.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			defer func() {
				if r := recover(); r != nil {
					log := logger.Error().
						Str("request_id", RequestIDFromContext(c)).
						Str("method", c.Request().Method).
						Str("path", c.Request().URL.Path)

					if recErr, ok := r.(error); ok {
						log = log.Err(recErr)
					}
					log.Msg("request panic recovered")

					status, body := sharederrors.HTTPError(sharederrors.ErrInternal)
					_ = c.JSON(status, body)
				}
			}()

			return next(c)
		}
	}
}
