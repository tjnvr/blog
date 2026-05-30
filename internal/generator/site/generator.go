package site

import (
	"errors"
	"fmt"

	"github.com/tjnvr/blog/internal/generator/page"
	"github.com/tjnvr/blog/internal/generator/page/html/substitution/content"
	"github.com/tjnvr/blog/internal/generator/section"

	"github.com/spf13/afero"
)

type (
	PageGenerator interface {
		Generate(fromMarkdownFilePath, toHTMLFilePath string) error

		Validate(HTMLFilePath string) error
	}

	pageGeneratorFactory func(sourceMDPath, toHTMLPath, buildDir, pageSection string, assetsPathTranslater, linksPathTranslater pathResolver, sections []section.Section, skipURLValidation bool) PageGenerator

	sectionsLister interface {
		ListSections(fs afero.Fs, sectionExtractor sectionExtractor, contentDir string) ([]section.Section, error)
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
		Create(toHTMLFilePath string, fromMarkdownFilePath string, assetsPathTranslater, linksPathTranslater content.PathTranslater, sections []section.Section, pageSection string) page.HTMLSubstituer
	}

	HTMLValidationFactory interface {
		Create(fs afero.Fs, destinationHTMLPath, buildDir string, sections []section.Section, skipURLValidation bool) page.PageValidator
	}

	sectionExtractor func(dir, filePath string) (string, error)

	Option func(*Generator)
)

type Generator struct {
	// All files and directories attributes are relative to the project root.
	fs             afero.Fs
	sectionsLister sectionsLister
	dirCopier      dirCopier
	pagesLister    pagesLister

	markdownToHTMLPathTranslater pathTranslater
	sectionExtractor             sectionExtractor

	markdownSubstituerFactory  markdownSubstituerFactory
	markdownConverterFactory   markdownConverterFactory
	HTMLSubstituerFactory      HTMLSubstituerFactory
	assetsPathResolverFactory  pathResolverFactory
	scriptsPathResolverFactory pathResolverFactory
	HTMLValidationFactory      HTMLValidationFactory

	skipURLValidation bool

	pageGeneratorFactory               pageGeneratorFactory
	sections                           []section.Section
	generatorsBySourceMarkdownFilePath map[string]PageGenerator
}

// WithSkipURLValidation returns an Option that disables external URL validation.
func WithSkipURLValidation(skip bool) Option {
	return func(g *Generator) { g.skipURLValidation = skip }
}

func NewGenerator(fs afero.Fs, opts ...Option) (*Generator, error) {
	g := &Generator{
		sections:                           make([]section.Section, 0),
		pageGeneratorFactory:               newPageGeneratorFactory(fs),
		generatorsBySourceMarkdownFilePath: make(map[string]PageGenerator),
		fs:                                 fs,
	}

	for _, opt := range opts {
		opt(g)
	}

	return g, nil
}

// Generate process markdown files to gerne
//
// sourceMarkdownFilePath is the markdown file content to process
// destinationHTMLFilePath is the HTML file that will be produced
func (g *Generator) Generate(contentDir, buildDir, assetsDir, assetsOutDir, scriptsDir, scriptsOutDir string) error {
	sections, err := g.sectionsLister.ListSections(g.fs, g.sectionExtractor, contentDir)
	if err != nil {
		return fmt.Errorf("ListSections err: %v", err)
	}

	if err := g.dirCopier.CopyDir(g.fs, assetsDir, assetsOutDir); err != nil {
		return fmt.Errorf("dirCopier err: %v", err)
	}

	if err := g.dirCopier.CopyDir(g.fs, scriptsDir, scriptsOutDir); err != nil {
		return fmt.Errorf("dirCopier err: %v", err)
	}

	pagesFilePaths, err := g.pagesLister.ListPages(g.fs, contentDir)
	if err != nil {
		return fmt.Errorf("ListPages err: %v", err)
	}

	pagesGeneratorsByPath := make(map[string]*page.Generator)
	for _, markdownPageFilePath := range pagesFilePaths {
		markdownSubstituer := g.markdownSubstituerFactory.Create(g.fs, g.markdownToHTMLPathTranslater.GetNewPath(markdownPageFilePath))
		markdownToHTMLConverter := g.markdownConverterFactory.Create()
		assetsPathResolver := g.assetsPathResolverFactory.Create(assetsDir, assetsOutDir)
		scriptsPathResolver := g.scriptsPathResolverFactory.Create(scriptsDir, scriptsOutDir)
		section, err := g.sectionExtractor(contentDir, markdownPageFilePath)
		if err != nil {
			return fmt.Errorf("sectionExtractor err: %v", err)
		}

		HTMLSubstituer := g.HTMLSubstituerFactory.Create(
			g.markdownToHTMLPathTranslater.GetNewPath(markdownPageFilePath),
			markdownPageFilePath,
			assetsPathResolver,
			scriptsPathResolver,
			sections,
			section,
		)
		HTMLValidator := g.HTMLValidationFactory.Create(g.fs, g.markdownToHTMLPathTranslater.GetNewPath(markdownPageFilePath), buildDir, sections, g.skipURLValidation)
		pagesGeneratorsByPath[markdownPageFilePath] = page.NewGenerator(g.fs, markdownSubstituer, markdownToHTMLConverter, HTMLSubstituer, HTMLValidator)
	}

	errs := make([]error, 0)
	for markdownPageFilePath, pg := range pagesGeneratorsByPath {
		if err := pg.Generate(markdownPageFilePath, g.markdownToHTMLPathTranslater.GetNewPath(markdownPageFilePath)); err != nil {
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
		if err := pg.Validate(g.markdownToHTMLPathTranslater.GetNewPath(markdownPageFilePath)); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}
