// Package sitemap defines the sitemaps.org XML shape shared by the site
// generator, which writes sitemap.xml, and the post-deploy validator, which
// reads it back. Keeping one definition avoids the two independently
// agreeing, by coincidence, on the same wire format.
package sitemap

import "encoding/xml"

// XMLNS is the sitemaps.org namespace declared on the root <urlset> element.
const XMLNS = "http://www.sitemaps.org/schemas/sitemap/0.9"

// URLSet is the root element of a sitemap.xml document.
type URLSet struct {
	XMLName xml.Name `xml:"urlset"`
	Xmlns   string   `xml:"xmlns,attr"`
	URLs    []URL    `xml:"url"`
}

// URL describes one page listed in a sitemap.
type URL struct {
	// Loc is the page's absolute URL.
	Loc string `xml:"loc"`
	// LastMod is the page's last-modified date, in YYYY-MM-DD form. Empty
	// when the page declares no creation date.
	LastMod string `xml:"lastmod,omitempty"`
}
