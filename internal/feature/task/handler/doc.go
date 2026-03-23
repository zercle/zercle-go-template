// Package handler provides HTTP handlers for the task feature.
// It implements the task.Handler interface using Echo v4 framework.
//
// The handlers follow RESTful conventions and use JSend response format:
//
//   - POST   /tasks        - Create a new task
//   - GET    /tasks        - List tasks with optional filtering and pagination
//   - GET    /tasks/:id    - Get a task by ID
//   - PUT    /tasks/:id    - Update a task
//   - DELETE /tasks/:id    - Delete a task
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
//   - ErrTaskNotFound → 404 Not Found
//   - ErrTaskAlreadyExists → 409 Conflict
//   - ErrInvalidTaskStatus → 400 Bad Request
//   - ErrInvalidTaskPriority → 400 Bad Request
//   - Validation errors → 400 Bad Request
package handler
