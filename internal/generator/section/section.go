package section

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
)

type FileSystem interface {
	Walk(root string, fn filepath.WalkFunc) error
	ReadFile(path string) ([]byte, error)
}

// Section represents a top-level site section which will be included in the navigation bar of the site
// All markdown files inside the root markdown content directory are site sections
type Section struct {
	DirName     string // directory name (used for URL path construction)
	DisplayName string // display name shown in navigation (from # title in index.md)
}

func ListSections(fs afero.Fs, rootDirectory string) ([]Section, error) {
	sections := make([]Section, 0)
	return sections, afero.Walk(fs, rootDirectory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(rootDirectory, path)
		if err != nil {
			return err
		}

		// Only process top-level directories (and the root itself)
		if strings.Contains(relPath, "/") {
			return nil
		}

		// Root content directory: home section has an empty DirName
		if relPath == "." {
			displayName := extractSectionTitle(fs, filepath.Join(path, "index.md"), "home")
			sections = append(sections, Section{
				DirName:     "",
				DisplayName: displayName,
			})
			return nil
		}

		// Sub-section: read display name from # title in section's index.md
		displayName := extractSectionTitle(fs, filepath.Join(path, "index.md"), relPath)
		sections = append(sections, Section{
			DirName:     relPath,
			DisplayName: displayName,
		})
		return nil
	})
}

// extractSectionTitle reads the # title from the index.md of a section directory.
// Falls back to the capitalized dirName if no index.md or no title is found.
func extractSectionTitle(fs afero.Fs, indexMDPath, dirName string) string {
	content, err := afero.ReadFile(fs, indexMDPath)
	if err != nil {
		return strings.ToUpper(dirName[:1]) + dirName[1:]
	}
	for line := range strings.SplitSeq(string(content), "\n") {
		if after, ok := strings.CutPrefix(line, "# "); ok {
			return strings.TrimSpace(after)
		}
	}
	return strings.ToUpper(dirName[:1]) + dirName[1:]
}

// extractSection returns the section (subdirectory path) for a file relative to the content directory.
// For example, if contentDir is "content/markdown" and filePath is "content/markdown/blog/post.md",
// the section returned is "blog". Returns empty string if the file is at the root of contentDir.
func ExtractSection(contentDir, filePath string) (string, error) {
	relPath, err := filepath.Rel(contentDir, filePath)
	if err != nil {
		return "", fmt.Errorf("failed to get relative path for %s: %w", filePath, err)
	}
	section := filepath.Dir(relPath)
	if section == "." {
		return "", nil
	}
	return section, nil
}
