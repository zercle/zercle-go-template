package port

import (
	"context"

	"github.com/google/uuid"
	"github.com/zercle/zercle-go-template/pkg/dto"
)

// UserService defines the input port for User operations.
type UserService interface {
	Register(ctx context.Context, req *dto.RegisterRequest) (*dto.UserResponse, error)
	Login(ctx context.Context, req *dto.LoginRequest) (string, error) // Returns JWT token
	GetProfile(ctx context.Context, userID uuid.UUID) (*dto.UserResponse, error)
}

// PostService defines the input port for Post operations.
type PostService interface {
	CreatePost(ctx context.Context, userID uuid.UUID, req *dto.CreatePostRequest) (*dto.PostResponse, error)
	GetPost(ctx context.Context, postID uuid.UUID) (*dto.PostResponse, error)
	ListPosts(ctx context.Context) ([]*dto.PostResponse, error)
}
