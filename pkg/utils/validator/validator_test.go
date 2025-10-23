package validator_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zercle/zercle-go-template/pkg/dto"
	"github.com/zercle/zercle-go-template/pkg/utils/validator"
)

func TestValidate_RegisterRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     dto.RegisterRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request",
			req: dto.RegisterRequest{
				Email:    "test@example.com",
				Password: "password123",
				Name:     "Test User",
			},
			wantErr: false,
		},
		{
			name: "missing email",
			req: dto.RegisterRequest{
				Password: "password123",
				Name:     "Test User",
			},
			wantErr: true,
			errMsg:  "email is required",
		},
		{
			name: "invalid email",
			req: dto.RegisterRequest{
				Email:    "not-an-email",
				Password: "password123",
				Name:     "Test User",
			},
			wantErr: true,
			errMsg:  "email must be a valid email address",
		},
		{
			name: "password too short",
			req: dto.RegisterRequest{
				Email:    "test@example.com",
				Password: "short",
				Name:     "Test User",
			},
			wantErr: true,
			errMsg:  "password must be at least 8 characters long",
		},
		{
			name: "name too short",
			req: dto.RegisterRequest{
				Email:    "test@example.com",
				Password: "password123",
				Name:     "A",
			},
			wantErr: true,
			errMsg:  "name must be at least 2 characters long",
		},
		{
			name:    "all fields missing",
			req:     dto.RegisterRequest{},
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
		req     dto.LoginRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request",
			req: dto.LoginRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			wantErr: false,
		},
		{
			name: "missing email",
			req: dto.LoginRequest{
				Password: "password123",
			},
			wantErr: true,
			errMsg:  "email is required",
		},
		{
			name: "invalid email format",
			req: dto.LoginRequest{
				Email:    "invalid-email",
				Password: "password123",
			},
			wantErr: true,
			errMsg:  "email must be a valid email address",
		},
		{
			name: "missing password",
			req: dto.LoginRequest{
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
		req     dto.CreatePostRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request",
			req: dto.CreatePostRequest{
				Title:   "Test Post",
				Content: "This is test content with enough characters",
			},
			wantErr: false,
		},
		{
			name: "title too short",
			req: dto.CreatePostRequest{
				Title:   "AB",
				Content: "This is test content",
			},
			wantErr: true,
			errMsg:  "title must be at least 3 characters long",
		},
		{
			name: "content too short",
			req: dto.CreatePostRequest{
				Title:   "Test Post",
				Content: "short",
			},
			wantErr: true,
			errMsg:  "content must be at least 10 characters long",
		},
		{
			name:    "missing required fields",
			req:     dto.CreatePostRequest{},
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
		req     dto.UpdateUserRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request with name",
			req: dto.UpdateUserRequest{
				Name: "New Name",
			},
			wantErr: false,
		},
		{
			name:    "empty request (all optional)",
			req:     dto.UpdateUserRequest{},
			wantErr: false,
		},
		{
			name: "name too short when provided",
			req: dto.UpdateUserRequest{
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
		req     dto.UpdatePostRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request with both fields",
			req: dto.UpdatePostRequest{
				Title:   "Updated Title",
				Content: "Updated content with enough length",
			},
			wantErr: false,
		},
		{
			name: "valid request with title only",
			req: dto.UpdatePostRequest{
				Title: "Updated Title",
			},
			wantErr: false,
		},
		{
			name:    "empty request (all optional)",
			req:     dto.UpdatePostRequest{},
			wantErr: false,
		},
		{
			name: "title too short when provided",
			req: dto.UpdatePostRequest{
				Title: "AB",
			},
			wantErr: true,
			errMsg:  "title must be at least 3 characters long",
		},
		{
			name: "content too short when provided",
			req: dto.UpdatePostRequest{
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
