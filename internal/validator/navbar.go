package main

import (
	"fmt"
	"regexp"
	"strings"
)

// navRegex matches the same <nav>...</nav> shape as the generator's build-time
// navigation.Validator. This check only confirms a <nav> is present and
// non-empty (catches truncation, a stale cache, or the wrong page being
// served) — structural correctness (right sections, right order, right
// labels) is a pre-deploy-only concern, already proven and gated before
// deploy runs.
var navRegex = regexp.MustCompile(`(?s)<nav[^>]*>(.*?)</nav>`)

func checkNavbar(pageURL string, body []byte) error {
	match := navRegex.FindSubmatch(body)
	if match == nil {
		return fmt.Errorf("%s: missing <nav> element", pageURL)
	}
	if strings.TrimSpace(string(match[1])) == "" {
		return fmt.Errorf("%s: <nav> element is empty", pageURL)
	}
	return nil
}
