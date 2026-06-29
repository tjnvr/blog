package listing

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestSubstituer_NewSubstituer(t *testing.T) {
	// given
	placeholder := "{{placeholder}}"
	lister := fakeLister{}
	separator := "\n"

	// test
	s := NewSubstituer(placeholder, lister, separator)

	// expect
	assert.Equal(t, placeholder, s.Placeholder())
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
			// given
			lister := fakeLister{printers: tt.printers, err: tt.listerErr}
			s := NewSubstituer("{{x}}", lister, tt.separator)

			// test
			got, err := s.Resolve()

			// expect
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
