package fs

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/afero"
)

type (
	dirCopier struct {
		fs afero.Fs
	}

	// DirCopier defines the interface for copying a directory's contents.
	DirCopier interface {
		// CopyDir copies files from the 'from' directory to the 'to' directory
		// using the provided FilesFinder to filter and locate files.
		CopyDir(filesFinder FilesFinder, from, to string) error
	}
)

// NewDirCopier creates a new DirCopier instance using the provided afero.Fs.
func NewDirCopier(fs afero.Fs) DirCopier {
	return dirCopier{
		fs: fs,
	}
}

func (c dirCopier) CopyDir(filesFinder FilesFinder, from, to string) error {
	files, err := filesFinder.FindFiles(from)
	if err != nil {
		return fmt.Errorf("FindFiles err: %v", err)
	}

	for _, file := range files {
		sourcePath := filepath.Join(from, file)
		content, err := afero.ReadFile(c.fs, sourcePath)
		if err != nil {
			return fmt.Errorf("afero.ReadFile err for '%s': %v", sourcePath, err)
		}

		destinationPath := filepath.Join(to, file)
		if err := c.fs.MkdirAll(filepath.Dir(destinationPath), 0744); err != nil {
			return fmt.Errorf("MkdirAll err: %v", err)
		}
		err = afero.WriteFile(c.fs, destinationPath, content, 0644)
		if err != nil {
			return fmt.Errorf("afero.WriteFile err for '%s': %v", destinationPath, err)
		}
	}

	return nil
}
