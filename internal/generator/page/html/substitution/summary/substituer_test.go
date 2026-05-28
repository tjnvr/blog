package summary

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSubstituter_Placeholder(t *testing.T) {
	s := NewSubstituer()
	assert.Equal(t, "<p>{{summary}}</p>", s.Placeholder())
}

func TestSubstituter_Resolve(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		contains []string
		empty    bool
	}{
		{
			name: "h2 and h3 headings produce anchor links",
			content: `<h1 id="title">Title<a href="#title" class="heading-anchor">#</a></h1>` +
				`<h2 id="intro">Introduction<a href="#intro" class="heading-anchor">#</a></h2>` +
				`<h3 id="details">Details<a href="#details" class="heading-anchor">#</a></h3>`,
			contains: []string{
				`<a href="#intro" class="text-sm">Introduction</a>`,
				`<a href="#details" class="text-xs">Details</a>`,
			},
		},
		{
			name: "h1 is ignored",
			content: `<h1 id="page-title">Page Title<a href="#page-title" class="heading-anchor">#</a></h1>` +
				`<h2 id="section">Section<a href="#section" class="heading-anchor">#</a></h2>`,
			contains: []string{`<a href="#section" class="text-sm">Section</a>`},
		},
		{
			name:    "h1 only returns empty string",
			content: `<h1 id="page-title">Page Title<a href="#page-title" class="heading-anchor">#</a></h1>`,
			empty:   true,
		},
		{
			name:    "no headings returns empty string",
			content: `<p>Some paragraph text.</p>`,
			empty:   true,
		},
		{
			name:     "anchor symbol stripped from heading text",
			content:  `<h2 id="foo">My Section<a href="#foo" class="heading-anchor">#</a></h2>`,
			contains: []string{`My Section</a>`},
		},
		{
			name: "h2 only produces flat list",
			content: `<h2 id="a">A<a href="#a" class="heading-anchor">#</a></h2>` +
				`<h2 id="b">B<a href="#b" class="heading-anchor">#</a></h2>`,
			contains: []string{
				`<a href="#a" class="text-sm">A</a>`,
				`<a href="#b" class="text-sm">B</a>`,
			},
		},
		{
			name: "h3 nested under h2",
			content: `<h2 id="intro">Intro<a href="#intro" class="heading-anchor">#</a></h2>` +
				`<h3 id="details">Details<a href="#details" class="heading-anchor">#</a></h3>`,
			contains: []string{
				`<li><a href="#intro" class="text-sm">Intro</a><ul><li><a href="#details" class="text-xs">Details</a></li></ul></li>`,
			},
		},
		{
			name: "h2 after h3 returns to top level",
			content: `<h2 id="a">A<a href="#a" class="heading-anchor">#</a></h2>` +
				`<h3 id="b">B<a href="#b" class="heading-anchor">#</a></h3>` +
				`<h2 id="c">C<a href="#c" class="heading-anchor">#</a></h2>`,
			contains: []string{
				`</ul></li><li><a href="#c" class="text-sm">C</a></li>`,
			},
		},
		{
			name: "h4 under h2 opens two nested ul",
			content: `<h2 id="a">A<a href="#a" class="heading-anchor">#</a></h2>` +
				`<h4 id="b">B<a href="#b" class="heading-anchor">#</a></h4>`,
			contains: []string{
				`<a href="#a" class="text-sm">A</a><ul><ul><li><a href="#b" class="text-xs">B</a>`,
			},
		},
		{
			name: "text-sm on first level, text-xs on second+",
			content: `<h2 id="a">A<a href="#a" class="heading-anchor">#</a></h2>` +
				`<h3 id="b">B<a href="#b" class="heading-anchor">#</a></h3>` +
				`<h4 id="c">C<a href="#c" class="heading-anchor">#</a></h4>`,
			contains: []string{
				`class="text-sm"`,
				`class="text-xs"`,
				`class="text-xs"`,
			},
		},
		{
			name:     "output wrapped in nav and ul",
			content:  `<h2 id="section">Section<a href="#section" class="heading-anchor">#</a></h2>`,
			contains: []string{`<nav><ul>`, "</ul></nav>"},
		},
		{
			name: "all levels h2-h6 included",
			content: `<h2 id="h2">H2<a href="#h2" class="heading-anchor">#</a></h2>` +
				`<h3 id="h3">H3<a href="#h3" class="heading-anchor">#</a></h3>` +
				`<h4 id="h4">H4<a href="#h4" class="heading-anchor">#</a></h4>` +
				`<h5 id="h5">H5<a href="#h5" class="heading-anchor">#</a></h5>` +
				`<h6 id="h6">H6<a href="#h6" class="heading-anchor">#</a></h6>`,
			contains: []string{
				`href="#h2"`, `href="#h3"`, `href="#h4"`, `href="#h5"`, `href="#h6"`,
			},
		},
	}

	s := NewSubstituer()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := s.Resolve(tt.content)
			require.NoError(t, err)

			if tt.empty {
				assert.Empty(t, result)
				return
			}

			for _, substr := range tt.contains {
				assert.Contains(t, result, substr)
			}
		})
	}
}
