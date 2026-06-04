package section

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"github.com/spf13/afero"
	"github.com/tjnvr/blog/internal/generator/backbone/pages"
)

// Section represents a top-level site section.
type (
	Section struct {
		DirName     string // directory name (used for URL path construction)
		DisplayName string // display name shown in navigation (from # title in index.md)
	}

	defaultPageSectionResolver struct {
		fs            afero.Fs
		pagesResolver pages.Resolver
		contentDir    string
		sections      []Section
	}
)

func NewDefaultPageSectionResolver(fs afero.Fs, contentDir string, pagesResolver pages.Resolver) *defaultPageSectionResolver {
	return &defaultPageSectionResolver{fs: fs, contentDir: contentDir, sections: make([]Section, 0)}
}

func (p *defaultPageSectionResolver) ResolveAll() ([]Section, error) {
	pages, err := p.pagesResolver.ResolveAll()
	if err != nil {
		return nil, fmt.Errorf("cannot resolve pages: %v", err)
	}
	for _, page := range pages {
		pageSection, err := p.Resolve(page)
		if err != nil {
			return nil, fmt.Errorf("cannot resolve page section for page (%s) : %v", page, err)
		}

		if !slices.ContainsFunc(p.sections, func(e Section) bool {
			return e == pageSection
		}) {
			p.sections = append(p.sections, pageSection)
		}
	}

	return p.sections, nil
}

func (p *defaultPageSectionResolver) Resolve(pageFilePath string) (Section, error) {
	sectionDirName, err := extractSectionDirName(p.contentDir, pageFilePath)
	if err != nil {
		return Section{}, fmt.Errorf("extractSectionDirName err: %v", err)
	}

	rootSectionPageFilePath := filepath.Join(p.contentDir, sectionDirName, "index.md")
	sectionTitle, err := extractSectionTitle(p.fs, rootSectionPageFilePath)
	if err != nil {
		return Section{}, fmt.Errorf("extractSectionTitle err: %v", err)
	}

	return Section{DirName: sectionDirName, DisplayName: sectionTitle}, nil
}

// extractSectionDirName returns the section (subdirectory path) for a file relative to the content directory.
// For example, if contentDir is "content/markdown" and filePath is "content/markdown/blog/post.md",
// the section returned is "blog". Returns empty string if the file is at the root of contentDir.
func extractSectionDirName(contentDir, pageFilePath string) (string, error) {
	relPath, err := filepath.Rel(contentDir, pageFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to get relative path for %s: %w", pageFilePath, err)
	}
	section := filepath.Dir(relPath)
	if section == "." {
		return "", nil
	}
	return section, nil
}

// extractSectionTitle reads the # title from the index.md of a section directory.
// Falls back to the capitalized dirName if no index.md or no title is found.
func extractSectionTitle(fs afero.Fs, rootSectionPageFilePath string) (string, error) {
	content, err := afero.ReadFile(fs, rootSectionPageFilePath)
	if err != nil {
		return "", fmt.Errorf("ReadFile err for file (%s): %v", rootSectionPageFilePath, err)
	}
	lines := strings.Split(string(content), "\n")
	if len(lines) > 0 && strings.HasPrefix(lines[0], "# ") {
		return strings.TrimSpace(strings.TrimPrefix(lines[0], "# ")), nil
	}

	return "", fmt.Errorf("'#' not found in first line of %s", rootSectionPageFilePath)
}
