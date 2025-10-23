package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/zercle/zercle-go-template/internal/core/domain"
	domerrors "github.com/zercle/zercle-go-template/internal/core/domain/errors"
	"github.com/zercle/zercle-go-template/internal/core/port"
	"github.com/zercle/zercle-go-template/pkg/dto"
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

// CreatePost creates a new post for the given user.
func (s *postService) CreatePost(ctx context.Context, userID uuid.UUID, req *dto.CreatePostRequest) (*dto.PostResponse, error) {
	postID, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}

	post := &domain.Post{
		ID:        postID,
		Title:     req.Title,
		Content:   req.Content,
		AuthorID:  userID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.repo.Create(ctx, post); err != nil {
		return nil, err
	}

	return s.mapToDTO(post), nil
}

// GetPost retrieves a post by its ID.
func (s *postService) GetPost(ctx context.Context, postID uuid.UUID) (*dto.PostResponse, error) {
	post, err := s.repo.GetByID(ctx, postID)
	if err != nil {
		return nil, domerrors.ErrNotFound
	}
	return s.mapToDTO(post), nil
}

// ListPosts returns all posts.
func (s *postService) ListPosts(ctx context.Context) ([]*dto.PostResponse, error) {
	posts, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	res := make([]*dto.PostResponse, len(posts))
	for i, p := range posts {
		res[i] = s.mapToDTO(p)
	}
	return res, nil
}

func (s *postService) mapToDTO(post *domain.Post) *dto.PostResponse {
	return &dto.PostResponse{
		ID:        post.ID.String(),
		Title:     post.Title,
		Content:   post.Content,
		AuthorID:  post.AuthorID.String(),
		CreatedAt: post.CreatedAt,
		UpdatedAt: post.UpdatedAt,
	}
}
