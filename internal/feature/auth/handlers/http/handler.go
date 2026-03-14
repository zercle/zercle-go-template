package http

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"

	"github.com/zercle/zercle-go-template/api/dtos/auth"
	"github.com/zercle/zercle-go-template/internal/feature/auth/domain"
	"github.com/zercle/zercle-go-template/internal/feature/auth/ports"
)

// Handler handles HTTP requests for authentication.
type Handler struct {
	authService ports.AuthService
}

// NewHandler creates a new HTTP authentication handler.
func NewHandler(authService ports.AuthService) *Handler {
	return &Handler{authService: authService}
}

// Register handles user registration.
func (h *Handler) Register(c *echo.Context) error {
	var req auth.RegisterRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	input := ports.RegisterInput{
		Username:    req.Username,
		Email:       req.Email,
		Password:    req.Password,
		DisplayName: req.DisplayName,
	}

	result, err := h.authService.Register(c.Request().Context(), input)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusCreated, auth.Response{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		User:         toUserResponse(result.User),
		ExpiresAt:    result.ExpiresAt,
	})
}

// Login handles user login.
func (h *Handler) Login(c *echo.Context) error {
	var req auth.LoginRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	input := ports.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	}

	result, err := h.authService.Login(c.Request().Context(), input)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid credentials")
	}

	return c.JSON(http.StatusOK, auth.Response{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		User:         toUserResponse(result.User),
		ExpiresAt:    result.ExpiresAt,
	})
}

// GetCurrentUser returns the current authenticated user.
func (h *Handler) GetCurrentUser(c *echo.Context) error {
	userIDStr := c.Get("user_id")
	if userIDStr == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	userID, ok := userIDStr.(uuid.UUID)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
	}

	user, err := h.authService.ValidateToken(c.Request().Context(), c.Request().Header.Get("Authorization"))
	if err != nil || user.ID != userID {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
	}

	return c.JSON(http.StatusOK, toUserResponse(user))
}

// Logout handles user logout.
func (h *Handler) Logout(c *echo.Context) error {
	userIDStr := c.Get("user_id")
	if userIDStr == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	userID, ok := userIDStr.(uuid.UUID)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
	}

	if err := h.authService.Logout(c.Request().Context(), userID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusNoContent)
}

// RefreshToken handles token refresh.
func (h *Handler) RefreshToken(c *echo.Context) error {
	var req auth.RefreshRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	result, err := h.authService.RefreshToken(c.Request().Context(), req.RefreshToken)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid refresh token")
	}

	return c.JSON(http.StatusOK, auth.RefreshResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		ExpiresAt:    result.ExpiresAt,
	})
}

func toUserResponse(user *domain.User) *auth.UserResponse {
	if user == nil {
		return nil
	}
	return &auth.UserResponse{
		ID:          user.ID.String(),
		Username:    user.Username,
		Email:       user.Email,
		DisplayName: user.DisplayName,
		AvatarURL:   user.AvatarURL,
		Status:      user.Status,
		CreatedAt:   user.CreatedAt,
	}
}
