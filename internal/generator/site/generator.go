package site

import (
	"errors"
	"fmt"

	"github.com/spf13/afero"
	"github.com/tjnvr/blog/internal/generator/section"
	mfs "github.com/tjnvr/blog/internal/io/fs"
	"github.com/tjnvr/blog/internal/relpath"
)

type (
	PageGenerator interface {
		Generate() error
		Validate() error
	}

	pageGeneratorFactory func(fs afero.Fs, sourceMDPath, destinationHTMLPath, buildDir, pageSection string, assetsPathTranslater, linksPathTranslater relpath.Resolver, sections []section.Section, skipURLValidation bool) PageGenerator

	Option func(*Generator)
)

// All files and directories attributes are relative to the project root directory.
type Config struct {
	ContentDir    string
	BuildDir      string
	AssetsDir     string
	AssetsOutDir  string
	ScriptsDir    string
	ScriptsOutDir string
}

// Generator is the site generator which allows to generate and validate the site
type Generator struct {
	Config
	skipURLValidation    bool
	dirCopier            mfs.DirCopier
	pageGeneratorFactory pageGeneratorFactory
	sections             []section.Section
	pagesGenerators      []PageGenerator
	fs                   afero.Fs
}

// WithSkipURLValidation returns an Option that disables external URL validation.
func WithSkipURLValidation(skip bool) Option {
	return func(g *Generator) { g.skipURLValidation = skip }
}

func NewGenerator(fs afero.Fs, cfg Config, opts ...Option) (*Generator, error) {
	if cfg.ContentDir == "" {
		return nil, errors.New("ContentDir is mandatory")
	}
	if cfg.BuildDir == "" {
		return nil, errors.New("BuildDir is mandatory")
	}
	if cfg.AssetsDir == "" {
		return nil, errors.New("AssetsDir is mandatory")
	}
	if cfg.AssetsOutDir == "" {
		return nil, errors.New("AssetsOutDir is mandatory")
	}
	if cfg.ScriptsDir == "" {
		return nil, errors.New("ScriptsDir is mandatory")
	}
	if cfg.ScriptsOutDir == "" {
		return nil, errors.New("ScriptsOutDir is mandatory")
	}

	g := &Generator{
		Config:               cfg,
		sections:             make([]section.Section, 0),
		pagesGenerators:      make([]PageGenerator, 0),
		dirCopier:            mfs.NewDirCopier(fs),
		pageGeneratorFactory: defaultPageGeneratorFactory,
		fs:                   fs,
	}

	for _, opt := range opts {
		opt(g)
	}

	return g, nil
}

func (g *Generator) Generate() error {
	assetsPathTranslater := relpath.NewResolver(g.AssetsDir, g.AssetsOutDir)
	linksPathTranslater := relpath.NewResolver(g.ContentDir, g.BuildDir)

	if err := g.listSections(); err != nil {
		return fmt.Errorf("failed to list site sections: %w", err)
	}

	if err := g.dirCopier.CopyDir(mfs.NewFilesFinder(g.fs), g.AssetsDir, g.AssetsOutDir); err != nil {
		return fmt.Errorf("CopyDir assets err: %v", err)
	}

	// Fixed: Copying JS scripts from ScriptsDir to ScriptsOutDir instead of AssetsDir
	if err := g.dirCopier.CopyDir(mfs.NewFilesFinder(g.fs, mfs.WithExtension(".js")), g.ScriptsDir, g.ScriptsOutDir); err != nil {
		return fmt.Errorf("CopyDir scripts err: %v", err)
	}

	if err := g.generatePages(assetsPathTranslater, linksPathTranslater); err != nil {
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
