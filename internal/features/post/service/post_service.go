package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/zercle/zercle-go-template/internal/core/port"
	postDomain "github.com/zercle/zercle-go-template/internal/features/post/domain"
	postDto "github.com/zercle/zercle-go-template/internal/features/post/dto"
	sharedDomain "github.com/zercle/zercle-go-template/internal/shared/domain"
	sharederrors "github.com/zercle/zercle-go-template/internal/shared/errors"
)

type postService struct {
	repo port.PostRepository
}

// NewPostService creates a new instance of PostService.
func NewPostService(repo port.PostRepository) port.PostService {
	return &postService{
		repo: repo,
	}
}

// CreatePost creates a new post.
func (s *postService) CreatePost(ctx context.Context, userID uuid.UUID, req *postDto.CreatePostRequest) (*postDto.PostResponse, error) {
	postID, err := uuid.NewV7()
	if err != nil {
		return nil, sharederrors.ErrInternalServer
	}

	post := &postDomain.Post{
		ID:        postID,
		Title:     req.Title,
		Content:   req.Content,
		AuthorID:  sharedDomain.UserID(userID),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.repo.Create(ctx, post); err != nil {
		return nil, err
	}

	return s.mapToDTO(post), nil
}

// GetPost retrieves a post by ID.
func (s *postService) GetPost(ctx context.Context, postID uuid.UUID) (*postDto.PostResponse, error) {
	post, err := s.repo.GetByID(ctx, postID)
	if err != nil {
		return nil, sharederrors.ErrNotFound
	}

	return s.mapToDTO(post), nil
}

// ListPosts retrieves all posts.
func (s *postService) ListPosts(ctx context.Context) ([]*postDto.PostResponse, error) {
	posts, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	dtos := make([]*postDto.PostResponse, len(posts))
	for i, post := range posts {
		dtos[i] = s.mapToDTO(post)
	}

	return dtos, nil
}

func (s *postService) mapToDTO(post *postDomain.Post) *postDto.PostResponse {
	return &postDto.PostResponse{
		ID:        post.ID.String(),
		Title:     post.Title,
		Content:   post.Content,
		AuthorID:  post.AuthorID.String(),
		CreatedAt: post.CreatedAt,
		UpdatedAt: post.UpdatedAt,
	}
}
