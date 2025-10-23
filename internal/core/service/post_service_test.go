package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/zercle/zercle-go-template/internal/core/domain"
	"github.com/zercle/zercle-go-template/internal/core/service"
	"github.com/zercle/zercle-go-template/pkg/dto"
	"github.com/zercle/zercle-go-template/test/mocks"
	"go.uber.org/mock/gomock"
)

func TestPostService_CreatePost(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockPostRepository(ctrl)
	svc := service.NewPostService(mockRepo)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		userID := uuid.New()
		req := &dto.CreatePostRequest{
			Title:   "New Post",
			Content: "This is a new post content.",
		}

		mockRepo.EXPECT().Create(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, p *domain.Post) error {
			assert.Equal(t, req.Title, p.Title)
			assert.Equal(t, userID, p.AuthorID)
			return nil
		})

		res, err := svc.CreatePost(ctx, userID, req)
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, req.Title, res.Title)
	})
}

func TestPostService_GetPost(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockPostRepository(ctrl)
	svc := service.NewPostService(mockRepo)
	ctx := context.Background()

	postID := uuid.New()
	mockPost := &domain.Post{
		ID:        postID,
		Title:     "Test Post",
		Content:   "Content",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockRepo.EXPECT().GetByID(ctx, postID).Return(mockPost, nil)

	res, err := svc.GetPost(ctx, postID)
	assert.NoError(t, err)
	assert.Equal(t, mockPost.ID.String(), res.ID)
}

func TestPostService_ListPosts(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockPostRepository(ctrl)
	svc := service.NewPostService(mockRepo)
	ctx := context.Background()

	mockPosts := []*domain.Post{
		{Title: "Post 1"},
		{Title: "Post 2"},
	}

	mockRepo.EXPECT().GetAll(ctx).Return(mockPosts, nil)

	res, err := svc.ListPosts(ctx)
	assert.NoError(t, err)
	assert.Len(t, res, 2)
}
