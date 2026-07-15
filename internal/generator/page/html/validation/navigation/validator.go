package navigation

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/tjnvr/blog/internal/generator/backbone/section"
	"github.com/tjnvr/blog/internal/relpath"
)

// Validator checks that the generated HTML contains a <nav> element with links
// to all expected sections.
type Validator struct {
	sectionResolver section.Resolver
	pathResolver    relpath.Resolver
	htmlPath        string
	navRegex        *regexp.Regexp
}

// NewValidator creates a navigation validator for the page at htmlPath that
// checks for nav links to every section known to sectionResolver. pathResolver
// must resolve a section's markdown HomePath into the same href the navigation
// substituter generates, so expected and generated hrefs stay in sync. htmlPath
// is fixed for the page and captured here rather than passed to Validate.
func NewValidator(sectionResolver section.Resolver, pathResolver relpath.Resolver, htmlPath string) *Validator {
	return &Validator{
		sectionResolver: sectionResolver,
		pathResolver:    pathResolver,
		htmlPath:        htmlPath,
		navRegex:        regexp.MustCompile(`(?s)<nav[^>]*>(.*?)</nav>`),
	}
}

// Validate checks content for a <nav> element containing links to all sections.
func (v *Validator) Validate(content []byte) []error {
	var errs []error
	html := string(content)

	navMatch := v.navRegex.FindStringSubmatch(html)
	if len(navMatch) < 2 {
		return append(errs, fmt.Errorf("%s: missing <nav> element", v.htmlPath))
	}
	navContent := navMatch[1]

	sections, err := v.sectionResolver.Resolve()
	if err != nil {
		return append(errs, fmt.Errorf("sectionResolver.Resolve err: %v", err))
	}

	for _, s := range sections {
		expectedHref, err := v.pathResolver.Resolve(s.HomePath, v.htmlPath)
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: cannot resolve expected href for section %q: %v", v.htmlPath, s.HomePath, err))
			continue
		}

		if !strings.Contains(navContent, fmt.Sprintf(`href="%s"`, expectedHref)) {
			errs = append(errs, fmt.Errorf("%s: navigation missing link to section %q (expected href %q)", v.htmlPath, s.HomePath, expectedHref))
		}
		if !strings.Contains(navContent, s.DisplayName) {
			errs = append(errs, fmt.Errorf("%s: navigation missing display name %q for section %q", v.htmlPath, s.DisplayName, s.HomePath))
		}
	}

	return errs
}
