// Package handler provides Fiber HTTP handlers.
package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/zercle/zercle-go-template/internal/core/port"
	"github.com/zercle/zercle-go-template/internal/middleware"
	"github.com/zercle/zercle-go-template/pkg/dto"
	"github.com/zercle/zercle-go-template/pkg/utils/response"
)

// UserHandler handles HTTP requests for user operations.
type UserHandler struct {
	svc port.UserService
}

// NewUserHandler creates a new UserHandler instance.
func NewUserHandler(svc port.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

// RegisterRoutes registers user-related routes to the Fiber app.
func (h *UserHandler) RegisterRoutes(router fiber.Router) {
	// Auth routes
	auth := router.Group("/auth")
	auth.Post("/register", h.Register)
	auth.Post("/login", h.Login)

	// User routes
	users := router.Group("/users")
	users.Get("/me", h.GetProfile)
}

// Register godoc
// @Summary Register a new user
// @Tags users
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "Register Request"
// @Success 201 {object} response.Response{data=dto.UserResponse}
// @Failure 400 {object} response.Response
// @Router /auth/register [post]
func (h *UserHandler) Register(c *fiber.Ctx) error {
	var req dto.RegisterRequest
	if err := middleware.ParseAndValidate(c, &req); err != nil {
		return err
	}

	res, err := h.svc.Register(c.Context(), &req)
	if err != nil {
		return response.HandleError(c, err)
	}

	return response.Success(c, fiber.StatusCreated, res)
}

// Login godoc
// @Summary Login user
// @Tags users
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "Login Request"
// @Success 200 {object} response.Response{data=map[string]string}
// @Failure 401 {object} response.Response
// @Router /auth/login [post]
func (h *UserHandler) Login(c *fiber.Ctx) error {
	var req dto.LoginRequest
	if err := middleware.ParseAndValidate(c, &req); err != nil {
		return err
	}

	token, err := h.svc.Login(c.Context(), &req)
	if err != nil {
		return response.HandleError(c, err)
	}

	return response.Success(c, fiber.StatusOK, fiber.Map{"token": token})
}

// GetProfile godoc
// @Summary Get user profile
// @Tags users
// @Produce json
// @Success 200 {object} response.Response{data=dto.UserResponse}
// @Router /users/me [get]
func (h *UserHandler) GetProfile(c *fiber.Ctx) error {
	// Assuming Middleware sets "user_id" in Locals
	userIDStr, ok := c.Locals("user_id").(string)
	if !ok {
		return response.Fail(c, fiber.StatusUnauthorized, fiber.Map{"error": "unauthorized"})
	}
	uid, _ := uuid.Parse(userIDStr)

	res, err := h.svc.GetProfile(c.Context(), uid)
	if err != nil {
		return response.HandleError(c, err)
	}

	return response.Success(c, fiber.StatusOK, res)
}
