package main

import (
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/tjnvr/blog/internal/sitemap"
)

// placeholderBaseURL is the literal placeholder generated in sitemap.xml and
// robots.txt, meant to be replaced by the real base URL during deploy:sync's
// sed step. Finding it still present post-deploy means that substitution
// never happened.
const placeholderBaseURL = "__BASE_URL__"

func checkRobots(f *fetcher, baseURL string) error {
	target, err := resolve(baseURL, "/robots.txt")
	if err != nil {
		return err
	}
	body, err := f.get(target)
	if err != nil {
		return fmt.Errorf("robots.txt: %w", err)
	}
	if len(body) == 0 {
		return fmt.Errorf("robots.txt: empty body")
	}
	if strings.Contains(string(body), placeholderBaseURL) {
		return fmt.Errorf("robots.txt: unsubstituted %s placeholder found", placeholderBaseURL)
	}
	return nil
}

func fetchSitemap(f *fetcher, baseURL string) ([]string, error) {
	target, err := resolve(baseURL, "/sitemap.xml")
	if err != nil {
		return nil, err
	}
	body, err := f.get(target)
	if err != nil {
		return nil, fmt.Errorf("sitemap.xml: %w", err)
	}

	var urlset sitemap.URLSet
	if err := xml.Unmarshal(body, &urlset); err != nil {
		return nil, fmt.Errorf("sitemap.xml: invalid XML: %w", err)
	}

	pages := make([]string, 0, len(urlset.URLs))
	for _, u := range urlset.URLs {
		if strings.Contains(u.Loc, placeholderBaseURL) {
			return nil, fmt.Errorf("sitemap.xml: unsubstituted %s placeholder found in %q", placeholderBaseURL, u.Loc)
		}
		pages = append(pages, u.Loc)
	}
	if len(pages) == 0 {
		return nil, fmt.Errorf("sitemap.xml: no <url> entries found")
	}
	return pages, nil
}
