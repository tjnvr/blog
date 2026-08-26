package description

import (
	"testing"

	"github.com/tjnvr/blog/internal/backbone/metadata"
)

const wantPlaceholder = `<meta name="description" content="{{description}}">`

func TestSubstituter_Placeholder(t *testing.T) {
	s := NewSubstituer(metadata.Metadata{})
	if got := s.Placeholder(); got != wantPlaceholder {
		t.Errorf("Placeholder() = %q, want %q", got, wantPlaceholder)
	}
}

func TestSubstituter_Resolve(t *testing.T) {
	tests := []struct {
		name string
		m    metadata.Metadata
		want string
	}{
		{
			name: "returns the meta tag when a metadata description is set",
			m:    metadata.Metadata{Description: "A hand-written description."},
			want: `<meta name="description" content="A hand-written description.">`,
		},
		{
			name: "escapes special characters",
			m:    metadata.Metadata{Description: `Tom & Jerry say "hi"`},
			want: `<meta name="description" content="Tom &amp; Jerry say &#34;hi&#34;">`,
		},
		{
			name: "returns empty string when no metadata description is set, leaving detection to Google",
			m:    metadata.Metadata{},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewSubstituer(tt.m)
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
