//go:build unit

package archguard

import (
	"strings"
	"testing"
)

// importRules constrain what a package may import at all, independently of
// the layer-to-layer direction: the wire contract stays a leaf, the outbound
// port references only its own domain, and the published facade stays
// outward-only.
var importRules = []rule{
	{
		name: "published-contract-is-outward-only",
		why:  "internal code must not import the pkg/api/v1 facade; depend on the feature contract packages directly",
		applies: func(string) bool {
			return true
		},
		denied: func(_, imp string) bool {
			return imp == pkgAPIV1 || strings.HasPrefix(imp, pkgAPIV1+"/")
		},
	},
	{
		name: "contract-is-leaf",
		why:  "the wire contract must stay dependency-free so the published facade drags in nothing",
		applies: func(rel string) bool {
			segs := segments(rel)
			return len(segs) == 3 && segs[0] == "features" && segs[2] == "contract"
		},
		denied: func(_, imp string) bool {
			return !isStdlib(imp)
		},
	},
	{
		name: "repository-interface-depends-only-on-domain",
		why:  "the outbound repository interface may reference only its own feature's domain",
		applies: func(rel string) bool {
			segs := segments(rel)
			return len(segs) == 3 && segs[0] == "features" && segs[2] == "repository"
		},
		denied: func(rel, imp string) bool {
			if isStdlib(imp) {
				return false
			}
			if isThirdParty(imp) {
				return true
			}
			allowed := "features/" + feature(rel) + "/domain"
			return internalRel(imp) != allowed
		},
	},
}

// TestImports fails when a package imports something it may not import.
func TestImports(t *testing.T) {
	t.Parallel()

	checkRules(t, importRules)
}
