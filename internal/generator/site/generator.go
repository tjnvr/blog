package site

import (
	"errors"
	"fmt"

	"github.com/spf13/afero"
	"github.com/tjnvr/blog/internal/abspath"
	"github.com/tjnvr/blog/internal/backbone/section"
	"github.com/tjnvr/blog/internal/hrefpath"
	mfs "github.com/tjnvr/blog/internal/io/fs"
	"github.com/tjnvr/blog/internal/relpath"
)

type (
	PageGenerator interface {
		Generate() error
		Validate() error
	}

	pageGeneratorFactory func(fs afero.Fs, markdownPath, HTMLPath string, absolutePathsResolverFactory abspath.ResolverFactory, sectionResolver section.Resolver, hrefPathsResolver hrefpath.Resolver, assetPathsResolver relpath.Resolver, skipURLValidation bool) PageGenerator

	Option func(*Generator)
)

// All files and directories attributes are relative to the project root directory.
type Config struct {
	ContentDir    string
	PublicDir     string
	AssetsDir     string
	AssetsOutDir  string
	ScriptsDir    string
	ScriptsOutDir string
}

// Generator is the site generator which allows to generate and validate the site
type Generator struct {
	Config
	skipURLValidation            bool
	dirCopier                    mfs.DirCopier
	pagesFinder                  mfs.FilesFinder
	sectionResolver              section.Resolver
	absolutePathsResolverFactory abspath.ResolverFactory
	pagePathsResolver            relpath.Resolver
	hrefPathsResolver            hrefpath.Resolver
	assetPathsResolver           relpath.Resolver
	pageGeneratorFactory         pageGeneratorFactory
	pagesGenerators              []PageGenerator
	fs                           afero.Fs
}

// WithSkipURLValidation returns an Option that disables external URL validation.
func WithSkipURLValidation(skip bool) Option {
	return func(g *Generator) { g.skipURLValidation = skip }
}

func NewGenerator(fs afero.Fs, cfg Config, opts ...Option) (*Generator, error) {
	if cfg.ContentDir == "" {
		return nil, errors.New("ContentDir is mandatory")
	}
	if cfg.PublicDir == "" {
		return nil, errors.New("PublicDir is mandatory")
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
		Config:                       cfg,
		pagesGenerators:              make([]PageGenerator, 0),
		dirCopier:                    mfs.NewDirCopier(fs),
		pagesFinder:                  mfs.NewFilesFinder(fs, mfs.WithExtension(".md")),
		sectionResolver:              section.NewResolver(fs, cfg.ContentDir),
		absolutePathsResolverFactory: abspath.NewResolverFactory(fs, cfg.PublicDir),
		pagePathsResolver:            relpath.NewResolver(cfg.ContentDir, cfg.PublicDir, relpath.WithExtension("")),
		hrefPathsResolver:            hrefpath.NewResolver(cfg.ContentDir),
		assetPathsResolver:           relpath.NewResolver(cfg.AssetsDir, cfg.AssetsOutDir),
		pageGeneratorFactory:         defaultPageGeneratorFactory,
		fs:                           fs,
	}

	for _, opt := range opts {
		opt(g)
	}

	return g, nil
}

func (g *Generator) Generate() error {
	if err := g.dirCopier.CopyDir(mfs.NewFilesFinder(g.fs), g.AssetsDir, g.AssetsOutDir); err != nil {
		return fmt.Errorf("CopyDir assets err: %v", err)
	}

	if err := g.dirCopier.CopyDir(mfs.NewFilesFinder(g.fs, mfs.WithExtension(".js")), g.ScriptsDir, g.ScriptsOutDir); err != nil {
		return fmt.Errorf("CopyDir scripts err: %v", err)
	}

	if err := g.generatePages(g.assetPathsResolver, g.pagePathsResolver, g.hrefPathsResolver); err != nil {
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
