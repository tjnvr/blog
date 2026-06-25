package fs

import (
	"fmt"
	"io/fs"
	"path/filepath"
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

	FilesFinder interface {
		FindFiles(dir string) ([]string, error)
	}

	filesFinderOptionsFunc func(*filesFinderOptions)
)

func NewFilesFinder(fs afero.Fs, opts ...filesFinderOptionsFunc) FilesFinder {
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

		// Calculate current depth level relative to the starting directory
		if f.opts != nil && f.opts.level != nil {
			rel, err := filepath.Rel(baseDir, path)
			if err != nil {
				return err
			}

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
			if f.opts.pattern != nil && !strings.Contains(info.Name(), *f.opts.pattern) {
				return nil
			}
		}

		files = append(files, info.Name())
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("afero.Walk err: %v", err)
	}

	return files, nil
}

func WithLevel(level int) filesFinderOptionsFunc {
	return func(opts *filesFinderOptions) {
		opts.level = &level
	}
}

func WithExtension(ext string) filesFinderOptionsFunc {
	return func(opts *filesFinderOptions) {
		opts.extension = &ext
	}
}

func WithPattern(pattern string) filesFinderOptionsFunc {
	return func(opts *filesFinderOptions) {
		opts.pattern = &pattern
	}
}
