package site

import (
	"errors"
	"fmt"

	"github.com/tjnvr/blog/internal/generator/backbone/htmlpath"
	"github.com/tjnvr/blog/internal/generator/backbone/pages"
	"github.com/tjnvr/blog/internal/generator/backbone/relpath"
	"github.com/tjnvr/blog/internal/generator/backbone/section"
	"github.com/tjnvr/blog/internal/generator/page"

	"github.com/spf13/afero"
)

type (
	PageGenerator interface {
		Generate(fromMarkdownFilePath, toHTMLFilePath string) error

		Validate(HTMLFilePath string) error
	}

	dirCopier interface {
		CopyDir(fs afero.Fs, from, to string) error
	}

	pagesLister interface {
		ListPages(fs afero.Fs, contentDir string) ([]string, error)
	}

	markdownSubstituerFactory interface {
		Create(fs afero.Fs, filePath string) page.MarkdownSubstituer
	}

	markdownConverterFactory interface {
		Create() page.MarkdownConverter
	}

	pathTranslater interface {
		GetNewPath(filePath string) string
	}

	HTMLSubstituerFactory interface {
		Create(
			toHTMLFilePath string,
			fromMarkdownFilePath string,
			assetsPathTranslater,
			linksPathTranslater relpath.Resolver,
			sectionExtractor section.Resolver,
		) page.HTMLSubstituer
	}

	HTMLValidationFactory interface {
		Create(
			fs afero.Fs,
			destinationHTMLPath,
			buildDir string,
			sections []section.Section,
			skipURLValidation bool,
		) page.PageValidator
	}

	pageGeneratorFactory interface {
		Create(
			fs afero.Fs,
			markdownSubstituer page.MarkdownSubstituer,
			markdownToHTMLConverter page.MarkdownConverter,
			HTMLSubstituer page.HTMLSubstituer,
			HTMLValidator page.PageValidator,
		) PageGenerator
	}
)

type Generator struct {
	// Site generator dependencies
	fs        afero.Fs
	dirCopier dirCopier

	// Site generator dependencies
	pagesResolver    pages.Resolver
	HTMLPathResolver htmlpath.Resolver
	sectionResolver  section.Resolver

	// Page generator dependencies
	markdownSubstituerFactory  markdownSubstituerFactory
	markdownConverterFactory   markdownConverterFactory
	HTMLSubstituerFactory      HTMLSubstituerFactory
	assetsPathResolverFactory  relpath.ResolverFactory
	scriptsPathResolverFactory relpath.ResolverFactory
	HTMLValidationFactory      HTMLValidationFactory
	pageGeneratorFactory       pageGeneratorFactory

	// Site generator options
	skipURLValidation bool

	generatorsBySourceMarkdownFilePath map[string]PageGenerator
}

func NewGenerator(
	fs afero.Fs,
	dirCopier dirCopier,
	pagesResolver pages.Resolver,
	HTMLPathResolver htmlpath.Resolver,
	sectionResolver section.Resolver,
	markdownSubstituerFactory markdownSubstituerFactory,
	markdownConverterFactory markdownConverterFactory,
	HTMLSubstituerFactory HTMLSubstituerFactory,
	assetsPathResolverFactory relpath.ResolverFactory,
	scriptsPathResolverFactory relpath.ResolverFactory,
	HTMLValidationFactory HTMLValidationFactory,
	pageGeneratorFactory pageGeneratorFactory,
	opts ...Option,
) (*Generator, error) {
	g := &Generator{
		fs:                                 fs,
		dirCopier:                          dirCopier,
		pagesResolver:                      pagesResolver,
		HTMLPathResolver:                   HTMLPathResolver,
		sectionResolver:                    sectionResolver,
		markdownSubstituerFactory:          markdownSubstituerFactory,
		markdownConverterFactory:           markdownConverterFactory,
		HTMLSubstituerFactory:              HTMLSubstituerFactory,
		assetsPathResolverFactory:          assetsPathResolverFactory,
		scriptsPathResolverFactory:         scriptsPathResolverFactory,
		HTMLValidationFactory:              HTMLValidationFactory,
		pageGeneratorFactory:               pageGeneratorFactory,
		generatorsBySourceMarkdownFilePath: make(map[string]PageGenerator),
	}

	for _, opt := range opts {
		opt(g)
	}

	return g, nil
}

// Generate process markdown files to generate HTML pages
func (g *Generator) Generate(contentDir, buildDir, assetsDir, assetsOutDir, scriptsDir, scriptsOutDir string) error {
	if err := g.dirCopier.CopyDir(g.fs, assetsDir, assetsOutDir); err != nil {
		return fmt.Errorf("dir copy err: %v", err)
	}

	if err := g.dirCopier.CopyDir(g.fs, scriptsDir, scriptsOutDir); err != nil {
		return fmt.Errorf("dir copy err: %v", err)
	}

	pagesFilePaths, err := g.pagesResolver.ResolveAll()
	if err != nil {
		return fmt.Errorf("pages listing err: %v", err)
	}

	pagesGeneratorsByPath := make(map[string]PageGenerator)
	for _, markdownPageFilePath := range pagesFilePaths {
		HTMLPageFilePath, err := g.HTMLPathResolver.Resolve(markdownPageFilePath)
		if err != nil {
			return fmt.Errorf("cannot resolve HTML path for page (%s): %v", markdownPageFilePath, err)
		}
		markdownSubstituer := g.markdownSubstituerFactory.Create(g.fs, HTMLPageFilePath)
		markdownToHTMLConverter := g.markdownConverterFactory.Create()
		assetsPathResolver := g.assetsPathResolverFactory.Create(assetsDir, assetsOutDir)
		scriptsPathResolver := g.scriptsPathResolverFactory.Create(scriptsDir, scriptsOutDir)
		HTMLSubstituer := g.HTMLSubstituerFactory.Create(
			HTMLPageFilePath,
			markdownPageFilePath,
			assetsPathResolver,
			scriptsPathResolver,
			g.sectionResolver,
		)
		HTMLValidator := g.HTMLValidationFactory.Create(g.fs, HTMLPageFilePath, buildDir, sections, g.skipURLValidation)
		pagesGeneratorsByPath[markdownPageFilePath] = g.pageGeneratorFactory.Create(g.fs, markdownSubstituer, markdownToHTMLConverter, HTMLSubstituer, HTMLValidator)
	}

	errs := make([]error, 0)
	for markdownPageFilePath, pg := range pagesGeneratorsByPath {
		HTMLPageFilePath, _ := g.HTMLPathResolver.Resolve(markdownPageFilePath)
		if err := pg.Generate(markdownPageFilePath, HTMLPageFilePath); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) != 0 {
		return errors.Join(errs...)
	}

	return nil
}

func (g *Generator) Validate() error {
	errs := make([]error, 0)
	for markdownPageFilePath, pg := range g.generatorsBySourceMarkdownFilePath {
		HTMLPageFilePath, err := g.HTMLPathResolver.Resolve(markdownPageFilePath)
		if err != nil {
			return fmt.Errorf("cannot resolve HTML file path for page (%s): %v", markdownPageFilePath, err)
		}

		if err := pg.Validate(HTMLPageFilePath); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}
