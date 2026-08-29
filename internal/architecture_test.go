//go:build unit

// Executable dependency gates for the clean-architecture layering. Each rule
// scans the import statements of every non-test, non-generated Go file under
// internal/ and fails with the violated rule's rationale. The rules mirror
// the dependency direction documented in README.md: driving adapters depend
// on the application port, the use case depends on outbound ports and the
// domain, adapters satisfy ports structurally, and the published contract
// facade pkg/api/v1 is importable only from outside internal/.
package internal

import (
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

const modulePath = "github.com/zercle/zercle-go-template"

const pkgAPIV1 = modulePath + "/pkg/api/v1"

type rule struct {
	name string
	why  string
	// applies reports whether the rule governs the package at rel (the
	// package directory relative to internal/, slash-separated; "." for
	// internal's root).
	applies func(rel string) bool
	// denied reports whether the import violates the rule.
	denied func(rel string, imp string) bool
}

// segments splits a package path into its path segments.
func segments(rel string) []string {
	if rel == "." {
		return nil
	}
	return strings.Split(rel, "/")
}

// feature returns the feature name for a features/<name>/... package, or "".
func feature(rel string) string {
	segs := segments(rel)
	if len(segs) >= 2 && segs[0] == "features" {
		return segs[1]
	}
	return ""
}

// isStdlib reports whether an import path is stdlib (no dot in the first
// segment, so "context" and "net/http" qualify, "github.com/..." does not).
func isStdlib(imp string) bool {
	first, _, _ := strings.Cut(imp, "/")
	return !strings.Contains(first, ".")
}

// internalRel maps a module import path to its path relative to internal/, or
// "" when the import is outside internal/.
func internalRel(imp string) string {
	prefix := modulePath + "/internal/"
	if !strings.HasPrefix(imp, prefix) {
		return ""
	}
	rel := strings.TrimPrefix(imp, prefix)
	if rel == "" {
		return "."
	}
	return rel
}

// isThirdParty reports whether the import is external to this module and not
// stdlib.
func isThirdParty(imp string) bool {
	return !isStdlib(imp) && !strings.HasPrefix(imp, modulePath+"/")
}

// allowedThirdParty lists the external packages the inner layers may use.
const allowedUUID = "github.com/google/uuid"

var rules = []rule{
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
		name: "domain-is-innermost",
		why:  "domain may depend on nothing but the standard library and uuid",
		applies: func(rel string) bool {
			segs := segments(rel)
			return len(segs) == 3 && segs[0] == "features" && segs[2] == "domain"
		},
		denied: func(_, imp string) bool {
			if isStdlib(imp) {
				return false
			}
			if isThirdParty(imp) {
				return imp != allowedUUID
			}
			return internalRel(imp) != ""
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
		name: "port-depends-only-on-domain",
		why:  "outbound ports may reference only their own feature's domain",
		applies: func(rel string) bool {
			segs := segments(rel)
			return len(segs) == 3 && segs[0] == "features" && segs[2] == "port"
		},
		denied: func(rel, imp string) bool {
			if isStdlib(imp) {
				return false
			}
			if isThirdParty(imp) {
				return imp != allowedUUID
			}
			allowed := "features/" + feature(rel) + "/domain"
			return internalRel(imp) != allowed
		},
	},
	{
		name: "application-depends-on-domain-port-contract",
		why:  "use cases orchestrate their own feature's domain, ports, and wire contract, nothing else",
		applies: func(rel string) bool {
			segs := segments(rel)
			return len(segs) == 3 && segs[0] == "features" && segs[2] == "application"
		},
		denied: func(rel, imp string) bool {
			if isStdlib(imp) {
				return false
			}
			if isThirdParty(imp) {
				return imp != allowedUUID
			}
			f := feature(rel)
			allowed := map[string]bool{
				"features/" + f + "/domain":   true,
				"features/" + f + "/port":     true,
				"features/" + f + "/contract": true,
			}
			return !allowed[internalRel(imp)]
		},
	},
	{
		name: "driven-adapters-ignore-application",
		why:  "adapter/out satisfies ports structurally and must not know about the application layer or driving adapters",
		applies: func(rel string) bool {
			return strings.HasPrefix(rel, "features/") && strings.Contains(rel, "/adapter/out")
		},
		denied: func(_, imp string) bool {
			rel := internalRel(imp)
			if rel == "" {
				return false
			}
			segs := segments(rel)
			if len(segs) == 3 && segs[0] == "features" && segs[2] == "application" {
				return true
			}
			return strings.Contains(rel, "/adapter/in")
		},
	},
	{
		name: "driving-adapters-ignore-ports-and-driven-adapters",
		why:  "adapter/in talks to the application port only, never to outbound ports or other adapters",
		applies: func(rel string) bool {
			return strings.HasPrefix(rel, "features/") && strings.Contains(rel, "/adapter/in")
		},
		denied: func(_, imp string) bool {
			rel := internalRel(imp)
			if rel == "" {
				return false
			}
			segs := segments(rel)
			if len(segs) == 3 && segs[0] == "features" && segs[2] == "port" {
				return true
			}
			return strings.Contains(rel, "/adapter/out")
		},
	},
	{
		name: "platform-ignores-features",
		why:  "cross-cutting platform code must stay feature-agnostic; features depend on platform, never the reverse",
		applies: func(rel string) bool {
			return strings.HasPrefix(rel, "platform/")
		},
		denied: func(_, imp string) bool {
			rel := internalRel(imp)
			return strings.HasPrefix(rel, "features/")
		},
	},
}

func TestArchitecture(t *testing.T) {
	t.Parallel()

	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot locate test source file")
	}
	internalRoot := filepath.Dir(thisFile)

	err := filepath.WalkDir(internalRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == "mock" {
				return filepath.SkipDir // generated mocks re-export everything
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		rel := filepath.ToSlash(strings.TrimPrefix(filepath.Dir(path), internalRoot+string(filepath.Separator)))
		if rel == filepath.Dir(path) || filepath.Dir(path) == internalRoot {
			rel = "."
		}

		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		for _, spec := range file.Imports {
			imp := strings.Trim(spec.Path.Value, `"`)
			for _, r := range rules {
				if r.applies(rel) && r.denied(rel, imp) {
					t.Errorf("%s: package %q violates %s: imports %q (%s)",
						path, rel, r.name, imp, r.why)
				}
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk internal tree: %v", err)
	}
}
