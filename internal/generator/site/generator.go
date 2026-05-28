package site

import (
	"errors"
	"fmt"

	"github.com/tjnvr/blog/internal/generator/section"

	"github.com/spf13/afero"
)

type (
	PageGenerator interface {
		Generate(sourceMarkdownFilePath, destinationHTMLFilePath string) error
		Validate(HTMLFilePath string) error
	}

	pageGeneratorFactory func(buildDir, pageSection string, assetsPathTranslater, linksPathTranslater newPathResolver, sections []section.Section, skipURLValidation bool) PageGenerator

	Option func(*Generator)
)

// Generator is the site generator which allows to generate and validate the site
// All files and directories attributes are relative to the project root.
type Generator struct {
	contentDir           string
	assetsDir            string
	assetsOutDir         string
	buildDir             string
	scriptsDir           string
	scriptsOutDir        string
	skipURLValidation    bool
	pageGeneratorFactory pageGeneratorFactory
	sections             []section.Section
	pagesGenerators      []PageGenerator
	fs                   afero.Fs
}

// WithSkipURLValidation returns an Option that disables external URL validation.
func WithSkipURLValidation(skip bool) Option {
	return func(g *Generator) { g.skipURLValidation = skip }
}

func NewGenerator(fs afero.Fs, opts ...Option) (*Generator, error) {
	g := &Generator{
		contentDir:           "./content/markdown",
		buildDir:             "./target/build",
		assetsDir:            "./content/assets",
		assetsOutDir:         "./target/build/assets",
		scriptsDir:           "./scripts",
		scriptsOutDir:        "./target/build/scripts",
		sections:             make([]section.Section, 0),
		pagesGenerators:      make([]PageGenerator, 0),
		pageGeneratorFactory: newPageGeneratorFactory(fs),
		fs:                   fs,
	}

	for _, opt := range opts {
		opt(g)
	}

	return g, nil
}

func (g *Generator) Generate() error {
	if err := g.makeAllDirectories(); err != nil {
		return fmt.Errorf("failed to create output directories: %w", err)
	}

	if err := g.listSections(); err != nil {
		return fmt.Errorf("failed to list site sections: %w", err)
	}

	if err := g.copyAssets(); err != nil {
		return fmt.Errorf("failed to copy assets: %w", err)
	}

	if err := g.copyScripts(); err != nil {
		return fmt.Errorf("failed to copy scripts: %w", err)
	}

	if err := g.generatePages(); err != nil {
		return fmt.Errorf("failed to generate pages: %w", err)
	}

	return nil
}

func (g *Generator) Validate() error {
	errs := make([]error, 0)
	for _, pg := range g.pagesGenerators {
		if err := pg.Validate(); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}
