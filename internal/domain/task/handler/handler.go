package handler

import (
	"errors"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/zercle/zercle-go-template/internal/domain/task"
	"github.com/zercle/zercle-go-template/internal/domain/task/request"
	"github.com/zercle/zercle-go-template/internal/domain/task/usecase"
	"github.com/zercle/zercle-go-template/internal/infrastructure/logger"
	"github.com/zercle/zercle-go-template/pkg/middleware"
	"github.com/zercle/zercle-go-template/pkg/response"
)

type taskHandler struct {
	usecase task.TaskService
	log     *logger.Logger
}

func NewTaskHandler(usecase task.TaskService, log *logger.Logger) task.TaskHandler {
	return &taskHandler{
		usecase: usecase,
		log:     log,
	}
}

func (h *taskHandler) CreateTask(c echo.Context) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return response.Unauthorized(c, "Invalid token")
	}

	id, err := uuid.Parse(userID)
	if err != nil {
		return response.BadRequest(c, "Invalid user ID", nil)
	}

	var req request.CreateTask
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "Invalid request body", nil)
	}

	if err := c.Validate(&req); err != nil {
		return response.BadRequest(c, "Validation failed", middleware.ValidationErrors(err))
	}

	result, err := h.usecase.CreateTask(c.Request().Context(), id, req)
	if err != nil {
		h.log.Error("Failed to create task", "error", err, "request_id", middleware.GetRequestID(c))
		return response.InternalError(c, "Failed to create task")
	}

	return response.Created(c, result)
}

func (h *taskHandler) GetTask(c echo.Context) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return response.Unauthorized(c, "Invalid token")
	}

	uid, err := uuid.Parse(userID)
	if err != nil {
		return response.BadRequest(c, "Invalid user ID", nil)
	}

	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid task ID", nil)
	}

	result, err := h.usecase.GetTask(c.Request().Context(), uid, taskID)
	if err != nil {
		h.log.Error("Failed to get task", "error", err, "request_id", middleware.GetRequestID(c))
		if errors.Is(err, usecase.ErrTaskNotFound) {
			return response.NotFound(c, "Task not found")
		}
		if errors.Is(err, usecase.ErrTaskNotOwned) {
			return response.Forbidden(c, "You don't have access to this task")
		}
		return response.InternalError(c, "Failed to get task")
	}

	return response.OK(c, result)
}

func (h *taskHandler) ListTasks(c echo.Context) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return response.Unauthorized(c, "Invalid token")
	}

	uid, err := uuid.Parse(userID)
	if err != nil {
		return response.BadRequest(c, "Invalid user ID", nil)
	}

	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	offset, _ := strconv.Atoi(c.QueryParam("offset"))
	if offset < 0 {
		offset = 0
	}

	result, err := h.usecase.ListTasks(c.Request().Context(), uid, limit, offset)
	if err != nil {
		h.log.Error("Failed to list tasks", "error", err, "request_id", middleware.GetRequestID(c))
		return response.InternalError(c, "Failed to list tasks")
	}

	return response.OK(c, result)
}

func (h *taskHandler) UpdateTask(c echo.Context) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return response.Unauthorized(c, "Invalid token")
	}

	uid, err := uuid.Parse(userID)
	if err != nil {
		return response.BadRequest(c, "Invalid user ID", nil)
	}

	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid task ID", nil)
	}

	var req request.UpdateTask
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "Invalid request body", nil)
	}

	if err := c.Validate(&req); err != nil {
		return response.BadRequest(c, "Validation failed", middleware.ValidationErrors(err))
	}

	result, err := h.usecase.UpdateTask(c.Request().Context(), uid, taskID, req)
	if err != nil {
		h.log.Error("Failed to update task", "error", err, "request_id", middleware.GetRequestID(c))
		if errors.Is(err, usecase.ErrTaskNotFound) {
			return response.NotFound(c, "Task not found")
		}
		if errors.Is(err, usecase.ErrTaskNotOwned) {
			return response.Forbidden(c, "You don't have access to this task")
		}
		if errors.Is(err, usecase.ErrInvalidTaskStatus) {
			return response.BadRequest(c, "Invalid task status", nil)
		}
		if errors.Is(err, usecase.ErrInvalidTaskPriority) {
			return response.BadRequest(c, "Invalid task priority", nil)
		}
		return response.InternalError(c, "Failed to update task")
	}

	return response.OK(c, result)
}

func (h *taskHandler) DeleteTask(c echo.Context) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return response.Unauthorized(c, "Invalid token")
	}

	uid, err := uuid.Parse(userID)
	if err != nil {
		return response.BadRequest(c, "Invalid user ID", nil)
	}

	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid task ID", nil)
	}

	if err := h.usecase.DeleteTask(c.Request().Context(), uid, taskID); err != nil {
		h.log.Error("Failed to delete task", "error", err, "request_id", middleware.GetRequestID(c))
		if errors.Is(err, usecase.ErrTaskNotFound) {
			return response.NotFound(c, "Task not found")
		}
		if errors.Is(err, usecase.ErrTaskNotOwned) {
			return response.Forbidden(c, "You don't have access to this task")
		}
		return response.InternalError(c, "Failed to delete task")
	}

	return response.NoContent(c)
}

func (h *taskHandler) RegisterRoutes(protected *echo.Group) {
	protected.POST("/tasks", h.CreateTask)
	protected.GET("/tasks", h.ListTasks)
	protected.GET("/tasks/:id", h.GetTask)
	protected.PUT("/tasks/:id", h.UpdateTask)
	protected.DELETE("/tasks/:id", h.DeleteTask)
}
