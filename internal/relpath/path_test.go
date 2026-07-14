package relpath

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolve(t *testing.T) {
	tests := []struct {
		name             string
		oldPath          string
		oldPathDirectory string
		newPathDirectory string
		fromPath         string
		want             string
		wantErr          bool
	}{
		{
			name:             "from same directory",
			oldPath:          "from/file.png",
			oldPathDirectory: "from/",
			newPathDirectory: "to/",
			fromPath:         "to/page.html",
			want:             "file.png",
		},
		{
			name:             "from a parent directory",
			oldPath:          "from/file.png",
			oldPathDirectory: "from/",
			newPathDirectory: "to/",
			fromPath:         "to/other/page.html",
			want:             "../file.png",
		},
		{
			name:             "from multiple parent directories",
			oldPath:          "from/file.png",
			oldPathDirectory: "from/",
			newPathDirectory: "to/",
			fromPath:         "to/other/one/page.html",
			want:             "../../file.png",
		},
		{
			name:             "from multiple parent directories with nested old path",
			oldPath:          "from/another/file.png",
			oldPathDirectory: "from/",
			newPathDirectory: "to/",
			fromPath:         "to/other/one/page.html",
			want:             "../../another/file.png",
		},
		{
			name:             "error when oldPath is not inside oldPathDirectory",
			oldPath:          "other/file.png",
			oldPathDirectory: "from/",
			newPathDirectory: "to/",
			fromPath:         "to/page.html",
			wantErr:          true,
		},
		{
			name:             "error when oldPath is absolute and oldPathDirectory is relative",
			oldPath:          "/absolute/file.png",
			oldPathDirectory: "from/",
			newPathDirectory: "to/",
			fromPath:         "to/page.html",
			wantErr:          true,
		},
		{
			name:             "error when oldPathDirectory is absolute and oldPath is relative",
			oldPath:          "from/file.png",
			oldPathDirectory: "/absolute/from/",
			newPathDirectory: "to/",
			fromPath:         "to/page.html",
			wantErr:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			resolver := NewResolver(tt.oldPathDirectory, tt.newPathDirectory)

			// test
			got, err := resolver.Resolve(tt.oldPath, tt.fromPath)

			// expect
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestResolve_WithExtension(t *testing.T) {
	tests := []struct {
		name     string
		oldPath  string
		fromPath string
		want     string
	}{
		{
			name:     "replaces a markdown source extension with the target extension",
			oldPath:  "content/index.md",
			fromPath: "target/index.html",
			want:     "index.html",
		},
		{
			name:     "replaces a nested source file extension with the target extension",
			oldPath:  "content/posts/index.md",
			fromPath: "target/index.html",
			want:     "posts/index.html",
		},
		{
			name:     "replaces an extension even when the source has none",
			oldPath:  "content/README",
			fromPath: "target/index.html",
			want:     "README.html",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			resolver := NewResolver("content", "target", WithExtension(".html"))

			// test
			got, err := resolver.Resolve(tt.oldPath, tt.fromPath)

			// expect
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
