package article

import (
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestArticlePrint(t *testing.T) {
	tests := []struct {
		name string
		a    Article
		want string
	}{
		{
			name: "no date",
			a:    Article{name: "Hello", filePath: "hello.md"},
			want: "- [Hello](hello.md)",
		},
		{
			name: "with date",
			a:    Article{name: "Hello", filePath: "hello.md", createdAt: "2024-03-15"},
			want: "- [Hello](hello.md) · *2024-03-15*",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.a.Print())
		})
	}
}

func TestExtractTitle(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "empty content", input: "", want: ""},
		{name: "no heading", input: "some text\nno heading here", want: ""},
		{name: "hash without space", input: "#notATitle", want: ""},
		{name: "valid h1", input: "# My Title\nsome content", want: "My Title"},
		{name: "multiple h1s returns first", input: "# First\n# Second", want: "First"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, extractTitle([]byte(tt.input)))
		})
	}
}

func TestListPrinters(t *testing.T) {
	tests := []struct {
		name      string
		files     map[string]string
		indexFile string
		wantNames []string
		wantErr   bool
	}{
		{
			name:      "empty directory",
			files:     map[string]string{"index.md": "# Index"},
			indexFile: "index.md",
			wantNames: []string{},
		},
		{
			name: "two articles",
			files: map[string]string{
				"index.md":  "# Index",
				"hello.md":  "# Hello World\nsome content",
				"second.md": "# Second Post\nmore content",
			},
			indexFile: "index.md",
			wantNames: []string{"Hello World", "Second Post"},
		},
		{
			name: "skips file without h1",
			files: map[string]string{
				"index.md":   "# Index",
				"notitle.md": "no heading here",
				"hello.md":   "# Hello\ncontent",
			},
			indexFile: "index.md",
			wantNames: []string{"Hello"},
		},
		{
			name: "skips subdirectory",
			files: map[string]string{
				"index.md":     "# Index",
				"sub/child.md": "# Child",
				"hello.md":     "# Hello\ncontent",
			},
			indexFile: "index.md",
			wantNames: []string{"Hello"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			dir := "/content"
			for relPath, content := range tt.files {
				fullPath := filepath.Join(dir, relPath)
				require.NoError(t, fs.MkdirAll(filepath.Dir(fullPath), 0755))
				require.NoError(t, afero.WriteFile(fs, fullPath, []byte(content), 0644))
			}

			lister := NewPageArticlesLister(filepath.Join(dir, tt.indexFile), fs)
			articles, err := lister.ListPrinters()
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Len(t, articles, len(tt.wantNames))

			actualNames := make([]string, 0, len(articles))
			for _, a := range articles {
				actualNames = append(actualNames, a.name)
			}
			assert.ElementsMatch(t, tt.wantNames, actualNames)
		})
	}
}
