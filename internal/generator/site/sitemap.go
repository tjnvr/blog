package site

import (
	"encoding/xml"
	"fmt"
	"path/filepath"

	"github.com/spf13/afero"
)

const (
	sitemapFileName           = "sitemap.xml"
	sitemapBaseURLPlaceholder = "__BASE_URL__"
	sitemapXMLNS              = "http://www.sitemaps.org/schemas/sitemap/0.9"
)

type sitemapEntry struct {
	href    string
	lastMod string
}

type sitemapURLSet struct {
	XMLName xml.Name     `xml:"urlset"`
	Xmlns   string       `xml:"xmlns,attr"`
	URLs    []sitemapURL `xml:"url"`
}

type sitemapURL struct {
	Loc     string `xml:"loc"`
	LastMod string `xml:"lastmod,omitempty"`
}

func (g *Generator) generateSitemap(entries []sitemapEntry) error {
	urlset := sitemapURLSet{Xmlns: sitemapXMLNS, URLs: make([]sitemapURL, 0, len(entries))}
	for _, e := range entries {
		urlset.URLs = append(urlset.URLs, sitemapURL{
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
