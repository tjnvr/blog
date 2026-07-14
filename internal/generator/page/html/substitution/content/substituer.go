package content

import (
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/tjnvr/blog/internal/relpath"
)

type (
	substituter struct {
		HTMLPath                              string
		markdownPath                          string
		pagePathsResolver, assetPathsResolver relpath.Resolver
	}
)

func NewSubstituer(HTMLPath, markdownSourcePath string, assetPathsTranslater relpath.Resolver, pagePathsResolver relpath.Resolver) substituter {
	return substituter{
		HTMLPath:           HTMLPath,
		markdownPath:       markdownSourcePath,
		pagePathsResolver:  pagePathsResolver,
		assetPathsResolver: assetPathsTranslater,
	}
}

func (s substituter) Placeholder() string {
	return "{{content}}"
}

func (s substituter) Resolve(htmlContent string) (string, error) {
	var errs error
	var err error

	htmlContent, err = s.convertMdLinksPath(htmlContent, s.HTMLPath)
	if err != nil {
		errs = errors.Join(errs, fmt.Errorf("s.convertMdLinksPath err: %w", err))
	}

	htmlContent, err = s.convertAssetsPath(htmlContent, s.HTMLPath)
	if err != nil {
		errs = errors.Join(errs, fmt.Errorf("s.convertAssetsPath err: %w", err))
	}

	return htmlContent, errs
}

func (s substituter) replacePaths(html, HTMLPath string, re *regexp.Regexp, relpathResolver relpath.Resolver, modifyPath func(string) string) (string, error) {
	var errs error

	result := re.ReplaceAllStringFunc(html, func(match string) string {
		submatch := re.FindStringSubmatch(match)
		if len(submatch) < 4 {
			return match
		}

		prefix := submatch[1]
		src := submatch[2]

		// Skip external URLs and absolute paths
		if strings.HasPrefix(src, "http://") || strings.HasPrefix(src, "https://") || strings.HasPrefix(src, "/") {
			return match
		}

		// Calculate fullOldPath from root directory
		fullOldPath := filepath.Join(filepath.Dir(s.markdownPath), src)
		if modifyPath != nil {
			fullOldPath = modifyPath(fullOldPath)
		}

		newPath, err := relpathResolver.Resolve(fullOldPath, HTMLPath)
		if err != nil {
			errs = errors.Join(errs, err)
			return match
		}

		return fmt.Sprintf(`%s%s"`, prefix, newPath)
	})

	return result, errs
}

func (s substituter) convertMdLinksPath(html string, HTMLPath string) (string, error) {
	re := regexp.MustCompile(`(href=")([^"]*\.md)(")`)
	return s.replacePaths(html, HTMLPath, re, s.pagePathsResolver, nil)
}

func (s substituter) convertAssetsPath(html string, HTMLPath string) (string, error) {
	re := regexp.MustCompile(`(<img[^>]+src=")([^"]+)(")`)
	return s.replacePaths(html, HTMLPath, re, s.assetPathsResolver, nil)
}
