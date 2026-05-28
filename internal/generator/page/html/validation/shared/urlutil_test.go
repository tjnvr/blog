package shared

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsExternalURL(t *testing.T) {
	tests := []struct {
		src  string
		want bool
	}{
		{"https://example.com/page", true},
		{"http://example.com/page", true},
		{"/assets/image.png", false},
		{"../relative/path.html", false},
		{"index.html", false},
		{"mailto:user@example.com", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.src, func(t *testing.T) {
			got := IsExternalURL(tt.src)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestResolveLocalPath(t *testing.T) {
	tests := []struct {
		name     string
		src      string
		htmlPath string
		buildDir string
		want     string
	}{
		{
			name:     "absolute path resolved from build root",
			src:      "/assets/image.png",
			htmlPath: "/build/posts/article.html",
			buildDir: "/build",
			want:     "/build/assets/image.png",
		},
		{
			name:     "relative path resolved from html directory",
			src:      "image.png",
			htmlPath: "/build/posts/article.html",
			buildDir: "/build",
			want:     "/build/posts/image.png",
		},
		{
			name:     "relative path with parent traversal",
			src:      "../assets/style.css",
			htmlPath: "/build/posts/article.html",
			buildDir: "/build",
			want:     "/build/assets/style.css",
		},
		{
			name:     "absolute path at build root",
			src:      "/index.html",
			htmlPath: "/build/index.html",
			buildDir: "/build",
			want:     "/build/index.html",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ResolveLocalPath(tt.src, tt.htmlPath, tt.buildDir)
			assert.Equal(t, tt.want, got)
		})
	}
}
