// Package archguard holds the executable dependency gates for the
// clean-architecture layering. The gates live in this package's *_test.go
// files: layering_test.go constrains the dependency direction between the
// layers, imports_test.go constrains what a package may import at all.
//
// The rules mirror the dependency direction documented in README.md: the
// handler depends on the use-case service, the use case depends on the
// outbound repository interface and the domain, the repository implementation
// satisfies that interface structurally, the shared error catalog stays
// feature-agnostic, and the published contract facade pkg/api/v1 is importable
// only from outside internal/.
package archguard
