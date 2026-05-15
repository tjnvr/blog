package substitution

import (
	"fmt"
	"strings"

	"github.com/tjnvr/blog/internal/generator/page/html/substitution/content"
	"github.com/tjnvr/blog/internal/generator/page/html/substitution/navigation"
	"github.com/tjnvr/blog/internal/generator/page/html/substitution/summary"
	"github.com/tjnvr/blog/internal/generator/page/html/substitution/title"
	"github.com/tjnvr/blog/internal/generator/section"
)

// Registry manages substitutions and applies them to templates
type Registry struct {
	substitutions []Substituer
}

// NewRegistry creates a new substitution registry with default substituters
func NewRegistry(filePath, markdownSourcePath string, assetsPathTranslater, markdownPathTranslater content.PathTranslater, sections []section.Section, currentSection string) *Registry {
	return NewRegistryWithSubstituters(
		content.NewSubstituer(filePath, markdownSourcePath, assetsPathTranslater, markdownPathTranslater),
		summary.NewSubstituer(),
		title.NewSubstituer(),
		navigation.NewSubstituer(sections, currentSection),
	)
}

// NewRegistryWithSubstituters creates a registry with custom substituters
func NewRegistryWithSubstituters(subs ...Substituer) *Registry {
	return &Registry{
		substitutions: subs,
	}
}

// Apply applies all registered substitutions in the template at placeholder with content value resolved
func (r Registry) Apply(template, content string) (string, error) {
	result := template
	for _, s := range r.substitutions {
		resolution, err := s.Resolve(content)
		if err != nil {
			return "", fmt.Errorf("failed to resolve substitution: %w", err)
		}
		result = strings.ReplaceAll(result, s.Placeholder(), resolution)
	}
	return result, nil
}
