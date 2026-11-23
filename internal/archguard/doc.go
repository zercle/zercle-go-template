// Package archguard holds the executable dependency gates for the
// clean-architecture layering of the final layout. The gates live in this
// package's *_test.go files: layering_test.go constrains the dependency
// direction inside a feature and toward the postgres driven adapter, and
// imports_test.go constrains what a package may import at all.
//
// The rules mirror the dependency direction documented in README.md: a
// feature's domain depends on nothing, its use case depends only on its own
// domain and port, its port only on its own domain, its handler never reaches
// its own port or any infrastructure, and the postgres driven adapter never
// knows about a use-case or handler. Cross-cutting infrastructure and the
// shared error catalog stay feature-agnostic, features never reach sideways
// into other features, the published contract pkg/api is a leaf, and only cmd/
// may import the composition root internal/app.
package archguard
