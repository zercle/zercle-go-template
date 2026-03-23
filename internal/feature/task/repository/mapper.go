package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	task_entity "github.com/zercle/zercle-go-template/internal/feature/task"
	task_sqlc "github.com/zercle/zercle-go-template/internal/infrastructure/database/sqlc/task"
)

// sqlcTaskToDomain converts a SQLC Task model to a domain Task entity.
func sqlcTaskToDomain(m *task_sqlc.Task) (*task_entity.Task, error) {
	createdAt := time.Time{}
	if m.CreatedAt.Valid {
		createdAt = m.CreatedAt.Time
	}
	updatedAt := time.Time{}
	if m.UpdatedAt.Valid {
		updatedAt = m.UpdatedAt.Time
	}

	var description string
	if m.Description.Valid {
		description = m.Description.String
	}

	return task_entity.NewTaskWithID(
		task_entity.TaskID(m.ID.String()),
		m.Title,
		description,
		task_entity.TaskStatus(m.Status),
		task_entity.TaskPriority(m.Priority),
		m.UserID.String(),
		createdAt,
		updatedAt,
	)
}

// domainTaskToCreateParams converts a domain Task to SQLC CreateTaskParams.
func domainTaskToCreateParams(t *task_entity.Task) task_sqlc.CreateTaskParams {
	return task_sqlc.CreateTaskParams{
		ID:          uuid.MustParse(string(t.ID)),
		Title:       t.Title,
		Description: pgtypeText(t.Description),
		Status:      string(t.Status),
		Priority:    string(t.Priority),
		UserID:      uuid.MustParse(t.UserID),
		CreatedAt:   pgtypeTimestamptz(t.CreatedAt),
		UpdatedAt:   pgtypeTimestamptz(t.UpdatedAt),
	}
}

// domainTaskToUpdateParams converts a domain Task to SQLC UpdateTaskParams.
func domainTaskToUpdateParams(t *task_entity.Task) task_sqlc.UpdateTaskParams {
	return task_sqlc.UpdateTaskParams{
		ID:          uuid.MustParse(string(t.ID)),
		Title:       t.Title,
		Description: pgtypeText(t.Description),
		Status:      string(t.Status),
		Priority:    string(t.Priority),
		UpdatedAt:   pgtypeTimestamptz(t.UpdatedAt),
	}
}

// domainFilterToListParams converts domain ListParams to SQLC ListTasksParams.
func domainFilterToListParams(params *task_entity.ListParams) *task_sqlc.ListTasksParams {
	var userIDFilter pgtype.UUID
	if params.Filter.UserID != "" {
		userIDFilter = pgtype.UUID{
			Bytes: uuid.MustParse(params.Filter.UserID),
			Valid: true,
		}
	}

	var statusFilter string
	if params.Filter.Status != nil {
		statusFilter = string(*params.Filter.Status)
	}

	var priorityFilter string
	if params.Filter.Priority != nil {
		priorityFilter = string(*params.Filter.Priority)
	}

	return &task_sqlc.ListTasksParams{
		Column1: userIDFilter,
		Column2: statusFilter,
		Column3: priorityFilter,
		Limit:   params.Limit,
		Offset:  params.Offset,
	}
}

// domainFilterToCountParams converts domain TaskFilter to SQLC CountTasksParams.
func domainFilterToCountParams(filter task_entity.TaskFilter) *task_sqlc.CountTasksParams {
	var userIDFilter pgtype.UUID
	if filter.UserID != "" {
		userIDFilter = pgtype.UUID{
			Bytes: uuid.MustParse(filter.UserID),
			Valid: true,
		}
	}

	var statusFilter string
	if filter.Status != nil {
		statusFilter = string(*filter.Status)
	}

	var priorityFilter string
	if filter.Priority != nil {
		priorityFilter = string(*filter.Priority)
	}

	return &task_sqlc.CountTasksParams{
		Column1: userIDFilter,
		Column2: statusFilter,
		Column3: priorityFilter,
	}
}

// pgtypeText converts a string to pgtype.Text.
func pgtypeText(s string) pgtype.Text {
	if s == "" {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{
		String: s,
		Valid:  true,
	}
}

// pgtypeTimestamptz converts a time.Time to pgtype.Timestamptz.
func pgtypeTimestamptz(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{
		Time:  t,
		Valid: true,
	}
}
