package site

import (
	"errors"
	"fmt"

	"github.com/timtimjnvr/blog/internal/generator/page/filesystem"
	"github.com/timtimjnvr/blog/internal/generator/page/styling"
	"github.com/timtimjnvr/blog/internal/generator/section"
)

type (
	PageGenerator interface {
		Generate() error
		Validate() error
	}

	pageGeneratorFactory func(sourceMDPath, destinationHTMLPath, buildDir, pageSection string, assetsPathTranslater, linksPathTranslater newPathResolver, stylingConfig *styling.Config, sections []section.Section, skipURLValidation bool) PageGenerator

	Option func(*Generator)
)

// Generator is the site generator which allows to generate and validate the site
// All files and directories attributes are relative to the project root.
type Generator struct {
	contentDir                string
	assetsDir                 string
	assetsOutDir              string
	optionalStylingConfigPath string
	buildDir                  string
	scriptsDir                string
	scriptsOutDir             string
	skipURLValidation         bool
	pageGeneratorFactory      pageGeneratorFactory
	sections                  []section.Section
	pagesGenerators           []PageGenerator
	stylingConfig             *styling.Config
	fs                        filesystem.FileSystem
}

// WithSkipURLValidation returns an Option that disables external URL validation.
func WithSkipURLValidation(skip bool) Option {
	return func(g *Generator) { g.skipURLValidation = skip }
}

func NewGenerator(opts ...Option) (*Generator, error) {
	g := &Generator{
		contentDir:                "./content/markdown",
		buildDir:                  "./target/build",
		assetsDir:                 "./content/assets",
		assetsOutDir:              "./target/build/assets",
		scriptsDir:                "./scripts",
		scriptsOutDir:             "./target/build/scripts",
		optionalStylingConfigPath: "./styles/styles.json",
		sections:                  make([]section.Section, 0),
		pagesGenerators:           make([]PageGenerator, 0),
		pageGeneratorFactory:      defaultPageGeneratorFactory,
		fs:                        filesystem.NewOSFileSystem(),
	}

	for _, opt := range opts {
		opt(g)
	}

	return g, nil
}

func (g *Generator) Generate() error {
	if err := g.loadStylingConfig(); err != nil {
		return fmt.Errorf("failed to load styling configuration: %w", err)
	}

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
