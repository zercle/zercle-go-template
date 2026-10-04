// STUB FEATURE — delete internal/features/example to start your project.

package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
	"uuid"

	"github.com/zercle/zercle-go-template/internal/features/example/contract"
	"github.com/zercle/zercle-go-template/internal/features/example/domain"
	"github.com/zercle/zercle-go-template/internal/features/example/repository"
)

const (
	defaultPageSizeFallback int32 = 20
	maxPageSizeFallback     int32 = 100
	maxNameLengthFallback   int32 = 255

	timeFormat = time.RFC3339
)

// Usecase implements the Service inbound use-case interface.
type Usecase struct {
	repo            repository.Repository
	defaultPageSize int32
	maxPageSize     int32
	maxNameLength   int32
}

// NewUsecase returns a Usecase backed by the provided repository. The limit
// arguments override the package fallback defaults; pass <= 0 to use the
// built-in defaults (20/100/255).
func NewUsecase(repo repository.Repository, defaultPageSize, maxPageSize, maxNameLength int32) *Usecase {
	if defaultPageSize <= 0 {
		defaultPageSize = defaultPageSizeFallback
	}
	if maxPageSize <= 0 {
		maxPageSize = maxPageSizeFallback
	}
	if maxNameLength <= 0 {
		maxNameLength = maxNameLengthFallback
	}
	return &Usecase{
		repo:            repo,
		defaultPageSize: defaultPageSize,
		maxPageSize:     maxPageSize,
		maxNameLength:   maxNameLength,
	}
}

// Create validates the name, persists a new item, and returns its wire form.
func (u *Usecase) Create(ctx context.Context, req *contract.CreateItemRequest) (*contract.ItemResponse, error) {
	var name string
	if req != nil {
		name = strings.TrimSpace(req.Name)
	}
	if name == "" || utf8.RuneCountInString(name) > int(u.maxNameLength) {
		return nil, domain.ErrInvalidName
	}

	now := time.Now().UTC()
	item := &domain.Item{
		ID:        uuid.New(),
		Name:      name,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := u.repo.Create(ctx, item); err != nil {
		return nil, fmt.Errorf("create item: %w", err)
	}

	resp := newItemResponse(item)
	return &resp, nil
}

// Get retrieves an item by ID, passing through domain.ErrItemNotFound. The
// wire id string is parsed here so both driving adapters share one validation
// path. A syntactically valid but absent id (including the nil UUID) is a
// not-found, which the repository decides.
func (u *Usecase) Get(ctx context.Context, id string) (*contract.ItemResponse, error) {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return nil, domain.ErrInvalidID
	}
	item, err := u.repo.GetByID(ctx, parsed)
	if err != nil {
		if errors.Is(err, domain.ErrItemNotFound) {
			return nil, domain.ErrItemNotFound
		}
		return nil, fmt.Errorf("get item: %w", err)
	}

	resp := newItemResponse(item)
	return &resp, nil
}

// List returns a paginated list of items. It enforces safe defaults so a
// zero-value limit (e.g. no query parameter) never produces LIMIT 0.
//
// Pagination is offset-based, so a concurrent insert between two page requests
// can shift rows across a page boundary (a page may repeat or skip an item).
// Keyset/cursor pagination on (created_at, id) would be stable but changes the
// published request contract, so it is left as a documented limitation here.
func (u *Usecase) List(ctx context.Context, req *contract.ListItemsRequest) (*contract.ListItemsResponse, error) {
	var limit, offset int32
	if req != nil {
		limit, offset = req.Limit, req.Offset
	}
	if limit <= 0 {
		limit = u.defaultPageSize
	}
	if limit > u.maxPageSize {
		limit = u.maxPageSize
	}

	items, err := u.repo.List(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list items: %w", err)
	}

	resp := &contract.ListItemsResponse{Items: make([]contract.ItemResponse, len(items))}
	for i := range items {
		resp.Items[i] = newItemResponse(&items[i])
	}
	return resp, nil
}

func newItemResponse(item *domain.Item) contract.ItemResponse {
	if item == nil {
		return contract.ItemResponse{}
	}
	return contract.ItemResponse{
		ID:        item.ID.String(),
		Name:      item.Name,
		CreatedAt: item.CreatedAt.Format(timeFormat),
		UpdatedAt: item.UpdatedAt.Format(timeFormat),
	}
}
