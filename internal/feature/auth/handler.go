package auth

import (
	"errors"

	"github.com/labstack/echo/v5"
	"github.com/rs/zerolog"

	infraAuth "github.com/zercle/zercle-go-template/internal/infrastructure/auth"
	http_transport "github.com/zercle/zercle-go-template/internal/transport/http"
)

// Handler handles HTTP requests for authentication.
type Handler struct {
	service *Service
	logger  zerolog.Logger
}

// NewHandler creates a new auth HTTP handler.
func NewHandler(service *Service, logger zerolog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// RegisterPublicRoutes registers public authentication routes.
func (h *Handler) RegisterPublicRoutes(g *echo.Group) {
	g.POST("/register", h.Register)
	g.POST("/login", h.Login)
	g.POST("/refresh", h.Refresh)
}

// RegisterProtectedRoutes registers protected authentication routes.
func (h *Handler) RegisterProtectedRoutes(g *echo.Group) {
	g.POST("/logout", h.Logout)
	g.GET("/me", h.Me)
}

// Register handles user registration requests.
func (h *Handler) Register(c *echo.Context) error {
	var req RegisterRequest
	if err := c.Bind(&req); err != nil {
		return http_transport.JSONBadRequest(c, "Invalid request body", err)
	}

	if err := validateRegisterRequest(&req); err != nil {
		return http_transport.JSONBadRequest(c, "Validation failed: "+err.Error(), err)
	}

	resp, err := h.service.Register(c.Request().Context(), &req)
	if err != nil {
		if errors.Is(err, ErrEmailAlreadyExists) {
			return http_transport.JSONConflict(c, "Email already exists")
		}
		return http_transport.JSONInternalError(c, "Internal server error")
	}

	return http_transport.JSONCreated(c, resp)
}

// Login handles user login requests.
func (h *Handler) Login(c *echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return http_transport.JSONBadRequest(c, "Invalid request body", err)
	}

	if err := validateLoginRequest(&req); err != nil {
		return http_transport.JSONBadRequest(c, "Validation failed: "+err.Error(), err)
	}

	resp, err := h.service.Login(c.Request().Context(), &req)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			return http_transport.JSONUnauthorized(c, "Invalid credentials")
		}
		return http_transport.JSONInternalError(c, "Internal server error")
	}

	return http_transport.JSONSuccess(c, resp)
}

// Refresh handles token refresh requests.
func (h *Handler) Refresh(c *echo.Context) error {
	var req RefreshRequest
	if err := c.Bind(&req); err != nil {
		return http_transport.JSONBadRequest(c, "Invalid request body", err)
	}

	if err := validateRefreshRequest(&req); err != nil {
		return http_transport.JSONBadRequest(c, "Validation failed: "+err.Error(), err)
	}

	resp, err := h.service.Refresh(c.Request().Context(), &req)
	if err != nil {
		if errors.Is(err, ErrTokenExpired) || errors.Is(err, ErrInvalidToken) {
			return http_transport.JSONUnauthorized(c, "Invalid or expired token")
		}
		return http_transport.JSONInternalError(c, "Internal server error")
	}

	return http_transport.JSONSuccess(c, resp)
}

// Logout handles user logout requests.
func (h *Handler) Logout(c *echo.Context) error {
	var req LogoutRequest
	if err := c.Bind(&req); err != nil {
		return http_transport.JSONBadRequest(c, "Invalid request body", err)
	}

	if err := validateLogoutRequest(&req); err != nil {
		return http_transport.JSONBadRequest(c, "Validation failed: "+err.Error(), err)
	}

	if err := h.service.Logout(c.Request().Context(), req.RefreshToken); err != nil {
		if errors.Is(err, ErrInvalidToken) {
			return http_transport.JSONUnauthorized(c, "Invalid token")
		}
		return http_transport.JSONInternalError(c, "Internal server error")
	}

	return http_transport.JSONSuccess(c, map[string]string{"message": "Logged out successfully"})
}

// Me returns the current authenticated user's profile.
func (h *Handler) Me(c *echo.Context) error {
	userID := getUserIDFromContext(c)
	if userID == "" {
		return http_transport.JSONUnauthorized(c, "Unauthorized")
	}

	resp, err := h.service.Me(c.Request().Context(), userID)
	if err != nil {
		return http_transport.JSONInternalError(c, "Internal server error")
	}

	return http_transport.JSONSuccess(c, resp)
}

func getUserIDFromContext(c *echo.Context) string {
	claimsVal := c.Get(infraAuth.ContextKeyUserClaims)
	if claimsVal == nil {
		return ""
	}
	claims, ok := claimsVal.(infraAuth.TokenClaims)
	if !ok {
		return ""
	}
	return claims.UserID
}

func validateRegisterRequest(req *RegisterRequest) error {
	if req.Email == "" {
		return errors.New("email is required")
	}
	if req.Password == "" {
		return errors.New("password is required")
	}
	if len(req.Password) < 8 {
		return errors.New("password must be at least 8 characters")
	}
	if req.FirstName == "" {
		return errors.New("first_name is required")
	}
	if req.LastName == "" {
		return errors.New("last_name is required")
	}
	return nil
}

func validateLoginRequest(req *LoginRequest) error {
	if req.Email == "" {
		return errors.New("email is required")
	}
	if req.Password == "" {
		return errors.New("password is required")
	}
	return nil
}

func validateRefreshRequest(req *RefreshRequest) error {
	if req.RefreshToken == "" {
		return errors.New("refresh_token is required")
	}
	return nil
}

func validateLogoutRequest(req *LogoutRequest) error {
	if req.RefreshToken == "" {
		return errors.New("refresh_token is required")
	}
	return nil
}
