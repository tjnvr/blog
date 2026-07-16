package abspath

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestResolver_Resolve(t *testing.T) {
	tests := []struct {
		name     string
		buildDir string
		htmlPath string
		ref      string
		want     string
	}{
		{
			name:     "resolves a relative ref from the HTML directory",
			buildDir: "/build",
			htmlPath: "/build/posts/article.html",
			ref:      "../assets/logo.png",
			want:     "/build/assets/logo.png",
		},
		{
			name:     "resolves an absolute ref from the build root",
			buildDir: "/build",
			htmlPath: "/build/posts/article.html",
			ref:      "/assets/logo.png",
			want:     "/build/assets/logo.png",
		},
		{
			name:     "ignores a URL fragment",
			buildDir: "/build",
			htmlPath: "/build/index.html",
			ref:      "/posts/article.html#section",
			want:     "/build/posts/article.html",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// setup
			resolver := NewResolver(afero.NewMemMapFs(), tt.buildDir, tt.htmlPath)

			// test
			got := resolver.Resolve(tt.ref)

			// expect
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestResolver_Exists(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(fs afero.Fs)
		buildDir string
		htmlPath string
		ref      string
		want     bool
	}{
		{
			name: "returns true when the file exists",
			setup: func(fs afero.Fs) {
				assert.NoError(t, afero.WriteFile(fs, "/build/assets/logo.png", []byte("x"), 0644))
			},
			buildDir: "/build",
			htmlPath: "/build/posts/article.html",
			ref:      "/assets/logo.png",
			want:     true,
		},
		{
			name:     "returns false when the file is missing",
			buildDir: "/build",
			htmlPath: "/build/posts/article.html",
			ref:      "/assets/missing.png",
			want:     false,
		},
		{
			name:     "returns true for a same-page fragment",
			buildDir: "/build",
			htmlPath: "/build/index.html",
			ref:      "#section",
			want:     true,
		},
		{
			name: "returns true when the target is an existing directory",
			setup: func(fs afero.Fs) {
				assert.NoError(t, afero.WriteFile(fs, "/build/posts/index.html", []byte("x"), 0644))
			},
			buildDir: "/build",
			htmlPath: "/build/index.html",
			ref:      "/posts",
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			fs := afero.NewMemMapFs()
			if tt.setup != nil {
				tt.setup(fs)
			}

			// setup
			resolver := NewResolver(fs, tt.buildDir, tt.htmlPath)

			// test
			ok, err := resolver.Exists(tt.ref)

			// expect
			assert.NoError(t, err)
			assert.Equal(t, tt.want, ok)
		})
	}
}
