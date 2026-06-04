package section

import (
	"github.com/spf13/afero"
	"github.com/tjnvr/blog/internal/generator/backbone/pages"
)

type (
	Resolver interface {
		Resolve(pageFilePath string) (Section, error)

		ResolveAll() ([]Section, error)
	}

	ResolverFactory interface {
		Create(fs afero.Fs, contentDir string, pagesResolver pages.Resolver) Resolver
	}
)
