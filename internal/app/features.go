package app

import (
	"fmt"

	"github.com/samber/do/v2"
)

// feature describes one feature's wiring contribution: its name and the
// function that registers its providers, sentinel mappings, and routes.
type feature struct {
	Name     string
	Register func(do.Injector) error
}

// features is the ordered list of features this composition root wires. Add
// one entry per feature; the list order is the registration order.
var features = []feature{
	{Name: "catalog", Register: registerCatalog},
	{Name: "machines", Register: registerMachines},
	{Name: "sales", Register: registerSales},
	{Name: "reporting", Register: registerReporting},
}

// registerFeatures wires every feature in list order into the injector.
func registerFeatures(c do.Injector) error {
	for _, f := range features {
		if err := f.Register(c); err != nil {
			return fmt.Errorf("register %s feature: %w", f.Name, err)
		}
	}
	return nil
}
