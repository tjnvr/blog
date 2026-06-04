package relpath

import (
	"fmt"
	"path/filepath"
	"strings"
)

type defaultResolver struct {
	oldPathDirectory, newPathDirectory string
}

func NewDefaultResolver(oldPathDirectory, newPathDirectory string) defaultResolver {
	return defaultResolver{
		oldPathDirectory: oldPathDirectory,
		newPathDirectory: newPathDirectory,
	}
}

// Resolve returns the relative path from file at fromPath of a file originally at oldPath.
// The path resolver instance can only be used to resolve paths relative to the project root.
// oldPath must be located inside oldPathDirectory, otherwise the function returns an error.
// It is used to compute relative path of links and local assets and pages in the html pages.
func (p defaultResolver) Resolve(oldPath, fromPath string) (string, error) {
	oldPathRelToOldPathDir, err := filepath.Rel(p.oldPathDirectory, oldPath)
	if err != nil {
		return "", err
	}

	if strings.HasPrefix(oldPathRelToOldPathDir, "..") {
		return "", fmt.Errorf("oldPath %q is not inside oldPathDirectory %q", oldPath, p.oldPathDirectory)
	}

	newPathFromRootDir := filepath.Join(p.newPathDirectory, oldPathRelToOldPathDir)

	// new path relative to fromPath (from fromPath directory)
	result, err := filepath.Rel(filepath.Dir(fromPath), newPathFromRootDir)
	if err != nil {
		return "", err
	}

	return result, nil
}
