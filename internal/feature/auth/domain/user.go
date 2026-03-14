package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/zercle/zercle-go-template/pkg/uuidgen"
)

// User represents a user entity.
type User struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	Username    string     `json:"username" db:"username"`
	Email       string     `json:"email" db:"email"`
	Password    string     `json:"-" db:"password_hash"`
	DisplayName string     `json:"display_name" db:"display_name"`
	AvatarURL   string     `json:"avatar_url" db:"avatar_url"`
	Status      string     `json:"status" db:"status"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// NewUser creates a new user instance.
func NewUser(username, email, password, displayName string) *User {
	now := time.Now()
	return &User{
		ID:          uuidgen.New(),
		Username:    username,
		Email:       email,
		Password:    password,
		DisplayName: displayName,
		Status:      "offline",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// Validate validates the user data.
func (u *User) Validate() error {
	if u.Username == "" {
		return ErrUsernameRequired
	}
	if u.Email == "" {
		return ErrEmailRequired
	}
	if u.Password != "" && len(u.Password) < 8 {
		return ErrPasswordTooShort
	}
	return nil
}

// UserStatus represents the online status of a user.
type UserStatus string

// User status constants.
const (
	StatusOnline  UserStatus = "online"
	StatusAway    UserStatus = "away"
	StatusOffline UserStatus = "offline"
)

// PublicUser represents public user information.
type PublicUser struct {
	ID          uuid.UUID `json:"id"`
	Username    string    `json:"username"`
	Email       string    `json:"email"`
	DisplayName string    `json:"display_name"`
	AvatarURL   string    `json:"avatar_url"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

// ToPublic converts a User to a PublicUser.
func (u *User) ToPublic() *PublicUser {
	if u == nil {
		return nil
	}
	return &PublicUser{
		ID:          u.ID,
		Username:    u.Username,
		Email:       u.Email,
		DisplayName: u.DisplayName,
		AvatarURL:   u.AvatarURL,
		Status:      u.Status,
		CreatedAt:   u.CreatedAt,
	}
}
