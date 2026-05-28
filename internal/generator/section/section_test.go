package section

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestListSections(t *testing.T) {
	tests := []struct {
		name             string
		files            map[string]string
		dirs             []string
		expectedSections []Section
	}{
		{
			name:             "no sections",
			files:            map[string]string{},
			expectedSections: []Section{{DirName: "", DisplayName: "Home"}},
		},
		{
			name:             "home title read from index.md",
			files:            map[string]string{"index.md": "# My Blog\n"},
			expectedSections: []Section{{DirName: "", DisplayName: "My Blog"}},
		},
		{
			name:             "section without index.md falls back to capitalized dir name",
			dirs:             []string{"blog"},
			expectedSections: []Section{{DirName: "", DisplayName: "Home"}, {DirName: "blog", DisplayName: "Blog"}},
		},
		{
			name:             "section title read from index.md",
			files:            map[string]string{"blog/index.md": "# Tech Blog\n"},
			expectedSections: []Section{{DirName: "", DisplayName: "Home"}, {DirName: "blog", DisplayName: "Tech Blog"}},
		},
		{
			name: "multiple sections",
			files: map[string]string{
				"articles/index.md": "# Articles\n",
				"projects/index.md": "# Projects\n",
			},
			expectedSections: []Section{
				{DirName: "", DisplayName: "Home"},
				{DirName: "articles", DisplayName: "Articles"},
				{DirName: "projects", DisplayName: "Projects"},
			},
		},
		{
			name:             "nested directories are not treated as sections",
			files:            map[string]string{"blog/2024/post.md": "# A Post"},
			expectedSections: []Section{{DirName: "", DisplayName: "Home"}, {DirName: "blog", DisplayName: "Blog"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			memoryFs, err := newMemoryFs(tt.files, tt.dirs)
			assert.Nil(t, err)
			sections, err := ListSections(memoryFs, ".")
			assert.Nil(t, err)
			assert.Equal(t, tt.expectedSections, sections)
		})
	}
}

func newMemoryFs(files map[string]string, dirs []string) (afero.Fs, error) {
	fs := afero.NewMemMapFs()
	for _, dir := range dirs {
		if err := fs.MkdirAll(dir, 0755); err != nil {
			return nil, err
		}
	}
	for k, v := range files {
		if err := afero.WriteFile(fs, k, []byte(v), 0644); err != nil {
			return nil, err
		}
	}
	return fs, nil
}
