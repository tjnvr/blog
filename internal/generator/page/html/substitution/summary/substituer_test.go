package summary

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListHeadings(t *testing.T) {
	// given
	tests := []struct {
		name     string
		content  string
		expected []Heading
	}{
		{
			name:     "empty slice when no headings",
			content:  "<p>Some content without headings</p>",
			expected: nil,
		},
		{
			name: "parse basic headings",
			content: `<h2 id="intro">Introduction</h2>
<h3 id="overview">Overview</h3>
<h4 id="details">Details</h4>`,
			expected: []Heading{
				{ID: "intro", Level: 2, Text: "Introduction"},
				{ID: "overview", Level: 3, Text: "Overview"},
				{ID: "details", Level: 4, Text: "Details"},
			},
		},
		{
			name: "ignore h1 tags",
			content: `<h1 id="title">Page Title</h1>
<h2 id="section1">Section 1</h2>
<h3 id="subsection">Subsection</h3>`,
			expected: []Heading{
				{ID: "section1", Level: 2, Text: "Section 1"},
				{ID: "subsection", Level: 3, Text: "Subsection"},
			},
		},
		{
			name: "handle headings with links",
			content: `<h2 id="intro">Introduction<a href="#intro" class="anchor">🔗</a></h2>
<h3 id="overview">Overview<a href="#overview">link</a></h3>`,
			expected: []Heading{
				{ID: "intro", Level: 2, Text: "Introduction"},
				{ID: "overview", Level: 3, Text: "Overview"},
			},
		},
		{
			name: "trim whitespace",
			content: `<h2 id="spaced">  Spaced Heading  </h2>
<h3 id="tabbed">	Tabbed Heading	</h3>`,
			expected: []Heading{
				{ID: "spaced", Level: 2, Text: "Spaced Heading"},
				{ID: "tabbed", Level: 3, Text: "Tabbed Heading"},
			},
		},
		{
			name: "handle headings with attributes",
			content: `<h2 class="heading-class" id="styled" data-attr="value">Styled Heading</h2>
<h3 style="color: red;" id="colored">Colored Heading</h3>`,
			expected: []Heading{
				{ID: "styled", Level: 2, Text: "Styled Heading"},
				{ID: "colored", Level: 3, Text: "Colored Heading"},
			},
		},
		{
			name:    "all heading levels h2-h6",
			content: `<h2 id="h2">Level 2</h2><h3 id="h3">Level 3</h3><h4 id="h4">Level 4</h4><h5 id="h5">Level 5</h5><h6 id="h6">Level 6</h6>`,
			expected: []Heading{
				{ID: "h2", Level: 2, Text: "Level 2"},
				{ID: "h3", Level: 3, Text: "Level 3"},
				{ID: "h4", Level: 4, Text: "Level 4"},
				{ID: "h5", Level: 5, Text: "Level 5"},
				{ID: "h6", Level: 6, Text: "Level 6"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// test
			result := listHeadings(tt.content)

			// expect
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBuildHeadingTrees(t *testing.T) {
	// given
	tests := []struct {
		name     string
		headings []Heading
		expected []*headingNode
	}{
		{
			name:     "empty slice when no headings",
			headings: nil,
			expected: nil,
		},
		{
			name: "single heading",
			headings: []Heading{
				{ID: "intro", Level: 2, Text: "Introduction"},
			},
			expected: []*headingNode{
				{
					Heading:  Heading{ID: "intro", Level: 2, Text: "Introduction"},
					Children: nil,
				},
			},
		},
		{
			name: "nested structure",
			headings: []Heading{
				{ID: "section1", Level: 2, Text: "Section 1"},
				{ID: "subsection1", Level: 3, Text: "Subsection 1"},
				{ID: "subsubsection1", Level: 4, Text: "Sub-subsection 1"},
				{ID: "section2", Level: 2, Text: "Section 2"},
			},
			expected: []*headingNode{
				{
					Heading: Heading{ID: "section1", Level: 2, Text: "Section 1"},
					Children: []*headingNode{
						{
							Heading: Heading{ID: "subsection1", Level: 3, Text: "Subsection 1"},
							Children: []*headingNode{
								{
									Heading:  Heading{ID: "subsubsection1", Level: 4, Text: "Sub-subsection 1"},
									Children: nil,
								},
							},
						},
					},
				},
				{
					Heading:  Heading{ID: "section2", Level: 2, Text: "Section 2"},
					Children: nil,
				},
			},
		},
		{
			name: "multiple siblings at same level",
			headings: []Heading{
				{ID: "section1", Level: 2, Text: "Section 1"},
				{ID: "subsection1", Level: 3, Text: "Subsection 1"},
				{ID: "subsection2", Level: 3, Text: "Subsection 2"},
				{ID: "section2", Level: 2, Text: "Section 2"},
			},
			expected: []*headingNode{
				{
					Heading: Heading{ID: "section1", Level: 2, Text: "Section 1"},
					Children: []*headingNode{
						{
							Heading:  Heading{ID: "subsection1", Level: 3, Text: "Subsection 1"},
							Children: nil,
						},
						{
							Heading:  Heading{ID: "subsection2", Level: 3, Text: "Subsection 2"},
							Children: nil,
						},
					},
				},
				{
					Heading:  Heading{ID: "section2", Level: 2, Text: "Section 2"},
					Children: nil,
				},
			},
		},
		{
			name: "skipped levels",
			headings: []Heading{
				{ID: "section1", Level: 2, Text: "Section 1"},
				{ID: "subsection1", Level: 5, Text: "Subsection 1"}, // Skips levels 3 and 4
				{ID: "section2", Level: 2, Text: "Section 2"},
			},
			expected: []*headingNode{
				{
					Heading: Heading{ID: "section1", Level: 2, Text: "Section 1"},
					Children: []*headingNode{
						{
							Heading:  Heading{ID: "subsection1", Level: 5, Text: "Subsection 1"},
							Children: nil,
						},
					},
				},
				{
					Heading:  Heading{ID: "section2", Level: 2, Text: "Section 2"},
					Children: nil,
				},
			},
		},
		{
			name: "complex nesting with backtracking",
			headings: []Heading{
				{ID: "h1", Level: 2, Text: "H1"},
				{ID: "h2", Level: 3, Text: "H2"},
				{ID: "h3", Level: 4, Text: "H3"},
				{ID: "h4", Level: 3, Text: "H4"}, // Back to level 3
				{ID: "h5", Level: 2, Text: "H5"}, // Back to level 2
			},
			expected: []*headingNode{
				{
					Heading: Heading{ID: "h1", Level: 2, Text: "H1"},
					Children: []*headingNode{
						{
							Heading: Heading{ID: "h2", Level: 3, Text: "H2"},
							Children: []*headingNode{
								{
									Heading:  Heading{ID: "h3", Level: 4, Text: "H3"},
									Children: nil,
								},
							},
						},
						{
							Heading:  Heading{ID: "h4", Level: 3, Text: "H4"},
							Children: nil,
						},
					},
				},
				{
					Heading:  Heading{ID: "h5", Level: 2, Text: "H5"},
					Children: nil,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// test
			result := buildHeadingTrees(tt.headings)

			// expect
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSubstituter_Placeholder_ShouldReturnCorrectPlaceholder(t *testing.T) {
	// given
	substituter := NewSubstituer()

	// test
	result := substituter.Placeholder()

	// expect
	assert.Equal(t, "<p>{{summary}}</p>", result)
}

func TestNewSubstituer_ShouldCreateValidSubstituter(t *testing.T) {
	// test
	substituter := NewSubstituer()

	// expect
	assert.NotNil(t, substituter.template)
}

func TestSubstituter_Resolve(t *testing.T) {
	substituter := NewSubstituer()

	tests := []struct {
		name             string
		content          string
		expectedEmpty    bool
		expectedError    bool
		expectedContains []string
	}{
		{
			name:          "empty string when no headings",
			content:       "<p>Content without headings</p>",
			expectedEmpty: true,
		},
		{
			name: "generate table of contents",
			content: `<div>
				<h2 id="introduction">Introduction</h2>
				<p>Some content</p>
				<h3 id="overview">Overview</h3>
				<p>More content</p>
				<h2 id="conclusion">Conclusion</h2>
			</div>`,
			expectedContains: []string{
				"Introduction", "Overview", "Conclusion",
				`href="#introduction"`, `href="#overview"`, `href="#conclusion"`,
				"<nav>", "</nav>",
			},
		},
		{
			name: "handle nested headings",
			content: `<div>
				<h2 id="chapter1">Chapter 1</h2>
				<h3 id="section1-1">Section 1.1</h3>
				<h4 id="subsection1-1-1">Subsection 1.1.1</h4>
				<h2 id="chapter2">Chapter 2</h2>
			</div>`,
			expectedContains: []string{
				"Chapter 1", "Section 1.1", "Subsection 1.1.1", "Chapter 2",
				"<ul>", "<li",
			},
		},
		{
			name: "apply correct styling",
			content: `<div>
				<h2 id="level2">Level 2 Heading</h2>
				<h3 id="level3">Level 3 Heading</h3>
			</div>`,
			expectedContains: []string{
				"text-sm", "text-xs",
			},
		},
		{
			name: "complex nested structure with multiple levels",
			content: `<div>
				<h2 id="part1">Part I</h2>
				<h3 id="chapter1">Chapter 1</h3>
				<h4 id="section1-1">Section 1.1</h4>
				<h5 id="subsection1-1-1">Subsection 1.1.1</h5>
				<h6 id="subsubsection1-1-1-1">Sub-subsection 1.1.1.1</h6>
				<h3 id="chapter2">Chapter 2</h3>
				<h2 id="part2">Part II</h2>
			</div>`,
			expectedContains: []string{
				"Part I", "Chapter 1", "Section 1.1", "Subsection 1.1.1", "Sub-subsection 1.1.1.1",
				"Chapter 2", "Part II",
				"<nav>", "</nav>", "<ul>", "<li",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// test
			result, err := substituter.Resolve(tt.content)

			// expect
			if tt.expectedError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)

			if tt.expectedEmpty {
				assert.Empty(t, result)
				return
			}

			for _, expected := range tt.expectedContains {
				assert.Contains(t, result, expected, "Expected result to contain: %s", expected)
			}
		})
	}
}
