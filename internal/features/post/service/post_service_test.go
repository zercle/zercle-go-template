package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/zercle/zercle-go-template/internal/core/port/mocks"
	postDomain "github.com/zercle/zercle-go-template/internal/features/post/domain"
	postDto "github.com/zercle/zercle-go-template/internal/features/post/dto"
	postService "github.com/zercle/zercle-go-template/internal/features/post/service"
	sharedDomain "github.com/zercle/zercle-go-template/internal/shared/domain"
	"go.uber.org/mock/gomock"
)

func TestPostService(t *testing.T) {
	t.Run("Create_Success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		// Arrange
		mockRepo := mocks.NewMockPostRepository(ctrl)
		svc := postService.NewPostService(mockRepo)
		ctx := context.Background()

		req := &postDto.CreatePostRequest{
			Title:   "Test Post",
			Content: "This is test content",
		}

		mockRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)

		// Act
		userID := uuid.New()
		result, err := svc.CreatePost(ctx, userID, req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, req.Title, result.Title)
		assert.Equal(t, req.Content, result.Content)
	})

	t.Run("Create_Error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		// Arrange
		mockRepo := mocks.NewMockPostRepository(ctrl)
		svc := postService.NewPostService(mockRepo)
		ctx := context.Background()

		req := &postDto.CreatePostRequest{
			Title:   "Test Post",
			Content: "Test content",
		}

		mockRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(assert.AnError)

		// Act
		userID := uuid.New()
		result, err := svc.CreatePost(ctx, userID, req)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("GetAll_Success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		// Arrange
		mockRepo := mocks.NewMockPostRepository(ctrl)
		svc := postService.NewPostService(mockRepo)
		ctx := context.Background()

		posts := []*postDomain.Post{
			{
				ID:        uuid.New(),
				Title:     "Post 1",
				Content:   "Content 1",
				AuthorID:  sharedDomain.UserID(uuid.New()),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			{
				ID:        uuid.New(),
				Title:     "Post 2",
				Content:   "Content 2",
				AuthorID:  sharedDomain.UserID(uuid.New()),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		}

		mockRepo.EXPECT().GetAll(gomock.Any()).Return(posts, nil)

		// Act
		result, err := svc.ListPosts(ctx)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result, 2)
	})

	t.Run("GetAll_Empty", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		// Arrange
		mockRepo := mocks.NewMockPostRepository(ctrl)
		svc := postService.NewPostService(mockRepo)
		ctx := context.Background()

		mockRepo.EXPECT().GetAll(gomock.Any()).Return([]*postDomain.Post{}, nil)

		// Act
		result, err := svc.ListPosts(ctx)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Empty(t, result)
	})

	t.Run("GetByID_Success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		// Arrange
		mockRepo := mocks.NewMockPostRepository(ctrl)
		svc := postService.NewPostService(mockRepo)
		ctx := context.Background()

		postID := uuid.New()
		post := &postDomain.Post{
			ID:        postID,
			Title:     "Test Post",
			Content:   "Test content",
			AuthorID:  sharedDomain.UserID(uuid.New()),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		mockRepo.EXPECT().GetByID(gomock.Any(), postID).Return(post, nil)

		// Act
		result, err := svc.GetPost(ctx, postID)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, postID.String(), result.ID)
	})

	t.Run("GetByID_NotFound", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		// Arrange
		mockRepo := mocks.NewMockPostRepository(ctrl)
		svc := postService.NewPostService(mockRepo)
		ctx := context.Background()

		postID := uuid.New()
		mockRepo.EXPECT().GetByID(gomock.Any(), postID).Return(nil, assert.AnError)

		// Act
		result, err := svc.GetPost(ctx, postID)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("Create_ContextCancellation", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		// Arrange
		mockRepo := mocks.NewMockPostRepository(ctrl)
		svc := postService.NewPostService(mockRepo)
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		req := &postDto.CreatePostRequest{
			Title:   "Test Post",
			Content: "Test content",
		}

		mockRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(context.Canceled)

		// Act
		userID := uuid.New()
		result, err := svc.CreatePost(ctx, userID, req)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("GetAll_ContextTimeout", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		// Arrange
		mockRepo := mocks.NewMockPostRepository(ctrl)
		svc := postService.NewPostService(mockRepo)
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()

		mockRepo.EXPECT().GetAll(gomock.Any()).Return(nil, context.DeadlineExceeded)

		// Act
		result, err := svc.ListPosts(ctx)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("Create_DifferentInputs", func(t *testing.T) {
		// Service doesn't validate at service layer - validation is at handler level
		tests := []struct {
			name    string
			title   string
			content string
		}{
			{"Valid", "Title", "Content"},
			{"ValidLongTitle", "A Very Long Title", "Content"},
			{"ValidLongContent", "Title", "A very long content"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				ctrl := gomock.NewController(t)
				defer ctrl.Finish()
				// Arrange
				mockRepo := mocks.NewMockPostRepository(ctrl)
				svc := postService.NewPostService(mockRepo)
				ctx := context.Background()

				req := &postDto.CreatePostRequest{
					Title:   tt.title,
					Content: tt.content,
				}

				mockRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)

				// Act
				userID := uuid.New()
				result, err := svc.CreatePost(ctx, userID, req)

				// Assert
				assert.NoError(t, err)
				assert.NotNil(t, result)
			})
		}
	})

	t.Run("GetByID_DifferentUUIDs", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		// Arrange
		mockRepo := mocks.NewMockPostRepository(ctrl)
		svc := postService.NewPostService(mockRepo)
		ctx := context.Background()

		// Test multiple different UUIDs
		for i := 0; i < 5; i++ {
			postID := uuid.New()
			post := &postDomain.Post{
				ID:        postID,
				Title:     "Test Post",
				Content:   "Test content",
				AuthorID:  sharedDomain.UserID(uuid.New()),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			mockRepo.EXPECT().GetByID(gomock.Any(), postID).Return(post, nil)

			// Act
			result, err := svc.GetPost(ctx, postID)

			// Assert
			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, postID.String(), result.ID)
		}
	})

	t.Run("Create_Concurrent", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		// Arrange
		mockRepo := mocks.NewMockPostRepository(ctrl)
		svc := postService.NewPostService(mockRepo)
		ctx := context.Background()

		mockRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil).Times(3)

		// Act - concurrent creates
		requests := []*postDto.CreatePostRequest{
			{Title: "Post 1", Content: "Content 1"},
			{Title: "Post 2", Content: "Content 2"},
			{Title: "Post 3", Content: "Content 3"},
		}

		results := make(chan *postDto.PostResponse, 3)
		errs := make(chan error, 3)

		for _, req := range requests {
			go func(r *postDto.CreatePostRequest) {
				userID := uuid.New()
				result, err := svc.CreatePost(ctx, userID, r)
				results <- result
				errs <- err
			}(req)
		}

		// Wait for all operations
		for i := 0; i < 3; i++ {
			err := <-errs
			result := <-results
			assert.NoError(t, err)
			assert.NotNil(t, result)
		}
	})

	t.Run("GetAll_Pagination", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		// Arrange
		mockRepo := mocks.NewMockPostRepository(ctrl)
		svc := postService.NewPostService(mockRepo)
		ctx := context.Background()

		// Create many posts
		posts := make([]*postDomain.Post, 100)
		for i := 0; i < 100; i++ {
			posts[i] = &postDomain.Post{
				ID:        uuid.New(),
				Title:     "Post " + string(rune(i)),
				Content:   "Content " + string(rune(i)),
				AuthorID:  sharedDomain.UserID(uuid.New()),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
		}

		mockRepo.EXPECT().GetAll(gomock.Any()).Return(posts, nil)

		// Act
		result, err := svc.ListPosts(ctx)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result, 100)
	})

	t.Run("Create_UUIDGeneration", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		// Arrange
		mockRepo := mocks.NewMockPostRepository(ctrl)
		svc := postService.NewPostService(mockRepo)
		ctx := context.Background()

		req := &postDto.CreatePostRequest{
			Title:   "Test Post",
			Content: "Test content",
		}
		mockRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)

		// Act
		userID := uuid.New()
		result, err := svc.CreatePost(ctx, userID, req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotEmpty(t, result.ID)
		assert.True(t, isValidUUID(result.ID))
	})

	t.Run("GetAll_DatabaseError", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		// Arrange
		mockRepo := mocks.NewMockPostRepository(ctrl)
		svc := postService.NewPostService(mockRepo)
		ctx := context.Background()

		mockRepo.EXPECT().GetAll(gomock.Any()).Return(nil, assert.AnError)

		// Act
		result, err := svc.ListPosts(ctx)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

// Helper function to validate UUID format
func isValidUUID(uuidStr string) bool {
	_, err := uuid.Parse(uuidStr)
	return err == nil
}
