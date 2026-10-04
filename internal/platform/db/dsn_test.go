//go:build unit

package db_test

import (
	"context"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"

	"github.com/zercle/zercle-go-template/internal/platform/config"
	"github.com/zercle/zercle-go-template/internal/platform/db"
)

// TestNewDB_NeverLeaksPasswordInError proves that a DSN parse/open failure does
// not echo the configured database password. A malformed host is used because
// it forces the driver to reject the DSN, which is exactly the path where a
// naive `fmt.Errorf("parse dsn: %w", url.Parse...)` used to embed the whole DSN
// (password included) into the error text and thence into logs.
func TestNewDB_NeverLeaksPasswordInError(t *testing.T) {
	const password = "s3cret-pw-should-never-appear"

	cfg := &config.Config{
		DB: config.DBConfig{
			Host:           "ho st", // contains a space: not a valid DSN host
			Port:           5432,
			Name:           "app",
			User:           "app",
			Password:       password,
			SSLMode:        "disable",
			MaxConns:       2,
			MaxIdleConns:   1,
			MaxConnIdle:    5 * time.Second,
			MaxConnLife:    10 * time.Second,
			ConnectTimeout: 1 * time.Second,
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	nop := zerolog.Nop()
	gormDB, err := db.NewDB(ctx, cfg, &nop)

	require.Nil(t, gormDB)
	require.NotEmpty(t, cfg.DBConnString(), "DSN must still be produced from typed fields")
	if err != nil {
		require.NotContains(t, err.Error(), password, "error text must never contain the DB password; got: %s", err.Error())
	}
}
