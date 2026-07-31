package site

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testConfig() Config {
	return Config{
		ContentDir:    "content",
		PublicDir:     "target",
		AssetsDir:     "assets",
		AssetsOutDir:  "target/assets",
		ScriptsDir:    "scripts",
		ScriptsOutDir: "target/scripts",
	}
}

func TestNewGenerator_ResolverWiring(t *testing.T) {
	tests := []struct {
		name         string
		markdownPath string
		wantHTMLPath string
		wantHref     string
	}{
		{
			name:         "root index.md writes extensionless and links to the root URL",
			markdownPath: "content/index.md",
			wantHTMLPath: "target/index",
			wantHref:     "/",
		},
		{
			name:         "section index.md writes into its own directory and links to that directory's URL",
			markdownPath: "content/posts/index.md",
			wantHTMLPath: "target/posts/index",
			wantHref:     "/posts",
		},
		{
			name:         "leaf page keeps the same shape on disk and in its href",
			markdownPath: "content/posts/java-learning-roadmap.md",
			wantHTMLPath: "target/posts/java-learning-roadmap",
			wantHref:     "/posts/java-learning-roadmap",
		},
	}

	g, err := NewGenerator(afero.NewMemMapFs(), testConfig())
	require.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// file-writing destination: WithExtension("") must still strip .md to nothing
			htmlPath, err := g.pagePathsResolver.Resolve(tt.markdownPath, ".")
			require.NoError(t, err)
			assert.Equal(t, tt.wantHTMLPath, htmlPath)

			// href: index.md renders as its containing directory, absolute
			href, err := g.hrefPathsResolver.Resolve(tt.markdownPath)
			require.NoError(t, err)
			assert.Equal(t, tt.wantHref, href)
		})
	}
}
