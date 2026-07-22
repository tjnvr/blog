package markdown

import (
	"bytes"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

// Converter wraps goldmark for markdown to HTML conversion
type Converter struct {
	md goldmark.Markdown
}

// NewConverter creates a new markdown converter with GFM extensions.
func NewConverter() *Converter {
	return &Converter{
		md: goldmark.New(
			goldmark.WithExtensions(extension.Footnote),
			goldmark.WithExtensions(extension.GFM),
			goldmark.WithParserOptions(
				parser.WithAttribute(),
				parser.WithAutoHeadingID(),
				parser.WithASTTransformers(
					util.Prioritized(&ImageAttributeTransformer{}, 100),
				),
			),
			goldmark.WithRendererOptions(
				renderer.WithNodeRenderers(
					util.Prioritized(&HeadingRenderer{}, 100),
				),
			),
		),
	}
}

// Convert converts markdown source to HTML
func (c *Converter) Convert(source []byte) (string, error) {
	var buf bytes.Buffer
	if err := c.md.Convert(source, &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}
