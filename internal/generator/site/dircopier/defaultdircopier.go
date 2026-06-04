package dircopier

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/afero"
)

type defaultDirCopier struct{}

func (d defaultDirCopier) CopyDir(fs afero.Fs, from, to string) error {
	return afero.Walk(fs, from, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(from, path)
		if err != nil {
			return err
		}

		outPath := filepath.Join(to, relPath)

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
