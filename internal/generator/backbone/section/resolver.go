package section

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
	mfs "github.com/tjnvr/blog/internal/io/fs"
)

type (
	sectionResolver struct {
		fs               afero.Fs
		filesFinder      mfs.FilesFinder
		contentDirectory string
	}

	// Resolver discovers site sections within a content directory.
	Resolver interface {
		// Resolve returns one Section for every index.md found under the
		// content directory, including the root section.
		Resolve() ([]Section, error)
		// ResolveForFile returns the Section that owns the Markdown file at
		// markdownPath, determined by its first-level directory under the
		// content directory.
		ResolveForFile(markdownPath string) (Section, error)
	}
)

// NewResolver returns a Resolver that reads sections from contentDirectory
// using fs. Section index files are located by matching the name "index.md".
func NewResolver(fs afero.Fs, contentDirectory string) Resolver {
	return &sectionResolver{
		fs:               fs,
		filesFinder:      mfs.NewFilesFinder(fs, mfs.WithPattern("index.md")),
		contentDirectory: contentDirectory,
	}
}

func (r *sectionResolver) Resolve() ([]Section, error) {
	sections := make([]Section, 0)
	homeSectionFiles, err := r.filesFinder.FindFiles(r.contentDirectory)
	if err != nil {
		return nil, fmt.Errorf("FindFiles err: %v", err)
	}
	for _, relHomeSectionFile := range homeSectionFiles {
		homeSectionFile := filepath.Join(r.contentDirectory, relHomeSectionFile)
		title, err := extractSectionTitle(r.fs, homeSectionFile)
		if err != nil {
			return nil, err
		}
		sections = append(sections, Section{HomePath: homeSectionFile, DisplayName: title})
	}

	return sections, nil
}

func (r *sectionResolver) ResolveForFile(markdownPath string) (Section, error) {
	markdownPathRelativeToContentDir, err := filepath.Rel(r.contentDirectory, markdownPath)
	if err != nil {
		return Section{}, fmt.Errorf("filepath.Rel(%s, %s) err: %v", r.contentDirectory, markdownPath, err)
	}

	directories := strings.SplitN(markdownPathRelativeToContentDir, string(filepath.Separator), 2)
	if len(directories) < 1 {
		return Section{}, fmt.Errorf("no section directory found for '%s'", markdownPathRelativeToContentDir)
	}

	// A file directly under contentDirectory (e.g. the root "index.md") belongs to the root section.
	sectionDirectory := ""
	if len(directories) == 2 {
		sectionDirectory = directories[0]
	}

	homeSectionFile := filepath.Join(r.contentDirectory, sectionDirectory, "index.md")
	sectionTitle, err := extractSectionTitle(r.fs, homeSectionFile)
	if err != nil {
		return Section{}, err
	}

	return Section{HomePath: homeSectionFile, DisplayName: sectionTitle}, nil
}

func extractSectionTitle(fs afero.Fs, homeSectionFile string) (string, error) {
	content, err := afero.ReadFile(fs, homeSectionFile)
	if err != nil {
		return "", fmt.Errorf("afero.ReadFile err:%v", err)
	}

	for line := range strings.SplitSeq(string(content), "\n") {
		if after, ok := strings.CutPrefix(line, "# "); ok {
			return strings.TrimSpace(after), nil
		}
	}

	return "", fmt.Errorf("no section title found")
}
