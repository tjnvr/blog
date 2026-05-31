package path

import (
	"testing"
)

func TestDefaultResolver_Resolve(t *testing.T) {
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
			resolver := NewDefaultResolver(tt.oldPathDirectory, tt.newPathDirectory)
			got, err := resolver.Resolve(tt.oldPath, tt.fromPath)
			if (err != nil) != tt.wantErr {
				t.Fatalf("GetNewPath() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("GetNewPath() = %q, want %q", got, tt.want)
			}
		})
	}
}
