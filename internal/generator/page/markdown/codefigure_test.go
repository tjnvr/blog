package markdown

import (
	"strings"
	"testing"
)

func TestCodeFigureTransformer_WrapsTitledCodeBlock(t *testing.T) {
	converter := NewConverter()

	result, err := converter.Convert([]byte("```sh {title=\"Size <of> project\"}\nls\n```"))
	if err != nil {
		t.Fatalf("Convert() error = %v", err)
	}

	expected := "<figure>\n<pre><code class=\"language-sh\">ls\n</code></pre>\n" +
		"<figcaption>Size &lt;of&gt; project</figcaption>\n</figure>\n"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestCodeFigureTransformer_TitleWithoutLanguage(t *testing.T) {
	converter := NewConverter()

	result, err := converter.Convert([]byte("```{title=\"T\"}\nls\n```"))
	if err != nil {
		t.Fatalf("Convert() error = %v", err)
	}

	if !strings.Contains(result, "<pre><code>ls\n</code></pre>") {
		t.Errorf("expected code block without language class, got %q", result)
	}
	if !strings.Contains(result, "<figcaption>T</figcaption>") {
		t.Errorf("expected figcaption, got %q", result)
	}
}

func TestCodeFigureTransformer_NoTitle_Unaffected(t *testing.T) {
	converter := NewConverter()

	for _, input := range []string{"```sh\nls\n```", "```sh {.other}\nls\n```", "```\nls\n```"} {
		result, err := converter.Convert([]byte(input))
		if err != nil {
			t.Fatalf("Convert() error = %v", err)
		}

		if strings.Contains(result, "<figure>") {
			t.Errorf("expected no figure for %q, got %q", input, result)
		}
	}
}

func TestCodeFigureTransformer_NestedInList(t *testing.T) {
	converter := NewConverter()

	result, err := converter.Convert([]byte("- item\n\n  ```go {title=\"T\"}\n  x\n  ```"))
	if err != nil {
		t.Fatalf("Convert() error = %v", err)
	}

	if !strings.Contains(result, "<li>") || !strings.Contains(result, "<figcaption>T</figcaption>") {
		t.Errorf("expected figure inside list item, got %q", result)
	}
}
