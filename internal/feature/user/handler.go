package user

import (
	"github.com/labstack/echo/v5"
)

//go:generate mockgen -source=$GOFILE -destination=./mock/$GOFILE -package=mock

// Handler defines the contract for HTTP handlers that process user-related requests.
// This interface follows the hexagonal/clean architecture pattern where the transport
// layer (HTTP handlers) is an adapter that implements the ports defined by the
// application layer.
type Handler interface {
	// RegisterRoutes registers all user routes with the given echo group.
	// Routes should follow RESTful conventions:
	//   - POST   /users          - Create a new user
	//   - GET    /users          - List users with optional filtering
	//   - GET    /users/:id      - Get a user by ID
	//   - GET    /users/by-email - Get a user by email
	//   - PUT    /users/:id      - Update a user
	//   - DELETE /users/:id      - Delete a user
	RegisterRoutes(g *echo.Group)
}
