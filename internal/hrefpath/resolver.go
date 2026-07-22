// Package hrefpath computes the absolute, root-relative URL used to link to
// a generated page from anywhere else on the site, given only the page's
// source markdown path.
//
// A file literally named index.md is treated as the home page of its
// containing directory: the trailing "index" segment is dropped, so
// "content/markdown/posts/index.md" resolves to "/posts" rather than
// "/posts/index".
package hrefpath

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"
)

type (
	// Resolver computes the href used to link to the page sourced from
	// oldPath.
	Resolver interface {
		Resolve(oldPath string) (string, error)
	}
	resolver struct {
		oldPathDirectory string
	}
)

// NewResolver creates a Resolver for markdown sources rooted at
// oldPathDirectory (e.g. the site's ContentDir).
func NewResolver(oldPathDirectory string) Resolver {
	return resolver{oldPathDirectory: oldPathDirectory}
}

func (r resolver) Resolve(oldPath string) (string, error) {
	oldPath = strings.TrimSuffix(oldPath, filepath.Ext(oldPath))

	oldPathRelToOldPathDir, err := filepath.Rel(r.oldPathDirectory, oldPath)
	if err != nil {
		return "", err
	}
	if strings.HasPrefix(oldPathRelToOldPathDir, "..") {
		return "", fmt.Errorf("oldPath %q is not inside oldPathDirectory %q", oldPath, r.oldPathDirectory)
	}

	slashPath := filepath.ToSlash(oldPathRelToOldPathDir)
	dir, base := path.Split(slashPath)
	if base == "index" {
		slashPath = strings.TrimSuffix(dir, "/")
	}

	return "/" + slashPath, nil
}
