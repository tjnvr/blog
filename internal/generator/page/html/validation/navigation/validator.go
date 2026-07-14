package navigation

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/tjnvr/blog/internal/generator/backbone/section"
	"github.com/tjnvr/blog/internal/relpath"
)

// Validator checks that the generated HTML contains a <nav> element
// with links to all expected sections
type Validator struct {
	sectionResolver section.Resolver
	pathResolver    relpath.Resolver
	navRegex        *regexp.Regexp
}

// NewValidator creates a new navigation validator that will check for the
// presence of nav links to all sections known to sectionResolver. pathResolver
// must resolve a section's markdown HomePath into the same href nav.Substituter
// would generate, so the expected and generated hrefs stay in sync.
func NewValidator(sectionResolver section.Resolver, pathResolver relpath.Resolver) *Validator {
	return &Validator{
		sectionResolver: sectionResolver,
		pathResolver:    pathResolver,
		navRegex:        regexp.MustCompile(`(?s)<nav[^>]*>(.*?)</nav>`),
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

	sections, err := v.sectionResolver.Resolve()
	if err != nil {
		errs = append(errs, fmt.Errorf("sectionResolver.Resolve err: %v", err))
		return errs
	}

	for _, s := range sections {
		expectedHref, err := v.pathResolver.Resolve(s.HomePath, htmlPath)
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: cannot resolve expected href for section %q: %v", htmlPath, s.HomePath, err))
			continue
		}

		if !strings.Contains(navContent, fmt.Sprintf(`href="%s"`, expectedHref)) {
			errs = append(errs, fmt.Errorf("%s: navigation missing link to section %q (expected href %q)", htmlPath, s.HomePath, expectedHref))
		}
		if !strings.Contains(navContent, s.DisplayName) {
			errs = append(errs, fmt.Errorf("%s: navigation missing display name %q for section %q", htmlPath, s.DisplayName, s.HomePath))
		}
	}

	return errs
}
