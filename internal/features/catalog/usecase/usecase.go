package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
	"uuid"

	"github.com/zercle/zercle-go-template/internal/features/catalog/contract"
	"github.com/zercle/zercle-go-template/internal/features/catalog/domain"
	"github.com/zercle/zercle-go-template/internal/features/catalog/repository"
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

// Create validates the name and price, persists a new product, and returns its
// wire form. The name is trimmed first so surrounding whitespace never counts
// toward the configured length limit.
func (u *Usecase) Create(ctx context.Context, req *contract.CreateProductRequest) (*contract.ProductResponse, error) {
	var name string
	var priceCents, stock int32
	if req != nil {
		name = strings.TrimSpace(req.Name)
		priceCents = req.PriceCents
		stock = req.Stock
	}
	if name == "" || utf8.RuneCountInString(name) > int(u.maxNameLength) {
		return nil, fmt.Errorf("%w: must be 1..%d characters", domain.ErrInvalidProductName, u.maxNameLength)
	}
	if priceCents <= 0 {
		return nil, domain.ErrInvalidPrice
	}

	now := time.Now().UTC()
	product := &domain.Product{
		ID:         uuid.New(),
		Name:       name,
		PriceCents: priceCents,
		Stock:      stock,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if err := u.repo.Create(ctx, product); err != nil {
		return nil, fmt.Errorf("create product: %w", err)
	}

	resp := newProductResponse(product)
	return &resp, nil
}

// Get retrieves a product by ID, passing through domain.ErrProductNotFound. The
// wire id string is parsed here so both driving adapters share one validation
// path. A syntactically valid but absent id (including the nil UUID) is a
// not-found, which the repository decides.
func (u *Usecase) Get(ctx context.Context, id string) (*contract.ProductResponse, error) {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return nil, domain.ErrInvalidID
	}
	product, err := u.repo.GetByID(ctx, parsed)
	if err != nil {
		if errors.Is(err, domain.ErrProductNotFound) {
			return nil, domain.ErrProductNotFound
		}
		return nil, fmt.Errorf("get product: %w", err)
	}

	resp := newProductResponse(product)
	return &resp, nil
}

// List returns a paginated list of products. It enforces safe defaults so a
// zero-value limit (e.g. no query parameter) never produces LIMIT 0 and a
// negative offset never reaches the query.
//
// Pagination is offset-based, so a concurrent insert between two page requests
// can shift rows across a page boundary (a page may repeat or skip a product).
// Keyset/cursor pagination on (created_at, id) would be stable but changes the
// published request contract, so it is left as a documented limitation here.
func (u *Usecase) List(ctx context.Context, req *contract.ListProductsRequest) (*contract.ListProductsResponse, error) {
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
	if offset < 0 {
		offset = 0
	}

	products, err := u.repo.List(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list products: %w", err)
	}

	resp := &contract.ListProductsResponse{Products: make([]contract.ProductResponse, len(products))}
	for i := range products {
		resp.Products[i] = newProductResponse(&products[i])
	}
	return resp, nil
}

func newProductResponse(product *domain.Product) contract.ProductResponse {
	if product == nil {
		return contract.ProductResponse{}
	}
	return contract.ProductResponse{
		ID:         product.ID.String(),
		Name:       product.Name,
		PriceCents: product.PriceCents,
		Stock:      product.Stock,
		CreatedAt:  product.CreatedAt.Format(timeFormat),
		UpdatedAt:  product.UpdatedAt.Format(timeFormat),
	}
}
