package sectionslister

import (
	"os"
	"slices"

	"github.com/spf13/afero"
	"github.com/tjnvr/blog/internal/generator/backbone/section"
)

type defaultSectionslister struct{}

func (d defaultSectionslister) ListSections(fs afero.Fs, sectionExtractor section.Resolver, contentDir string) ([]section.Section, error) {
	sections := make([]section.Section, 0)
	return sections, afero.Walk(fs, contentDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			return nil
		}

		s, err := sectionExtractor.Resolve(path)
		if err != nil {
			return err
		}

		if  !slices.ContainsFunc(sections, func(e section.Section) bool {
			return e == s
		}) {
			sections = append(sections, s)
		}

		return nil
	})
}
