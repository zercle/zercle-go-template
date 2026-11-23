// Package apiv1 is the canonical published wire contract of the /api/v1
// endpoints: the single source of the HTTP JSON shapes (request and response
// types) plus the error codes carried in the {"error": code, "message": msg}
// envelope. Other services import it to construct payloads and interpret the
// error envelope without depending on server internals.
//
// The package stays a LEAF: it imports nothing from this module except
// pkg/api/errcodes, so a consumer never drags the server in. A future v2
// contract is a new facade package, not a change here.
package apiv1

// CreateProductRequest is the payload for adding a product to the global
// catalog pool. Only structural constraints live here; the name-length limit
// is deployment-configurable and enforced in the usecase layer, so a
// hardcoded max= tag would drift from the actual limit.
type CreateProductRequest struct {
	Name       string `json:"name" validate:"required"`
	PriceCents int32  `json:"price_cents" validate:"required,gt=0"`
	Stock      int32  `json:"stock" validate:"gte=0"`
}

// ProductResponse is the JSON representation of a product.
type ProductResponse struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	PriceCents int32  `json:"price_cents"`
	Stock      int32  `json:"stock"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}

// ListProductsRequest carries pagination parameters for listing products. It
// is bound from the GET query string (the query tags); the json tags exist so
// the type still marshals to the published lower-cased shape, not because the
// handler reads a JSON body. Only structural constraints live here; the upper
// page-size limit is deployment-configurable and enforced in the usecase
// layer, so a hardcoded max= tag would drift.
type ListProductsRequest struct {
	Limit  int32 `json:"limit" query:"limit" validate:"min=0"`
	Offset int32 `json:"offset" query:"offset" validate:"min=0"`
}

// ListProductsResponse wraps a page of products.
type ListProductsResponse struct {
	Products []ProductResponse `json:"products"`
}
