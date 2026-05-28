package site

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tjnvr/blog/internal/generator/page"
	htmlsubstitutions "github.com/tjnvr/blog/internal/generator/page/html/substitution"
	"github.com/tjnvr/blog/internal/generator/page/html/validation"
	"github.com/tjnvr/blog/internal/generator/page/markdown"
	mdsubstitutions "github.com/tjnvr/blog/internal/generator/page/markdown/substitution"
	"github.com/tjnvr/blog/internal/generator/section"

	"github.com/spf13/afero"
)

func (g *Generator) generatePages(contentDir, assetsDir, buildDir string) error {
	assetsPathTranslater := NewPathResolver(assetsDir, filepath.Join(buildDir, "assets"))
	linksPathTranslater := NewPathResolver(contentDir, buildDir)

	errs := make([]error, 0)
	HTMLFilePathsBySourceMarkdownFilePath := make(map[string]string)
	err := afero.Walk(g.fs, contentDir, func(fromMarkdownFilePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Only Handling markdown files
		if !strings.HasSuffix(fromMarkdownFilePath, ".md") {
			errs = append(errs, fmt.Errorf("wrong extension for file in %s", fromMarkdownFilePath))
			return nil
		}

		// Page section is the directory between content dir and file name
		pageSection, err := section.ExtractSection(contentDir, fromMarkdownFilePath)
		if err != nil {
			errs = append(errs, err)
			return nil
		}

		pageFilePathRelToContentDir, err := filepath.Rel(contentDir, fromMarkdownFilePath)
		if err != nil {
			return fmt.Errorf("cannot compute relative path of %s from %s: %w", fromMarkdownFilePath, contentDir, err)
		}

		toHTMLFilePath := filepath.Join(buildDir, strings.TrimSuffix(pageFilePathRelToContentDir, ".md")+".html")
		HTMLFilePathsBySourceMarkdownFilePath[fromMarkdownFilePath] = toHTMLFilePath
		g.generatorsBySourceMarkdownFilePath[fromMarkdownFilePath] = g.pageGeneratorFactory(
			fromMarkdownFilePath,
			toHTMLFilePath,
			buildDir,
			pageSection,
			assetsPathTranslater,
			linksPathTranslater,
			g.sections,
			g.skipURLValidation,
		)
		return nil
	})

	if err != nil {
		errs = append(errs, err)
	}

	for markdownSourceFilePath, HTMLFilePath := range HTMLFilePathsBySourceMarkdownFilePath {
		pageGenerator, ok := g.generatorsBySourceMarkdownFilePath[markdownSourceFilePath]
		if !ok {
			return fmt.Errorf("cannot find a page generator for %s", markdownSourceFilePath)
		}
		if err := pageGenerator.Generate(markdownSourceFilePath, HTMLFilePath); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) != 0 {
		// Empty the page generators
		g.generatorsBySourceMarkdownFilePath = make(map[string]PageGenerator, 0)
		return errors.Join(errs...)
	}

	return nil
}

func newPageGeneratorFactory(fs afero.Fs) pageGeneratorFactory {
	return func(sourceMDPath, destinationHTMLPath, buildDir, pageSection string, assetsPathTranslater, linksPathTranslater newPathResolver, sections []section.Section, skipURLValidation bool) PageGenerator {
		var (
			markdownSubstitutions = mdsubstitutions.NewRegistry(sourceMDPath, fs)
			HTMLSubstitutions     = htmlsubstitutions.NewRegistry(destinationHTMLPath, sourceMDPath, assetsPathTranslater, linksPathTranslater, sections, pageSection)
			validations           = validation.NewRegistry(fs, destinationHTMLPath, buildDir, sections, skipURLValidation)
		)
		return page.NewGenerator(fs, markdownSubstitutions, markdown.NewConverter(), HTMLSubstitutions, validations)
	}
}
