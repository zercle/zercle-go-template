package health

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/zercle/zercle-go-template/internal/infrastructure/db"
	"github.com/zercle/zercle-go-template/pkg/response"
)

type Status struct {
	Checks    map[string]string `json:"checks,omitempty"`
	Status    string            `json:"status"`
	Timestamp string            `json:"timestamp"`
}

type Handler struct {
	database db.Database
}

func NewHandler(database db.Database) *Handler {
	return &Handler{
		database: database,
	}
}

func (h *Handler) Check(c echo.Context) error {
	checks := make(map[string]string)
	allHealthy := true

	if h.database != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := h.database.Ping(ctx); err != nil {
			checks["database"] = "unhealthy: " + err.Error()
			allHealthy = false
		} else {
			checks["database"] = "healthy"
		}
	} else {
		checks["database"] = "unhealthy: database not initialized"
		allHealthy = false
	}

	status := "healthy"
	if !allHealthy {
		status = "unhealthy"
	}

	data := map[string]any{
		"status":    status,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"checks":    checks,
	}

	if allHealthy {
		return response.OK(c, data)
	}

	return response.Success(c, http.StatusServiceUnavailable, data)
}
