//go:build unit

package archguard

import (
	"strings"
	"testing"
)

// layeringRules constrain the dependency direction inside a feature and toward
// the postgres driven adapter: the domain is the sink, the use case reaches
// only its own feature's domain and port, the port only its own domain, the
// handler never the port or any infrastructure, and the adapter never a
// use-case or handler.
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
		name: "usecase-depends-on-domain-and-port",
		why:  "use cases orchestrate only their own feature's domain and port; no wire types, adapters, or third-party",
		applies: func(rel string) bool {
			segs := segments(rel)
			return len(segs) == 3 && segs[0] == "features" && segs[2] == "usecase"
		},
		denied: func(rel, imp string) bool {
			if isStdlib(imp) {
				return false
			}
			allowed := map[string]bool{
				"features/" + feature(rel) + "/domain": true,
				"features/" + feature(rel) + "/port":   true,
			}
			return !allowed[internalRel(imp)]
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
			return internalRel(imp) != "features/"+feature(rel)+"/domain"
		},
	},
	{
		name: "handler-ignores-port-and-adapters",
		why:  "the driving handler talks to the use-case service only; it never reaches its own port interface or any infrastructure package",
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
}

// TestLayering fails when a package reaches a layer it may not depend on.
func TestLayering(t *testing.T) {
	t.Parallel()

	checkRules(t, layeringRules)
}
