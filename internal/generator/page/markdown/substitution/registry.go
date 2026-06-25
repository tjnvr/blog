package substitution

import (
	"strings"

	"github.com/spf13/afero"
	"github.com/tjnvr/blog/internal/io/fs"
	"github.com/tjnvr/blog/internal/generator/page/markdown/substitution/listing"
	"github.com/tjnvr/blog/internal/generator/page/markdown/substitution/listing/article"
)

// Registry manages substitutions and applies them to templates
type Registry struct {
	substitutions []Substituer
}

// NewRegistry creates a new substitution registry with default substituters
func NewRegistry(filePath string) *Registry {
	return NewRegistryWithSubstituters(
		listing.NewSubstituer(
			"{{list-child-articles}}",
			article.NewPageArticlesLister(afero.NewOsFs(), filePath, fs.NewFilesFinder(afero.NewOsFs(), fs.WithExtension(".md"))),
			"\n",
		),
	)
}

// NewRegistryWithSubstituters creates a registry with custom substituters
func NewRegistryWithSubstituters(subs ...Substituer) *Registry {
	return &Registry{
		substitutions: subs,
	}
}

// Apply replaces substititutions place holders with resolution in content
func (r *Registry) Apply(content string) (string, error) {
	for _, s := range r.substitutions {
		if !strings.Contains(content, s.Placeholder()) {
			return content, nil
		}

		resolution, err := s.Resolve()
		if err != nil {
			return "", err
		}

		content = strings.ReplaceAll(content, s.Placeholder(), resolution)
	}

	return content, nil
}
