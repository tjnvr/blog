package section

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockFilesFinder struct {
	FindFilesFunc func(dir string) ([]string, error)
}

func (m *mockFilesFinder) FindFiles(dir string) ([]string, error) {
	if m.FindFilesFunc != nil {
		return m.FindFilesFunc(dir)
	}
	return nil, nil
}

func TestSectionResolver_Resolve_ShouldReturnOneSectionPerIndexFile(t *testing.T) {
	// given
	contentDirectory := uuid.New().String()
	rootTitle := uuid.New().String()
	firstSection := uuid.New().String()
	firstTitle := uuid.New().String()
	secondSection := uuid.New().String()
	secondTitle := uuid.New().String()

	// setup
	fs := afero.NewMemMapFs()
	require.NoError(t, afero.WriteFile(fs, filepath.Join(contentDirectory, "index.md"), []byte("# "+rootTitle), 0644))
	require.NoError(t, afero.WriteFile(fs, filepath.Join(contentDirectory, firstSection, "index.md"), []byte("# "+firstTitle), 0644))
	require.NoError(t, afero.WriteFile(fs, filepath.Join(contentDirectory, secondSection, "index.md"), []byte("# "+secondTitle), 0644))
	require.NoError(t, afero.WriteFile(fs, filepath.Join(contentDirectory, firstSection, "post.md"), []byte("# not a section"), 0644))
	resolver := NewResolver(fs, contentDirectory)

	// test
	sections, err := resolver.Resolve()

	// expect
	require.NoError(t, err)
	assert.ElementsMatch(t, []Section{
		{HomePath: filepath.Join(contentDirectory, "index.md"), DisplayName: rootTitle},
		{HomePath: filepath.Join(contentDirectory, firstSection, "index.md"), DisplayName: firstTitle},
		{HomePath: filepath.Join(contentDirectory, secondSection, "index.md"), DisplayName: secondTitle},
	}, sections)
}

func TestSectionResolver_Resolve_ShouldReturnEmptySliceWhenNoIndexFile(t *testing.T) {
	// given
	contentDirectory := uuid.New().String()

	// setup
	fs := afero.NewMemMapFs()
	require.NoError(t, fs.MkdirAll(contentDirectory, 0755))
	resolver := NewResolver(fs, contentDirectory)

	// test
	sections, err := resolver.Resolve()

	// expect
	require.NoError(t, err)
	assert.Equal(t, []Section{}, sections)
}

func TestSectionResolver_Resolve_ShouldReturnErrorOnFindFilesFailure(t *testing.T) {
	// given
	contentDirectory := uuid.New().String()

	// setup
	resolver := &sectionResolver{
		fs: afero.NewMemMapFs(),
		filesFinder: &mockFilesFinder{
			FindFilesFunc: func(dir string) ([]string, error) {
				return nil, errors.New("mock finder error")
			},
		},
		contentDirectory: contentDirectory,
	}

	// test
	_, err := resolver.Resolve()

	// expect
	assert.ErrorContains(t, err, "FindFiles err")
}

func TestSectionResolver_Resolve_ShouldReturnErrorWhenIndexFileHasNoTitle(t *testing.T) {
	// given
	contentDirectory := uuid.New().String()

	// setup
	fs := afero.NewMemMapFs()
	require.NoError(t, afero.WriteFile(fs, filepath.Join(contentDirectory, "index.md"), []byte(uuid.New().String()), 0644))
	resolver := NewResolver(fs, contentDirectory)

	// test
	_, err := resolver.Resolve()

	// expect
	assert.ErrorContains(t, err, "no section title found")
}

func TestSectionResolver_Resolve_ShouldSortSectionsBySeq(t *testing.T) {
	// given
	contentDirectory := uuid.New().String()
	rootTitle := uuid.New().String()
	firstSection := uuid.New().String()
	firstTitle := uuid.New().String()
	secondSection := uuid.New().String()
	secondTitle := uuid.New().String()

	// setup
	fs := afero.NewMemMapFs()
	require.NoError(t, afero.WriteFile(fs, filepath.Join(contentDirectory, "index.md"), []byte("<!-- seq: 3 -->\n# "+rootTitle), 0644))
	require.NoError(t, afero.WriteFile(fs, filepath.Join(contentDirectory, firstSection, "index.md"), []byte("<!-- seq: 1 -->\n# "+firstTitle), 0644))
	require.NoError(t, afero.WriteFile(fs, filepath.Join(contentDirectory, secondSection, "index.md"), []byte("<!-- seq: 2 -->\n# "+secondTitle), 0644))
	resolver := NewResolver(fs, contentDirectory)

	// test
	sections, err := resolver.Resolve()

	// expect
	require.NoError(t, err)
	assert.Equal(t, []Section{
		{HomePath: filepath.Join(contentDirectory, firstSection, "index.md"), DisplayName: firstTitle, Seq: 1},
		{HomePath: filepath.Join(contentDirectory, secondSection, "index.md"), DisplayName: secondTitle, Seq: 2},
		{HomePath: filepath.Join(contentDirectory, "index.md"), DisplayName: rootTitle, Seq: 3},
	}, sections)
}

func TestSectionResolver_Resolve_ShouldSortUnsequencedSectionsLastByDisplayName(t *testing.T) {
	// given
	contentDirectory := uuid.New().String()
	sequencedSection := uuid.New().String()
	sequencedTitle := uuid.New().String()
	// Prefixed so the expected display name order does not depend on the random suffix.
	firstUnsequencedTitle := "a-" + uuid.New().String()
	secondUnsequencedTitle := "b-" + uuid.New().String()

	// setup
	fs := afero.NewMemMapFs()
	require.NoError(t, afero.WriteFile(fs, filepath.Join(contentDirectory, "index.md"), []byte("# "+secondUnsequencedTitle), 0644))
	require.NoError(t, afero.WriteFile(fs, filepath.Join(contentDirectory, sequencedSection, "index.md"), []byte("<!-- seq: 9 -->\n# "+sequencedTitle), 0644))
	require.NoError(t, afero.WriteFile(fs, filepath.Join(contentDirectory, uuid.New().String(), "index.md"), []byte("# "+firstUnsequencedTitle), 0644))
	resolver := NewResolver(fs, contentDirectory)

	// test
	sections, err := resolver.Resolve()

	// expect
	require.NoError(t, err)
	displayNames := make([]string, 0, len(sections))
	for _, s := range sections {
		displayNames = append(displayNames, s.DisplayName)
	}
	assert.Equal(t, []string{sequencedTitle, firstUnsequencedTitle, secondUnsequencedTitle}, displayNames)
}

func TestSectionResolver_ResolveForFile(t *testing.T) {
	contentDirectory := uuid.New().String()
	rootTitle := uuid.New().String()
	sectionDirectory := uuid.New().String()
	sectionTitle := uuid.New().String()
	subDirectory := uuid.New().String()

	tests := []struct {
		name         string
		markdownPath string
		expected     Section
	}{
		{
			name:         "file directly under the content directory belongs to the root section",
			markdownPath: filepath.Join(contentDirectory, "post.md"),
			expected:     Section{HomePath: filepath.Join(contentDirectory, "index.md"), DisplayName: rootTitle},
		},
		{
			name:         "file in a section directory belongs to that section",
			markdownPath: filepath.Join(contentDirectory, sectionDirectory, "post.md"),
			expected:     Section{HomePath: filepath.Join(contentDirectory, sectionDirectory, "index.md"), DisplayName: sectionTitle},
		},
		{
			name:         "nested file belongs to its top-level section",
			markdownPath: filepath.Join(contentDirectory, sectionDirectory, subDirectory, "post.md"),
			expected:     Section{HomePath: filepath.Join(contentDirectory, sectionDirectory, "index.md"), DisplayName: sectionTitle},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// setup
			fs := afero.NewMemMapFs()
			require.NoError(t, afero.WriteFile(fs, filepath.Join(contentDirectory, "index.md"), []byte("# "+rootTitle), 0644))
			require.NoError(t, afero.WriteFile(fs, filepath.Join(contentDirectory, sectionDirectory, "index.md"), []byte("# "+sectionTitle), 0644))
			resolver := NewResolver(fs, contentDirectory)

			// test
			section, err := resolver.ResolveForFile(tt.markdownPath)

			// expect
			require.NoError(t, err)
			assert.Equal(t, tt.expected, section)
		})
	}
}

// Callers match the section owning the current file against the resolved section
// list by struct equality, so ResolveForFile must fill every field exactly as
// Resolve does, Seq included.
func TestSectionResolver_ResolveForFile_ShouldReturnSectionEqualToItsResolveEntry(t *testing.T) {
	// given
	contentDirectory := uuid.New().String()
	sectionDirectory := uuid.New().String()
	sectionTitle := uuid.New().String()
	markdownPath := filepath.Join(contentDirectory, sectionDirectory, "post.md")

	// setup
	fs := afero.NewMemMapFs()
	require.NoError(t, afero.WriteFile(fs, filepath.Join(contentDirectory, "index.md"), []byte("<!-- seq: 1 -->\n# "+uuid.New().String()), 0644))
	require.NoError(t, afero.WriteFile(fs, filepath.Join(contentDirectory, sectionDirectory, "index.md"), []byte("<!-- seq: 2 -->\n# "+sectionTitle), 0644))
	require.NoError(t, afero.WriteFile(fs, markdownPath, []byte("# "+uuid.New().String()), 0644))
	resolver := NewResolver(fs, contentDirectory)

	// test
	currentSection, err := resolver.ResolveForFile(markdownPath)
	sections, resolveErr := resolver.Resolve()

	// expect
	require.NoError(t, err)
	require.NoError(t, resolveErr)
	assert.Equal(t, 2, currentSection.Seq)
	assert.Contains(t, sections, currentSection)
}

func TestSectionResolver_ResolveForFile_ShouldReturnErrorOnRelFailure(t *testing.T) {
	// given
	contentDirectory := uuid.New().String()
	markdownPath := string(filepath.Separator) + filepath.Join(uuid.New().String(), "post.md")

	// setup
	resolver := NewResolver(afero.NewMemMapFs(), contentDirectory)

	// test
	_, err := resolver.ResolveForFile(markdownPath)

	// expect
	assert.ErrorContains(t, err, "filepath.Rel")
}

func TestSectionResolver_ResolveForFile_ShouldReturnErrorWhenIndexFileMissing(t *testing.T) {
	// given
	contentDirectory := uuid.New().String()
	sectionDirectory := uuid.New().String()

	// setup
	fs := afero.NewMemMapFs()
	markdownPath := filepath.Join(contentDirectory, sectionDirectory, "post.md")
	require.NoError(t, afero.WriteFile(fs, markdownPath, []byte("# "+uuid.New().String()), 0644))
	resolver := NewResolver(fs, contentDirectory)

	// test
	_, err := resolver.ResolveForFile(markdownPath)

	// expect
	assert.ErrorContains(t, err, "afero.ReadFile err")
}

func TestExtractSection_ShouldReturnErrorWhenFileMissing(t *testing.T) {
	// given
	homeSectionFile := filepath.Join(uuid.New().String(), "index.md")

	// setup
	fs := afero.NewMemMapFs()

	// test
	_, err := extractSection(fs, homeSectionFile)

	// expect
	assert.ErrorContains(t, err, "afero.ReadFile err")
}
