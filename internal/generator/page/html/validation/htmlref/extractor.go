package htmlref

import "regexp"

// Extractor locates all references of a single kind within HTML content, in
// document order. It replaces the per-validator pattern of compiling a regexp,
// running FindAllSubmatch and reading the first capture group by hand.
type Extractor interface {
	// Extract returns the raw value of every matched attribute found in
	// content. It returns an empty (non-nil) slice when nothing matches.
	Extract(content []byte) []string
}

type tagAttrExtractor struct {
	re *regexp.Regexp
}

// NewTagAttrExtractor returns an Extractor that yields the value of the attr
// attribute on every tag element in HTML content. For example
// NewTagAttrExtractor("img", "src") matches <img ... src="VALUE" ...> and
// returns VALUE for each occurrence.
func NewTagAttrExtractor(tag, attr string) Extractor {
	re := regexp.MustCompile(`<` + regexp.QuoteMeta(tag) + `[^>]+` + regexp.QuoteMeta(attr) + `="([^"]+)"`)
	return tagAttrExtractor{re: re}
}

func (e tagAttrExtractor) Extract(content []byte) []string {
	matches := e.re.FindAllSubmatch(content, -1)
	refs := make([]string, 0, len(matches))
	for _, m := range matches {
		if len(m) < 2 {
			continue
		}
		refs = append(refs, string(m[1]))
	}
	return refs
}
