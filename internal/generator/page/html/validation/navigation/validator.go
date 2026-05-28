package navigation

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/tjnvr/blog/internal/generator/section"

	"github.com/spf13/afero"
)

// Validator checks that the generated HTML contains a <nav> element
// with links to all expected sections
type Validator struct {
	fs            afero.Fs
	sections      []section.Section
	navRegex      *regexp.Regexp
	homeHrefRegex *regexp.Regexp
	HTMLPath      string
}

// NewValidator creates a new navigation validator that will check
// for the presence of nav links to all given sections plus the home page
func NewValidator(fs afero.Fs, HTMLPath string, sections []section.Section) *Validator {
	return &Validator{
		fs:            fs,
		sections:      sections,
		navRegex:      regexp.MustCompile(`(?s)<nav[^>]*>(.*?)</nav>`),
		homeHrefRegex: regexp.MustCompile(`href="(\.\./)*index\.html"`),
		HTMLPath:      HTMLPath,
	}
}

// Validate checks the HTML content for a <nav> element containing links to all sections
func (v *Validator) Validate(content []byte) []error {
	var errs []error
	html := string(content)

	// Extract <nav> content
	navMatch := v.navRegex.FindStringSubmatch(html)
	if len(navMatch) < 2 {
		errs = append(errs, fmt.Errorf("%s: missing <nav> element", v.HTMLPath))
		return errs
	}

	navContent := navMatch[1]

	for _, s := range v.sections {
		if s.DirName == "" {
			// Home section: href may be prefixed with ../ depending on depth
			if !strings.Contains(navContent, s.DisplayName) {
				errs = append(errs, fmt.Errorf("%s: navigation missing home link (%s)", v.HTMLPath, s.DisplayName))
			}
			if !v.homeHrefRegex.MatchString(navContent) {
				errs = append(errs, fmt.Errorf("%s: navigation missing home href to index.html", v.HTMLPath))
			}
		} else {
			expectedHref := s.DirName + "/index.html"
			if !strings.Contains(navContent, expectedHref) {
				errs = append(errs, fmt.Errorf("%s: navigation missing link to section %q (expected href containing %q)", v.HTMLPath, s.DirName, expectedHref))
			}
			if !strings.Contains(navContent, s.DisplayName) {
				errs = append(errs, fmt.Errorf("%s: navigation missing display name %q for section %q", v.HTMLPath, s.DisplayName, s.DirName))
			}
		}
	}

	return errs
}
