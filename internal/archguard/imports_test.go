//go:build unit

package archguard

import (
	"strings"
	"testing"
)

// importRules constrain what a package may import at all, independently of
// the layer-to-layer direction: the published contract stays a leaf and the
// outbound port references only its own domain.
var importRules = []rule{
	{
		name: "published-contract-is-leaf",
		why:  "the published pkg/api contract is the single source of the wire shapes; it may import stdlib and pkg/api/errcodes only, so a consumer never drags the server in",
		applies: func(rel string) bool {
			return strings.HasPrefix(rel, "api/")
		},
		denied: func(_, imp string) bool {
			if isStdlib(imp) {
				return false
			}
			if imp == modulePath+"/pkg/api/errcodes" {
				return false
			}
			return true
		},
	},
	{
		name: "port-depends-only-on-domain",
		why:  "the outbound port interface may reference only its own feature's domain",
		applies: func(rel string) bool {
			segs := segments(rel)
			return len(segs) == 3 && segs[0] == "features" && segs[2] == "port"
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
