package site

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/spf13/afero"
	"github.com/tjnvr/blog/internal/abspath"
	"github.com/tjnvr/blog/internal/backbone/section"
	"github.com/tjnvr/blog/internal/generator/page"
	"github.com/tjnvr/blog/internal/generator/page/html/validation"
	"github.com/tjnvr/blog/internal/hrefpath"
	"github.com/tjnvr/blog/internal/relpath"
)

func (g *Generator) generatePages(assetsPathTranslater, linksPathTranslater relpath.Resolver, hrefPathTranslater hrefpath.Resolver) error {
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

		g.pagesGenerators = append(g.pagesGenerators, g.pageGeneratorFactory(g.fs, pagePath, HTMLPath, g.absolutePathsResolverFactory, g.sectionResolver, hrefPathTranslater, assetsPathTranslater, g.skipURLValidation))
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

func defaultPageGeneratorFactory(fs afero.Fs, sourceMDPath, destinationHTMLPath string, absolutePathsResolverFactory abspath.ResolverFactory, sectionResolver section.Resolver, hrefPathsResolver hrefpath.Resolver, assetPathsResolver relpath.Resolver, skipURLValidation bool) PageGenerator {
	validations := validation.NewRegistry(destinationHTMLPath, fs, absolutePathsResolverFactory, sectionResolver, hrefPathsResolver, skipURLValidation)

	return page.NewGenerator(sourceMDPath, destinationHTMLPath, fs, sectionResolver, hrefPathsResolver, assetPathsResolver, validations)
}
