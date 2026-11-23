// Package features is the single feature registry: it lists every feature and
// exposes the enumeration point the composition root needs, so adding or
// deleting a feature touches exactly one entry here instead of each call site.
package features

import (
	"fmt"

	"github.com/samber/do/v2"

	catalogdi "github.com/zercle/zercle-go-template/internal/features/catalog/di"
	machinesdi "github.com/zercle/zercle-go-template/internal/features/machines/di"
	reportingdi "github.com/zercle/zercle-go-template/internal/features/reporting/di"
	salesdi "github.com/zercle/zercle-go-template/internal/features/sales/di"
)

// Feature describes one feature's wiring contribution.
type Feature struct {
	Name     string
	Register func(do.Injector) error
}

// List is the ordered registry of features. Add one entry per feature.
var List = []Feature{
	{Name: "catalog", Register: catalogdi.Register},
	{Name: "machines", Register: machinesdi.Register},
	{Name: "sales", Register: salesdi.Register},
	{Name: "reporting", Register: reportingdi.Register},
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
