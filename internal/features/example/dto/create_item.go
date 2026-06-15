// STUB FEATURE — delete internal/features/example to start your project.

package dto

// CreateItemRequest is the payload for creating a new item.
type CreateItemRequest struct {
	Name string `json:"name" validate:"required,min=1,max=255"`
}

// ItemResponse is the JSON representation of an item.
type ItemResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}
