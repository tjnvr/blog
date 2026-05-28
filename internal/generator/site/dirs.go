package site

import (
	"fmt"
	"path/filepath"
)

func (g *Generator) makeAllDirectories(dirs ...string) error {
	for _, d := range dirs {
		if err := g.fs.MkdirAll(filepath.Dir(d), 0755); err != nil {
			return fmt.Errorf("creating directory %s: %w", d, err)
		}
	}

	return nil
}
