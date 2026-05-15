package navigation

import (
	"fmt"
	"strings"

	"github.com/tjnvr/blog/internal/generator/section"
)

// Substituter resolves {{navigation}} placeholder with an auto-generated nav bar
type Substituter struct {
	sections       []section.Section
	currentSection string
}

func NewSubstituer(sections []section.Section, currentSection string) Substituter {
	return Substituter{
		sections:       sections,
		currentSection: currentSection,
	}
}

func (n Substituter) Placeholder() string {
	return "{{navigation}}"
}

func (n Substituter) Resolve(_ string) (string, error) {
	prefix := relativePrefix(n.currentSection)

	var links []string
	for _, s := range n.sections {
		var href string
		if s.DirName == "" {
			href = prefix + "index.html"
		} else {
			href = prefix + s.DirName + "/index.html"
		}
		class := "hover:underline"
		if s.DirName == n.currentSection {
			class = "font-semibold underline"
		}
		links = append(links, fmt.Sprintf(`<a href="%s" class="%s">%s</a>`, href, class, s.DisplayName))
	}

	return fmt.Sprintf(`<nav class="flex flex-col sm:flex-row gap-4">%s</nav>`, strings.Join(links, "\n    ")), nil
}

// relativePrefix returns the "../" prefix needed to reach the site root from the current section.
func relativePrefix(currentSection string) string {
	if currentSection == "" {
		return ""
	}
	depth := strings.Count(currentSection, "/") + 1
	return strings.Repeat("../", depth)
}
