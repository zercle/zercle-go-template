package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/zercle/zercle-go-template/internal/core/port/mocks"
	postDto "github.com/zercle/zercle-go-template/internal/features/post/dto"
	postHandler "github.com/zercle/zercle-go-template/internal/features/post/handler"
	"go.uber.org/mock/gomock"
)

// mockAuthMiddleware simulates authentication middleware that sets user_id in locals
func mockAuthMiddleware(userID string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if userID != "" {
			c.Locals("user_id", userID)
		}
		return c.Next()
	}
}

func TestPostHandler_Create_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Arrange
	mockService := mocks.NewMockPostService(ctrl)
	handler := postHandler.NewPostHandler(mockService)

	userID := uuid.New()
	app := fiber.New()
	app.Post("/posts", mockAuthMiddleware(userID.String()), handler.CreatePost)

	reqBody := postDto.CreatePostRequest{
		Title:   "Test Post",
		Content: "This is a test post content",
	}
	body, _ := json.Marshal(reqBody)

	expectedPost := &postDto.PostResponse{
		ID:        uuid.New().String(),
		Title:     reqBody.Title,
		Content:   reqBody.Content,
		AuthorID:  userID.String(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockService.EXPECT().CreatePost(gomock.Any(), gomock.Eq(userID), gomock.Eq(&reqBody)).Return(expectedPost, nil)

	// Act
	req := httptest.NewRequest("POST", "/posts", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	// Assert
	assert.NoError(t, err)
	if resp != nil {
		defer resp.Body.Close()
		assert.Equal(t, fiber.StatusCreated, resp.StatusCode)

		var response map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&response)
		assert.Equal(t, "success", response["status"])
	}
}

func TestPostHandler_Create_Unauthorized(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Arrange
	mockService := mocks.NewMockPostService(ctrl)
	handler := postHandler.NewPostHandler(mockService)

	app := fiber.New()
	// No auth middleware - user_id not set
	app.Post("/posts", handler.CreatePost)

	reqBody := postDto.CreatePostRequest{
		Title:   "Test Post",
		Content: "Test content",
	}
	body, _ := json.Marshal(reqBody)

	// Act - request without authentication
	req := httptest.NewRequest("POST", "/posts", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	// Assert
	assert.NoError(t, err)
	if resp != nil {
		defer resp.Body.Close()
		assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
	}
}

func TestPostHandler_Create_InvalidJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Arrange
	mockService := mocks.NewMockPostService(ctrl)
	handler := postHandler.NewPostHandler(mockService)

	userID := uuid.New()
	app := fiber.New()
	app.Post("/posts", mockAuthMiddleware(userID.String()), handler.CreatePost)

	// Act - invalid JSON
	req := httptest.NewRequest("POST", "/posts", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	// Assert
	assert.NoError(t, err)
	if resp != nil {
		defer resp.Body.Close()
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	}
}

func TestPostHandler_Create_ValidationError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Arrange
	mockService := mocks.NewMockPostService(ctrl)
	handler := postHandler.NewPostHandler(mockService)

	userID := uuid.New()
	app := fiber.New()
	app.Post("/posts", mockAuthMiddleware(userID.String()), handler.CreatePost)

	reqBody := postDto.CreatePostRequest{
		Title:   "", // empty title
		Content: "", // empty content
	}
	body, _ := json.Marshal(reqBody)

	// Act
	req := httptest.NewRequest("POST", "/posts", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	// Assert
	assert.NoError(t, err)
	if resp != nil {
		defer resp.Body.Close()
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	}
}

func TestPostHandler_Create_ServiceError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Arrange
	mockService := mocks.NewMockPostService(ctrl)
	handler := postHandler.NewPostHandler(mockService)

	userID := uuid.New()
	app := fiber.New()
	app.Post("/posts", mockAuthMiddleware(userID.String()), handler.CreatePost)

	reqBody := postDto.CreatePostRequest{
		Title:   "Test Post",
		Content: "Test content",
	}
	body, _ := json.Marshal(reqBody)

	mockService.EXPECT().CreatePost(gomock.Any(), gomock.Eq(userID), gomock.Eq(&reqBody)).Return(nil, assert.AnError)

	// Act
	req := httptest.NewRequest("POST", "/posts", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	// Assert
	assert.NoError(t, err)
	if resp != nil {
		defer resp.Body.Close()
		assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
	}
}

func TestPostHandler_GetAll_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Arrange
	mockService := mocks.NewMockPostService(ctrl)
	handler := postHandler.NewPostHandler(mockService)

	app := fiber.New()
	app.Get("/posts", handler.ListPosts)

	expectedPosts := []*postDto.PostResponse{
		{
			ID:        uuid.New().String(),
			Title:     "Post 1",
			Content:   "Content 1",
			AuthorID:  uuid.New().String(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        uuid.New().String(),
			Title:     "Post 2",
			Content:   "Content 2",
			AuthorID:  uuid.New().String(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	mockService.EXPECT().ListPosts(gomock.Any()).Return(expectedPosts, nil)

	// Act
	req := httptest.NewRequest("GET", "/posts", nil)
	resp, err := app.Test(req)

	// Assert
	assert.NoError(t, err)
	if resp != nil {
		defer resp.Body.Close()
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		var response map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&response)
		assert.Equal(t, "success", response["status"])
	}
}

func TestPostHandler_GetAll_Empty(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Arrange
	mockService := mocks.NewMockPostService(ctrl)
	handler := postHandler.NewPostHandler(mockService)

	app := fiber.New()
	app.Get("/posts", handler.ListPosts)

	expectedPosts := []*postDto.PostResponse{}

	mockService.EXPECT().ListPosts(gomock.Any()).Return(expectedPosts, nil)

	// Act
	req := httptest.NewRequest("GET", "/posts", nil)
	resp, err := app.Test(req)

	// Assert
	assert.NoError(t, err)
	if resp != nil {
		defer resp.Body.Close()
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	}
}

func TestPostHandler_GetAll_ServiceError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Arrange
	mockService := mocks.NewMockPostService(ctrl)
	handler := postHandler.NewPostHandler(mockService)

	app := fiber.New()
	app.Get("/posts", handler.ListPosts)

	mockService.EXPECT().ListPosts(gomock.Any()).Return(nil, assert.AnError)

	// Act
	req := httptest.NewRequest("GET", "/posts", nil)
	resp, err := app.Test(req)

	// Assert
	assert.NoError(t, err)
	if resp != nil {
		defer resp.Body.Close()
		assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
	}
}

func TestPostHandler_GetByID_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Arrange
	mockService := mocks.NewMockPostService(ctrl)
	handler := postHandler.NewPostHandler(mockService)

	app := fiber.New()
	app.Get("/posts/:id", handler.GetPost)

	postID := uuid.New()
	expectedPost := &postDto.PostResponse{
		ID:        postID.String(),
		Title:     "Test Post",
		Content:   "Test content",
		AuthorID:  uuid.New().String(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockService.EXPECT().GetPost(gomock.Any(), postID).Return(expectedPost, nil)

	// Act
	req := httptest.NewRequest("GET", "/posts/"+postID.String(), nil)
	resp, err := app.Test(req)

	// Assert
	assert.NoError(t, err)
	if resp != nil {
		defer resp.Body.Close()
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		var response map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&response)
		assert.Equal(t, "success", response["status"])
	}
}

func TestPostHandler_GetByID_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Arrange
	mockService := mocks.NewMockPostService(ctrl)
	handler := postHandler.NewPostHandler(mockService)

	app := fiber.New()
	app.Get("/posts/:id", handler.GetPost)

	postID := uuid.New()
	mockService.EXPECT().GetPost(gomock.Any(), postID).Return(nil, assert.AnError)

	// Act
	req := httptest.NewRequest("GET", "/posts/"+postID.String(), nil)
	resp, err := app.Test(req)

	// Assert
	assert.NoError(t, err)
	if resp != nil {
		defer resp.Body.Close()
		assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
	}
}

func TestPostHandler_GetByID_InvalidUUID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Arrange
	mockService := mocks.NewMockPostService(ctrl)
	handler := postHandler.NewPostHandler(mockService)

	app := fiber.New()
	app.Get("/posts/:id", handler.GetPost)

	// The handler doesn't validate UUID, it will parse to uuid.Nil
	mockService.EXPECT().GetPost(gomock.Any(), uuid.Nil).Return(nil, assert.AnError)

	// Act
	req := httptest.NewRequest("GET", "/posts/invalid-uuid", nil)
	resp, err := app.Test(req)

	// Assert
	assert.NoError(t, err)
	if resp != nil {
		defer resp.Body.Close()
		assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
	}
}

func TestPostHandler_Handler_EdgeCases(t *testing.T) {
	tests := []struct {
		name         string
		method       string
		path         string
		body         interface{}
		headers      map[string]string
		withAuth     bool
		setupMock    func(*mocks.MockPostService)
		expectedCode int
	}{
		{
			name:         "Create Empty Body",
			method:       "POST",
			path:         "/posts",
			body:         nil,
			headers:      map[string]string{"Content-Type": "application/json"},
			withAuth:     true,
			setupMock:    func(m *mocks.MockPostService) {},
			expectedCode: fiber.StatusBadRequest,
		},
		{
			name:     "GetAll No Posts",
			method:   "GET",
			path:     "/posts",
			body:     nil,
			headers:  map[string]string{},
			withAuth: false,
			setupMock: func(m *mocks.MockPostService) {
				m.EXPECT().ListPosts(gomock.Any()).Return([]*postDto.PostResponse{}, nil)
			},
			expectedCode: fiber.StatusOK,
		},
		{
			name:     "GetByID NonExistent",
			method:   "GET",
			path:     "/posts/" + uuid.New().String(),
			body:     nil,
			headers:  map[string]string{},
			withAuth: false,
			setupMock: func(m *mocks.MockPostService) {
				m.EXPECT().GetPost(gomock.Any(), gomock.Any()).Return(nil, assert.AnError)
			},
			expectedCode: fiber.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			// Arrange
			mockService := mocks.NewMockPostService(ctrl)
			handler := postHandler.NewPostHandler(mockService)

			app := fiber.New()
			userID := uuid.New()
			if tt.withAuth {
				app.Post("/posts", mockAuthMiddleware(userID.String()), handler.CreatePost)
			} else {
				app.Post("/posts", handler.CreatePost)
			}
			app.Get("/posts", handler.ListPosts)
			app.Get("/posts/:id", handler.GetPost)

			var body *bytes.Reader
			if tt.body != nil {
				data, _ := json.Marshal(tt.body)
				body = bytes.NewReader(data)
			} else {
				body = bytes.NewReader(nil)
			}

			req := httptest.NewRequest(tt.method, tt.path, body)
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			// Setup mock if needed
			tt.setupMock(mockService)

			// Act
			resp, err := app.Test(req)

			// Assert
			assert.NoError(t, err)
			if resp != nil {
				defer resp.Body.Close()
				assert.Equal(t, tt.expectedCode, resp.StatusCode)
			}
		})
	}
}

func TestPostHandler_Create_ValidTitleAndContent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Arrange
	mockService := mocks.NewMockPostService(ctrl)
	handler := postHandler.NewPostHandler(mockService)

	userID := uuid.New()
	app := fiber.New()
	app.Post("/posts", mockAuthMiddleware(userID.String()), handler.CreatePost)

	reqBody := postDto.CreatePostRequest{
		Title:   "A Very Long Title That Is Valid And Meets All Requirements",
		Content: "This is a very long content that is valid and meets all the requirements for creating a post in the system",
	}
	body, _ := json.Marshal(reqBody)

	expectedPost := &postDto.PostResponse{
		ID:        uuid.New().String(),
		Title:     reqBody.Title,
		Content:   reqBody.Content,
		AuthorID:  userID.String(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockService.EXPECT().CreatePost(gomock.Any(), gomock.Eq(userID), gomock.Eq(&reqBody)).Return(expectedPost, nil)

	// Act
	req := httptest.NewRequest("POST", "/posts", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	// Assert
	assert.NoError(t, err)
	if resp != nil {
		defer resp.Body.Close()
		assert.Equal(t, fiber.StatusCreated, resp.StatusCode)
	}
}

func TestPostHandler_GetAll_Pagination(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Arrange
	mockService := mocks.NewMockPostService(ctrl)
	handler := postHandler.NewPostHandler(mockService)

	app := fiber.New()
	app.Get("/posts", handler.ListPosts)

	mockService.EXPECT().ListPosts(gomock.Any()).Return([]*postDto.PostResponse{}, nil)

	// Act & Assert - test query parameters if handler supports them
	req := httptest.NewRequest("GET", "/posts?page=1&limit=10", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	if resp != nil {
		defer resp.Body.Close()
		// Should still work even if pagination isn't implemented
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	}
}
