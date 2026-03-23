package handler

import (
	"github.com/labstack/echo/v5"
)

// RegisterRoutes registers all user routes with the given echo group.
// Routes follow RESTful conventions:
//   - POST   /users     - Create a new user
//   - GET    /users     - List users with optional filtering
//   - GET    /users/:id - Get a user by ID
//   - PUT    /users/:id - Update a user
//   - DELETE /users/:id - Delete a user
func (h *Handler) RegisterRoutes(g *echo.Group) {
	g.POST("", h.Create)
	g.GET("", h.List)
	g.GET("/:id", h.GetByID)
	g.PUT("/:id", h.Update)
	g.DELETE("/:id", h.Delete)
}
