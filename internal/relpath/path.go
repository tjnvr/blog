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

type Resolver struct {
	oldPathDirectory, newPathDirectory string
}

func NewResolver(oldPathDirectory, newPathDirectory string) Resolver {
	return Resolver{
		oldPathDirectory: oldPathDirectory,
		newPathDirectory: newPathDirectory,
	}
}

// Resolve computes the relative path to the moved file from the perspective
// of the file at fromPath.
//
// It takes the original path of the file (oldPath) and the location of the page
// referring to it (fromPath). It returns the relative path that should be used
// in the page to link to the file at its new location.
//
// oldPath must be within the resolver's oldPathDirectory, otherwise it returns
// an error.
func (r Resolver) Resolve(oldPath, fromPath string) (string, error) {
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
