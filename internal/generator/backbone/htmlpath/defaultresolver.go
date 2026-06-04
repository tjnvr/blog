package htmlpath

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
)

type defaultResolver struct {
	fs         afero.Fs
	contentDir string
	buildDir   string
}

func NewDefaultResolver(fs afero.Fs, contentDir, buildDir string) defaultResolver {
	return defaultResolver{
		fs:         fs,
		contentDir: contentDir,
		buildDir:   buildDir,
	}
}
func (r defaultResolver) Resolve(path string) (string, error) {
	pageFilePathRelToContentDir, err := filepath.Rel(r.contentDir, path)
	if err != nil {
		return "", fmt.Errorf("cannot compute relative path of %s from %s: %w", path, r.contentDir, err)
	}

	return filepath.Join(r.buildDir, strings.Replace(pageFilePathRelToContentDir, ".md", ".html", 1)), nil
}
