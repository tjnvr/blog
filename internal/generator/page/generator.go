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
		HTMLValidations         PageValidator
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
		Validate(HTMLPath string) error
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
		HTMLValidations:         validations,
	}
}

// Generate produces an HTML page using a template and the content provided in a markdown file
//
// sourceMarkdownFilePath is the markdown file content to process
// destinationHTMLFilePath is the HTML file that will be produced
func (g *Generator) Generate(fromMarkdownFilePath, toHTMLFilePath string) error {
	// Read markdown file content
	sourceMarkdownFilePathContent, err := afero.ReadFile(g.fs, fromMarkdownFilePath)
	if err != nil {
		return fmt.Errorf("reading %s: %w", fromMarkdownFilePath, err)
	}

	// Apply needed substitutions and generations inside the mardkdown content
	markdDownStringSourceContent, err := g.markdownSubstitutions.Apply(string(sourceMarkdownFilePathContent))
	if err != nil {
		return fmt.Errorf("failed to project content inside the page template: %w", err)
	}

	// Convert the markdown to HTML
	HTMLContent, err := g.markdownToHTMLConverter.Convert([]byte(markdDownStringSourceContent))
	if err != nil {
		return fmt.Errorf("failed to convert markdown content: %w", err)
	}

	// Project the HTML content into the page template (e.g., populating body, title).
	HTMLContent, err = g.HTMLSubstitutions.Apply(defaultTemplate, HTMLContent)
	if err != nil {
		return fmt.Errorf("failed to project content inside the page template: %w", err)
	}

	// Ensure output directory exists
	if err := g.fs.MkdirAll(filepath.Dir(toHTMLFilePath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Write HTML file result
	if err := afero.WriteFile(g.fs, toHTMLFilePath, []byte(HTMLContent), 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", toHTMLFilePath, err)
	}

	fmt.Printf("Generated: %s -> %s\n", fromMarkdownFilePath, toHTMLFilePath)
	return nil
}

func (g *Generator) Validate(HTMLFilePath string) error {
	return g.HTMLValidations.Validate(HTMLFilePath)
}
