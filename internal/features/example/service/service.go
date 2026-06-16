// STUB FEATURE — delete internal/features/example to start your project.

package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/zercle/zercle-go-template/internal/features/example/domain"
)

const (
	defaultPageSizeFallback int32 = 20
	maxPageSizeFallback     int32 = 100
	maxNameLengthFallback         = 255
)

// Service implements the domain.Service inbound use-case port.
type Service struct {
	repo            domain.Repository
	defaultPageSize int32
	maxPageSize     int32
	maxNameLength   int32
}

// NewService returns a Service backed by the provided repository. The limit
// arguments override the package fallback defaults; pass <= 0 to use the
// built-in defaults (20/100/255).
func NewService(repo domain.Repository, defaultPageSize, maxPageSize, maxNameLength int32) *Service {
	if defaultPageSize <= 0 {
		defaultPageSize = defaultPageSizeFallback
	}
	if maxPageSize <= 0 {
		maxPageSize = maxPageSizeFallback
	}
	if maxNameLength <= 0 {
		maxNameLength = maxNameLengthFallback
	}
	return &Service{
		repo:            repo,
		defaultPageSize: defaultPageSize,
		maxPageSize:     maxPageSize,
		maxNameLength:   maxNameLength,
	}
}

// Create validates the name and persists a new item.
func (s *Service) Create(ctx context.Context, name string) (*domain.Item, error) {
	name = strings.TrimSpace(name)
	if name == "" || int32(len(name)) > s.maxNameLength {
		return nil, domain.ErrInvalidName
	}

	now := time.Now().UTC()
	item := &domain.Item{
		ID:        uuid.New(),
		Name:      name,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.repo.Create(ctx, item); err != nil {
		return nil, fmt.Errorf("create item: %w", err)
	}

	return item, nil
}

// Get retrieves an item by ID, passing through domain.ErrItemNotFound.
func (s *Service) Get(ctx context.Context, id uuid.UUID) (*domain.Item, error) {
	item, err := s.repo.GetByID(ctx, id)
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
func (s *Service) List(ctx context.Context, limit, offset int32) ([]domain.Item, error) {
	if limit <= 0 {
		limit = s.defaultPageSize
	}
	if limit > s.maxPageSize {
		limit = s.maxPageSize
	}
	if offset < 0 {
		offset = 0
	}

	items, err := s.repo.List(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list items: %w", err)
	}

	return items, nil
}
