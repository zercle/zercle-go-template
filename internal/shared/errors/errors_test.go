package errors

import (
	"errors"
	"testing"
)

func TestSentinelErrors_AreDistinct(t *testing.T) {
	t.Parallel()
	errorVars := []error{
		ErrNotFound,
		ErrUnauthorized,
		ErrForbidden,
		ErrAlreadyExists,
		ErrInvalidInput,
		ErrInternalError,
		ErrUsernameRequired,
		ErrEmailRequired,
		ErrPasswordRequired,
		ErrPasswordTooShort,
		ErrInvalidEmail,
		ErrUserNotFound,
		ErrInvalidCredentials,
		ErrTokenExpired,
		ErrTokenInvalid,
		ErrRoomNotFound,
		ErrMessageNotFound,
		ErrRoomNameRequired,
		ErrInvalidRoomType,
		ErrMessageContentRequired,
		ErrAlreadyJoined,
		ErrNotMember,
	}

	for _, err := range errorVars {
		if err == nil {
			t.Errorf("expected non-nil error, got nil")
		}
	}

	for i := range len(errorVars) {
		for j := i + 1; j < len(errorVars); j++ {
			if errors.Is(errorVars[i], errorVars[j]) {
				t.Errorf("errors %v and %v are the same", errorVars[i], errorVars[j])
			}
			msg1 := errorVars[i].Error()
			msg2 := errorVars[j].Error()
			if msg1 == msg2 {
				t.Errorf("errors %v and %v have the same message: %s", errorVars[i], errorVars[j], msg1)
			}
		}
	}
}

func TestSentinelErrors_SupportErrorsIs(t *testing.T) {
	t.Parallel()
	errorVars := []error{
		ErrNotFound,
		ErrUnauthorized,
		ErrForbidden,
		ErrAlreadyExists,
		ErrInvalidInput,
		ErrInternalError,
		ErrUsernameRequired,
		ErrEmailRequired,
		ErrPasswordRequired,
		ErrPasswordTooShort,
		ErrInvalidEmail,
		ErrUserNotFound,
		ErrInvalidCredentials,
		ErrTokenExpired,
		ErrTokenInvalid,
		ErrRoomNotFound,
		ErrMessageNotFound,
		ErrRoomNameRequired,
		ErrInvalidRoomType,
		ErrMessageContentRequired,
		ErrAlreadyJoined,
		ErrNotMember,
	}

	for _, err := range errorVars {
		if !errors.Is(err, err) {
			t.Errorf("errors.Is(%v, %v) returned false", err, err)
		}
	}
}
