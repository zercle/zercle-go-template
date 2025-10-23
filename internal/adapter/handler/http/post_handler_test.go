package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	h "github.com/zercle/zercle-go-template/internal/adapter/handler/http"
	"github.com/zercle/zercle-go-template/pkg/dto"
	"github.com/zercle/zercle-go-template/test/mocks"
	"go.uber.org/mock/gomock"
)

func TestPostHandler_CreatePost(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mocks.NewMockPostService(ctrl)
	handler := h.NewPostHandler(mockSvc)

	app := fiber.New()
	// Mock middleware to set user_id
	app.Post("/posts", func(c *fiber.Ctx) error {
		c.Locals("user_id", "550e8400-e29b-41d4-a716-446655440000")
		return c.Next()
	}, handler.CreatePost)

	reqBody := dto.CreatePostRequest{
		Title:   "Valid Title",
		Content: "Valid Content Validation",
	}
	body, _ := json.Marshal(reqBody)

	resDTO := &dto.PostResponse{
		ID:        "post-uuid",
		Title:     reqBody.Title,
		Content:   reqBody.Content,
		AuthorID:  "550e8400-e29b-41d4-a716-446655440000",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockSvc.EXPECT().CreatePost(gomock.Any(), gomock.Any(), gomock.Any()).Return(resDTO, nil)

	req := httptest.NewRequest("POST", "/posts", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	if resp != nil {
		defer resp.Body.Close()
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	}
}

func TestPostHandler_ListPosts(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mocks.NewMockPostService(ctrl)
	handler := h.NewPostHandler(mockSvc)

	app := fiber.New()
	app.Get("/posts", handler.ListPosts)

	mockSvc.EXPECT().ListPosts(gomock.Any()).Return([]*dto.PostResponse{}, nil)

	req := httptest.NewRequest("GET", "/posts", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	if resp != nil {
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	}
}

func TestPostHandler_GetPost(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mocks.NewMockPostService(ctrl)
	handler := h.NewPostHandler(mockSvc)

	app := fiber.New()
	app.Get("/posts/:id", handler.GetPost)

	id := uuid.New()
	resDTO := &dto.PostResponse{
		ID:    id.String(),
		Title: "Found",
	}

	mockSvc.EXPECT().GetPost(gomock.Any(), id).Return(resDTO, nil)

	req := httptest.NewRequest("GET", "/posts/"+id.String(), nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	if resp != nil {
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	}
}
