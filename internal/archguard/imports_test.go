//go:build unit

package archguard

import (
	"strings"
	"testing"
)

// importRules constrain what a package may import at all, independently of the
// layer-to-layer direction: the published contract stays a leaf, features never
// reach sideways into other features, shared and feature-agnostic infrastructure
// stay out of the feature tree, and nothing under internal/ reaches the
// composition root.
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
		name: "features-do-not-import-other-features",
		why:  "features are self-contained bounded contexts; one feature must never reach into another feature's packages",
		applies: func(rel string) bool {
			segs := segments(rel)
			return len(segs) >= 2 && segs[0] == "features"
		},
		denied: func(rel, imp string) bool {
			target := internalRel(imp)
			tsegs := segments(target)
			return len(tsegs) >= 2 && tsegs[0] == "features" && tsegs[1] != feature(rel)
		},
	},
	{
		name: "infrastructure-is-feature-agnostic",
		why:  "cross-cutting infrastructure (config, valkey, telemetry, lifecycle, transport, the postgres client) stays feature-agnostic; only the postgres/<f> driven adapters may import a feature",
		applies: func(rel string) bool {
			segs := segments(rel)
			if len(segs) < 2 || segs[0] != "infrastructure" {
				return false
			}
			return !(len(segs) >= 3 && segs[1] == "postgres")
		},
		denied: func(_, imp string) bool {
			return strings.HasPrefix(internalRel(imp), "features/")
		},
	},
	{
		name: "shared-ignores-features",
		why:  "the shared error catalog is cross-feature; features depend on it, never the reverse",
		applies: func(rel string) bool {
			return strings.HasPrefix(rel, "shared/")
		},
		denied: func(_, imp string) bool {
			return strings.HasPrefix(internalRel(imp), "features/")
		},
	},
	{
		name: "composition-root-is-top",
		why:  "internal/app is the composition root and may only be imported by cmd/; nothing under internal/ reaches it",
		applies: func(rel string) bool {
			return !strings.HasPrefix(rel, "api/")
		},
		denied: func(_, imp string) bool {
			return internalRel(imp) == "app"
		},
	},
}

// TestImports fails when a package imports something it may not import.
func TestImports(t *testing.T) {
	t.Parallel()

	checkRules(t, importRules)
}
