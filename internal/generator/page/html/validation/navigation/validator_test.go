package navigation

import (
	"testing"

	"github.com/tjnvr/blog/internal/generator/section"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestValidator_Validate(t *testing.T) {
	tests := []struct {
		name       string
		sections   []section.Section
		html       string
		wantErrors int
		wantMsg    []string
	}{
		{
			name: "valid nav with all sections from root",
			sections: []section.Section{
				{DirName: "", DisplayName: "Accueil"},
				{DirName: "posts", DisplayName: "Posts"},
				{DirName: "about", DisplayName: "About"},
			},
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
			name: "valid nav with all sections from section depth",
			sections: []section.Section{
				{DirName: "", DisplayName: "Accueil"},
				{DirName: "posts", DisplayName: "Posts"},
				{DirName: "about", DisplayName: "About"},
			},
			html: `<html><body>
				<nav class="flex gap-4">
					<a href="../index.html">Accueil</a>
					<a href="../posts/index.html">Posts</a>
					<a href="../about/index.html">About</a>
				</nav>
				<p>Content</p>
			</body></html>`,
			wantErrors: 0,
		},
		{
			name:       "missing nav element entirely",
			sections:   []section.Section{{DirName: "", DisplayName: "Accueil"}, {DirName: "posts", DisplayName: "Posts"}},
			html:       `<html><body><p>No nav here</p></body></html>`,
			wantErrors: 1,
			wantMsg:    []string{"missing <nav> element"},
		},
		{
			name: "nav missing a section link",
			sections: []section.Section{
				{DirName: "", DisplayName: "Accueil"},
				{DirName: "posts", DisplayName: "Posts"},
				{DirName: "about", DisplayName: "About"},
			},
			html: `<html><body>
				<nav>
					<a href="index.html">Accueil</a>
					<a href="posts/index.html">Posts</a>
				</nav>
			</body></html>`,
			wantErrors: 2,
			wantMsg:    []string{"missing link to section \"about\"", "missing display name \"About\""},
		},
		{
			name:     "nav missing home link",
			sections: []section.Section{{DirName: "", DisplayName: "Accueil"}, {DirName: "posts", DisplayName: "Posts"}},
			html: `<html><body>
				<nav>
					<a href="posts/index.html">Posts</a>
				</nav>
			</body></html>`,
			wantErrors: 2,
			wantMsg:    []string{"missing home link (Accueil)", "missing home href"},
		},
		{
			name:     "nav with home but wrong display name",
			sections: []section.Section{{DirName: "", DisplayName: "Accueil"}},
			html: `<html><body>
				<nav>
					<a href="index.html">Home</a>
				</nav>
			</body></html>`,
			wantErrors: 1,
			wantMsg:    []string{"missing home link (Accueil)"},
		},
		{
			name:     "only home section",
			sections: []section.Section{{DirName: "", DisplayName: "Accueil"}},
			html: `<html><body>
				<nav>
					<a href="index.html">Accueil</a>
				</nav>
			</body></html>`,
			wantErrors: 0,
		},
		{
			name: "home display name from root index.md title",
			sections: []section.Section{
				{DirName: "", DisplayName: "Bienvenue sur mon blog"},
				{DirName: "posts", DisplayName: "All Articles"},
			},
			html: `<html><body>
				<nav>
					<a href="index.html">Bienvenue sur mon blog</a>
					<a href="posts/index.html">All Articles</a>
				</nav>
			</body></html>`,
			wantErrors: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator(afero.NewMemMapFs(), "test.html", tt.sections)
			errs := v.Validate([]byte(tt.html))
			assert.Len(t, errs, tt.wantErrors)
			if len(tt.wantMsg) > 0 {
				allErrs := ""
				for _, e := range errs {
					allErrs += e.Error() + "\n"
				}
				for _, msg := range tt.wantMsg {
					assert.Contains(t, allErrs, msg)
				}
			}
		})
	}
}
