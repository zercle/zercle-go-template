//go:build unit

package fsmerge_test

import (
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/require"

	"github.com/zercle/zercle-go-template/internal/fsmerge"
)

func TestMerge_MergesRootListings(t *testing.T) {
	t.Parallel()

	a := fstest.MapFS{"000001_a.up.sql": &fstest.MapFile{Data: []byte("a")}}
	b := fstest.MapFS{"000002_b.up.sql": &fstest.MapFile{Data: []byte("b")}}

	entries, err := fs.ReadDir(fsmerge.Merge(a, b), ".")
	require.NoError(t, err)
	require.Len(t, entries, 2)
}

func TestMerge_DeduplicatesByName(t *testing.T) {
	t.Parallel()

	a := fstest.MapFS{"shared.txt": &fstest.MapFile{Data: []byte("first")}}
	b := fstest.MapFS{"shared.txt": &fstest.MapFile{Data: []byte("second")}}

	entries, err := fs.ReadDir(fsmerge.Merge(a, b), ".")
	require.NoError(t, err)
	require.Len(t, entries, 1)

	content, err := fs.ReadFile(fsmerge.Merge(a, b), "shared.txt")
	require.NoError(t, err)
	require.Equal(t, "first", string(content))
}

func TestMerge_ResolvesFilesFromLaterSources(t *testing.T) {
	t.Parallel()

	a := fstest.MapFS{"000001_a.up.sql": &fstest.MapFile{Data: []byte("a")}}
	b := fstest.MapFS{"000002_b.up.sql": &fstest.MapFile{Data: []byte("b")}}

	content, err := fs.ReadFile(fsmerge.Merge(a, b), "000002_b.up.sql")
	require.NoError(t, err)
	require.Equal(t, "b", string(content))
}

func TestMerge_MissingPathIsErrNotExist(t *testing.T) {
	t.Parallel()

	a := fstest.MapFS{"000001_a.up.sql": &fstest.MapFile{Data: []byte("a")}}

	_, err := fsmerge.Merge(a).Open("missing.sql")
	require.ErrorIs(t, err, fs.ErrNotExist)

	_, err = fs.ReadDir(fsmerge.Merge(), ".")
	require.ErrorIs(t, err, fs.ErrNotExist)
}

func TestMerge_OpenRootReturnsDirFile(t *testing.T) {
	t.Parallel()

	a := fstest.MapFS{"000001_a.up.sql": &fstest.MapFile{Data: []byte("a")}}
	b := fstest.MapFS{"000002_b.up.sql": &fstest.MapFile{Data: []byte("b")}}

	f, err := fsmerge.Merge(a, b).Open(".")
	require.NoError(t, err)
	require.NoError(t, f.Close())

	dir, ok := f.(fs.ReadDirFile)
	require.True(t, ok)
	entries, err := dir.ReadDir(-1)
	require.NoError(t, err)
	require.Len(t, entries, 2)
}
