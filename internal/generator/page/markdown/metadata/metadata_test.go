package metadata

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtract(t *testing.T) {
	tests := []struct {
		name string
		data string
		want Metadata
	}{
		{
			name: "no metadata",
			data: "# Hello\nsome content",
			want: Metadata{},
		},
		{
			name: "creation-date",
			data: "<!-- creation-date: 2026-01-24 -->\n# Hello",
			want: Metadata{CreationDate: "2026-01-24"},
		},
		{
			name: "unknown key is ignored",
			data: "<!-- unknown-key: some-value -->\n# Hello",
			want: Metadata{},
		},
		{
			name: "multiple metadata",
			data: "<!-- creation-date: 2026-03-15 -->\n<!-- unknown-key: ignored -->\n# Hello",
			want: Metadata{CreationDate: "2026-03-15"},
		},
		{
			name: "extra spaces",
			data: "<!--   creation-date:   2026-01-24   -->\n# Hello",
			want: Metadata{CreationDate: "2026-01-24"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Extract([]byte(tt.data))
			assert.Equal(t, tt.want, got)
		})
	}
}
