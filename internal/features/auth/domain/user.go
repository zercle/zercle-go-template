package domain

import (
	"time"

	"github.com/google/uuid"

	"github.com/zercle/zercle-go-template/internal/shared/errors"
	"github.com/zercle/zercle-go-template/pkg/uuidgen"
)

// User represents an authenticated user entity.
type User struct {
	ID          uuid.UUID  `json:"id"`
	Username    string     `json:"username"`
	Email       string     `json:"email"`
	Password    string     `json:"-"`
	DisplayName string     `json:"display_name"`
	AvatarURL   string     `json:"avatar_url"`
	Status      string     `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

// NewUser creates a new User with the specified attributes.
func NewUser(username, email, password, displayName string) *User {
	now := time.Now()
	return &User{
		ID:          uuidgen.New(),
		Username:    username,
		Email:       email,
		Password:    password,
		DisplayName: displayName,
		Status:      string(StatusOffline),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// Validate checks if the user data is valid.
func (u *User) Validate() error {
	if u.Username == "" {
		return errors.ErrUsernameRequired
	}
	if u.Email == "" {
		return errors.ErrEmailRequired
	}
	if u.Password != "" && len(u.Password) < 8 {
		return errors.ErrPasswordTooShort
	}
	return nil
}

// Session represents an active user session.
type Session struct {
	UserID    uuid.UUID `json:"user_id"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

// UserStatus represents the online status of a user.
type UserStatus string

// User status constants.
const (
	StatusOnline  UserStatus = "online"
	StatusAway    UserStatus = "away"
	StatusOffline UserStatus = "offline"
)
