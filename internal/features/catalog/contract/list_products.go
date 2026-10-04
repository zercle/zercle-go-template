package contract

// ListProductsRequest carries pagination parameters for listing products. It
// is bound from the GET query string (the query tags); the json tags exist so
// the type still marshals to the published lower-cased shape as part of
// pkg/api/v1, not because the handler reads a JSON body. Only structural
// constraints live here; the upper page-size limit is deployment-configurable
// and enforced in the usecase layer, so a hardcoded max= tag would drift.
type ListProductsRequest struct {
	Limit  int32 `json:"limit" query:"limit" validate:"min=0"`
	Offset int32 `json:"offset" query:"offset" validate:"min=0"`
}

// ListProductsResponse wraps a page of products.
type ListProductsResponse struct {
	Products []ProductResponse `json:"products"`
}
