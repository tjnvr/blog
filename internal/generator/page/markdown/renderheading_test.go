package markdown

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHeadingRenderer_AllLevels(t *testing.T) {
	converter := NewConverter()

	for level := 1; level <= 6; level++ {
		t.Run(fmt.Sprintf("h%d", level), func(t *testing.T) {
			input := fmt.Sprintf("%s Heading", strings.Repeat("#", level))
			result, err := converter.Convert([]byte(input))
			require.NoError(t, err)

			assert.Contains(t, result, fmt.Sprintf("<h%d", level))
			assert.Contains(t, result, `<a href="#heading" class="heading-anchor">#</a>`)
			assert.Contains(t, result, fmt.Sprintf("</h%d>", level))
		})
	}
}

func TestHeadingRenderer_SlugifiedID(t *testing.T) {
	converter := NewConverter()

	tests := []struct {
		name       string
		input      string
		expectedID string
	}{
		{
			name:       "special characters are removed",
			input:      "## Hello, World!",
			expectedID: "hello-world",
		},
		{
			name:       "spaces become hyphens",
			input:      "## My Great Section",
			expectedID: "my-great-section",
		},
		{
			name:       "mixed case becomes lowercase",
			input:      "## CamelCase Title",
			expectedID: "camelcase-title",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := converter.Convert([]byte(tt.input))
			require.NoError(t, err)
			assert.Contains(t, result, fmt.Sprintf(`id="%s"`, tt.expectedID))
			assert.Contains(t, result, fmt.Sprintf(`href="#%s"`, tt.expectedID))
		})
	}
}

func TestHeadingRenderer_DuplicateHeadings(t *testing.T) {
	converter := NewConverter()

	input := "## Section\n\nSome text.\n\n## Section\n\nMore text.\n\n## Section"
	result, err := converter.Convert([]byte(input))
	require.NoError(t, err)

	assert.Contains(t, result, `id="section"`)
	assert.Contains(t, result, `id="section-1"`)
	assert.Contains(t, result, `id="section-2"`)
	assert.Contains(t, result, `href="#section-1"`)
}

func TestHeadingRenderer_CustomID(t *testing.T) {
	converter := NewConverter()

	input := "## Section {#custom-id}"
	result, err := converter.Convert([]byte(input))
	require.NoError(t, err)

	assert.Contains(t, result, `id="custom-id"`)
	assert.Contains(t, result, `href="#custom-id"`)
}

func TestHeadingRenderer_AnchorStructure(t *testing.T) {
	converter := NewConverter()

	input := "## My Section"
	result, err := converter.Convert([]byte(input))
	require.NoError(t, err)

	assert.Contains(t, result, `class="heading-anchor"`)
	assert.Contains(t, result, `">#</a>`)

	anchorIdx := strings.Index(result, `<a href="#my-section"`)
	textIdx := strings.Index(result, "My Section")
	if anchorIdx == -1 || textIdx == -1 || anchorIdx <= textIdx {
		t.Errorf("anchor should appear after heading text, got %q", result)
	}
}
