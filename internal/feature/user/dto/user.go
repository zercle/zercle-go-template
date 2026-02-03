// Package dto contains Data Transfer Objects for request/response handling for the user feature.
// DTOs decouple the API contract from internal domain models.
package dto

import (
	"time"

	"zercle-go-template/internal/feature/user/domain"
)

// CreateUserRequest represents the request to create a new user.
type CreateUserRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Name     string `json:"name" validate:"required,min=2,max=100"`
	Password string `json:"password" validate:"required,min=8"`
}

// UpdateUserRequest represents the request to update an existing user.
type UpdateUserRequest struct {
	Name string `json:"name" validate:"omitempty,min=2,max=100"`
}

// UpdatePasswordRequest represents the request to update a user's password.
type UpdatePasswordRequest struct {
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8"`
}

// UserResponse represents the user data in API responses.
type UserResponse struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ListUsersResponse represents a paginated list of users.
type ListUsersResponse struct {
	Users []UserResponse `json:"users"`
	Total int            `json:"total"`
	Page  int            `json:"page"`
	Limit int            `json:"limit"`
}

// UserLoginRequest represents the request for user login.
type UserLoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// UserLoginResponse represents the response for successful login.
type UserLoginResponse struct {
	User  UserResponse `json:"user"`
	Token string       `json:"token"`
}

// ToUserResponse converts a domain.User to UserResponse.
func ToUserResponse(user *domain.User) UserResponse {
	return UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}

// ToUserListResponse converts a slice of domain.User to ListUsersResponse.
func ToUserListResponse(users []*domain.User, total, page, limit int) ListUsersResponse {
	response := ListUsersResponse{
		Users: make([]UserResponse, len(users)),
		Total: total,
		Page:  page,
		Limit: limit,
	}

	for i, user := range users {
		response.Users[i] = ToUserResponse(user)
	}

	return response
}
