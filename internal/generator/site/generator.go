package site

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/tjnvr/blog/internal/generator/section"

	"github.com/spf13/afero"
)

type (
	PageGenerator interface {
		Generate(fromMarkdownFilePath, toHTMLFilePath string) error

		Validate(HTMLFilePath string) error
	}

	pageGeneratorFactory func(sourceMDPath, toHTMLPath, buildDir, pageSection string, assetsPathTranslater, linksPathTranslater newPathResolver, sections []section.Section, skipURLValidation bool) PageGenerator

	Option func(*Generator)
)

type Generator struct {
	// All files and directories attributes are relative to the project root.
	fs                                    afero.Fs
	skipURLValidation                     bool
	pageGeneratorFactory                  pageGeneratorFactory
	sections                              []section.Section
	HTMLFilePathsBySourceMarkdownFilePath map[string]string
	generatorsBySourceMarkdownFilePath    map[string]PageGenerator
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
	if err := g.listSections(contentDir); err != nil {
		return fmt.Errorf("failed to list site sections: %w", err)
	}

	if err := copyDir(g.fs, assetsDir, filepath.Join(buildDir, "assets")); err != nil {
		return fmt.Errorf("failed to copy assets: %w", err)
	}

	if err := copyDir(g.fs, scriptsDir, filepath.Join(buildDir, "scripts")); err != nil {
		return fmt.Errorf("failed to copy assets: %w", err)
	}

	if err := g.generatePages(contentDir, assetsDir, buildDir); err != nil {
		return fmt.Errorf("failed to generate pages: %w", err)
	}

	return nil
}

func (g *Generator) Validate() error {
	errs := make([]error, 0)
	for markdownFilePath, HTMLFilePath := range g.HTMLFilePathsBySourceMarkdownFilePath {
		pageGenerator, ok := g.generatorsBySourceMarkdownFilePath[markdownFilePath]
		if !ok {
			return fmt.Errorf("cannot find a page generator for %s", HTMLFilePath)
		}
		if err := pageGenerator.Validate(HTMLFilePath); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}
