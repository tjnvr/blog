package page

import (
	_ "embed"
	"fmt"
	"path/filepath"

	"github.com/spf13/afero"

	"github.com/tjnvr/blog/internal/backbone/metadata"
	"github.com/tjnvr/blog/internal/backbone/section"
	"github.com/tjnvr/blog/internal/generator/page/html/substitution"
	"github.com/tjnvr/blog/internal/generator/page/html/validation"
	"github.com/tjnvr/blog/internal/generator/page/markdown"
	"github.com/tjnvr/blog/internal/hrefpath"
	"github.com/tjnvr/blog/internal/relpath"
)

//go:embed page.html
var defaultTemplate string

type Generator struct {
	htmlPageTemplate    string
	sourceMDPath        string
	htmlContentBytes    []byte
	destinationHTMLPath string
	fs                  afero.Fs
	sectionResolver     section.Resolver
	hrefPathsResolver   hrefpath.Resolver
	assetPathsResolver  relpath.Resolver
	validations         *validation.Registry
}

func NewGenerator(
	markdownPath string,
	HTMLPath string,
	fs afero.Fs,
	sectionResolver section.Resolver,
	hrefPathsResolver hrefpath.Resolver,
	assetPathsResolver relpath.Resolver,
	validations *validation.Registry,
) *Generator {
	return &Generator{
		htmlPageTemplate:    defaultTemplate,
		sourceMDPath:        markdownPath,
		destinationHTMLPath: HTMLPath,
		fs:                  fs,
		sectionResolver:     sectionResolver,
		hrefPathsResolver:   hrefPathsResolver,
		assetPathsResolver:  assetPathsResolver,
		validations:         validations,
	}
}

// Generate generates an HTML page by projecting the markdown file into the HTML template.
func (g *Generator) Generate() error {
	// Read markdown file using afero utility
	markdownSourceContent, err := afero.ReadFile(g.fs, g.sourceMDPath)
	if err != nil {
		return fmt.Errorf("reading %s: %w", g.sourceMDPath, err)
	}

	// Convert markdown to HTML
	htmlContent, err := markdown.NewConverter().Convert(markdownSourceContent)
	if err != nil {
		return fmt.Errorf("failed to convert markdown content: %w", err)
	}

	// Project the raw HTML content inside the page template
	htmlContent, err = substitution.NewRegistry(
		g.fs,
		g.destinationHTMLPath,
		g.sourceMDPath,
		metadata.Extract(markdownSourceContent),
		g.sectionResolver,
		g.hrefPathsResolver,
		g.assetPathsResolver,
	).Apply(g.htmlPageTemplate, htmlContent)
	if err != nil {
		return fmt.Errorf("failed to project content inside the page template for %s: %w", g.sourceMDPath, err)
	}

	if err := g.fs.MkdirAll(filepath.Dir(g.destinationHTMLPath), 0744); err != nil {
		return fmt.Errorf("failed to MkdirAll for %s: %w", filepath.Dir(g.destinationHTMLPath), err)
	}

	// Write HTML file using afero utility
	htmlContentBytes := []byte(htmlContent)
	if err := afero.WriteFile(g.fs, g.destinationHTMLPath, htmlContentBytes, 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", g.destinationHTMLPath, err)
	}

	fmt.Printf("Generated: %s -> %s\n", g.sourceMDPath, g.destinationHTMLPath)
	g.htmlContentBytes = htmlContentBytes
	return nil
}

// Validate triggers structural validation checks on the generated HTML content.
// Note: This must be called after a successful invocation of Generate().
func (g *Generator) Validate() error {
	return g.validations.Validate(g.htmlContentBytes)
}
