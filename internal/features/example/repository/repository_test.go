//go:build unit

// STUB FEATURE — delete internal/features/example to start your project.

package repository_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zercle/zercle-go-template/internal/features/example/domain"
	"github.com/zercle/zercle-go-template/internal/features/example/repository"
	sqlcdb "github.com/zercle/zercle-go-template/internal/infrastructure/db/sqlc"
)

// mockDBTX implements sqlc.DBTX for in-memory repository tests.
type mockDBTX struct {
	items     []sqlcdb.Item
	listItems []sqlcdb.Item
	listErr   error
	err       error
}

func (m *mockDBTX) Exec(_ context.Context, _ string, args ...interface{}) (pgconn.CommandTag, error) {
	if m.err != nil {
		return pgconn.CommandTag{}, m.err
	}
	id := args[0].(uuid.UUID)
	m.items = append(m.items, sqlcdb.Item{
		ID:        id,
		Name:      args[1].(string),
		CreatedAt: args[2].(time.Time),
		UpdatedAt: args[3].(time.Time),
	})
	return pgconn.CommandTag{}, nil
}

func (m *mockDBTX) Query(_ context.Context, _ string, _ ...interface{}) (pgx.Rows, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return &mockRows{items: m.listItems}, nil
}

func (m *mockDBTX) QueryRow(_ context.Context, _ string, _ ...interface{}) pgx.Row {
	return &mockRow{item: m.items, err: m.err}
}

// mockRow implements pgx.Row for the fake DBTX.
type mockRow struct {
	item []sqlcdb.Item
	err  error
}

func (r *mockRow) Scan(dest ...interface{}) error {
	if r.err != nil {
		return r.err
	}
	if len(r.item) == 0 {
		return pgx.ErrNoRows
	}
	i := r.item[0]
	*dest[0].(*uuid.UUID) = i.ID
	*dest[1].(*string) = i.Name
	*dest[2].(*time.Time) = i.CreatedAt
	*dest[3].(*time.Time) = i.UpdatedAt
	return nil
}

// mockRows implements pgx.Rows for the fake DBTX.
type mockRows struct {
	items []sqlcdb.Item
	idx   int
}

func (r *mockRows) Next() bool {
	if r.idx >= len(r.items) {
		return false
	}
	r.idx++
	return true
}

func (r *mockRows) Scan(dest ...interface{}) error {
	i := r.items[r.idx-1]
	*dest[0].(*uuid.UUID) = i.ID
	*dest[1].(*string) = i.Name
	*dest[2].(*time.Time) = i.CreatedAt
	*dest[3].(*time.Time) = i.UpdatedAt
	return nil
}

func (r *mockRows) Close()                                       {}
func (r *mockRows) Err() error                                   { return nil }
func (r *mockRows) CommandTag() pgconn.CommandTag                { return pgconn.CommandTag{} }
func (r *mockRows) FieldDescriptions() []pgconn.FieldDescription { return nil }
func (r *mockRows) Values() ([]any, error)                       { return nil, nil }
func (r *mockRows) RawValues() [][]byte                          { return nil }
func (r *mockRows) Conn() *pgx.Conn                              { return nil }

func TestRepository_Create(t *testing.T) {
	dbtx := &mockDBTX{}
	repo := repository.NewRepository(nil, sqlcdb.New(dbtx))

	item := &domain.Item{
		ID:        uuid.New(),
		Name:      "repo-item",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	err := repo.Create(context.Background(), item)
	require.NoError(t, err)
	assert.Len(t, dbtx.items, 1)
}

func TestRepository_Create_Error(t *testing.T) {
	dbtx := &mockDBTX{err: errors.New("exec failed")}
	repo := repository.NewRepository(nil, sqlcdb.New(dbtx))

	item := &domain.Item{ID: uuid.New(), Name: "x", CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}
	err := repo.Create(context.Background(), item)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "exec failed")
}

func TestRepository_GetByID(t *testing.T) {
	id := uuid.New()
	dbtx := &mockDBTX{items: []sqlcdb.Item{{ID: id, Name: "found", CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}}}
	repo := repository.NewRepository(nil, sqlcdb.New(dbtx))

	got, err := repo.GetByID(context.Background(), id)
	require.NoError(t, err)
	assert.Equal(t, id, got.ID)
}

func TestRepository_GetByID_NotFound(t *testing.T) {
	dbtx := &mockDBTX{}
	repo := repository.NewRepository(nil, sqlcdb.New(dbtx))

	got, err := repo.GetByID(context.Background(), uuid.New())
	assert.Nil(t, got)
	assert.True(t, errors.Is(err, domain.ErrItemNotFound))
}

func TestRepository_List(t *testing.T) {
	id := uuid.New()
	dbtx := &mockDBTX{listItems: []sqlcdb.Item{{ID: id, Name: "listed", CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}}}
	repo := repository.NewRepository(nil, sqlcdb.New(dbtx))

	items, err := repo.List(context.Background(), 10, 0)
	require.NoError(t, err)
	assert.Len(t, items, 1)
	assert.Equal(t, id, items[0].ID)
}

func TestRepository_List_Error(t *testing.T) {
	dbtx := &mockDBTX{listErr: errors.New("query failed")}
	repo := repository.NewRepository(nil, sqlcdb.New(dbtx))

	items, err := repo.List(context.Background(), 10, 0)
	assert.Error(t, err)
	assert.Nil(t, items)
}
