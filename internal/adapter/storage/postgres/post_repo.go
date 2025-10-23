package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/zercle/zercle-go-template/internal/core/domain"
	domErr "github.com/zercle/zercle-go-template/internal/core/domain/errors"
	"github.com/zercle/zercle-go-template/internal/core/port"
	db "github.com/zercle/zercle-go-template/internal/infrastructure/sqlc"
)

type postRepository struct {
	q  *db.Queries
	db *sql.DB
}

// NewPostRepository creates a new MySQL repository for Posts.
func NewPostRepository(d *sql.DB) port.PostRepository {
	return &postRepository{
		q:  db.New(d),
		db: d,
	}
}

func (r *postRepository) Create(ctx context.Context, p *domain.Post) error {
	return r.q.CreatePost(ctx, db.CreatePostParams{
		ID:       p.ID,
		Title:    p.Title,
		Content:  p.Content,
		AuthorID: p.AuthorID,
		// CreatedAt, UpdatedAt handled by DB default usually, but we can set strict
		CreatedAt: sql.NullTime{Time: p.CreatedAt, Valid: true},
		UpdatedAt: sql.NullTime{Time: p.UpdatedAt, Valid: true},
	})
}

func (r *postRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Post, error) {
	p, err := r.q.GetPost(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domErr.ErrNotFound
		}
		return nil, err
	}
	return mapPostToDomain(p), nil
}

func (r *postRepository) GetAll(ctx context.Context) ([]*domain.Post, error) {
	// ListPosts in query file doesn't take paging args in schema step?
	// I didn't verify queries arguments. I'll assume ListPosts takes no arg or I check query.
	// Step 154: SELECT * FROM posts ORDER BY created_at DESC;
	posts, err := r.q.ListPosts(ctx)
	if err != nil {
		return nil, err
	}

	res := make([]*domain.Post, len(posts))
	for i, p := range posts {
		res[i] = mapPostToDomain(p)
	}
	return res, nil
}

func (r *postRepository) Update(ctx context.Context, p *domain.Post) error {
	return r.q.UpdatePost(ctx, db.UpdatePostParams{
		ID:      p.ID,
		Title:   p.Title,
		Content: p.Content,
	})
}

func (r *postRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.q.DeletePost(ctx, id)
}

func mapPostToDomain(p db.Post) *domain.Post {
	return &domain.Post{
		ID:        p.ID,
		Title:     p.Title,
		Content:   p.Content,
		AuthorID:  p.AuthorID,
		CreatedAt: p.CreatedAt.Time,
		UpdatedAt: p.UpdatedAt.Time,
	}
}
