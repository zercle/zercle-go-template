package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/zercle/zercle-go-template/internal/core/port"
	"github.com/zercle/zercle-go-template/internal/middleware"
	"github.com/zercle/zercle-go-template/pkg/dto"
	"github.com/zercle/zercle-go-template/pkg/utils/response"
)

// PostHandler handles HTTP requests for post operations.
type PostHandler struct {
	svc port.PostService
}

// NewPostHandler creates a new PostHandler instance.
func NewPostHandler(svc port.PostService) *PostHandler {
	return &PostHandler{svc: svc}
}

// RegisterRoutes registers post-related routes to the Fiber app.
func (h *PostHandler) RegisterRoutes(router fiber.Router) {
	posts := router.Group("/posts")
	posts.Post("/", h.CreatePost)
	posts.Get("/", h.ListPosts)
	posts.Get("/:id", h.GetPost)
}

// CreatePost godoc
// @Summary Create a new post
// @Tags posts
// @Accept json
// @Produce json
// @Param request body dto.CreatePostRequest true "Create Post Request"
// @Success 201 {object} response.Response{data=dto.PostResponse}
// @Failure 401 {object} response.Response
// @Router /posts [post]
func (h *PostHandler) CreatePost(c *fiber.Ctx) error {
	userIDStr, ok := c.Locals("user_id").(string)
	if !ok {
		return response.Fail(c, fiber.StatusUnauthorized, fiber.Map{"error": "unauthorized"})
	}
	uid, _ := uuid.Parse(userIDStr)

	var req dto.CreatePostRequest
	if err := middleware.ParseAndValidate(c, &req); err != nil {
		return err
	}

	res, err := h.svc.CreatePost(c.Context(), uid, &req)
	if err != nil {
		return response.HandleError(c, err)
	}

	return response.Success(c, fiber.StatusCreated, res)
}

// ListPosts godoc
// @Summary List all posts
// @Tags posts
// @Produce json
// @Success 200 {object} response.Response{data=[]dto.PostResponse}
// @Router /posts [get]
func (h *PostHandler) ListPosts(c *fiber.Ctx) error {
	res, err := h.svc.ListPosts(c.Context())
	if err != nil {
		return response.HandleError(c, err)
	}
	return response.Success(c, fiber.StatusOK, res)
}

// GetPost godoc
// @Summary Get a post by ID
// @Tags posts
// @Produce json
// @Param id path string true "Post ID"
// @Success 200 {object} response.Response{data=dto.PostResponse}
// @Failure 404 {object} response.Response
// @Router /posts/{id} [get]
func (h *PostHandler) GetPost(c *fiber.Ctx) error {
	id := c.Params("id")
	uid, _ := uuid.Parse(id)

	res, err := h.svc.GetPost(c.Context(), uid)
	if err != nil {
		return response.HandleError(c, err)
	}
	return response.Success(c, fiber.StatusOK, res)
}
