// STUB FEATURE — delete internal/features/example to start your project.

package contract

// ListItemsRequest carries pagination parameters for listing items. Only
// structural constraints live here; the upper page-size limit is
// deployment-configurable (EXAMPLE_MAX_PAGE_SIZE) and enforced in the
// usecase layer, so a hardcoded max= tag would drift from the actual limit.
type ListItemsRequest struct {
	Limit  int32 `json:"limit" query:"limit" validate:"min=0"`
	Offset int32 `json:"offset" query:"offset" validate:"min=0"`
}

// ListItemsResponse wraps a page of items.
type ListItemsResponse struct {
	Items []ItemResponse `json:"items"`
}
