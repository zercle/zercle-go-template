//go:build unit

package main

import (
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/require"
)

func TestMergedFS_MergesRootListings(t *testing.T) {
	t.Parallel()

	a := fstest.MapFS{"000001_a.up.sql": &fstest.MapFile{Data: []byte("a")}}
	b := fstest.MapFS{"000002_b.up.sql": &fstest.MapFile{Data: []byte("b")}}

	entries, err := mergedFS{a, b}.ReadDir(".")
	require.NoError(t, err)
	require.Len(t, entries, 2)
}

func TestMergedFS_DeduplicatesByName(t *testing.T) {
	t.Parallel()

	a := fstest.MapFS{"shared.txt": &fstest.MapFile{Data: []byte("first")}}
	b := fstest.MapFS{"shared.txt": &fstest.MapFile{Data: []byte("second")}}

	entries, err := mergedFS{a, b}.ReadDir(".")
	require.NoError(t, err)
	require.Len(t, entries, 1)

	content, err := fs.ReadFile(mergedFS{a, b}, "shared.txt")
	require.NoError(t, err)
	require.Equal(t, "first", string(content))
}

func TestMergedFS_ResolvesFilesFromLaterSources(t *testing.T) {
	t.Parallel()

	a := fstest.MapFS{"000001_a.up.sql": &fstest.MapFile{Data: []byte("a")}}
	b := fstest.MapFS{"000002_b.up.sql": &fstest.MapFile{Data: []byte("b")}}

	content, err := fs.ReadFile(mergedFS{a, b}, "000002_b.up.sql")
	require.NoError(t, err)
	require.Equal(t, "b", string(content))
}

func TestMergedFS_MissingPathIsErrNotExist(t *testing.T) {
	t.Parallel()

	a := fstest.MapFS{"000001_a.up.sql": &fstest.MapFile{Data: []byte("a")}}

	_, err := mergedFS{a}.Open("missing.sql")
	require.ErrorIs(t, err, fs.ErrNotExist)

	_, err = mergedFS{}.ReadDir(".")
	require.ErrorIs(t, err, fs.ErrNotExist)
}

func TestMergedFS_OpenRootReturnsDirFile(t *testing.T) {
	t.Parallel()

	a := fstest.MapFS{"000001_a.up.sql": &fstest.MapFile{Data: []byte("a")}}
	b := fstest.MapFS{"000002_b.up.sql": &fstest.MapFile{Data: []byte("b")}}

	f, err := mergedFS{a, b}.Open(".")
	require.NoError(t, err)
	require.NoError(t, f.Close())

	dir, ok := f.(fs.ReadDirFile)
	require.True(t, ok)
	entries, err := dir.ReadDir(-1)
	require.NoError(t, err)
	require.Len(t, entries, 2)
}

func TestMigrationSourcesCoverSQLFiles(t *testing.T) {
	t.Parallel()

	entries, err := mergedFS(migrationSources()).ReadDir(".")
	require.NoError(t, err)
	require.NotEmpty(t, entries, "the example feature ships SQL migrations")
	for _, e := range entries {
		require.Equal(t, ".sql", e.Name()[len(e.Name())-4:], e.Name())
	}
}
