package site

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/afero"
)

// copyDir copies all files in srcDir to destDir using fs
func copyDir(fs afero.Fs, srcDir, destDir string) error {
	return afero.Walk(fs, srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
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

		if err := afero.WriteFile(fs, outPath, data, 0644); err != nil {
			return fmt.Errorf("writing %s: %w", outPath, err)
		}

		fmt.Printf("Copied: %s -> %s\n", relPath, outPath)
		return nil
	})
}
