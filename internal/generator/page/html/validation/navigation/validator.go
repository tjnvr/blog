package navigation

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/tjnvr/blog/internal/generator/section"
)

// Validator checks that the generated HTML contains a <nav> element
// with links to all expected sections
type Validator struct {
	sections      []section.Section
	navRegex      *regexp.Regexp
	homeHrefRegex *regexp.Regexp
}

// NewValidator creates a new navigation validator that will check
// for the presence of nav links to all given sections plus the home page
func NewValidator(sections []section.Section) *Validator {
	return &Validator{
		sections:      sections,
		navRegex:      regexp.MustCompile(`(?s)<nav[^>]*>(.*?)</nav>`),
		homeHrefRegex: regexp.MustCompile(`href="(\.\./)*index\.html"`),
	}
}

// Validate checks the HTML content for a <nav> element containing links to all sections
func (v *Validator) Validate(htmlPath, buildDir string, content []byte) []error {
	var errs []error
	html := string(content)

	// Extract <nav> content
	navMatch := v.navRegex.FindStringSubmatch(html)
	if len(navMatch) < 2 {
		errs = append(errs, fmt.Errorf("%s: missing <nav> element", htmlPath))
		return errs
	}

	navContent := navMatch[1]

	for _, s := range v.sections {
		if s.DirName == "" {
			// Home section: href may be prefixed with ../ depending on depth
			if !strings.Contains(navContent, s.DisplayName) {
				errs = append(errs, fmt.Errorf("%s: navigation missing home link (%s)", htmlPath, s.DisplayName))
			}
			if !v.homeHrefRegex.MatchString(navContent) {
				errs = append(errs, fmt.Errorf("%s: navigation missing home href to index.html", htmlPath))
			}
		} else {
			expectedHref := s.DirName + "/index.html"
			if !strings.Contains(navContent, expectedHref) {
				errs = append(errs, fmt.Errorf("%s: navigation missing link to section %q (expected href containing %q)", htmlPath, s.DirName, expectedHref))
			}
			if !strings.Contains(navContent, s.DisplayName) {
				errs = append(errs, fmt.Errorf("%s: navigation missing display name %q for section %q", htmlPath, s.DisplayName, s.DirName))
			}
		}
	}

	return errs
}
