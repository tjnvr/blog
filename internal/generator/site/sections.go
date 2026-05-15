package site

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tjnvr/blog/internal/generator/section"
)

func (g *Generator) listSections() error {
	return g.fs.Walk(g.contentDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(g.contentDir, path)
		if err != nil {
			return err
		}

		// Only process top-level directories (and the root itself)
		if strings.Contains(relPath, "/") {
			return nil
		}

		// Root content directory: home section has an empty DirName
		if relPath == "." {
			displayName := g.extractSectionTitle(filepath.Join(path, "index.md"), "home")
			g.sections = append(g.sections, section.Section{
				DirName:     "",
				DisplayName: displayName,
			})
			return nil
		}

		// Sub-section: read display name from # title in section's index.md
		displayName := g.extractSectionTitle(filepath.Join(path, "index.md"), relPath)
		g.sections = append(g.sections, section.Section{
			DirName:     relPath,
			DisplayName: displayName,
		})
		return nil
	})
}

// extractSectionTitle reads the # title from the index.md of a section directory.
// Falls back to the capitalized dirName if no index.md or no title is found.
func (g *Generator) extractSectionTitle(indexMDPath, dirName string) string {
	content, err := g.fs.ReadFile(indexMDPath)
	if err != nil {
		return strings.ToUpper(dirName[:1]) + dirName[1:]
	}
	for _, line := range strings.Split(string(content), "\n") {
		if strings.HasPrefix(line, "# ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "# "))
		}
	}
	return strings.ToUpper(dirName[:1]) + dirName[1:]
}

// extractSection returns the section (subdirectory path) for a file relative to the content directory.
// For example, if contentDir is "content/markdown" and filePath is "content/markdown/blog/post.md",
// the section returned is "blog". Returns empty string if the file is at the root of contentDir.
func extractSection(contentDir, filePath string) (string, error) {
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
