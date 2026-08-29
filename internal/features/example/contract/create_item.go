// STUB FEATURE — delete internal/features/example to start your project.

// Canonical inbound wire types for the example feature's /api/v1 endpoints.
// This package is the single source of the HTTP JSON shapes; the published
// contract facade pkg/api/v1 re-exports these types as aliases so other
// services can construct payloads without importing server internals.
package contract

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
