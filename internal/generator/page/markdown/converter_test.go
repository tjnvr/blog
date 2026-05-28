package markdown

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConverter(t *testing.T) {
	converter := NewConverter()
	require.NotNil(t, converter)
	require.NotNil(t, converter.md)
}

func TestConverter_Convert(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains []string
	}{
		{
			name:     "converts heading",
			input:    "# Hello World",
			contains: []string{`<h1 id="hello-world">Hello World<a href="#hello-world" class="heading-anchor">#</a></h1>`},
		},
		{
			name:     "converts paragraph",
			input:    "This is a paragraph.",
			contains: []string{"<p>This is a paragraph.</p>"},
		},
		{
			name:     "converts link",
			input:    "[Link](https://example.com)",
			contains: []string{`<a href="https://example.com">Link</a>`},
		},
		{
			name:     "converts bold text",
			input:    "**bold**",
			contains: []string{"<strong>bold</strong>"},
		},
		{
			name:     "converts italic text",
			input:    "*italic*",
			contains: []string{"<em>italic</em>"},
		},
		{
			name:     "converts code block with GFM",
			input:    "```go\nfunc main() {}\n```",
			contains: []string{`<pre><code class="language-go">`},
		},
		{
			name:     "converts inline code",
			input:    "Use `fmt.Println`",
			contains: []string{"<code>fmt.Println</code>"},
		},
		{
			name:     "converts unordered list",
			input:    "- item 1\n- item 2",
			contains: []string{"<ul>", "<li>item 1</li>", "<li>item 2</li>"},
		},
		{
			name:     "converts ordered list",
			input:    "1. first\n2. second",
			contains: []string{"<ol>", "<li>first</li>", "<li>second</li>"},
		},
		{
			name:     "handles empty input",
			input:    "",
			contains: []string{},
		},
		{
			name:     "converts GFM strikethrough",
			input:    "~~deleted~~",
			contains: []string{"<del>deleted</del>"},
		},
		{
			name:     "converts GFM table",
			input:    "| A | B |\n|---|---|\n| 1 | 2 |",
			contains: []string{"<table>", "<th>A</th>", "<td>1</td>"},
		},
	}

	converter := NewConverter()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := converter.Convert([]byte(tt.input))
			require.NoError(t, err)
			for _, substr := range tt.contains {
				assert.Contains(t, result, substr)
			}
		})
	}
}

func TestConverter_Convert_ReturnsValidHTML(t *testing.T) {
	converter := NewConverter()

	input := "# Title\n\nParagraph with **bold** and *italic*.\n\n- List item"
	result, err := converter.Convert([]byte(input))
	require.NoError(t, err)
	assert.NotEmpty(t, result)
	assert.Contains(t, result, `<h1 id="title">`)
	assert.Contains(t, result, "<p>")
	assert.Contains(t, result, "<ul>")
}

func TestConverter_InlineAttributes(t *testing.T) {
	converter := NewConverter()

	tests := []struct {
		name     string
		input    string
		contains []string
	}{
		{
			name:     "heading with class",
			input:    "# Title {.custom-class}",
			contains: []string{`<h1 class="custom-class"`, "Title"},
		},
		{
			name:     "heading with id",
			input:    "## Section {#my-section}",
			contains: []string{`<h2 id="my-section"`, "Section"},
		},
		{
			name:     "heading with class and id",
			input:    "### Header {.styled #header-id}",
			contains: []string{`class="styled"`, `id="header-id"`, "Header"},
		},
		{
			name:     "heading with multiple classes",
			input:    "# Big Title {.text-4xl .font-bold .text-red-500}",
			contains: []string{`class="text-4xl font-bold text-red-500"`, "Big Title"},
		},
		{
			name:     "heading with custom attribute",
			input:    "# Title {data-testid=main-title}",
			contains: []string{`data-testid="main-title"`, "Title"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := converter.Convert([]byte(tt.input))
			require.NoError(t, err)
			for _, substr := range tt.contains {
				assert.Contains(t, result, substr)
			}
		})
	}
}
