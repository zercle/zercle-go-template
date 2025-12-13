package integration

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	postDomain "github.com/zercle/zercle-go-template/internal/features/post/domain"
	postHandler "github.com/zercle/zercle-go-template/internal/features/post/handler"
	"github.com/zercle/zercle-go-template/internal/features/post/service"
	"go.uber.org/mock/gomock"
)

func TestPostEndpoints_Integration(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock repository
	mockRepo := SetupMockPostRepo(ctrl)

	// Create service with mock repository
	svc := service.NewPostService(mockRepo)

	// Create handler with service
	handler := postHandler.NewPostHandler(svc)

	// Create Fiber app
	app := fiber.New()
	posts := app.Group("/api/v1/posts")
	posts.Post("/", handler.CreatePost)
	posts.Get("/", handler.ListPosts)
	posts.Get("/:id", handler.GetPost)

	t.Run("POST /api/v1/posts - Success", func(t *testing.T) {
		// Setup mock expectations
		mockRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)

		// Create request with auth
		req := NewRequest("POST", "/api/v1/posts", ValidCreatePostRequest, true, MockUserID1)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		jsend := ValidateSuccessResponse(t, resp, fiber.StatusCreated)
		assert.NotNil(t, jsend.Data)
	})

	t.Run("POST /api/v1/posts - Empty Title", func(t *testing.T) {
		// Create request with auth
		req := NewRequest("POST", "/api/v1/posts", InvalidCreatePostRequest_EmptyTitle, true, MockUserID1)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		jsend := ValidateFailResponse(t, resp, fiber.StatusBadRequest)
		assert.NotNil(t, jsend.Data)
	})

	t.Run("POST /api/v1/posts - Short Title", func(t *testing.T) {
		// Create request with auth
		req := NewRequest("POST", "/api/v1/posts", InvalidCreatePostRequest_ShortTitle, true, MockUserID1)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		jsend := ValidateFailResponse(t, resp, fiber.StatusBadRequest)
		assert.NotNil(t, jsend.Data)
	})

	t.Run("POST /api/v1/posts - Empty Content", func(t *testing.T) {
		// Create request with auth
		req := NewRequest("POST", "/api/v1/posts", InvalidCreatePostRequest_EmptyContent, true, MockUserID1)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		jsend := ValidateFailResponse(t, resp, fiber.StatusBadRequest)
		assert.NotNil(t, jsend.Data)
	})

	t.Run("POST /api/v1/posts - Short Content", func(t *testing.T) {
		// Create request with auth
		req := NewRequest("POST", "/api/v1/posts", InvalidCreatePostRequest_ShortContent, true, MockUserID1)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		jsend := ValidateFailResponse(t, resp, fiber.StatusBadRequest)
		assert.NotNil(t, jsend.Data)
	})

	t.Run("GET /api/v1/posts - Success (Empty List)", func(t *testing.T) {
		// Setup mock expectations
		mockRepo.EXPECT().GetAll(gomock.Any()).Return([]*postDomain.Post{}, nil)

		req := httptest.NewRequest("GET", "/api/v1/posts", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		jsend := ValidateSuccessResponse(t, resp, fiber.StatusOK)
		assert.NotNil(t, jsend.Data)
		// Verify it's an empty array
		dataArray, ok := jsend.Data.([]interface{})
		assert.True(t, ok, "Expected data to be an array")
		assert.Equal(t, 0, len(dataArray))
	})

	t.Run("GET /api/v1/posts - Success (With Posts)", func(t *testing.T) {
		// Setup mock expectations
		mockRepo.EXPECT().GetAll(gomock.Any()).Return(MockPosts, nil)

		req := httptest.NewRequest("GET", "/api/v1/posts", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		jsend := ValidateSuccessResponse(t, resp, fiber.StatusOK)
		assert.NotNil(t, jsend.Data)
		// Verify it's an array with posts
		dataArray, ok := jsend.Data.([]interface{})
		assert.True(t, ok, "Expected data to be an array")
		assert.Equal(t, 3, len(dataArray))
	})

	t.Run("GET /api/v1/posts/:id - Success", func(t *testing.T) {
		// Setup mock expectations
		mockRepo.EXPECT().GetByID(gomock.Any(), MockPostID1).Return(MockPost1, nil)

		req := httptest.NewRequest("GET", "/api/v1/posts/"+MockPostID1.String(), nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		jsend := ValidateSuccessResponse(t, resp, fiber.StatusOK)
		assert.NotNil(t, jsend.Data)
	})

	t.Run("GET /api/v1/posts/:id - Not Found", func(t *testing.T) {
		// Setup mock expectations
		mockRepo.EXPECT().GetByID(gomock.Any(), MockPostID3).Return(nil, ErrNotFound)

		req := httptest.NewRequest("GET", "/api/v1/posts/"+MockPostID3.String(), nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		jsend := ValidateErrorResponse(t, resp, fiber.StatusNotFound)
		assert.NotNil(t, jsend.Message)
	})

	t.Run("GET /api/v1/posts/:id - Invalid UUID", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/posts/invalid-uuid", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		// Invalid UUID should result in an error response
		assert.True(t, resp.StatusCode >= fiber.StatusBadRequest)
	})
}
