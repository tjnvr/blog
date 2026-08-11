package description

import (
	"testing"

	"github.com/spf13/afero"
)

const wantPlaceholder = `<meta name="description" content="{{description}}">`

func TestSubstituer_Placeholder(t *testing.T) {
	s := NewSubstituer(afero.NewMemMapFs(), "content/index.md")
	if got := s.Placeholder(); got != wantPlaceholder {
		t.Errorf("Placeholder() = %q, want %q", got, wantPlaceholder)
	}
}

func TestSubstituer_Resolve(t *testing.T) {
	tests := []struct {
		name     string
		markdown string
		want     string
	}{
		{
			name:     "returns the meta tag when a metadata description is set",
			markdown: `<!-- description: A hand-written description. -->` + "\n# Hello",
			want:     `<meta name="description" content="A hand-written description.">`,
		},
		{
			name:     "escapes special characters",
			markdown: `<!-- description: Tom & Jerry say "hi" -->` + "\n# Hello",
			want:     `<meta name="description" content="Tom &amp; Jerry say &#34;hi&#34;">`,
		},
		{
			name:     "returns empty string when no metadata description is set, leaving detection to Google",
			markdown: "# Hello\n\nA paragraph that must NOT be used as a fallback description.",
			want:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			if err := afero.WriteFile(fs, "content/index.md", []byte(tt.markdown), 0644); err != nil {
				t.Fatalf("WriteFile err: %v", err)
			}

			s := NewSubstituer(fs, "content/index.md")
			got, err := s.Resolve("<h1>Hello</h1><p>Rendered content, unused by this substituter.</p>")
			if err != nil {
				t.Fatalf("Resolve() unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("Resolve() = %q, want %q", got, tt.want)
			}
		})
	}
}
