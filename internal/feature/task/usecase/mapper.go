package usecase

import (
	"github.com/zercle/zercle-go-template/internal/feature/task"
)

// domainToResponse converts a domain Task entity to a TaskResponse.
func domainToResponse(t *task.Task) *task.TaskResponse {
	if t == nil {
		return nil
	}
	return &task.TaskResponse{
		ID:          string(t.ID),
		Title:       t.Title,
		Description: t.Description,
		Status:      string(t.Status),
		Priority:    string(t.Priority),
		UserID:      t.UserID,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}

// createInputToDomain converts a CreateTaskInput to a domain Task entity.
func createInputToDomain(dto *task.CreateTaskInput) (*task.Task, error) {
	if dto == nil {
		return nil, nil
	}
	return task.NewTask(dto.Title, dto.Description, task.TaskPriority(dto.Priority), dto.UserID)
}

// listParamsToRepo converts ListParamsDTO to repository ListParams.
func listParamsToRepo(dto task.ListParamsDTO) task.ListParams {
	params := task.ListParams{
		Limit:  dto.Limit,
		Offset: dto.Offset,
	}

	if dto.UserID != "" {
		params.Filter.UserID = dto.UserID
	}
	if dto.Status != "" {
		status := task.TaskStatus(dto.Status)
		params.Filter.Status = &status
	}
	if dto.Priority != "" {
		priority := task.TaskPriority(dto.Priority)
		params.Filter.Priority = &priority
	}

	return params
}

// listResultToDTO converts a repository ListResult to a TaskListResponse.
func listResultToDTO(result *task.ListResult) *task.TaskListResponse {
	if result == nil {
		return nil
	}

	tasks := make([]*task.TaskResponse, 0, len(result.Tasks))
	for _, t := range result.Tasks {
		tasks = append(tasks, domainToResponse(t))
	}

	return &task.TaskListResponse{
		Tasks:  tasks,
		Total:  result.Total,
		Limit:  result.Limit,
		Offset: result.Offset,
	}
}
