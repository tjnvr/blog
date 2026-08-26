package substitution

import (
	"fmt"
	"strings"

	"github.com/spf13/afero"

	"github.com/tjnvr/blog/internal/backbone/metadata"
	"github.com/tjnvr/blog/internal/backbone/section"
	"github.com/tjnvr/blog/internal/generator/page/html/substitution/articlelisting"
	"github.com/tjnvr/blog/internal/generator/page/html/substitution/content"
	"github.com/tjnvr/blog/internal/generator/page/html/substitution/description"
	"github.com/tjnvr/blog/internal/generator/page/html/substitution/navigation"
	"github.com/tjnvr/blog/internal/generator/page/html/substitution/summary"
	"github.com/tjnvr/blog/internal/generator/page/html/substitution/title"
	"github.com/tjnvr/blog/internal/hrefpath"
	"github.com/tjnvr/blog/internal/relpath"
)

// Registry manages substitutions and applies them to templates
type Registry struct {
	substitutions []Substituer
}

// NewRegistry creates a new substitution registry with default substituters.
func NewRegistry(fs afero.Fs, HTMLPath, markdownSourcePath string, m metadata.Metadata, sectionResolver section.Resolver, pagesPathResolver hrefpath.Resolver, assetsPathResolver relpath.Resolver) *Registry {
	return newRegistryWithSubstituters(
		content.NewSubstituer(HTMLPath, markdownSourcePath, assetsPathResolver, pagesPathResolver),
		articlelisting.NewSubstituer(fs, HTMLPath, markdownSourcePath, assetsPathResolver, pagesPathResolver),
		summary.NewSubstituer(),
		title.NewSubstituer(m),
		description.NewSubstituer(m),
		navigation.NewSubstituer(pagesPathResolver, sectionResolver, markdownSourcePath, HTMLPath),
	)
}

// newRegistryWithSubstituters creates a registry with custom substituters
func newRegistryWithSubstituters(subs ...Substituer) *Registry {
	return &Registry{
		substitutions: subs,
	}
}

// Apply applies all registered substitutions in the template at placeholder with content value resolved.
// A substituter is only resolved when its placeholder is actually present, so pages that don't use a
// given placeholder don't pay for (or risk failing on) work whose result would be discarded anyway.
func (r Registry) Apply(template, content string) (string, error) {
	result := template
	for _, s := range r.substitutions {
		if !strings.Contains(result, s.Placeholder()) {
			continue
		}

		resolution, err := s.Resolve(content)
		if err != nil {
			return "", fmt.Errorf("failed to resolve substitution: %w", err)
		}
		result = strings.ReplaceAll(result, s.Placeholder(), resolution)
	}
	return result, nil
}
