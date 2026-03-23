package task

import (
	"github.com/labstack/echo/v5"
)

//go:generate mockgen -source=$GOFILE -destination=./mock/$GOFILE -package=mock

// Handler defines the contract for HTTP handlers that process task-related requests.
// This interface follows the hexagonal/clean architecture pattern where the transport
// layer (HTTP handlers) is an adapter that implements the ports defined by the
// application layer.
type Handler interface {
	// RegisterRoutes registers all task routes with the given echo group.
	// Routes should follow RESTful conventions:
	//   - POST   /tasks         - Create a new task
	//   - GET    /tasks         - List tasks with optional filtering
	//   - GET    /tasks/:id     - Get a task by ID
	//   - PUT    /tasks/:id     - Update a task
	//   - DELETE /tasks/:id     - Delete a task
	RegisterRoutes(g *echo.Group)
}
