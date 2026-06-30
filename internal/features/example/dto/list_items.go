// STUB FEATURE — delete internal/features/example to start your project.

package dto

// ListItemsRequest carries pagination parameters for listing items.
type ListItemsRequest struct {
	Limit  int32 `json:"limit" query:"limit" validate:"omitempty,min=0,max=100"`
	Offset int32 `json:"offset" query:"offset" validate:"omitempty,min=0"`
}

// ListItemsResponse wraps a page of items.
type ListItemsResponse struct {
	Items []ItemResponse `json:"items"`
}
