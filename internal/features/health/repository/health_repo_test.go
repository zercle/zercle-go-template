package repository_test

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	healthRepo "github.com/zercle/zercle-go-template/internal/features/health/repository"
)

func TestHealthRepository(t *testing.T) {
	t.Run("CheckDatabase_Success", func(t *testing.T) {
		// Arrange
		db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
		assert.NoError(t, err)
		defer db.Close()

		repo := healthRepo.NewHealthRepository(db)
		ctx := context.Background()

		expectedVersion := "PostgreSQL 16.0 on x86_64-pc-linux-gnu"
		rows := sqlmock.NewRows([]string{"version"}).
			AddRow(expectedVersion)

		mock.ExpectPing()
		mock.ExpectQuery("SELECT version()").WillReturnRows(rows)

		// Act
		result, err := repo.CheckDatabase(ctx)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, expectedVersion, result)

		// Verify all expectations were met
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("CheckDatabase_PingFailure", func(t *testing.T) {
		// Arrange
		db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
		assert.NoError(t, err)
		defer db.Close()

		repo := healthRepo.NewHealthRepository(db)
		ctx := context.Background()

		expectedError := errors.New("test error")
		mock.ExpectPing().WillReturnError(expectedError)

		// Act
		result, err := repo.CheckDatabase(ctx)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, "unreachable", result)
		assert.Contains(t, err.Error(), expectedError.Error())

		// Verify all expectations were met
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("CheckDatabase_QueryFailure", func(t *testing.T) {
		// Arrange
		db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
		assert.NoError(t, err)
		defer db.Close()

		repo := healthRepo.NewHealthRepository(db)
		ctx := context.Background()

		expectedError := errors.New("test error")
		mock.ExpectPing()
		mock.ExpectQuery("SELECT version()").WillReturnError(expectedError)

		// Act
		result, err := repo.CheckDatabase(ctx)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, "unreachable", result)
		assert.Contains(t, err.Error(), expectedError.Error())

		// Verify all expectations were met
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
