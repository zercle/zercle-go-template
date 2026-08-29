// Package errcodes defines the version-independent error codes carried in the
// {"error": code, "message": msg} response envelope. They live outside the
// versioned contract (pkg/api/v1) so a future v2 facade can re-export the
// same codes instead of forking them.
//
// The boundary sentinels in internal/platform/errors source their Code values
// from these constants, so published codes and served codes cannot drift.
// Consumers may import this package to branch on error codes without
// depending on server internals.
package errcodes

const (
	NotFound         = "NOT_FOUND"
	InvalidInput     = "INVALID_INPUT"
	Unauthorized     = "UNAUTHORIZED"
	Forbidden        = "FORBIDDEN"
	Conflict         = "CONFLICT"
	Canceled         = "CANCELED"
	DeadlineExceeded = "DEADLINE_EXCEEDED"
	Internal         = "INTERNAL"
)
