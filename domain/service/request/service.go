package request

import "github.com/go-playground/validator/v10"

// CreateService represents a request to create a new service
type CreateService struct {
	Name            string  `json:"name" validate:"required,min=2,max=100"`
	Description     string  `json:"description" validate:"max=500"`
	DurationMinutes int     `json:"duration_minutes" validate:"required,min=1,max=480"`
	Price           float64 `json:"price" validate:"required,gt=0"`
	MaxCapacity     int     `json:"max_capacity" validate:"required,min=1,max=50"`
	IsActive        bool    `json:"is_active"`
}

// Validate validates the CreateService request
func (r *CreateService) Validate(validate *validator.Validate) error {
	return validate.Struct(r)
}

// UpdateService represents a request to update an existing service
type UpdateService struct {
	Name            *string  `json:"name,omitempty" validate:"omitempty,min=2,max=100"`
	Description     *string  `json:"description,omitempty" validate:"omitempty,max=500"`
	DurationMinutes *int     `json:"duration_minutes,omitempty" validate:"omitempty,min=1,max=480"`
	Price           *float64 `json:"price,omitempty" validate:"omitempty,gt=0"`
	MaxCapacity     *int     `json:"max_capacity,omitempty" validate:"omitempty,min=1,max=50"`
	IsActive        *bool    `json:"is_active,omitempty"`
}

// Validate validates the UpdateService request
func (r *UpdateService) Validate(validate *validator.Validate) error {
	return validate.Struct(r)
}

// ListServices represents a request to list services with pagination
type ListServices struct {
	IsActive bool `query:"is_active"`
	Limit    int  `query:"limit" validate:"min=1,max=100"`
	Offset   int  `query:"offset" validate:"min=0"`
}

// Validate validates the ListServices request
func (r *ListServices) Validate(validate *validator.Validate) error {
	return validate.Struct(r)
}
