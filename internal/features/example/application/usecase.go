// STUB FEATURE — delete internal/features/example to start your project.

package application

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"

	"github.com/zercle/zercle-go-template/internal/features/example/domain"
	"github.com/zercle/zercle-go-template/internal/features/example/port"
)

const (
	defaultPageSizeFallback int32 = 20
	maxPageSizeFallback     int32 = 100
	maxNameLengthFallback   int32 = 255
)

// Usecase implements the Service inbound use-case port.
type Usecase struct {
	repo            port.Repository
	defaultPageSize int32
	maxPageSize     int32
	maxNameLength   int32
}

// NewUsecase returns a Usecase backed by the provided repository. The limit
// arguments override the package fallback defaults; pass <= 0 to use the
// built-in defaults (20/100/255).
func NewUsecase(repo port.Repository, defaultPageSize, maxPageSize, maxNameLength int32) *Usecase {
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

// Create validates the name and persists a new item.
func (u *Usecase) Create(ctx context.Context, name string) (*domain.Item, error) {
	name = strings.TrimSpace(name)
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

	return item, nil
}

// Get retrieves an item by ID, passing through domain.ErrItemNotFound.
func (u *Usecase) Get(ctx context.Context, id uuid.UUID) (*domain.Item, error) {
	if id == uuid.Nil {
		return nil, domain.ErrInvalidID
	}
	item, err := u.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrItemNotFound) {
			return nil, domain.ErrItemNotFound
		}
		return nil, fmt.Errorf("get item: %w", err)
	}

	return item, nil
}

// List returns a paginated list of items. It enforces safe defaults so a
// zero-value limit (e.g. no query parameter) never produces LIMIT 0.
func (u *Usecase) List(ctx context.Context, limit, offset int32) ([]domain.Item, error) {
	if limit <= 0 {
		limit = u.defaultPageSize
	}
	if limit > u.maxPageSize {
		limit = u.maxPageSize
	}
	if offset < 0 {
		offset = 0
	}

	items, err := u.repo.List(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list items: %w", err)
	}

	return items, nil
}
