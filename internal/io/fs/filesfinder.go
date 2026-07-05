// Package fs provides utilities for searching files and copying directories
// within a filesystem using the afero.Fs abstraction. See github.com/spf13/afero for details.
//
// It includes a FilesFinder for filtering files by directory depth, extension, and name patterns,
// as well as a DirCopier for copying directory contents.
package fs

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/afero"
)

type (
	filesFinder struct {
		fs   afero.Fs
		opts *filesFinderOptions
	}

	filesFinderOptions struct {
		level     *int
		extension *string
		pattern   *string
	}

	// FilesFinder defines the interface for searching files within a directory.
	FilesFinder interface {
		// FindFiles searches for files within the specified directory based on the
		// options provided during instantiation. It returns a slice of file paths
		// relative to the dir argument.
		FindFiles(dir string) ([]string, error)
	}

	// FilesFinderOption defines a configuration function for FilesFinder.
	FilesFinderOption func(*filesFinderOptions)
)

// NewFilesFinder creates a new FilesFinder instance using the provided
// afero.Fs and optional configuration functions.
func NewFilesFinder(fs afero.Fs, opts ...FilesFinderOption) FilesFinder {
	f := &filesFinder{
		fs:   fs,
		opts: &filesFinderOptions{},
	}

	for _, opt := range opts {
		opt(f.opts)
	}

	return f
}

func (f *filesFinder) FindFiles(dir string) ([]string, error) {
	files := make([]string, 0)

	baseDir := filepath.Clean(dir)

	err := afero.Walk(f.fs, baseDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Calculate relative path
		rel, err := filepath.Rel(baseDir, path)
		if err != nil {
			return err
		}

		// Calculate current depth level relative to the starting directory
		if f.opts != nil && f.opts.level != nil {
			// Count the depth level based on path separators
			var currentLevel int
			if rel != "." {
				currentLevel = len(strings.Split(rel, string(filepath.Separator)))
			}

			// If a directory itself exceeds the maximum allowed level, skip traversing it entirely
			if info.IsDir() && currentLevel >= *f.opts.level {
				return filepath.SkipDir
			}

			// If a file exceeds the level limit, skip it
			if currentLevel > *f.opts.level {
				return nil
			}
		}

		if info.IsDir() {
			return nil
		}

		if f.opts != nil {
			// Extension filter
			if f.opts.extension != nil && filepath.Ext(info.Name()) != *f.opts.extension {
				return nil
			}

			// Pattern filter
			if f.opts.pattern != nil {
				matched, err := regexp.MatchString(*f.opts.pattern, info.Name())
				if err != nil || !matched {
					return nil
				}
			}
		}

		files = append(files, rel)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("afero.Walk err: %v", err)
	}

	return files, nil
}

// WithLevel sets the maximum depth level for searching files.
func WithLevel(level int) FilesFinderOption {
	return func(opts *filesFinderOptions) {
		opts.level = &level
	}
}

// WithExtension sets the file extension filter for searching files.
func WithExtension(ext string) FilesFinderOption {
	return func(opts *filesFinderOptions) {
		opts.extension = &ext
	}
}

// WithPattern sets the name pattern filter for searching files.
func WithPattern(pattern string) FilesFinderOption {
	return func(opts *filesFinderOptions) {
		opts.pattern = &pattern
	}
}
