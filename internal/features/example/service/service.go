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
	defaultPageSize int32 = 20
	maxPageSize     int32 = 100
	maxNameLength         = 255
)

// Service implements the domain.Service inbound use-case port.
type Service struct {
	repo domain.Repository
}

// NewService returns a Service backed by the provided repository.
func NewService(repo domain.Repository) *Service {
	return &Service{repo: repo}
}

// Create validates the name and persists a new item.
func (s *Service) Create(ctx context.Context, name string) (*domain.Item, error) {
	name = strings.TrimSpace(name)
	if name == "" || len(name) > maxNameLength {
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
		limit = defaultPageSize
	}
	if limit > maxPageSize {
		limit = maxPageSize
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
