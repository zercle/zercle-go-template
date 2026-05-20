package errors

import "errors"

// Common sentinel errors for general API usage.
var (
	ErrNotFound      = errors.New("resource not found")
	ErrUnauthorized  = errors.New("unauthorized")
	ErrForbidden     = errors.New("forbidden")
	ErrAlreadyExists = errors.New("resource already exists")
	ErrInvalidInput  = errors.New("invalid input")
	ErrInternalError = errors.New("internal error")
)

// Authentication and user-related errors.
var (
	ErrUsernameRequired       = errors.New("username is required")
	ErrEmailRequired          = errors.New("email is required")
	ErrPasswordRequired       = errors.New("password is required")
	ErrPasswordTooShort       = errors.New("password must be at least 8 characters")
	ErrInvalidEmail           = errors.New("invalid email format")
	ErrUserNotFound           = errors.New("user not found")
	ErrInvalidCredentials     = errors.New("invalid credentials")
	ErrTokenExpired           = errors.New("token expired")
	ErrTokenInvalid           = errors.New("token invalid")
	ErrRoomNotFound           = errors.New("room not found")
	ErrMessageNotFound        = errors.New("message not found")
	ErrRoomNameRequired       = errors.New("room name is required")
	ErrInvalidRoomType        = errors.New("invalid room type")
	ErrMessageContentRequired = errors.New("message content is required")
	ErrAlreadyJoined          = errors.New("already joined room")
	ErrNotMember              = errors.New("not a member of room")
)
