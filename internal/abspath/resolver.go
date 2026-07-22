// Package abspath resolves references found in generated HTML to paths in the
// build output.
package abspath

import (
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
)

// Resolver resolves a reference to a path in the build output and reports
// whether the target exists.
//
// The filesystem, the build output root and the HTML page being validated are
// fixed for a Resolver instance and captured at construction; only the
// reference varies per call. Resolution matches how a browser reads relative
// versus absolute paths: an absolute reference ("/assets/x") is resolved from
// the build root, a relative one ("../x") from the directory of the page.
type Resolver interface {
	// Resolve returns the cleaned filesystem path for ref. Any URL fragment
	// ("#...") is ignored.
	Resolve(ref string) string
	// Exists reports whether the resolved target for ref exists. A same-page
	// reference (a bare fragment) reports true. It returns an error only for
	// an unexpected filesystem failure, not for a plain "not found".
	Exists(ref string) (ok bool, err error)
}

type resolver struct {
	fs       afero.Fs
	buildDir string
	htmlPath string
}

// NewResolver returns a Resolver backed by fs that resolves relative
// references from the directory of htmlPath and absolute ones from buildDir.
func NewResolver(fs afero.Fs, buildDir, htmlPath string) Resolver {
	return &resolver{fs: fs, buildDir: buildDir, htmlPath: htmlPath}
}

// ResolverFactory builds a Resolver for the page at htmlPath, reusing the
// filesystem and build root it was constructed with.
type ResolverFactory func(htmlPath string) Resolver

// NewResolverFactory returns a ResolverFactory that builds resolvers backed by
// fs and buildDir.
func NewResolverFactory(fs afero.Fs, buildDir string) ResolverFactory {
	return func(htmlPath string) Resolver {
		return NewResolver(fs, buildDir, htmlPath)
	}
}

func (r *resolver) Resolve(ref string) string {
	ref = stripFragment(ref)

	var resolved string
	if strings.HasPrefix(ref, "/") {
		resolved = filepath.Join(r.buildDir, ref)
	} else {
		resolved = filepath.Join(filepath.Dir(r.htmlPath), ref)
	}
	return filepath.Clean(resolved)
}

func (r *resolver) Exists(ref string) (bool, error) {
	if stripFragment(ref) == "" {
		return true, nil // same-page reference
	}
	return afero.Exists(r.fs, r.Resolve(ref))
}

func stripFragment(ref string) string {
	if before, _, ok := strings.Cut(ref, "#"); ok {
		return before
	}
	return ref
}
