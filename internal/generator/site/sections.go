package site

import (
	"github.com/tjnvr/blog/internal/generator/section"
)

func (g *Generator) listSections(contentDir string) error {
	sections, err := section.ListSections(g.fs, contentDir)
	if err != nil {
		return err
	}
	g.sections = sections
	return nil
}
