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
	"github.com/zercle/zercle-go-template/internal/features/example/contract"
	"github.com/zercle/zercle-go-template/pkg/api/errcodes"
)

// CreateItemRequest is the payload for POST /api/v1/items.
type CreateItemRequest = contract.CreateItemRequest

// ItemResponse is the wire representation of an item.
type ItemResponse = contract.ItemResponse

// ListItemsRequest carries the pagination query parameters for GET /api/v1/items.
type ListItemsRequest = contract.ListItemsRequest

// ListItemsResponse wraps a page of items.
type ListItemsResponse = contract.ListItemsResponse

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
