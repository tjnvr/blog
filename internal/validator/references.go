package main

import (
	"fmt"
	"regexp"
	"strings"

	jsparser "github.com/dop251/goja/parser"
)

var skipPrefixes = []string{"#", "mailto:", "tel:", "javascript:", "data:"}

func skippable(ref string) bool {
	for _, p := range skipPrefixes {
		if strings.HasPrefix(ref, p) {
			return true
		}
	}
	return false
}

func extractAttr(body []byte, tag, attr string) []string {
	re := regexp.MustCompile(`<` + regexp.QuoteMeta(tag) + `[^>]+` + regexp.QuoteMeta(attr) + `="([^"]+)"`)
	matches := re.FindAllSubmatch(body, -1)
	refs := make([]string, 0, len(matches))
	for _, m := range matches {
		if len(m) < 2 {
			continue
		}
		refs = append(refs, string(m[1]))
	}
	return refs
}

func checkImages(f *fetcher, pageURL string, body []byte) (errs, warnings []error) {
	return checkRefsReachable(f, pageURL, body, "img", "src")
}

func checkLinks(f *fetcher, pageURL string, body []byte) (errs, warnings []error) {
	return checkRefsReachable(f, pageURL, body, "a", "href")
}

func checkRefsReachable(f *fetcher, pageURL string, body []byte, tag, attr string) (errs, warnings []error) {
	for _, ref := range extractAttr(body, tag, attr) {
		if skippable(ref) {
			continue
		}
		target, err := resolve(pageURL, ref)
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %s %s: %w", pageURL, tag, ref, err))
			continue
		}
		if err := f.head(target); err != nil {
			wrapped := fmt.Errorf("%s: %s %s not accessible: %w", pageURL, tag, ref, err)
			if isExternal(pageURL, target) {
				warnings = append(warnings, wrapped)
			} else {
				errs = append(errs, wrapped)
			}
		}
	}
	return errs, warnings
}

// The fetch is needed for reachability anyway, so parsing the body as
// JavaScript to catch transit corruption is free.
func checkScripts(f *fetcher, pageURL string, body []byte) (errs, warnings []error) {
	for _, ref := range extractAttr(body, "script", "src") {
		if skippable(ref) {
			continue
		}
		target, err := resolve(pageURL, ref)
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: script %s: %w", pageURL, ref, err))
			continue
		}
		external := isExternal(pageURL, target)

		scriptBody, err := f.get(target)
		if err != nil {
			wrapped := fmt.Errorf("%s: script %s not accessible: %w", pageURL, ref, err)
			if external {
				warnings = append(warnings, wrapped)
			} else {
				errs = append(errs, wrapped)
			}
			continue
		}

		if _, err := jsparser.ParseFile(nil, target, string(scriptBody), 0); err != nil {
			wrapped := fmt.Errorf("%s: script %s: JavaScript syntax error: %w", pageURL, ref, err)
			if external {
				warnings = append(warnings, wrapped)
			} else {
				errs = append(errs, wrapped)
			}
		}
	}
	return errs, warnings
}
