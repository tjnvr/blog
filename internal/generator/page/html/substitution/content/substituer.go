package content

import (
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

type (
	PathTranslater interface {
		Resolve(oldPath, fromPath string) (string, error)
	}

	Substituter struct {
		filePath              string
		markdownSourcePath    string
		assetsPathsTranslater PathTranslater
		linksPathTranslater   PathTranslater
	}
)

func NewSubstituer(filePath, markdownSourcePath string, assetsPathsTranslater PathTranslater, linksPathTranslater PathTranslater) Substituter {
	return Substituter{
		filePath:              filePath,
		markdownSourcePath:    markdownSourcePath,
		assetsPathsTranslater: assetsPathsTranslater,
		linksPathTranslater:   linksPathTranslater,
	}
}

func (s Substituter) Placeholder() string {
	return "{{content}}"
}

func (s Substituter) Resolve(htmlContent string) (string, error) {
	var errs error
	var err error

	htmlContent, err = s.convertMdLinksPath(htmlContent, s.filePath)
	if err != nil {
		errs = errors.Join(errs, fmt.Errorf("s.convertMdLinksPath err: %w", err))
	}

	htmlContent, err = s.convertAssetsPath(htmlContent, s.filePath)
	if err != nil {
		errs = errors.Join(errs, fmt.Errorf("s.convertAssetsPath err: %w", err))
	}

	return htmlContent, errs
}

func (s Substituter) replacePaths(html, filePath string, re *regexp.Regexp, translater PathTranslater, modifyPath func(string) string) (string, error) {
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
		fullOldPath := filepath.Join(filepath.Dir(s.markdownSourcePath), src)
		if modifyPath != nil {
			fullOldPath = modifyPath(fullOldPath)
		}

		newPath, err := translater.Resolve(fullOldPath, filePath)
		if err != nil {
			errs = errors.Join(errs, err)
			return match
		}

		return fmt.Sprintf(`%s%s"`, prefix, newPath)
	})

	return result, errs
}

func (s Substituter) convertMdLinksPath(html string, filePath string) (string, error) {
	re := regexp.MustCompile(`(href=")([^"]*\.md)(")`)
	return s.replacePaths(html, filePath, re, s.linksPathTranslater, func(p string) string {
		return strings.TrimSuffix(p, ".md") + ".html"
	})
}

func (s Substituter) convertAssetsPath(html string, filePath string) (string, error) {
	re := regexp.MustCompile(`(<img[^>]+src=")([^"]+)(")`)
	return s.replacePaths(html, filePath, re, s.assetsPathsTranslater, nil)
}
