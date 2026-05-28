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

func (g *Generator) generatePages() error {
	assetsPathTranslater := NewPathResolver(g.assetsDir, filepath.Join(g.buildDir, "assets"))
	linksPathTranslater := NewPathResolver(g.contentDir, g.buildDir)

	errs := make([]error, 0)
	err := afero.Walk(g.fs, g.contentDir, func(markDownFilePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Only Handling markdown files
		if !strings.HasSuffix(markDownFilePath, ".md") {
			errs = append(errs, fmt.Errorf("wrong extension for file in %s", markDownFilePath))
			return nil
		}

		// Page section is the directory between content dir and file name
		pageSection, err := section.ExtractSection(g.contentDir, markDownFilePath)
		if err != nil {
			errs = append(errs, err)
			return nil
		}

		pageFilePathRelToContentDir, err := filepath.Rel(g.contentDir, markDownFilePath)
		if err != nil {
			return fmt.Errorf("cannot compute relative path of %s from %s: %w", markDownFilePath, g.contentDir, err)
		}

		htmlOutputPath := filepath.Join(g.buildDir, strings.TrimSuffix(pageFilePathRelToContentDir, ".md")+".html")
		g.pagesGenerators = append(g.pagesGenerators, g.pageGeneratorFactory(markDownFilePath, htmlOutputPath, g.buildDir, pageSection, assetsPathTranslater, linksPathTranslater, g.sections, g.skipURLValidation))
		return nil
	})

	if err != nil {
		errs = append(errs, err)
	}

	for _, generator := range g.pagesGenerators {
		if err := generator.Generate(); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) != 0 {
		// Empty the page generators
		g.pagesGenerators = make([]PageGenerator, 0)
		return errors.Join(errs...)
	}

	return nil
}

func newPageGeneratorFactory(fs afero.Fs) pageGeneratorFactory {
	return func(sourceMDPath, destinationHTMLPath, buildDir, pageSection string, assetsPathTranslater, linksPathTranslater newPathResolver, sections []section.Section, skipURLValidation bool) PageGenerator {
		var (
			markdownSubstitutions = mdsubstitutions.NewRegistry(sourceMDPath, fs)
			HTMLSubstitutions     = htmlsubstitutions.NewRegistry(destinationHTMLPath, sourceMDPath, assetsPathTranslater, linksPathTranslater, sections, pageSection)
			validations           = validation.NewRegistry(fs, sections, skipURLValidation)
		)
		return page.NewGenerator(buildDir, pageSection, fs, markdownSubstitutions, markdown.NewConverter(), HTMLSubstitutions, validations)
	}
}
