package fs_test

import (
	"fmt"

	"github.com/spf13/afero"
	"github.com/tjnvr/blog/internal/io/fs"
)

func ExampleDirCopier() {
	memFs := afero.NewMemMapFs()

	// Create some source files
	_ = afero.WriteFile(memFs, "source/docs/readme.txt", []byte("Hello World"), 0644)
	_ = afero.WriteFile(memFs, "source/docs/hidden.tmp", []byte("Ignore me"), 0644)
	_ = afero.WriteFile(memFs, "source/notes.txt", []byte("Some notes"), 0644)

	// Initialize the DirCopier and a FilesFinder
	// In this example, we only copy ".txt" files.
	copier := fs.NewDirCopier(memFs)
	finder := fs.NewFilesFinder(memFs, fs.WithExtension(".txt"))

	// Perform the copy operation
	err := copier.CopyDir(finder, "source", "destination")
	if err != nil {
		fmt.Printf("Copy failed: %v\n", err)
		return
	}

	// Verify the destination
	finder = fs.NewFilesFinder(memFs)
	files, err := finder.FindFiles("destination")
	if err != nil {
		fmt.Printf("FindFiles failed: %v\n", err)
		return
	}
	// Output will contain the paths relative to the destination directory
	for _, f := range files {
		fmt.Println(f)
	}

	// Unordered output:
	// docs/readme.txt
	// notes.txt
}
