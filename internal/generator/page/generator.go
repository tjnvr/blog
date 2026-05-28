package page

import (
	_ "embed"
	"fmt"
	"path/filepath"

	"github.com/spf13/afero"
)

//go:embed page.html
var defaultTemplate string

type (
	Generator struct {
		fs                      afero.Fs
		markdownSubstitutions   MarkdownSubstituer
		markdownToHTMLConverter MarkdownConverter
		HTMLSubstitutions       HTMLSubstituer
		validations             PageValidator
	}

	MarkdownSubstituer interface {
		Apply(content string) (string, error)
	}

	MarkdownConverter interface {
		Convert(content []byte) (string, error)
	}

	HTMLSubstituer interface {
		Apply(template, content string) (string, error)
	}

	PageValidator interface {
		Validate(htmlPath string) error
	}
)

func NewGenerator(
	fs afero.Fs,
	markdownSubstitutions MarkdownSubstituer,
	markdownToHTMLConverter MarkdownConverter,
	HTMLSubstitutions HTMLSubstituer,
	validations PageValidator,
) *Generator {
	return &Generator{
		fs:                      fs,
		markdownSubstitutions:   markdownSubstitutions,
		markdownToHTMLConverter: markdownToHTMLConverter,
		HTMLSubstitutions:       HTMLSubstitutions,
		validations:             validations,
	}
}

// Generate generates an html page by projecting the markdown file in the HTML template.
func (g *Generator) Generate(sourceMarkdownFilePath, destinationHTMLFilePath string) error {
	// Read markdown file content
	sourceMarkdownFilePathContent, err := afero.ReadFile(g.fs, sourceMarkdownFilePath)
	if err != nil {
		return fmt.Errorf("reading %s: %w", sourceMarkdownFilePath, err)
	}

	// Apply needed substitutions and generations inside the mardkdown content
	markdDownStringSourceContent, err := g.markdownSubstitutions.Apply(string(sourceMarkdownFilePathContent))
	if err != nil {
		return fmt.Errorf("failed to project content inside the page template: %w", err)
	}

	// Convert the markdown to HTML
	htmlContent, err := g.markdownToHTMLConverter.Convert([]byte(markdDownStringSourceContent))
	if err != nil {
		return fmt.Errorf("failed to convert markdown content: %w", err)
	}

	// Project the HTML content into the page template (e.g., populating body, title).
	htmlContent, err = g.HTMLSubstitutions.Apply(defaultTemplate, htmlContent)
	if err != nil {
		return fmt.Errorf("failed to project content inside the page template: %w", err)
	}

	// Ensure output directory exists
	if err := g.fs.MkdirAll(filepath.Dir(destinationHTMLFilePath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Write HTML file result
	htmlContentBytes := []byte(htmlContent)
	if err := afero.WriteFile(g.fs, destinationHTMLFilePath, htmlContentBytes, 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", destinationHTMLFilePath, err)
	}

	fmt.Printf("Generated: %s -> %s\n", sourceMarkdownFilePath, destinationHTMLFilePath)
	return nil
}

func (g *Generator) Validate(HTMLFilePath string) error {
	return g.validations.Validate(HTMLFilePath)
}
