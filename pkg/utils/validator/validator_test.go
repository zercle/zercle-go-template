package validator_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	postDto "github.com/zercle/zercle-go-template/internal/features/post/dto"
	userDto "github.com/zercle/zercle-go-template/internal/features/user/dto"
	"github.com/zercle/zercle-go-template/pkg/utils/validator"
)

func TestValidate_RegisterRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     userDto.RegisterRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request",
			req: userDto.RegisterRequest{
				Email:    "test@example.com",
				Password: "password123",
				Name:     "Test User",
			},
			wantErr: false,
		},
		{
			name: "missing email",
			req: userDto.RegisterRequest{
				Password: "password123",
				Name:     "Test User",
			},
			wantErr: true,
			errMsg:  "email is required",
		},
		{
			name: "invalid email",
			req: userDto.RegisterRequest{
				Email:    "not-an-email",
				Password: "password123",
				Name:     "Test User",
			},
			wantErr: true,
			errMsg:  "email must be a valid email address",
		},
		{
			name: "password too short",
			req: userDto.RegisterRequest{
				Email:    "test@example.com",
				Password: "short",
				Name:     "Test User",
			},
			wantErr: true,
			errMsg:  "password must be at least 8 characters long",
		},
		{
			name: "name too short",
			req: userDto.RegisterRequest{
				Email:    "test@example.com",
				Password: "password123",
				Name:     "A",
			},
			wantErr: true,
			errMsg:  "name must be at least 2 characters long",
		},
		{
			name:    "all fields missing",
			req:     userDto.RegisterRequest{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(&tt.req)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidate_LoginRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     userDto.LoginRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request",
			req: userDto.LoginRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			wantErr: false,
		},
		{
			name: "missing email",
			req: userDto.LoginRequest{
				Password: "password123",
			},
			wantErr: true,
			errMsg:  "email is required",
		},
		{
			name: "invalid email format",
			req: userDto.LoginRequest{
				Email:    "invalid-email",
				Password: "password123",
			},
			wantErr: true,
			errMsg:  "email must be a valid email address",
		},
		{
			name: "missing password",
			req: userDto.LoginRequest{
				Email: "test@example.com",
			},
			wantErr: true,
			errMsg:  "password is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(&tt.req)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidate_CreatePostRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     postDto.CreatePostRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request",
			req: postDto.CreatePostRequest{
				Title:   "Test Post",
				Content: "This is test content with enough characters",
			},
			wantErr: false,
		},
		{
			name: "title too short",
			req: postDto.CreatePostRequest{
				Title:   "AB",
				Content: "This is test content",
			},
			wantErr: true,
			errMsg:  "title must be at least 3 characters long",
		},
		{
			name: "content too short",
			req: postDto.CreatePostRequest{
				Title:   "Test Post",
				Content: "short",
			},
			wantErr: true,
			errMsg:  "content must be at least 10 characters long",
		},
		{
			name:    "missing required fields",
			req:     postDto.CreatePostRequest{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(&tt.req)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidate_UpdateUserRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     userDto.UpdateUserRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request with name",
			req: userDto.UpdateUserRequest{
				Name: "New Name",
			},
			wantErr: false,
		},
		{
			name:    "empty request (all optional)",
			req:     userDto.UpdateUserRequest{},
			wantErr: false,
		},
		{
			name: "name too short when provided",
			req: userDto.UpdateUserRequest{
				Name: "A",
			},
			wantErr: true,
			errMsg:  "name must be at least 2 characters long",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(&tt.req)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidate_UpdatePostRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     postDto.UpdatePostRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request with both fields",
			req: postDto.UpdatePostRequest{
				Title:   "Updated Title",
				Content: "Updated content with enough length",
			},
			wantErr: false,
		},
		{
			name: "valid request with title only",
			req: postDto.UpdatePostRequest{
				Title: "Updated Title",
			},
			wantErr: false,
		},
		{
			name:    "empty request (all optional)",
			req:     postDto.UpdatePostRequest{},
			wantErr: false,
		},
		{
			name: "title too short when provided",
			req: postDto.UpdatePostRequest{
				Title: "AB",
			},
			wantErr: true,
			errMsg:  "title must be at least 3 characters long",
		},
		{
			name: "content too short when provided",
			req: postDto.UpdatePostRequest{
				Content: "short",
			},
			wantErr: true,
			errMsg:  "content must be at least 10 characters long",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(&tt.req)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
