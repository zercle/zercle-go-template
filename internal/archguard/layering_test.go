//go:build unit

package archguard

import (
	"strings"
	"testing"
)

// layeringRules constrain the dependency direction between the layers: the
// domain is the sink, the use case reaches only its own feature's domain,
// repository port, and contract, and the layers never reach sideways.
var layeringRules = []rule{
	{
		name: "domain-is-innermost",
		why:  "domain may depend on nothing but the standard library",
		applies: func(rel string) bool {
			segs := segments(rel)
			return len(segs) == 3 && segs[0] == "features" && segs[2] == "domain"
		},
		denied: func(_, imp string) bool {
			return isThirdParty(imp) || internalRel(imp) != ""
		},
	},
	{
		name: "usecase-depends-on-domain-repository-contract",
		why:  "use cases orchestrate their own feature's domain, repository interface, and wire contract, nothing else",
		applies: func(rel string) bool {
			segs := segments(rel)
			return len(segs) == 3 && segs[0] == "features" && segs[2] == "usecase"
		},
		denied: func(rel, imp string) bool {
			if isStdlib(imp) {
				return false
			}
			if isThirdParty(imp) {
				return true
			}
			f := feature(rel)
			allowed := map[string]bool{
				"features/" + f + "/domain":     true,
				"features/" + f + "/repository": true,
				"features/" + f + "/contract":   true,
			}
			return !allowed[internalRel(imp)]
		},
	},
	{
		name: "repository-impl-ignores-usecase-and-handler",
		why:  "the repository implementation satisfies the repository interface structurally and must not know about the use-case or handler layers",
		applies: func(rel string) bool {
			segs := segments(rel)
			return len(segs) >= 4 && segs[0] == "features" && segs[2] == "repository"
		},
		denied: func(_, imp string) bool {
			rel := internalRel(imp)
			if rel == "" {
				return false
			}
			segs := segments(rel)
			return len(segs) >= 3 && segs[0] == "features" && (segs[2] == "usecase" || segs[2] == "handler")
		},
	},
	{
		name: "handler-ignores-repository",
		why:  "the driving handler talks to the use-case service only, never to the outbound repository interface or its implementation",
		applies: func(rel string) bool {
			segs := segments(rel)
			return len(segs) == 3 && segs[0] == "features" && segs[2] == "handler"
		},
		denied: func(_, imp string) bool {
			rel := internalRel(imp)
			if rel == "" {
				return false
			}
			segs := segments(rel)
			return len(segs) >= 3 && segs[0] == "features" && segs[2] == "repository"
		},
	},
	{
		name: "features-registry-imports-only-own-features",
		why:  "the feature registry enumerates features; it must not depend on infrastructure or anything outside the features tree (samber/do excepted)",
		applies: func(rel string) bool {
			return rel == "features"
		},
		denied: func(_, imp string) bool {
			if isStdlib(imp) || imp == "github.com/samber/do/v2" {
				return false
			}
			rel := internalRel(imp)
			if rel == "" {
				return true
			}
			return !strings.HasPrefix(rel, "features/")
		},
	},
	{
		name: "infrastructure-ignores-features",
		why:  "cross-cutting infrastructure must stay feature-agnostic; features depend on infrastructure, never the reverse",
		applies: func(rel string) bool {
			return strings.HasPrefix(rel, "infrastructure/")
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
}

// TestLayering fails when a package reaches a layer it may not depend on.
func TestLayering(t *testing.T) {
	t.Parallel()

	checkRules(t, layeringRules)
}
