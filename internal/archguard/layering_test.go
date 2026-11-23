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
		name: "usecase-depends-on-domain-port-contract",
		why:  "use cases orchestrate their own feature's domain, port interface, and the published wire contract (pkg/api/v1), nothing else",
		applies: func(rel string) bool {
			segs := segments(rel)
			return len(segs) == 3 && segs[0] == "features" && segs[2] == "usecase"
		},
		denied: func(rel, imp string) bool {
			if isStdlib(imp) {
				return false
			}
			if imp == modulePath+"/pkg/api/v1" {
				return false
			}
			if isThirdParty(imp) {
				return true
			}
			f := feature(rel)
			allowed := map[string]bool{
				"features/" + f + "/domain": true,
				"features/" + f + "/port":   true,
			}
			return !allowed[internalRel(imp)]
		},
	},
	{
		name: "driven-adapter-ignores-usecase-and-handler",
		why:  "the postgres driven adapter satisfies the port interface structurally and must not know about any feature's use-case or handler layers",
		applies: func(rel string) bool {
			segs := segments(rel)
			return len(segs) >= 3 && segs[0] == "infrastructure" && segs[1] == "postgres"
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
		name: "handler-ignores-port",
		why:  "the driving handler talks to the use-case service only, never to the outbound port interface or any infrastructure package",
		applies: func(rel string) bool {
			segs := segments(rel)
			return len(segs) == 3 && segs[0] == "features" && segs[2] == "handler"
		},
		denied: func(rel, imp string) bool {
			target := internalRel(imp)
			if target == "" {
				return false
			}
			if strings.HasPrefix(target, "infrastructure/") {
				return true
			}
			segs := segments(target)
			return len(segs) >= 3 && segs[0] == "features" && segs[1] == feature(rel) && segs[2] == "port"
		},
	},
	{
		name: "features-registry-imports-only-own-features",
		why:  "the feature registry enumerates features; it must not depend on anything outside the features tree or the embedded postgres migrations (stdlib and samber/do excepted)",
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
			return !strings.HasPrefix(rel, "features/") && !strings.HasPrefix(rel, "infrastructure/postgres/")
		},
	},
	{
		name: "infrastructure-ignores-features",
		why:  "cross-cutting infrastructure must stay feature-agnostic; only the postgres feature adapters may import their own feature's domain and port",
		applies: func(rel string) bool {
			return strings.HasPrefix(rel, "infrastructure/")
		},
		denied: func(rel, imp string) bool {
			target := internalRel(imp)
			if !strings.HasPrefix(target, "features/") {
				return false
			}
			segs := segments(rel)
			if len(segs) >= 3 && segs[0] == "infrastructure" && segs[1] == "postgres" {
				f := segs[2]
				tsegs := segments(target)
				return !(len(tsegs) == 3 && tsegs[1] == f && (tsegs[2] == "domain" || tsegs[2] == "port"))
			}
			return true
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
