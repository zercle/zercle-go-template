//go:build unit

package archguard

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

// internalRoot is the internal/ directory the gates scan.
func internalRoot(t *testing.T) string {
	t.Helper()

	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot locate test source file")
	}
	return filepath.Dir(filepath.Dir(thisFile))
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

// rule is one executable gate: it governs the packages matched by applies and
// fails on the imports matched by denied.
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

// checkRules walks the internal tree and reports every rule violation.
func checkRules(t *testing.T, rules []rule) {
	t.Helper()

	root := internalRoot(t)

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
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

		rel := filepath.ToSlash(strings.TrimPrefix(filepath.Dir(path), root+string(filepath.Separator)))
		if rel == filepath.Dir(path) || filepath.Dir(path) == root {
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
