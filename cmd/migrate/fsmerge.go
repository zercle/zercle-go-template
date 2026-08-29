// Migrations source composition: each feature owns its embedded migrations,
// and the runner merges every feature's FS into one namespace so a single
// golang-migrate source driver sees them all.
package main

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path"
	"time"

	examplemigrations "github.com/zercle/zercle-go-template/internal/features/example/adapter/out/postgres/migrations"
)

// migrationSources lists every feature's embedded migrations in evaluation
// order. Add a feature's FS here when it introduces persistence; deleting a
// feature removes its entry (and its schema files) with it.
func migrationSources() []fs.FS {
	return []fs.FS{
		examplemigrations.FS,
	}
}

// mergedFS concatenates the roots of several fs.FS instances into one
// namespace. Files resolve to the first source that contains them; directory
// listings merge and de-duplicate by name. It exists because golang-migrate
// takes a single source, while migrations are owned per feature.
type mergedFS []fs.FS

// Open resolves name against each source in order. Files delegate to the
// owning source's real file handle; directories synthesize a ReadDirFile over
// the merged listing — delegating a directory to one source would leak that
// source's unmerged listing.
func (m mergedFS) Open(name string) (fs.File, error) {
	if !fs.ValidPath(name) {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrInvalid}
	}
	for _, source := range m {
		f, err := source.Open(name)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				continue
			}
			return nil, fmt.Errorf("open %q in migration source: %w", name, err)
		}
		st, statErr := f.Stat()
		if statErr != nil {
			_ = f.Close()
			return nil, fmt.Errorf("stat %q in migration source: %w", name, statErr)
		}
		if !st.IsDir() {
			return f, nil
		}
		_ = f.Close()
	}

	entries, err := m.ReadDir(name)
	if err != nil {
		return nil, err
	}
	return &mergedDir{name: name, entries: entries}, nil
}

// ReadDir merges the directory listings of name across all sources. Duplicate
// entry names resolve to the first source that lists them.
func (m mergedFS) ReadDir(name string) ([]fs.DirEntry, error) {
	if !fs.ValidPath(name) {
		return nil, &fs.PathError{Op: "readdir", Path: name, Err: fs.ErrInvalid}
	}
	var (
		entries []fs.DirEntry
		seen    = make(map[string]bool)
		found   bool
	)
	for _, source := range m {
		dirEntries, err := fs.ReadDir(source, name)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				continue
			}
			return nil, fmt.Errorf("readdir %q in migration source: %w", name, err)
		}
		found = true
		for _, e := range dirEntries {
			if seen[e.Name()] {
				continue
			}
			seen[e.Name()] = true
			entries = append(entries, e)
		}
	}
	if !found {
		return nil, &fs.PathError{Op: "readdir", Path: name, Err: fs.ErrNotExist}
	}
	return entries, nil
}

// mergedDir is a synthesized directory file over a pre-merged listing.
type mergedDir struct {
	name    string
	entries []fs.DirEntry
	offset  int
}

func (d *mergedDir) Read([]byte) (int, error) {
	return 0, &fs.PathError{Op: "read", Path: d.name, Err: fs.ErrInvalid}
}

func (d *mergedDir) Close() error { return nil }

func (d *mergedDir) Stat() (fs.FileInfo, error) {
	return mergedDirInfo{name: d.name}, nil
}

func (d *mergedDir) ReadDir(n int) ([]fs.DirEntry, error) {
	remaining := len(d.entries) - d.offset
	if n <= 0 {
		out := d.entries[d.offset:]
		d.offset = len(d.entries)
		return out, nil
	}
	if remaining == 0 {
		return nil, io.EOF
	}
	take := min(n, remaining)
	out := d.entries[d.offset : d.offset+take]
	d.offset += take
	return out, nil
}

type mergedDirInfo struct {
	name string
}

func (i mergedDirInfo) Name() string { return path.Base(i.name) }

func (i mergedDirInfo) Size() int64 { return 0 }

func (i mergedDirInfo) Mode() fs.FileMode { return fs.ModeDir | 0o555 }

func (i mergedDirInfo) ModTime() time.Time { return time.Time{} }

func (i mergedDirInfo) IsDir() bool { return true }

func (i mergedDirInfo) Sys() any { return nil }
