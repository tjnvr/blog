package section

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
)

// Section represents a top-level site section.
type Section struct {
	DirName     string // directory name (used for URL path construction)
	DisplayName string // display name shown in navigation (from # title in index.md)
}

type defaultPageSectionResolver struct {
	fs         afero.Fs
	contentDir string
}

func NewDefaultPageSectionResolver(fs afero.Fs, contentDir string) defaultPageSectionResolver {
	return defaultPageSectionResolver{fs: fs, contentDir: contentDir}
}

func (p defaultPageSectionResolver) Resolve(pageFilePath string) (Section, error) {
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
