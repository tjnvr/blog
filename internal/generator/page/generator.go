package page

import (
	_ "embed"

	"fmt"
	"path/filepath"

	"github.com/tjnvr/blog/internal/generator/page/filesystem"
	htmlsubstitution "github.com/tjnvr/blog/internal/generator/page/html/substitution"
	"github.com/tjnvr/blog/internal/generator/page/html/validation"
	"github.com/tjnvr/blog/internal/generator/page/markdown"
	mdsubstitution "github.com/tjnvr/blog/internal/generator/page/markdown/substitution"
)

//go:embed page.html
var defaultTemplate string

type Generator struct {
	htmlPageTemplate      string
	sourceMDPath          string
	buildDir              string
	htmlContentBytes      []byte
	destinationHTMLPath   string
	sectionName           string
	fs                    filesystem.FileSystem
	markdownSubstitutions *mdsubstitution.Registry
	HTMLSubstitutions     *htmlsubstitution.Registry
	validations           *validation.Registry
}

func NewGenerator(
	markdownSourcePath string,
	htmlOutputPath string,
	buildDir string,
	sectionName string,
	fs filesystem.FileSystem,
	markdownSubstitutions *mdsubstitution.Registry,
	HTMLSubstitutions *htmlsubstitution.Registry,
	validations *validation.Registry,
) *Generator {
	return &Generator{
		htmlPageTemplate:      defaultTemplate,
		sourceMDPath:          markdownSourcePath,
		destinationHTMLPath:   htmlOutputPath,
		buildDir:              buildDir,
		sectionName:           sectionName,
		fs:                    fs,
		markdownSubstitutions: markdownSubstitutions,
		HTMLSubstitutions:     HTMLSubstitutions,
		validations:           validations,
	}
}

// Generate generates an html page by projecting the markdown file in the HTML template.
func (g *Generator) Generate() error {
	// Read markdown file
	markdDownSourceContent, err := g.fs.ReadFile(g.sourceMDPath)
	if err != nil {
		return fmt.Errorf("reading %s: %w", g.sourceMDPath, err)
	}

	// Apply needed substitutions and generation in mardkdown
	markdDownStringSourceContent, err := g.markdownSubstitutions.Apply(string(markdDownSourceContent))
	if err != nil {
		return fmt.Errorf("failed to project content inside the page template: %w", err)
	}

	// Convert marddown to HTML
	htmlContent, err := markdown.NewConverter().Convert([]byte(markdDownStringSourceContent))
	if err != nil {
		return fmt.Errorf("failed to convert markdown content: %w", err)
	}

	// Project result inside the page template
	htmlContent, err = g.HTMLSubstitutions.Apply(g.htmlPageTemplate, htmlContent)
	if err != nil {
		return fmt.Errorf("failed to project content inside the page template: %w", err)
	}

	// Ensure output directory exists
	if err := g.fs.MkdirAll(filepath.Dir(g.destinationHTMLPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Write HTML file
	htmlContentBytes := []byte(htmlContent)
	if err := g.fs.WriteFile(g.destinationHTMLPath, htmlContentBytes, 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", g.destinationHTMLPath, err)
	}

	fmt.Printf("Generated: %s -> %s\n", g.sourceMDPath, g.destinationHTMLPath)
	g.htmlContentBytes = htmlContentBytes
	return nil
}

func (g *Generator) Validate() error {
	return g.validations.Validate(g.destinationHTMLPath, g.buildDir, g.htmlContentBytes)
}
