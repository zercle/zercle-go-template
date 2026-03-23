package handler

import (
	"github.com/labstack/echo/v5"
)

// RegisterRoutes registers all task routes with the given echo group.
// Routes follow RESTful conventions:
//   - POST   /tasks      - Create a new task
//   - GET    /tasks      - List tasks with optional filtering
//   - GET    /tasks/:id  - Get a task by ID
//   - PUT    /tasks/:id  - Update a task
//   - DELETE /tasks/:id  - Delete a task
func (h *handler) RegisterRoutes(g *echo.Group) {
	g.POST("", h.Create)
	g.GET("", h.List)
	g.GET("/:id", h.GetByID)
	g.PUT("/:id", h.Update)
	g.DELETE("/:id", h.Delete)
}
