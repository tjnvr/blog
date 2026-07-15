package site

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/spf13/afero"
	"github.com/tjnvr/blog/internal/generator/backbone/section"
	"github.com/tjnvr/blog/internal/generator/page"
	htmlsubstitutions "github.com/tjnvr/blog/internal/generator/page/html/substitution"
	"github.com/tjnvr/blog/internal/generator/page/html/validation"
	mdsubstitutions "github.com/tjnvr/blog/internal/generator/page/markdown/substitution"
	"github.com/tjnvr/blog/internal/relpath"
)

func (g *Generator) generatePages(assetsPathTranslater, linksPathTranslater relpath.Resolver) error {
	errs := make([]error, 0)
	pagePaths, err := g.pagesFinder.FindFiles(g.ContentDir)
	if err != nil {
		return fmt.Errorf("pagesFinder.FindFiles err: %v", err)
	}

	for _, pagePath := range pagePaths {
		pagePath := filepath.Join(g.ContentDir, pagePath)
		HTMLPath, err := linksPathTranslater.Resolve(pagePath, ".")
		if err != nil {
			errs = append(errs, fmt.Errorf("Resolve('%s') err: %v", pagePath, err))
			continue
		}

		g.pagesGenerators = append(g.pagesGenerators, g.pageGeneratorFactory(g.fs, pagePath, HTMLPath, g.BuildDir, g.sectionResolver, linksPathTranslater, assetsPathTranslater, g.skipURLValidation))
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

func defaultPageGeneratorFactory(fs afero.Fs, sourceMDPath, destinationHTMLPath, buildDir string, sectionResolver section.Resolver, pagePathsResolver, assetPathsResolver relpath.Resolver, skipURLValidation bool) PageGenerator {
	var (
		markdownSubstitutions = mdsubstitutions.NewRegistry(sourceMDPath)
		HTMLSubstitutions     = htmlsubstitutions.NewRegistry(destinationHTMLPath, sourceMDPath, sectionResolver, pagePathsResolver, assetPathsResolver)
		validations           = validation.NewRegistry(destinationHTMLPath, buildDir, fs, sectionResolver, pagePathsResolver, skipURLValidation)
	)

	return page.NewGenerator(sourceMDPath, destinationHTMLPath, fs, markdownSubstitutions, HTMLSubstitutions, validations)
}
