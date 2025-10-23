package request

import "time"

type CreateTask struct {
	Title       string     `json:"title" validate:"required,min=1,max=255"`
	Description string     `json:"description" validate:"max=1000"`
	Priority    string     `json:"priority" validate:"omitempty,oneof=low medium high urgent"`
	DueDate     *time.Time `json:"due_date"`
}

type UpdateTask struct {
	Title       *string    `json:"title" validate:"omitempty,min=1,max=255"`
	Description *string    `json:"description" validate:"omitempty,max=1000"`
	Status      *string    `json:"status" validate:"omitempty,oneof=pending in_progress completed cancelled"`
	Priority    *string    `json:"priority" validate:"omitempty,oneof=low medium high urgent"`
	DueDate     *time.Time `json:"due_date"`
}
