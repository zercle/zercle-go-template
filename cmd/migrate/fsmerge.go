// Migrations source composition: each feature owns its embedded migrations,
// and the runner merges every feature's FS into one namespace (internal/fsmerge)
// so a single golang-migrate source driver sees them all.
package main

import (
	"io/fs"

	"github.com/zercle/zercle-go-template/internal/features"
)

// migrationSources returns every registered feature's embedded migrations in
// registry order. Add a feature to the registry (internal/features/features.go)
// and its schema travels with it; deleting the entry removes its migrations.
func migrationSources() []fs.FS {
	return features.MigrationSources()
}
