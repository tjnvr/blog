package title

import (
	"testing"

	"github.com/tjnvr/blog/internal/backbone/metadata"
)

const wantPlaceholder = "{{title}}"

func TestSubstituter_Placeholder(t *testing.T) {
	s := NewSubstituer(metadata.Metadata{})
	if got := s.Placeholder(); got != wantPlaceholder {
		t.Errorf("Placeholder() = %q, want %q", got, wantPlaceholder)
	}
}

func TestSubstituter_Resolve(t *testing.T) {
	tests := []struct {
		name    string
		m       metadata.Metadata
		want    string
		wantErr bool
	}{
		{
			name: "returns the title",
			m:    metadata.Metadata{Title: "My Page Title"},
			want: "My Page Title",
		},
		{
			name:    "returns error when there is no title",
			m:       metadata.Metadata{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewSubstituer(tt.m)
			got, err := s.Resolve("<h1>Rendered content, unused by this substituter.</h1>")

			if tt.wantErr {
				if err == nil {
					t.Error("Resolve() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Resolve() unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("Resolve() = %q, want %q", got, tt.want)
			}
		})
	}
}
