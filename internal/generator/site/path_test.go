package site

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetNewPath(t *testing.T) {
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
			resolver := NewPathResolver(tt.oldPathDirectory, tt.newPathDirectory)
			got, err := resolver.GetNewPath(tt.oldPath, tt.fromPath)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
