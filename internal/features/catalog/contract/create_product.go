// Canonical inbound wire types for the catalog feature's /api/v1 endpoints.
// This package is the single source of the HTTP JSON shapes; the published
// contract facade pkg/api/v1 re-exports these types as aliases so other
// services can construct payloads without importing server internals.
package contract

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
