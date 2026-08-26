package metadata

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMetadata_Extract(t *testing.T) {
	tests := []struct {
		name string
		data string
		want Metadata
	}{
		{
			name: "no metadata",
			data: "# Hello\nsome content",
			want: Metadata{Title: "Hello"},
		},
		{
			name: "no title",
			data: "some content without a heading",
			want: Metadata{},
		},
		{
			name: "title with extra spaces is trimmed",
			data: "#   Hello World   \nsome content",
			want: Metadata{Title: "Hello World"},
		},
		{
			name: "title is the first level-1 heading only",
			data: "some intro\n# Hello\n# Ignored second heading",
			want: Metadata{Title: "Hello"},
		},
		{
			name: "creation-date",
			data: "<!-- creation-date: 2026-01-24 -->\n# Hello",
			want: Metadata{Title: "Hello", CreationDate: "2026-01-24"},
		},
		{
			name: "unknown key is ignored",
			data: "<!-- unknown-key: some-value -->\n# Hello",
			want: Metadata{Title: "Hello"},
		},
		{
			name: "multiple metadata",
			data: "<!-- creation-date: 2026-03-15 -->\n<!-- unknown-key: ignored -->\n# Hello",
			want: Metadata{Title: "Hello", CreationDate: "2026-03-15"},
		},
		{
			name: "extra spaces",
			data: "<!--   creation-date:   2026-01-24   -->\n# Hello",
			want: Metadata{Title: "Hello", CreationDate: "2026-01-24"},
		},
		{
			name: "seq",
			data: "<!-- seq: 2 -->\n# Hello",
			want: Metadata{Title: "Hello", Seq: 2},
		},
		{
			name: "seq alongside creation-date",
			data: "<!-- seq: 3 -->\n<!-- creation-date: 2026-01-24 -->\n# Hello",
			want: Metadata{Title: "Hello", CreationDate: "2026-01-24", Seq: 3},
		},
		{
			name: "non numeric seq is ignored",
			data: "<!-- seq: first -->\n# Hello",
			want: Metadata{Title: "Hello"},
		},
		{
			name: "description",
			data: "<!-- description: A short summary of the page. -->\n# Hello",
			want: Metadata{Title: "Hello", Description: "A short summary of the page."},
		},
		{
			name: "image",
			data: "<!-- image: ../assets/images/cover.png -->\n# Hello",
			want: Metadata{Title: "Hello", Image: "../assets/images/cover.png"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// test
			got := Extract([]byte(tt.data))

			// expect
			assert.Equal(t, tt.want, got)
		})
	}
}
