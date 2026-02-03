package http

import (
	"net/http"

	"github.com/labstack/echo/v5"

	"github.com/zercle/zercle-go-template/internal/features/auth/dto"
	"github.com/zercle/zercle-go-template/internal/features/auth/service"
	"github.com/zercle/zercle-go-template/internal/shared/middleware"
)

type AuthHandler struct {
	authService service.AuthServiceInterface
}

func NewAuthHandler(authService service.AuthServiceInterface) *AuthHandler {
	return &AuthHandler{authService: authService}
}

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
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusCreated, dto.AuthResponse{
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
	})
}

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
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid credentials")
	}

	return c.JSON(http.StatusOK, dto.AuthResponse{
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
	})
}

func (h *AuthHandler) GetCurrentUser(c *echo.Context) error {
	_, err := middleware.GetUserID(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}

	user, err := h.authService.ValidateToken(c.Request().Context(), c.Request().Header.Get("Authorization"))
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
	}

	return c.JSON(http.StatusOK, dto.UserDTO{
		ID:          user.ID,
		Username:    user.Username,
		Email:       user.Email,
		DisplayName: user.DisplayName,
		AvatarURL:   user.AvatarURL,
		Status:      user.Status,
	})
}

func (h *AuthHandler) Logout(c *echo.Context) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}

	if err := h.authService.Logout(c.Request().Context(), userID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *AuthHandler) RefreshToken(c *echo.Context) error {
	var req dto.RefreshRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	result, err := h.authService.RefreshToken(c.Request().Context(), req.RefreshToken)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid refresh token")
	}

	return c.JSON(http.StatusOK, dto.AuthResponse{
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
	})
}
