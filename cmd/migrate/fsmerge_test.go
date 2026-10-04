//go:build unit

package main

import (
	"io/fs"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/zercle/zercle-go-template/internal/fsmerge"
)

func TestMigrationSourcesCoverSQLFiles(t *testing.T) {
	t.Parallel()

	entries, err := fs.ReadDir(fsmerge.Merge(migrationSources()...), ".")
	require.NoError(t, err)
	require.NotEmpty(t, entries, "the example feature ships SQL migrations")
	for _, e := range entries {
		require.Equal(t, ".sql", e.Name()[len(e.Name())-4:], e.Name())
	}
}
