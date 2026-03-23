// Package handler provides HTTP handlers for the user feature.
// It implements the user.Handler interface using Echo v4 framework.
//
// The handlers follow RESTful conventions and use JSend response format:
//
//   - POST   /users     - Create a new user
//   - GET    /users     - List users with optional filtering and pagination
//   - GET    /users/:id - Get a user by ID
//   - PUT    /users/:id - Update a user
//   - DELETE /users/:id - Delete a user
//
// All handlers return JSend-formatted responses:
//
// Success response:
//
//	{
//	  "status": "success",
//	  "data": { ... }
//	}
//
// Fail response (client error):
//
//	{
//	  "status": "fail",
//	  "data": { ... }
//	}
//
// Error response (server error):
//
//	{
//	  "status": "error",
//	  "message": "...",
//	  "code": 5000
//	}
//
// Error mapping:
//   - ErrUserNotFound → 404 Not Found
//   - ErrDuplicateEmail → 409 Conflict
//   - ErrInvalidCredentials → 401 Unauthorized
//   - Validation errors → 400 Bad Request
package handler
