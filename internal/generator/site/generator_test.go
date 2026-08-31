package site

import (
	"encoding/xml"
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

// TestGenerate_WritesSitemapWithPlaceholderAndPageHrefs proves sitemap.xml is
// produced automatically as part of Generate(), with every page's href
// prefixed by the deploy-time base URL placeholder and lastmod populated
// only when a page declares creation-date metadata.
func TestGenerate_WritesSitemapWithPlaceholderAndPageHrefs(t *testing.T) {
	// given
	fs := afero.NewMemMapFs()
	require.NoError(t, afero.WriteFile(fs, "content/index.md", []byte("# Home\n\n<!-- creation-date: 2024-01-15 -->"), 0644))
	require.NoError(t, afero.WriteFile(fs, "content/about.md", []byte("# About"), 0644))
	require.NoError(t, fs.MkdirAll("assets", 0744))
	require.NoError(t, fs.MkdirAll("scripts", 0744))

	g, err := NewGenerator(fs, testConfig(), WithSkipURLValidation(true))
	require.NoError(t, err)

	// test
	require.NoError(t, g.Generate())

	// expect
	sitemapContent, err := afero.ReadFile(fs, "target/sitemap.xml")
	require.NoError(t, err)
	assert.Equal(t, xml.Header, string(sitemapContent[:len(xml.Header)]), "sitemap must start with the standard XML prolog")

	var urlset sitemapURLSet
	require.NoError(t, xml.Unmarshal(sitemapContent[len(xml.Header):], &urlset))
	assert.Equal(t, sitemapXMLNS, urlset.Xmlns)
	assert.ElementsMatch(t, []sitemapURL{
		{Loc: "__BASE_URL__/", LastMod: "2024-01-15"},
		{Loc: "__BASE_URL__/about", LastMod: ""},
	}, urlset.URLs)
}

func TestGenerate_WritesRobotsTxt(t *testing.T) {
	// given
	fs := afero.NewMemMapFs()
	require.NoError(t, afero.WriteFile(fs, "content/index.md", []byte("# Home"), 0644))
	require.NoError(t, fs.MkdirAll("assets", 0744))
	require.NoError(t, fs.MkdirAll("scripts", 0744))

	g, err := NewGenerator(fs, testConfig(), WithSkipURLValidation(true))
	require.NoError(t, err)

	// test
	require.NoError(t, g.Generate())

	// expect
	content, err := afero.ReadFile(fs, "target/robots.txt")
	require.NoError(t, err)
	assert.Equal(t, "User-agent: *\nDisallow:\nSitemap: __BASE_URL__/sitemap.xml", string(content))
}
