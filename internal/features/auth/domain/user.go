package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/zercle/zercle-go-template/internal/errors"
	"github.com/zercle/zercle-go-template/pkg/uuidgen"
)

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

type Session struct {
	UserID    uuid.UUID `json:"user_id"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

type UserStatus string

const (
	StatusOnline  UserStatus = "online"
	StatusAway    UserStatus = "away"
	StatusOffline UserStatus = "offline"
)
