package site

import (
	"encoding/xml"
	"fmt"
	"path/filepath"

	"github.com/spf13/afero"

	"github.com/tjnvr/blog/internal/sitemap"
)

const (
	sitemapFileName           = "sitemap.xml"
	sitemapBaseURLPlaceholder = "__BASE_URL__"
)

type sitemapEntry struct {
	href    string
	lastMod string
}

func (g *Generator) generateSitemap(entries []sitemapEntry) error {
	urlset := sitemap.URLSet{Xmlns: sitemap.XMLNS, URLs: make([]sitemap.URL, 0, len(entries))}
	for _, e := range entries {
		urlset.URLs = append(urlset.URLs, sitemap.URL{
			Loc:     sitemapBaseURLPlaceholder + e.href,
			LastMod: e.lastMod,
		})
	}

	body, err := xml.MarshalIndent(urlset, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal sitemap: %w", err)
	}
	content := append([]byte(xml.Header), body...)

	if err := afero.WriteFile(g.fs, filepath.Join(g.PublicDir, sitemapFileName), content, 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", sitemapFileName, err)
	}
	return nil
}
