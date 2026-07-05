package page

import (
	_ "embed"
	"fmt"

	"github.com/spf13/afero"
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
	fs                    afero.Fs
	markdownSubstitutions *mdsubstitution.Registry
	HTMLSubstitutions     *htmlsubstitution.Registry
	validations           *validation.Registry
}

func NewGenerator(
	markdownSourcePath string,
	htmlOutputPath string,
	buildDir string,
	sectionName string,
	fs afero.Fs,
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

// Generate generates an HTML page by projecting the markdown file into the HTML template.
func (g *Generator) Generate() error {
	// Read markdown file using afero utility
	markdownSourceContent, err := afero.ReadFile(g.fs, g.sourceMDPath)
	if err != nil {
		return fmt.Errorf("reading %s: %w", g.sourceMDPath, err)
	}

	// Apply needed substitutions and generation in markdown
	markdownStringSourceContent, err := g.markdownSubstitutions.Apply(string(markdownSourceContent))
	if err != nil {
		return fmt.Errorf("failed to apply markdown substitutions: %w", err)
	}

	// Convert markdown to HTML
	htmlContent, err := markdown.NewConverter().Convert([]byte(markdownStringSourceContent))
	if err != nil {
		return fmt.Errorf("failed to convert markdown content: %w", err)
	}

	// Project result inside the page template
	htmlContent, err = g.HTMLSubstitutions.Apply(g.htmlPageTemplate, htmlContent)
	if err != nil {
		return fmt.Errorf("failed to project content inside the page template: %w", err)
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
	return g.validations.Validate(g.destinationHTMLPath, g.buildDir, g.htmlContentBytes)
}
