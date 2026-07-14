// Package relpath provides utilities for resolving relative file paths
// between different locations within a project, primarily used for generating
// correct links, assets, and page paths in HTML output.
//
// When a file is moved from an old location to a new one, this package calculates
// the necessary relative path from a given "from" page to the file's new location,
// ensuring links remain valid across different directory depths.
package relpath

import (
	"fmt"
	"path/filepath"
	"strings"
)

type (
	// Resolve computes the relative path to the moved file from the perspective
	// of the file at fromPath.
	//
	// It takes the original path of the file (oldPath) and the location of the page
	// referring to it (fromPath). It returns the relative path that should be used
	// in the page to link to the file at its new location.
	//
	// oldPath must be within the resolver's oldPathDirectory, otherwise it returns
	// an error.
	Resolver interface {
		Resolve(oldPath, fromPath string) (string, error)
	}
	resolver struct {
		oldPathDirectory, newPathDirectory string
		targetExtension                    string
	}

	// Option configures a Resolver created via NewResolver.
	Option func(*resolver)
)

// WithExtension makes the resolver replace oldPath's file extension with ext
// before resolving, for translating a source file (e.g. markdown) to the path
// of its generated output (e.g. HTML).
func WithExtension(ext string) Option {
	return func(r *resolver) {
		r.targetExtension = ext
	}
}

func NewResolver(oldPathDirectory, newPathDirectory string, opts ...Option) Resolver {
	r := resolver{
		oldPathDirectory: oldPathDirectory,
		newPathDirectory: newPathDirectory,
	}
	for _, opt := range opts {
		opt(&r)
	}
	return r
}

func (r resolver) Resolve(oldPath, fromPath string) (string, error) {
	if r.targetExtension != "" {
		oldPath = strings.TrimSuffix(oldPath, filepath.Ext(oldPath)) + r.targetExtension
	}

	oldPathRelToOldPathDir, err := filepath.Rel(r.oldPathDirectory, oldPath)
	if err != nil {
		return "", err
	}

	if strings.HasPrefix(oldPathRelToOldPathDir, "..") {
		return "", fmt.Errorf("oldPath %q is not inside oldPathDirectory %q", oldPath, r.oldPathDirectory)
	}

	newPathFromRootDir := filepath.Join(r.newPathDirectory, oldPathRelToOldPathDir)

	// new path relative to fromPath (from fromPath directory)
	result, err := filepath.Rel(filepath.Dir(fromPath), newPathFromRootDir)
	if err != nil {
		return "", err
	}

	return result, nil
}
