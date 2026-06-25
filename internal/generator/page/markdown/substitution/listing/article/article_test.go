package article

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/tjnvr/blog/internal/io/fs"
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
			if got := tt.a.Print(); got != tt.want {
				t.Errorf("Print() = %q, want %q", got, tt.want)
			}
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
			got := extractTitle([]byte(tt.input))
			if got != tt.want {
				t.Errorf("extractTitle() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestListPrinters(t *testing.T) {
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
			memFs := afero.NewMemMapFs()
			for relPath, content := range tt.files {
				assert.Nil(t, afero.WriteFile(memFs, relPath, []byte(content), 0644))
			}
			lister := NewPageArticlesLister(memFs, tt.indexFile, fs.NewFilesFinder(memFs, fs.WithExtension(".md"), fs.WithLevel(1)))
			articles, err := lister.ListPrinters()
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(articles) != len(tt.wantNames) {
				t.Fatalf("got %d articles, want %d", len(articles), len(tt.wantNames))
			}

			gotNames := make(map[string]bool)
			for _, a := range articles {
				gotNames[a.name] = true
			}
			for _, name := range tt.wantNames {
				if !gotNames[name] {
					t.Errorf("missing article with name %q", name)
				}
			}
		})
	}
}
