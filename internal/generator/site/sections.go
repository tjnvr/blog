package site

import (
	"github.com/tjnvr/blog/internal/generator/section"
)

func (g *Generator) listSections() error {
	sections, err := section.ListSections(g.fs, g.contentDir)
	if err != nil {
		return err
	}
	g.sections = sections
	return nil
}
