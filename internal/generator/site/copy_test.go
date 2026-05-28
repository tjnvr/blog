package site

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCopyDir(t *testing.T) {
	tests := []struct {
		name          string
		srcFiles      map[string]string
		filter        func(string) bool
		expectedFiles []string
		wantErr       bool
	}{
		{
			name: "copy all files without filter",
			srcFiles: map[string]string{
				"file1.txt": "content1",
				"file2.txt": "content2",
			},
			filter:        nil,
			expectedFiles: []string{"file1.txt", "file2.txt"},
		},
		{
			name: "preserve directory structure",
			srcFiles: map[string]string{
				"root.txt":          "root",
				"sub/nested.txt":    "nested",
				"sub/deep/file.txt": "deep",
			},
			filter:        nil,
			expectedFiles: []string{"root.txt", "sub/nested.txt", "sub/deep/file.txt"},
		},
		{
			name:          "empty source directory",
			srcFiles:      map[string]string{},
			filter:        nil,
			expectedFiles: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			srcDir := "/src"
			destDir := "/dest"

			require.NoError(t, fs.MkdirAll(srcDir, 0755))

			for relPath, content := range tt.srcFiles {
				fullPath := filepath.Join(srcDir, relPath)
				require.NoError(t, fs.MkdirAll(filepath.Dir(fullPath), 0755))
				require.NoError(t, afero.WriteFile(fs, fullPath, []byte(content), 0644))
			}

			err := copyDir(fs, srcDir, destDir)

			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			for _, expectedFile := range tt.expectedFiles {
				destPath := filepath.Join(destDir, expectedFile)
				if _, err := fs.Stat(destPath); err != nil {
					assert.NoError(t, err, "expected file %s not found in dest", expectedFile)
					continue
				}

				srcContent, _ := afero.ReadFile(fs, filepath.Join(srcDir, expectedFile))
				destContent, _ := afero.ReadFile(fs, destPath)
				assert.Equal(t, string(srcContent), string(destContent))
			}

			var copiedFiles []string
			_ = afero.Walk(fs, destDir, func(path string, info os.FileInfo, err error) error {
				if err != nil || info.IsDir() {
					return err
				}
				relPath, _ := filepath.Rel(destDir, path)
				copiedFiles = append(copiedFiles, relPath)
				return nil
			})
			assert.Len(t, copiedFiles, len(tt.expectedFiles))
		})
	}
}

func TestCopyDir_NonExistentSource(t *testing.T) {
	fs := afero.NewMemMapFs()
	err := copyDir(fs, "/nonexistent", "/dest")
	assert.Error(t, err)
}
