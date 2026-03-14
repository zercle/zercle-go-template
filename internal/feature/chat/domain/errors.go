package domain

import "errors"

// Domain-specific error variables.
var (
	ErrRoomNameRequired       = errors.New("room name is required")
	ErrInvalidRoomType        = errors.New("invalid room type")
	ErrMessageContentRequired = errors.New("message content is required")
	ErrRoomNotFound           = errors.New("room not found")
	ErrMessageNotFound        = errors.New("message not found")
	ErrAlreadyJoined          = errors.New("already joined room")
	ErrNotMember              = errors.New("not a member of room")
)
