// Package apiv1 is the published inbound contract of the /api/v1 endpoints:
// the request and response wire types, plus the error codes, that other
// services may import to construct payloads and interpret error envelopes
// without depending on server internals.
//
// It is a facade of type aliases over the canonical types in the owning
// feature's contract package. Internal code must not import this package —
// the dependency is strictly outward-only (enforced by the architecture
// test). A future v2 contract is a new facade package, not a change here.
package apiv1

import (
	catalogcontract "github.com/zercle/zercle-go-template/internal/features/catalog/contract"
	machinescontract "github.com/zercle/zercle-go-template/internal/features/machines/contract"
	salescontract "github.com/zercle/zercle-go-template/internal/features/sales/contract"
	"github.com/zercle/zercle-go-template/pkg/api/errcodes"
)

// --- catalog ---------------------------------------------------------------

// CreateProductRequest is the payload for POST /api/v1/products.
type CreateProductRequest = catalogcontract.CreateProductRequest

// ProductResponse is the wire representation of a product.
type ProductResponse = catalogcontract.ProductResponse

// ListProductsRequest carries the pagination query parameters for
// GET /api/v1/products.
type ListProductsRequest = catalogcontract.ListProductsRequest

// ListProductsResponse wraps a page of products.
type ListProductsResponse = catalogcontract.ListProductsResponse

// --- machines --------------------------------------------------------------

// CreateMachineRequest is the payload for POST /api/v1/machines.
type CreateMachineRequest = machinescontract.CreateMachineRequest

// MachineResponse is the wire representation of a machine and its coin bank.
type MachineResponse = machinescontract.MachineResponse

// RestockBankRequest is the payload for POST /api/v1/machines/{id}/bank.
type RestockBankRequest = machinescontract.RestockBankRequest

// ListMachinesRequest carries the pagination query parameters for
// GET /api/v1/machines.
type ListMachinesRequest = machinescontract.ListMachinesRequest

// ListMachinesResponse wraps a page of machines.
type ListMachinesResponse = machinescontract.ListMachinesResponse

// --- sales -----------------------------------------------------------------

// PurchaseRequest is the payload for POST /api/v1/purchases.
type PurchaseRequest = salescontract.PurchaseRequest

// PurchaseResponse is the wire representation of a completed purchase.
type PurchaseResponse = salescontract.PurchaseResponse

// Error codes carried in the {"error": code, "message": msg} response
// envelope, re-exported from the version-independent errcodes package.
const (
	ErrCodeNotFound         = errcodes.NotFound
	ErrCodeInvalidInput     = errcodes.InvalidInput
	ErrCodeUnauthorized     = errcodes.Unauthorized
	ErrCodeForbidden        = errcodes.Forbidden
	ErrCodeConflict         = errcodes.Conflict
	ErrCodeCanceled         = errcodes.Canceled
	ErrCodeDeadlineExceeded = errcodes.DeadlineExceeded
	ErrCodeInternal         = errcodes.Internal
	ErrCodeMethodNotAllowed = errcodes.MethodNotAllowed
	ErrCodePayloadTooLarge  = errcodes.PayloadTooLarge
)
