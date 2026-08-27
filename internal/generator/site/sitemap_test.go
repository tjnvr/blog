package site

import (
	"encoding/xml"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateSitemap_WritesExpectedXML(t *testing.T) {
	// given
	fs := afero.NewMemMapFs()
	g := &Generator{Config: Config{PublicDir: "target"}, fs: fs}
	entries := []sitemapEntry{
		{href: "/", lastMod: "2024-01-01"},
		{href: "/posts", lastMod: ""},
	}

	// test
	err := g.generateSitemap(entries)

	// expect
	require.NoError(t, err)
	content, err := afero.ReadFile(fs, "target/sitemap.xml")
	require.NoError(t, err)
	assert.Equal(t, xml.Header, string(content[:len(xml.Header)]))

	var urlset sitemapURLSet
	require.NoError(t, xml.Unmarshal(content[len(xml.Header):], &urlset))
	assert.Equal(t, sitemapXMLNS, urlset.Xmlns)
	assert.Equal(t, []sitemapURL{
		{Loc: "__BASE_URL__/", LastMod: "2024-01-01"},
		{Loc: "__BASE_URL__/posts", LastMod: ""},
	}, urlset.URLs)
}

func TestGenerateSitemap_WritesEmptyURLSetWhenNoEntries(t *testing.T) {
	// given
	fs := afero.NewMemMapFs()
	g := &Generator{Config: Config{PublicDir: "target"}, fs: fs}

	// test
	err := g.generateSitemap(nil)

	// expect
	require.NoError(t, err)
	content, err := afero.ReadFile(fs, "target/sitemap.xml")
	require.NoError(t, err)

	var urlset sitemapURLSet
	require.NoError(t, xml.Unmarshal(content[len(xml.Header):], &urlset))
	assert.Empty(t, urlset.URLs)
}
