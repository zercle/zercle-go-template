package http

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v5"

	"github.com/zercle/zercle-go-template/internal/features/auth/dto"
	"github.com/zercle/zercle-go-template/internal/features/auth/service"
	"github.com/zercle/zercle-go-template/internal/shared/middleware"
)

// AuthHandler handles HTTP requests for authentication operations.
type AuthHandler struct {
	authService service.AuthServiceInterface
}

// NewAuthHandler creates a new AuthHandler with the given auth service.
func NewAuthHandler(authService service.AuthServiceInterface) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Register godoc
// @Summary Register a new user
// @Description Create a new user account with email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "Registration details"
// @Success 201 {object} dto.AuthResponse "Successfully registered"
// @Failure 400 {object} map[string]string "Invalid request body"
// @Failure 409 {object} map[string]string "User already exists"
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *echo.Context) error {
	var req dto.RegisterRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	input := service.RegisterInput{
		Username:    req.Username,
		Email:       req.Email,
		Password:    req.Password,
		DisplayName: req.DisplayName,
	}

	result, err := h.authService.Register(c.Request().Context(), input)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("failed to register user: %w", err).Error())
	}

	if err := c.JSON(http.StatusCreated, dto.AuthResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		User: &dto.UserDTO{
			ID:          result.User.ID,
			Username:    result.User.Username,
			Email:       result.User.Email,
			DisplayName: result.User.DisplayName,
			AvatarURL:   result.User.AvatarURL,
			Status:      result.User.Status,
		},
		ExpiresAt: result.ExpiresAt,
	}); err != nil {
		return fmt.Errorf("failed to send register response: %w", err)
	}
	return nil
}

// Login godoc
// @Summary User login
// @Description Authenticate user with email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "Login credentials"
// @Success 200 {object} dto.AuthResponse "Successfully logged in"
// @Failure 400 {object} map[string]string "Invalid request body"
// @Failure 401 {object} map[string]string "Invalid credentials"
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *echo.Context) error {
	var req dto.LoginRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	input := service.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	}

	result, err := h.authService.Login(c.Request().Context(), input)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, fmt.Errorf("failed to login user: %w", err).Error())
	}

	if err := c.JSON(http.StatusOK, dto.AuthResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		User: &dto.UserDTO{
			ID:          result.User.ID,
			Username:    result.User.Username,
			Email:       result.User.Email,
			DisplayName: result.User.DisplayName,
			AvatarURL:   result.User.AvatarURL,
			Status:      result.User.Status,
		},
		ExpiresAt: result.ExpiresAt,
	}); err != nil {
		return fmt.Errorf("failed to send login response: %w", err)
	}
	return nil
}

// GetCurrentUser godoc
// @Summary Get current user profile
// @Description Get the authenticated user's profile information
// @Tags auth
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} dto.UserDTO "User profile"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Router /auth/me [get]
func (h *AuthHandler) GetCurrentUser(c *echo.Context) error {
	_, err := middleware.GetUserID(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}

	authHeader := c.Request().Header.Get("Authorization")
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	user, err := h.authService.ValidateToken(c.Request().Context(), tokenString)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, fmt.Errorf("failed to validate token: %w", err).Error())
	}

	if err := c.JSON(http.StatusOK, dto.UserDTO{
		ID:          user.ID,
		Username:    user.Username,
		Email:       user.Email,
		DisplayName: user.DisplayName,
		AvatarURL:   user.AvatarURL,
		Status:      user.Status,
	}); err != nil {
		return fmt.Errorf("failed to send user response: %w", err)
	}
	return nil
}

// Logout godoc
// @Summary User logout
// @Description Invalidate the current user session
// @Tags auth
// @Accept json
// @Produce json
// @Security Bearer
// @Success 204 "Successfully logged out"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *echo.Context) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}

	if err := h.authService.Logout(c.Request().Context(), userID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("failed to logout user: %w", err).Error())
	}

	if err := c.NoContent(http.StatusNoContent); err != nil {
		return fmt.Errorf("failed to send logout response: %w", err)
	}
	return nil
}

// RefreshToken godoc
// @Summary Refresh access token
// @Description Get a new access token using a refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.RefreshRequest true "Refresh token"
// @Success 200 {object} dto.AuthResponse "New tokens issued"
// @Failure 400 {object} map[string]string "Invalid request body"
// @Failure 401 {object} map[string]string "Invalid refresh token"
// @Router /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *echo.Context) error {
	var req dto.RefreshRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	result, err := h.authService.RefreshToken(c.Request().Context(), req.RefreshToken)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, fmt.Errorf("failed to refresh token: %w", err).Error())
	}

	if err := c.JSON(http.StatusOK, dto.AuthResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		User: &dto.UserDTO{
			ID:          result.User.ID,
			Username:    result.User.Username,
			Email:       result.User.Email,
			DisplayName: result.User.DisplayName,
			AvatarURL:   result.User.AvatarURL,
			Status:      result.User.Status,
		},
		ExpiresAt: result.ExpiresAt,
	}); err != nil {
		return fmt.Errorf("failed to send refresh response: %w", err)
	}
	return nil
}
