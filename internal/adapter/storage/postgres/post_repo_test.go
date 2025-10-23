package postgres_test

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/zercle/zercle-go-template/internal/adapter/storage/postgres"
	"github.com/zercle/zercle-go-template/internal/core/domain"
)

func TestPostRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := postgres.NewPostRepository(db)
	ctx := context.Background()

	post := &domain.Post{
		ID:        uuid.New(),
		Title:     "Test Title",
		Content:   "Test Content",
		AuthorID:  uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO posts")).
		WithArgs(post.ID, post.Title, post.Content, post.AuthorID, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Create(ctx, post)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostRepository_GetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := postgres.NewPostRepository(db)
	ctx := context.Background()

	id := uuid.New()
	authorID := uuid.New()
	now := time.Now()

	rows := sqlmock.NewRows([]string{"id", "title", "content", "author_id", "created_at", "updated_at"}).
		AddRow(id.String(), "Title", "Content", authorID.String(), now, now)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, title, content, author_id, created_at, updated_at FROM posts WHERE id = $1")).
		WithArgs(id).
		WillReturnRows(rows)

	p, err := repo.GetByID(ctx, id)
	assert.NoError(t, err)
	assert.Equal(t, "Title", p.Title)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostRepository_GetAll(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := postgres.NewPostRepository(db)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"id", "title", "content", "author_id", "created_at", "updated_at"}).
		AddRow(uuid.New().String(), "Title 1", "Content 1", uuid.New().String(), time.Now(), time.Now()).
		AddRow(uuid.New().String(), "Title 2", "Content 2", uuid.New().String(), time.Now(), time.Now())

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, title, content, author_id, created_at, updated_at FROM posts ORDER BY created_at DESC")).
		WillReturnRows(rows)

	posts, err := repo.GetAll(ctx)
	assert.NoError(t, err)
	assert.Len(t, posts, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}
