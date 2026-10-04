// Package features is the single feature registry: it lists every feature and
// exposes the two enumeration points the composition root and the migration
// runner need, so adding or deleting a feature touches exactly one entry here
// instead of each call site.
package features

import (
	"fmt"
	"io/fs"

	"github.com/samber/do/v2"

	catalogdi "github.com/zercle/zercle-go-template/internal/features/catalog/di"
	catalogmigrations "github.com/zercle/zercle-go-template/internal/features/catalog/repository/postgres/migrations"
	machinesdi "github.com/zercle/zercle-go-template/internal/features/machines/di"
	machinesmigrations "github.com/zercle/zercle-go-template/internal/features/machines/repository/postgres/migrations"
	salesdi "github.com/zercle/zercle-go-template/internal/features/sales/di"
	salesmigrations "github.com/zercle/zercle-go-template/internal/features/sales/repository/postgres/migrations"
)

// Feature describes one feature's wiring and persistence contributions.
type Feature struct {
	Name       string
	Register   func(do.Injector) error
	Migrations fs.FS
}

// List is the ordered registry of features. The order is also the migration
// order: catalog owns schema version 1, machines 2, and sales 3. Add one entry
// per feature; leave Migrations nil when the feature owns no schema.
var List = []Feature{
	{
		Name:       "catalog",
		Register:   catalogdi.Register,
		Migrations: catalogmigrations.FS,
	},
	{
		Name:       "machines",
		Register:   machinesdi.Register,
		Migrations: machinesmigrations.FS,
	},
	{
		Name:       "sales",
		Register:   salesdi.Register,
		Migrations: salesmigrations.FS,
	},
}

// RegisterAll wires every registered feature into the injector in list order.
func RegisterAll(c do.Injector) error {
	for _, f := range List {
		if err := f.Register(c); err != nil {
			return fmt.Errorf("register %s feature: %w", f.Name, err)
		}
	}
	return nil
}

// MigrationSources returns the embedded migrations of every registered feature
// that owns schema, in list order.
func MigrationSources() []fs.FS {
	sources := make([]fs.FS, 0, len(List))
	for _, f := range List {
		if f.Migrations != nil {
			sources = append(sources, f.Migrations)
		}
	}
	return sources
}
