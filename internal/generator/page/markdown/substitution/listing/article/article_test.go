package article

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestArticle_Print(t *testing.T) {
	tests := []struct {
		name string
		a    Article
		want string
	}{
		{
			name: "no date",
			a:    Article{name: "Hello", markdownPath: "hello.md"},
			want: "- [Hello](hello.md)",
		},
		{
			name: "with date",
			a:    Article{name: "Hello", markdownPath: "hello.md", createdAt: "2024-03-15"},
			want: "- [Hello](hello.md) · *2024-03-15*",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// test
			got := tt.a.Print()

			// expect
			assert.Equal(t, tt.want, got)
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
			// test
			got := extractTitle([]byte(tt.input))

			// expect
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPageArticlesLister_ListPrinters(t *testing.T) {
	tests := []struct {
		name      string
		files     map[string]string // relative path -> content
		indexFile string            // relative path of the index file
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
			// given
			memFs := afero.NewMemMapFs()

			// setup
			for relPath, content := range tt.files {
				assert.Nil(t, afero.WriteFile(memFs, relPath, []byte(content), 0644))
			}
			lister := NewPageArticlesLister(memFs, tt.indexFile)

			// test
			articles, err := lister.ListPrinters()

			// expect
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, len(tt.wantNames), len(articles))

				gotNames := make(map[string]bool)
				for _, a := range articles {
					gotNames[a.name] = true
				}
				for _, name := range tt.wantNames {
					assert.True(t, gotNames[name], "missing article with name %q", name)
				}
			}
		})
	}
}
