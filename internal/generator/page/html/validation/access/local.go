package access

import (
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
)

// LocalResolver resolves a reference to a path in the build output and reports
// whether the target exists.
//
// The filesystem, the build output root and the HTML page being validated are
// fixed for a validator instance and captured at construction; only the
// reference varies per call. Resolution matches how a browser reads relative
// versus absolute paths: an absolute reference ("/assets/x") is resolved from
// the build root, a relative one ("../x") from the directory of the page.
type LocalResolver interface {
	// Resolve returns the cleaned filesystem path for ref. Any URL fragment
	// ("#...") is ignored.
	Resolve(ref string) string
	// Exists reports whether the resolved target for ref exists. A same-page
	// reference (a bare fragment) reports true. It returns an error only for
	// an unexpected filesystem failure, not for a plain "not found".
	Exists(ref string) (ok bool, err error)
}

type localResolver struct {
	fs       afero.Fs
	buildDir string
	htmlPath string
}

// NewResolver returns a LocalResolver backed by fs that resolves relative
// references from the directory of htmlPath and absolute ones from buildDir.
func NewResolver(fs afero.Fs, buildDir, htmlPath string) LocalResolver {
	return &localResolver{fs: fs, buildDir: buildDir, htmlPath: htmlPath}
}

func (r *localResolver) Resolve(ref string) string {
	ref = stripFragment(ref)

	var resolved string
	if strings.HasPrefix(ref, "/") {
		resolved = filepath.Join(r.buildDir, ref)
	} else {
		resolved = filepath.Join(filepath.Dir(r.htmlPath), ref)
	}
	return filepath.Clean(resolved)
}

func (r *localResolver) Exists(ref string) (bool, error) {
	if stripFragment(ref) == "" {
		return true, nil // same-page reference
	}
	return afero.Exists(r.fs, r.Resolve(ref))
}

func stripFragment(ref string) string {
	if i := strings.IndexByte(ref, '#'); i >= 0 {
		return ref[:i]
	}
	return ref
}
