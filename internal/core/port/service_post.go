package port

import (
	"context"

	"github.com/google/uuid"
	postDto "github.com/zercle/zercle-go-template/internal/features/post/dto"
)

//go:generate mockgen -destination=./mocks/$GOFILE -package=mocks -source=$GOFILE

// PostService defines the input port for Post operations.
type PostService interface {
	CreatePost(ctx context.Context, userID uuid.UUID, req *postDto.CreatePostRequest) (*postDto.PostResponse, error)
	GetPost(ctx context.Context, postID uuid.UUID) (*postDto.PostResponse, error)
	ListPosts(ctx context.Context) ([]*postDto.PostResponse, error)
}
