package fs_test

import (
	"fmt"
	"sort"

	"github.com/spf13/afero"
	"github.com/tjnvr/blog/internal/io/fs"
)

func ExampleNewFilesFinder() {
	appFs := afero.NewMemMapFs()
	_ = afero.WriteFile(appFs, "file1.txt", []byte(""), 0644)
	_ = afero.WriteFile(appFs, "file2.txt", []byte(""), 0644)

	finder := fs.NewFilesFinder(appFs)
	files, _ := finder.FindFiles(".")
	sort.Strings(files)

	for _, f := range files {
		fmt.Println(f)
	}
	// Output:
	// file1.txt
	// file2.txt
}

func ExampleWithLevel() {
	appFs := afero.NewMemMapFs()
	_ = afero.WriteFile(appFs, "root.txt", []byte(""), 0644)
	_ = appFs.Mkdir("sub", 0755)
	_ = afero.WriteFile(appFs, "sub/sub.txt", []byte(""), 0644)

	finder := fs.NewFilesFinder(appFs, fs.WithLevel(1))
	files, _ := finder.FindFiles(".")
	sort.Strings(files)

	for _, f := range files {
		fmt.Println(f)
	}
	// Output:
	// root.txt
}

func ExampleWithExtension() {
	appFs := afero.NewMemMapFs()
	_ = afero.WriteFile(appFs, "a.txt", []byte(""), 0644)
	_ = afero.WriteFile(appFs, "b.go", []byte(""), 0644)

	finder := fs.NewFilesFinder(appFs, fs.WithExtension(".go"))
	files, _ := finder.FindFiles(".")
	sort.Strings(files)

	for _, f := range files {
		fmt.Println(f)
	}
	// Output:
	// b.go
}

func ExampleWithPattern() {
	appFs := afero.NewMemMapFs()
	_ = afero.WriteFile(appFs, "bar.txt", []byte(""), 0644)
	_ = afero.WriteFile(appFs, "foobar.go", []byte(""), 0644)

	finder := fs.NewFilesFinder(appFs, fs.WithPattern("foo"))
	files, _ := finder.FindFiles(".")
	sort.Strings(files)

	for _, f := range files {
		fmt.Println(f)
	}
	// Output:
	// foobar.go
}

func ExampleFilesFinder_FindFiles() {
	appFs := afero.NewMemMapFs()
	_ = appFs.Mkdir("data", 0755)
	_ = afero.WriteFile(appFs, "data/a.txt", []byte(""), 0644)
	_ = afero.WriteFile(appFs, "data/b.txt", []byte(""), 0644)

	finder := fs.NewFilesFinder(appFs)
	// Find files relative to the "data" directory
	files, _ := finder.FindFiles("data")
	sort.Strings(files)

	for _, f := range files {
		fmt.Println(f)
	}
	// Output:
	// a.txt
	// b.txt
}
