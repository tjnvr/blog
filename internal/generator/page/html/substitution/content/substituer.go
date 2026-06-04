package content

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/tjnvr/blog/internal/generator/backbone/relpath"
)

type (
	// Substituter resolves the {{content}} template placeholder
	// it replaces links and assets with their real path in the build directory
	Substituter struct {
		filePath              string
		markdownSourcePath    string
		assetsPathsTranslater path.Resolver
		linksPathTranslater   path.Resolver
	}
)

func NewSubstituer(filePath, markdownSourcePath string, assetsPathsTranslater, linksPathTranslater path.Resolver) Substituter {
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
	var err error
	htmlContent, err = s.convertMdLinksPath(htmlContent, s.filePath)
	if err != nil {
		return "", fmt.Errorf("s.convertMdLinksPath err: %w", err)
	}
	htmlContent, err = s.convertAssetsPath(htmlContent, s.filePath)
	if err != nil {
		return "", fmt.Errorf("s.convertAssetsPath err: %w", err)
	}
	return htmlContent, nil
}

func (s Substituter) convertMdLinksPath(html string, filePath string) (string, error) {
	// Match href attributes pointing to .md files
	re := regexp.MustCompile(`(href=")([^"]*\.md)(")`)
	var firstErr error
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

		// From root directory
		fullOldPath := strings.TrimSuffix(filepath.Join(filepath.Dir(s.markdownSourcePath), src), ".md") + ".html"

		newPath, err := s.linksPathTranslater.Resolve(fullOldPath, filePath)
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			return match
		}
		return fmt.Sprintf(`%s%s"`, prefix, newPath)
	})

	return result, firstErr
}

func (s Substituter) convertAssetsPath(html string, filePath string) (string, error) {
	// Match img src attributes with relative paths
	re := regexp.MustCompile(`(<img[^>]+src=")([^"]+)(")`)
	var firstErr error
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

		// From root directory
		fullOldPath := filepath.Join(filepath.Dir(s.markdownSourcePath), src)

		newPath, err := s.assetsPathsTranslater.Resolve(fullOldPath, filePath)
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			return match
		}
		return fmt.Sprintf(`%s%s"`, prefix, newPath)
	})
	return result, firstErr
}
