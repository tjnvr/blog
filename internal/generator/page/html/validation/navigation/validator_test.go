package navigation

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tjnvr/blog/internal/backbone/section"
	"github.com/tjnvr/blog/internal/relpath"
)

type fakeSectionResolver struct {
	sections []section.Section
	err      error
}

func (f fakeSectionResolver) Resolve() ([]section.Section, error) {
	return f.sections, f.err
}

func (f fakeSectionResolver) ResolveForFile(string) (section.Section, error) {
	return section.Section{}, nil
}

func TestNewValidator_ShouldCreateValidValidator(t *testing.T) {
	// setup
	v := NewValidator(fakeSectionResolver{}, relpath.NewResolver("content", "target"))

	// expect
	assert.NotNil(t, v)
}

func TestValidator_Validate(t *testing.T) {
	// given
	sections := []section.Section{
		{HomePath: "content/markdown/index.md", DisplayName: "Accueil"},
		{HomePath: "content/markdown/posts/index.md", DisplayName: "Posts"},
		{HomePath: "content/markdown/about/index.md", DisplayName: "About"},
	}
	pathResolver := relpath.NewResolver("content/markdown", "target/build", relpath.WithExtension(".html"))

	tests := []struct {
		name       string
		sections   []section.Section
		resolveErr error
		htmlPath   string
		html       string
		wantErrors int
		wantMsg    []string
	}{
		{
			name:     "valid nav from the root page",
			sections: sections,
			htmlPath: "target/build/index.html",
			html: `<html><body>
				<nav class="flex gap-4">
					<a href="index.html">Accueil</a>
					<a href="posts/index.html">Posts</a>
					<a href="about/index.html">About</a>
				</nav>
				<p>Content</p>
			</body></html>`,
			wantErrors: 0,
		},
		{
			name:     "valid nav from a nested section page",
			sections: sections,
			htmlPath: "target/build/about/index.html",
			html: `<html><body>
				<nav class="flex gap-4">
					<a href="../index.html">Accueil</a>
					<a href="../posts/index.html">Posts</a>
					<a href="index.html">About</a>
				</nav>
				<p>Content</p>
			</body></html>`,
			wantErrors: 0,
		},
		{
			name:     "nav links rendered in a different order than the sections",
			sections: sections,
			htmlPath: "target/build/index.html",
			html: `<html><body>
				<nav class="flex gap-4">
					<a href="index.html">Accueil</a>
					<a href="about/index.html">About</a>
					<a href="posts/index.html">Posts</a>
				</nav>
			</body></html>`,
			wantErrors: 1,
			wantMsg: []string{
				`link to section "content/markdown/about/index.md" is out of order`,
			},
		},
		{
			name:       "missing nav element entirely",
			sections:   sections,
			htmlPath:   "target/build/index.html",
			html:       `<html><body><p>No nav here</p></body></html>`,
			wantErrors: 1,
			wantMsg:    []string{"missing <nav> element"},
		},
		{
			name:     "nav missing a section link and display name",
			sections: sections,
			htmlPath: "target/build/index.html",
			html: `<html><body>
				<nav>
					<a href="index.html">Accueil</a>
					<a href="posts/index.html">Posts</a>
				</nav>
			</body></html>`,
			wantErrors: 2,
			wantMsg: []string{
				`missing link to section "content/markdown/about/index.md"`,
				`missing display name "About"`,
			},
		},
		{
			name:       "propagates sectionResolver error",
			resolveErr: errors.New("boom"),
			htmlPath:   "target/build/index.html",
			html:       `<html><body><nav></nav></body></html>`,
			wantErrors: 1,
			wantMsg:    []string{"sectionResolver.Resolve err"},
		},
		{
			name: "propagates pathResolver error for a section outside contentDirectory",
			sections: []section.Section{
				{HomePath: "other/index.md", DisplayName: "Elsewhere"},
			},
			htmlPath:   "target/build/index.html",
			html:       `<html><body><nav></nav></body></html>`,
			wantErrors: 1,
			wantMsg:    []string{"cannot resolve expected href"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// setup
			v := NewValidator(fakeSectionResolver{sections: tt.sections, err: tt.resolveErr}, pathResolver)

			// test
			errs := v.Validate(tt.htmlPath, "target/build", []byte(tt.html))

			// expect
			assert.Len(t, errs, tt.wantErrors)
			allErrs := ""
			for _, e := range errs {
				allErrs += e.Error() + "\n"
			}
			for _, msg := range tt.wantMsg {
				assert.Contains(t, allErrs, msg)
			}
		})
	}
}
