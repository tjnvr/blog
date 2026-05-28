package site

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
)

func (g *Generator) copyAssets() error {
	return copyDir(g.fs, g.assetsDir, filepath.Join(g.buildDir, "assets"), nil)
}

func (g *Generator) copyScripts() error {
	return copyDir(g.fs, g.scriptsDir, filepath.Join(g.buildDir, "scripts"), func(path string) bool {
		return strings.HasSuffix(path, ".js")
	})
}

// copyDir copies files from srcDir to destDir using fs, optionally filtering by the provided function.
// If filter is nil, all files are copied. If filter returns true, the file is copied.
func copyDir(fs afero.Fs, srcDir, destDir string, filter func(path string) bool) error {
	return afero.Walk(fs, srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if filter != nil && !filter(path) {
			return nil
		}

		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}

		outPath := filepath.Join(destDir, relPath)

		data, err := afero.ReadFile(fs, path)
		if err != nil {
			return fmt.Errorf("reading %s: %w", path, err)
		}

		if err := fs.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
			return fmt.Errorf("creating directory for %s: %w", outPath, err)
		}

		if err := afero.WriteFile(fs, outPath, data, 0644); err != nil {
			return fmt.Errorf("writing %s: %w", outPath, err)
		}
		fmt.Printf("Copied: %s -> %s\n", relPath, outPath)
		return nil
	})
}
