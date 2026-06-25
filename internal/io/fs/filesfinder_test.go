package fs

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestFindFiles(t *testing.T) {
	tests := []struct {
		name          string
		setupFs       func(t *testing.T) afero.Fs
		searchDir     string
		expectedFiles []string
		expectErr     bool
	}{
		{
			name: "one level directory",
			setupFs: func(t *testing.T) afero.Fs {
				fs := afero.NewMemMapFs()
				assert.Nil(t, afero.WriteFile(fs, "file1.txt", []byte("content1"), 0644))
				assert.Nil(t, afero.WriteFile(fs, "file2.log", []byte("content2"), 0644))
				return fs
			},
			searchDir:     ".",
			expectedFiles: []string{"file1.txt", "file2.log"},
		},
		{
			name: "multiple levels directories",
			setupFs: func(t *testing.T) afero.Fs {
				fs := afero.NewMemMapFs()
				assert.Nil(t, afero.WriteFile(fs, "a/b/c/file1.txt", []byte("content"), 0644))
				assert.Nil(t, afero.WriteFile(fs, "a/b/file2.txt", []byte("content"), 0644))
				return fs
			},
			searchDir:     ".",
			expectedFiles: []string{"file1.txt", "file2.txt"},
		},
		{
			name: "empty directory",
			setupFs: func(t *testing.T) afero.Fs {
				fs := afero.NewMemMapFs()
				assert.Nil(t, fs.Mkdir("empty", 0755))
				return fs
			},
			searchDir:     "empty",
			expectedFiles: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := tt.setupFs(t)
			finder := NewFilesFinder(fs)
			files, err := finder.FindFiles(tt.searchDir)
			assert.Nil(t, err)
			assert.EqualValues(t, tt.expectedFiles, files)
		})
	}
}

func TestFindFiles_Options(t *testing.T) {
	tests := []struct {
		name           string
		setupFs        func(t *testing.T) afero.Fs
		setupFinder    func(t *testing.T, fs afero.Fs) FilesFinder
		searchDir      string
		expectedFiles  []string
		expectErr      bool
	}{
		{
			name: "with level 1",
			setupFs: func(t *testing.T) afero.Fs {
				fs := afero.NewMemMapFs()
				assert.Nil(t, afero.WriteFile(fs, "file1.txt", []byte("content1"), 0644))
				assert.Nil(t, afero.WriteFile(fs, "subdir/file2.txt", []byte("content2"), 0644))
				return fs
			},
			setupFinder: func(t *testing.T, fs afero.Fs) FilesFinder {
				return NewFilesFinder(fs, WithLevel(1))
			},
			searchDir:     ".",
			expectedFiles: []string{"file1.txt"},
		},
		{
			name: "with extension .txt",
			setupFs: func(t *testing.T) afero.Fs {
				fs := afero.NewMemMapFs()
				assert.Nil(t, afero.WriteFile(fs, "file1.txt", []byte("content1"), 0644))
				assert.Nil(t, afero.WriteFile(fs, "file2.log", []byte("content2"), 0644))
				return fs
			},
			setupFinder: func(t *testing.T, fs afero.Fs) FilesFinder {
				return NewFilesFinder(fs, WithExtension(".txt"))
			},
			searchDir:     ".",
			expectedFiles: []string{"file1.txt"},
		},

		{
			name: "with extension .txt",
			setupFs: func(t *testing.T) afero.Fs {
				fs := afero.NewMemMapFs()
				assert.Nil(t, afero.WriteFile(fs, "file1.txt", []byte("content1"), 0644))
				assert.Nil(t, afero.WriteFile(fs, "file2.log", []byte("content2"), 0644))
				return fs
			},
			setupFinder: func(t *testing.T, fs afero.Fs) FilesFinder {
				return NewFilesFinder(fs, WithExtension(".txt"))
			},
			searchDir:     ".",
			expectedFiles: []string{"file1.txt"},
		},

	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := tt.setupFs(t)
			finder := tt.setupFinder(t, fs)
			files, err := finder.FindFiles(tt.searchDir)
			if tt.expectErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.EqualValues(t, tt.expectedFiles, files)
			}
		})
	}
}
