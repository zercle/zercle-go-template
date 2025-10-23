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

func TestUserRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := postgres.NewUserRepository(db)
	ctx := context.Background()

	user := &domain.User{
		ID:       uuid.New(),
		Email:    "test@example.com",
		Password: "hashed",
		Name:     "Test",
	}

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO users")).
		WithArgs(user.ID, user.Email, user.Password, user.Name, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Create(ctx, user)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetByEmail(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := postgres.NewUserRepository(db)
	ctx := context.Background()

	email := "test@example.com"
	id := uuid.New().String()
	now := time.Now()

	rows := sqlmock.NewRows([]string{"id", "email", "password_hash", "first_name", "last_name", "is_active", "created_at", "updated_at", "deleted_at"}).
		AddRow(id, email, "hashed", "Test", "", true, now, now, nil)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, email, password_hash")).
		WithArgs(email).
		WillReturnRows(rows)

	u, err := repo.GetByEmail(ctx, email)
	assert.NoError(t, err)
	assert.Equal(t, email, u.Email)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := postgres.NewUserRepository(db)
	ctx := context.Background()

	id := uuid.New()
	email := "test@example.com"
	now := time.Now()

	rows := sqlmock.NewRows([]string{"id", "email", "password_hash", "first_name", "last_name", "is_active", "created_at", "updated_at", "deleted_at"}).
		AddRow(id, email, "hashed", "Test", "", true, now, now, nil)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, email, password_hash")).
		WithArgs(id).
		WillReturnRows(rows)

	u, err := repo.GetByID(ctx, id)
	assert.NoError(t, err)
	assert.Equal(t, id, u.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}
