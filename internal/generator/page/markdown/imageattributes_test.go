package markdown

import (
	"strings"
	"testing"
)

func TestImageAttributeTransformer_AppliesClassAttribute(t *testing.T) {
	converter := NewConverter()

	result, err := converter.Convert([]byte(`![x](a.png){.small-centered}`))
	if err != nil {
		t.Fatalf("Convert() error = %v", err)
	}

	if !strings.Contains(result, `<img src="a.png" alt="x" class="small-centered">`) {
		t.Errorf("expected img with class attribute, got %q", result)
	}
	if strings.Contains(result, "{.small-centered}") {
		t.Errorf("expected brace syntax to be removed from output, got %q", result)
	}
}

func TestImageAttributeTransformer_PreservesTrailingText(t *testing.T) {
	converter := NewConverter()

	result, err := converter.Convert([]byte(`![x](a.png){.small-centered} trailing text`))
	if err != nil {
		t.Fatalf("Convert() error = %v", err)
	}

	if !strings.Contains(result, `class="small-centered"`) {
		t.Errorf("expected class attribute applied, got %q", result)
	}
	if !strings.Contains(result, "trailing text") {
		t.Errorf("expected trailing text preserved, got %q", result)
	}
}

func TestImageAttributeTransformer_NoAttributeSyntax_Unaffected(t *testing.T) {
	converter := NewConverter()

	result, err := converter.Convert([]byte(`![x](a.png)`))
	if err != nil {
		t.Fatalf("Convert() error = %v", err)
	}

	if !strings.Contains(result, `<img src="a.png" alt="x">`) {
		t.Errorf("expected plain img tag unaffected, got %q", result)
	}
	if strings.Contains(result, "class=") {
		t.Errorf("expected no class attribute, got %q", result)
	}
}

func TestImageAttributeTransformer_NoFollowingSibling_NoPanic(t *testing.T) {
	converter := NewConverter()

	result, err := converter.Convert([]byte(`text before ![x](a.png)`))
	if err != nil {
		t.Fatalf("Convert() error = %v", err)
	}

	if !strings.Contains(result, `<img src="a.png" alt="x">`) {
		t.Errorf("expected plain img tag at end of paragraph, got %q", result)
	}
}
