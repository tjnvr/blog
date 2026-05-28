package listing

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakePrinter struct {
	output string
}

func (f fakePrinter) Print() string { return f.output }

type fakeLister struct {
	printers []fakePrinter
	err      error
}

func (f fakeLister) ListPrinters() ([]fakePrinter, error) {
	return f.printers, f.err
}

func TestNewSubstituer(t *testing.T) {
	s := NewSubstituer("{{placeholder}}", fakeLister{}, "\n")
	assert.Equal(t, "{{placeholder}}", s.Placeholder())
}

func TestSubstituer_Resolve(t *testing.T) {
	tests := []struct {
		name      string
		printers  []fakePrinter
		separator string
		listerErr error
		want      string
		wantErr   bool
	}{
		{
			name:      "empty list returns empty string",
			printers:  []fakePrinter{},
			separator: "\n",
			want:      "",
		},
		{
			name:      "single item includes separator",
			printers:  []fakePrinter{{output: "- [Post](post.md)"}},
			separator: "\n",
			want:      "- [Post](post.md)\n",
		},
		{
			name: "multiple items joined with separator",
			printers: []fakePrinter{
				{output: "- [First](first.md)"},
				{output: "- [Second](second.md)"},
			},
			separator: "\n",
			want:      "- [First](first.md)\n- [Second](second.md)\n",
		},
		{
			name:      "lister error propagated",
			printers:  nil,
			separator: "\n",
			listerErr: fmt.Errorf("read error"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lister := fakeLister{printers: tt.printers, err: tt.listerErr}
			s := NewSubstituer("{{x}}", lister, tt.separator)
			got, err := s.Resolve()
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
