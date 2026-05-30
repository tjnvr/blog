package site

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/tjnvr/blog/internal/generator/section"
)

type skeleton struct {
	contentDir,
	buildDir,
	assetsDir,
	assetsOutDir,
	scriptsDir,
	scriptsOutDir string
}

func NewSkeleton(
	contentDir,
	buildDir,
	assetsDir,
	assetsOutDir,
	scriptsDir,
	scriptsOutDir string,
) (skeleton, error) {

	// verify that all inputs are valid
	panic("TODO: not implemented")

	return skeleton{
		contentDir:    contentDir,
		buildDir:      buildDir,
		assetsDir:     assetsDir,
		assetsOutDir:  assetsOutDir,
		scriptsDir:    scriptsDir,
		scriptsOutDir: scriptsOutDir,
	}, nil
}

func ExtractSection(markdownSourcePath string) (section.Section, error) {
	panic("TODO: not implemented")
	return section.Section{}, nil
}

func (s skeleton) GetHTMLFilePath(markdownFilePath string) (string, error) {
	pageFilePathRelToContentDir, err := filepath.Rel(s.contentDir, markdownFilePath)
	if err != nil {
		return "", fmt.Errorf("cannot compute relative path of %s from %s: %w", markdownFilePath, s.contentDir, err)
	}

	return filepath.Join(s.buildDir, strings.TrimSuffix(pageFilePathRelToContentDir, ".md")+".html"), nil
}

func GetAssetFilePath(assetFilePath string) (string, error) {
	panic("TODO: not implemented")
	return "", nil
}

func GetScriptFilePath(scriptFilePath string) (string, error) {
	panic("TODO: not implemented")
	return "", nil
}
