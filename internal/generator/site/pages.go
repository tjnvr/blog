package site

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
	"github.com/tjnvr/blog/internal/generator/page"
	htmlsubstitutions "github.com/tjnvr/blog/internal/generator/page/html/substitution"
	"github.com/tjnvr/blog/internal/generator/page/html/validation"
	mdsubstitutions "github.com/tjnvr/blog/internal/generator/page/markdown/substitution"
	"github.com/tjnvr/blog/internal/generator/section"
	"github.com/tjnvr/blog/internal/relpath"
)

func (g *Generator) generatePages(assetsPathTranslater, linksPathTranslater relpath.Resolver) error {
	errs := make([]error, 0)

	// Modified: Using afero.Walk to navigate the virtual/real file system seamlessly
	err := afero.Walk(g.fs, g.ContentDir, func(markDownFilePath string, info os.FileInfo, err error) error {
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
		pageSection, err := extractSection(g.ContentDir, markDownFilePath)
		if err != nil {
			errs = append(errs, err)
			return nil
		}

		pageFilePathRelToContentDir, err := filepath.Rel(g.ContentDir, markDownFilePath)
		if err != nil {
			return fmt.Errorf("cannot compute relative path of %s from %s: %w", markDownFilePath, g.ContentDir, err)
		}

		htmlOutputPath := filepath.Join(g.BuildDir, strings.TrimSuffix(pageFilePathRelToContentDir, ".md")+".html")
		g.pagesGenerators = append(g.pagesGenerators, g.pageGeneratorFactory(g.fs, markDownFilePath, htmlOutputPath, g.BuildDir, pageSection, assetsPathTranslater, linksPathTranslater, g.sections, g.skipURLValidation))
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

func defaultPageGeneratorFactory(fs afero.Fs, sourceMDPath, destinationHTMLPath, buildDir, pageSection string, assetsPathTranslater, linksPathTranslater relpath.Resolver, sections []section.Section, skipURLValidation bool) PageGenerator {
	var (
		markdownSubstitutions = mdsubstitutions.NewRegistry(sourceMDPath)
		HTMLSubstitutions     = htmlsubstitutions.NewRegistry(destinationHTMLPath, sourceMDPath, assetsPathTranslater, linksPathTranslater, sections, pageSection)
		validations           = validation.NewRegistry(sections, skipURLValidation)
	)

	return page.NewGenerator(sourceMDPath, destinationHTMLPath, buildDir, pageSection, fs, markdownSubstitutions, HTMLSubstitutions, validations)
}
