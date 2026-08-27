package site

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/spf13/afero"
	"github.com/tjnvr/blog/internal/abspath"
	"github.com/tjnvr/blog/internal/backbone/metadata"
	"github.com/tjnvr/blog/internal/backbone/section"
	"github.com/tjnvr/blog/internal/generator/page"
	"github.com/tjnvr/blog/internal/generator/page/html/validation"
	"github.com/tjnvr/blog/internal/hrefpath"
	"github.com/tjnvr/blog/internal/relpath"
)

func (g *Generator) generatePages(assetsPathTranslater, linksPathTranslater relpath.Resolver, hrefPathTranslater hrefpath.Resolver) ([]sitemapEntry, error) {
	errs := make([]error, 0)
	entries := make([]sitemapEntry, 0)
	pagePaths, err := g.pagesFinder.FindFiles(g.ContentDir)
	if err != nil {
		return nil, fmt.Errorf("pagesFinder.FindFiles err: %v", err)
	}

	for _, pagePath := range pagePaths {
		pagePath := filepath.Join(g.ContentDir, pagePath)
		HTMLPath, err := linksPathTranslater.Resolve(pagePath, ".")
		if err != nil {
			errs = append(errs, fmt.Errorf("Resolve('%s') err: %v", pagePath, err))
			continue
		}

		href, err := hrefPathTranslater.Resolve(pagePath)
		if err != nil {
			errs = append(errs, fmt.Errorf("Resolve('%s') err: %v", pagePath, err))
			continue
		}

		markdownContent, err := afero.ReadFile(g.fs, pagePath)
		if err != nil {
			errs = append(errs, fmt.Errorf("reading '%s' err: %v", pagePath, err))
			continue
		}
		entries = append(entries, sitemapEntry{href: href, lastMod: metadata.Extract(markdownContent).CreationDate})

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
		return nil, errors.Join(errs...)
	}

	return entries, nil
}

func defaultPageGeneratorFactory(fs afero.Fs, sourceMDPath, destinationHTMLPath string, absolutePathsResolverFactory abspath.ResolverFactory, sectionResolver section.Resolver, hrefPathsResolver hrefpath.Resolver, assetPathsResolver relpath.Resolver, skipURLValidation bool) PageGenerator {
	validations := validation.NewRegistry(destinationHTMLPath, fs, absolutePathsResolverFactory, sectionResolver, hrefPathsResolver, skipURLValidation)

	return page.NewGenerator(sourceMDPath, destinationHTMLPath, fs, sectionResolver, hrefPathsResolver, assetPathsResolver, validations)
}
