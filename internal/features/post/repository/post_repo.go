package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/zercle/zercle-go-template/internal/core/port"
	postDomain "github.com/zercle/zercle-go-template/internal/features/post/domain"
	db "github.com/zercle/zercle-go-template/internal/infrastructure/sqlc"
	sharedDomain "github.com/zercle/zercle-go-template/internal/shared/domain"
)

type postRepository struct {
	q  *db.Queries
	db *sql.DB
}

// NewPostRepository creates a new PostgreSQL repository for Posts.
func NewPostRepository(d *sql.DB) port.PostRepository {
	return &postRepository{
		q:  db.New(d),
		db: d,
	}
}

func (r *postRepository) Create(ctx context.Context, p *postDomain.Post) error {
	err := r.q.CreatePost(ctx, db.CreatePostParams{
		ID:       p.ID,
		Title:    p.Title,
		Content:  p.Content,
		AuthorID: uuid.UUID(p.AuthorID),
	})
	return err
}

func (r *postRepository) GetByID(ctx context.Context, id uuid.UUID) (*postDomain.Post, error) {
	post, err := r.q.GetPost(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return mapPostToDomain(post), nil
}

func (r *postRepository) GetAll(ctx context.Context) ([]*postDomain.Post, error) {
	posts, err := r.q.ListPosts(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*postDomain.Post, 0, len(posts))
	for _, post := range posts {
		result = append(result, mapPostToDomain(post))
	}

	return result, nil
}

func (r *postRepository) Update(ctx context.Context, p *postDomain.Post) error {
	err := r.q.UpdatePost(ctx, db.UpdatePostParams{
		Title:   p.Title,
		Content: p.Content,
		ID:      p.ID,
	})
	return err
}

func (r *postRepository) Delete(ctx context.Context, id uuid.UUID) error {
	err := r.q.DeletePost(ctx, id)
	return err
}

func mapPostToDomain(p db.Post) *postDomain.Post {
	return &postDomain.Post{
		ID:        p.ID,
		Title:     p.Title,
		Content:   p.Content,
		AuthorID:  sharedDomain.UserID(p.AuthorID),
		CreatedAt: p.CreatedAt.Time,
		UpdatedAt: p.UpdatedAt.Time,
	}
}
