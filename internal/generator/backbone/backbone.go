package backbone

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
	"github.com/tjnvr/blog/internal/generator/backbone/path"
	"github.com/tjnvr/blog/internal/generator/backbone/section"
)

type (
	backBone struct {
		contentDir string
		assetsDir  string
		buildDir   string
		scriptsDir string

		fs                  afero.Fs
		pageSectionResolver PageSectionResolver
		assetsPathResolver  PathResolver
		linksPathResolver   PathResolver
		pagesPathResolver   PathResolver
	}

	PageSectionResolver interface {
		Resolve(sourcePageFilePath string) (section.Section, error)
	}

	PathResolver interface {
		Resolve(oldPath, fromPath string) (string, error)
	}
)

func NewBackBone(fs afero.Fs, contentDir string, assetsDir string, assetsOutDir string, buildDir string, scriptsDir string, scriptsOutDir string) backBone {
	return backBone{
		contentDir:          contentDir,
		assetsDir:           assetsDir,
		buildDir:            buildDir,
		scriptsDir:          scriptsDir,
		fs:                  fs,
		pageSectionResolver: section.NewDefaultPageSectionResolver(fs, contentDir),
		assetsPathResolver:  path.NewDefaultResolver(assetsDir, assetsOutDir),
		linksPathResolver:   path.NewDefaultResolver(scriptsDir, scriptsOutDir),
		pagesPathResolver:   path.NewDefaultResolver(contentDir, buildDir),
	}
}

func (b backBone) GetHTMLPageFilePathDestination(sourceMarkdownPageFilePath string) (string, error) {
	pageFilePathRelToContentDir, err := filepath.Rel(b.contentDir, sourceMarkdownPageFilePath)
	if err != nil {
		return "", fmt.Errorf("cannot compute relative path of %s from %s: %w", sourceMarkdownPageFilePath, b.contentDir, err)
	}

	return filepath.Join(b.buildDir, strings.Replace(pageFilePathRelToContentDir, ".md", ".html", 1)), nil
}

func (b backBone) GetPageSection(sourcePageFilePath string) (section.Section, error) {
	return b.pageSectionResolver.Resolve(sourcePageFilePath)
}

func (b backBone) GetTargetAbsoluteFilePath(relativeTargetFilePath string, fromFilePath string) string {
	return ""
}

func (b backBone) ResolveNewAssetPath(oldPath, fromPath string) (string, error) {
	return b.assetsPathResolver.Resolve(oldPath, fromPath)
}

func (b backBone) ResolveNewMarkdownPath(oldPath, fromPath string) (string, error) {
	return b.linksPathResolver.Resolve(oldPath, fromPath)
}

func (b backBone) ResolveNewScriptPath(oldPath, fromPath string) (string, error) {
	return "", nil
}
