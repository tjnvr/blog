package fs

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFilesFinder_FindFiles_Success(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(fs afero.Fs)
		expected []string
	}{
		{
			name: "single level",
			setup: func(fs afero.Fs) {
				_ = afero.WriteFile(fs, "f1.txt", []byte(""), 0644)
			},
			expected: []string{"f1.txt"},
		},
		{
			name: "multiple levels",
			setup: func(fs afero.Fs) {
				_ = afero.WriteFile(fs, "f1.txt", []byte(""), 0644)
				_ = afero.WriteFile(fs, "a/f2.txt", []byte(""), 0644)
				_ = afero.WriteFile(fs, "a/b/f3.txt", []byte(""), 0644)
			},
			expected: []string{"f1.txt", "a/f2.txt", "a/b/f3.txt"},
		},
		{
			name: "no files",
			setup: func(fs afero.Fs) {
			},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			fs := afero.NewMemMapFs()
			tt.setup(fs)
			finder := NewFilesFinder(fs)

			// test
			files, err := finder.FindFiles(".")

			// expect
			require.NoError(t, err)
			assert.ElementsMatch(t, tt.expected, files)
		})
	}
}

func TestFilesFinder_FindFiles_Failure(t *testing.T) {
	// given
	fs := afero.NewMemMapFs()
	finder := NewFilesFinder(fs)

	// test
	_, err := finder.FindFiles("non-existent")

	// expect
	assert.Error(t, err)
}

func TestFilesFinder_FindFiles_Options(t *testing.T) {
	tests := []struct {
		name     string
		opts     []FilesFinderOption
		setup    func(fs afero.Fs)
		expected []string
	}{
		{
			name: "WithLevel",
			opts: []FilesFinderOption{WithLevel(1)},
			setup: func(fs afero.Fs) {
				_ = afero.WriteFile(fs, "root.txt", []byte(""), 0644)
				_ = afero.WriteFile(fs, "sub/sub.txt", []byte(""), 0644)
			},
			expected: []string{"root.txt"},
		},
		{
			name: "WithExtension",
			opts: []FilesFinderOption{WithExtension(".go")},
			setup: func(fs afero.Fs) {
				_ = afero.WriteFile(fs, "a.txt", []byte(""), 0644)
				_ = afero.WriteFile(fs, "b.go", []byte(""), 0644)
			},
			expected: []string{"b.go"},
		},
		{
			name: "WithPattern",
			opts: []FilesFinderOption{WithPattern("^foo")},
			setup: func(fs afero.Fs) {
				_ = afero.WriteFile(fs, "bar.txt", []byte(""), 0644)
				_ = afero.WriteFile(fs, "foobar.go", []byte(""), 0644)
				_ = afero.WriteFile(fs, "barfoo.go", []byte(""), 0644)
			},
			expected: []string{"foobar.go"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			fs := afero.NewMemMapFs()
			tt.setup(fs)
			finder := NewFilesFinder(fs, tt.opts...)

			// test
			files, err := finder.FindFiles(".")

			// expect
			require.NoError(t, err)
			assert.ElementsMatch(t, tt.expected, files)
		})
	}
}
