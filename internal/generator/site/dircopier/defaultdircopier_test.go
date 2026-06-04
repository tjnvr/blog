package dircopier

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultDirCopier_CopyDir(t *testing.T) {
	tests := []struct {
		name          string
		from          string
		to            string
		srcFiles      map[string]string
		expectedFiles []string
		wantErr       bool
	}{
		{
			name: "copy all files without filter",
			from: "/src",
			to:   "/dest",
			srcFiles: map[string]string{
				"file1.txt": "content1",
				"file2.txt": "content2",
			},
			expectedFiles: []string{"file1.txt", "file2.txt"},
		},
		{
			name: "preserve directory structure",
			from: "/src",
			to:   "/dest",
			srcFiles: map[string]string{
				"root.txt":          "root",
				"sub/nested.txt":    "nested",
				"sub/deep/file.txt": "deep",
			},
			expectedFiles: []string{"root.txt", "sub/nested.txt", "sub/deep/file.txt"},
		},
		{
			name:          "empty source directory",
			from:          "/src",
			to:            "/dest",
			srcFiles:      map[string]string{},
			expectedFiles: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()

			require.NoError(t, fs.MkdirAll(tt.from, 0755))

			for relPath, content := range tt.srcFiles {
				fullPath := filepath.Join(tt.from, relPath)
				require.NoError(t, fs.MkdirAll(filepath.Dir(fullPath), 0755))
				require.NoError(t, afero.WriteFile(fs, fullPath, []byte(content), 0644))
			}

			copier := defaultDirCopier{}
			err := copier.CopyDir(fs, tt.from, tt.to)

			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			for _, expectedFile := range tt.expectedFiles {
				destPath := filepath.Join(tt.to, expectedFile)
				if _, err := fs.Stat(destPath); err != nil {
					assert.NoError(t, err, "expected file %s not found in dest", expectedFile)
					continue
				}

				srcContent, _ := afero.ReadFile(fs, filepath.Join(tt.from, expectedFile))
				destContent, _ := afero.ReadFile(fs, destPath)
				assert.Equal(t, string(srcContent), string(destContent))
			}

			var copiedFiles []string
			_ = afero.Walk(fs, tt.to, func(path string, info os.FileInfo, err error) error {
				if err != nil || info.IsDir() {
					return err
				}
				relPath, _ := filepath.Rel(tt.to, path)
				copiedFiles = append(copiedFiles, relPath)
				return nil
			})
			assert.Len(t, copiedFiles, len(tt.expectedFiles))
		})
	}
}

func TestDefaultDirCopier_CopyDir_Random(t *testing.T) {
	fs := afero.NewMemMapFs()
	from := fmt.Sprintf("/src-%d", rand.Intn(10000))
	to := fmt.Sprintf("/dest-%d", rand.Intn(10000))

	numFiles := 5
	expectedFiles := make([]string, numFiles)
	for i := range numFiles {
		relPath := fmt.Sprintf("file-%d.txt", rand.Intn(100))
		content := fmt.Sprintf("content-%d", rand.Intn(100))

		fullPath := filepath.Join(from, relPath)
		require.NoError(t, fs.MkdirAll(filepath.Dir(fullPath), 0755))
		require.NoError(t, afero.WriteFile(fs, fullPath, []byte(content), 0644))

		expectedFiles[i] = relPath
	}

	copier := defaultDirCopier{}
	require.NoError(t, copier.CopyDir(fs, from, to))

	for _, relPath := range expectedFiles {
		destPath := filepath.Join(to, relPath)
		_, err := fs.Stat(destPath)
		assert.NoError(t, err, "expected file %s not found in dest", relPath)
	}
}

func TestDefaultDirCopier_CopyDir_NonExistentSource(t *testing.T) {
	fs := afero.NewMemMapFs()
	copier := defaultDirCopier{}
	err := copier.CopyDir(fs, "/nonexistent", "/dest")
	assert.Error(t, err)
}
