package section

import (
	"cmp"
	"fmt"
	"math"
	"path/filepath"
	"slices"
	"strings"

	"github.com/spf13/afero"
	"github.com/tjnvr/blog/internal/backbone/metadata"
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
		// content directory, including the root section, ordered by ascending
		// Seq. Unsequenced sections come last, ordered by display name.
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
		s, err := extractSection(r.fs, filepath.Join(r.contentDirectory, relHomeSectionFile))
		if err != nil {
			return nil, err
		}
		sections = append(sections, s)
	}

	slices.SortFunc(sections, compareSections)

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

	return extractSection(r.fs, filepath.Join(r.contentDirectory, sectionDirectory, "index.md"))
}

// extractSection builds the Section described by homeSectionFile. Both Resolve
// and ResolveForFile go through it so that every field, Seq included, is filled
// the same way: callers compare the two results for equality.
func extractSection(fs afero.Fs, homeSectionFile string) (Section, error) {
	content, err := afero.ReadFile(fs, homeSectionFile)
	if err != nil {
		return Section{}, fmt.Errorf("afero.ReadFile err:%v", err)
	}

	m := metadata.Extract(content)
	if m.Title == "" {
		return Section{}, fmt.Errorf("no section title found")
	}

	return Section{
		HomePath:    homeSectionFile,
		DisplayName: m.Title,
		Seq:         m.Seq,
	}, nil
}

// compareSections orders sections by ascending Seq, then by display name.
func compareSections(a, b Section) int {
	return cmp.Or(
		cmp.Compare(sequence(a), sequence(b)),
		cmp.Compare(a.DisplayName, b.DisplayName),
	)
}

// sequence reports the sort rank of s. Unsequenced sections rank last.
func sequence(s Section) int {
	if s.Seq == 0 {
		return math.MaxInt
	}
	return s.Seq
}
