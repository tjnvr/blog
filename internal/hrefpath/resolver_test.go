package hrefpath

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
		want             string
		wantErr          bool
	}{
		{
			name:             "root index.md resolves to the root URL",
			oldPath:          "content/markdown/index.md",
			oldPathDirectory: "content/markdown",
			want:             "/",
		},
		{
			name:             "section index.md resolves to its directory URL",
			oldPath:          "content/markdown/posts/index.md",
			oldPathDirectory: "content/markdown",
			want:             "/posts",
		},
		{
			name:             "another section index.md resolves to its directory URL",
			oldPath:          "content/markdown/about/index.md",
			oldPathDirectory: "content/markdown",
			want:             "/about",
		},
		{
			name:             "non-index page keeps its full path, extension dropped",
			oldPath:          "content/markdown/posts/java-learning-roadmap.md",
			oldPathDirectory: "content/markdown",
			want:             "/posts/java-learning-roadmap",
		},
		{
			name:             "error when oldPath is not inside oldPathDirectory",
			oldPath:          "other/file.md",
			oldPathDirectory: "content/markdown",
			wantErr:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			resolver := NewResolver(tt.oldPathDirectory)

			// test
			got, err := resolver.Resolve(tt.oldPath)

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
