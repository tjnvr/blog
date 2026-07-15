package jssyntax_test

import (
	"fmt"

	"github.com/spf13/afero"

	"github.com/tjnvr/blog/internal/generator/page/html/validation/jssyntax"
)

func ExampleNewChecker() {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "/build/scripts/app.js", []byte("const answer = 42;"), 0644)

	checker := jssyntax.NewChecker(fs)

	fmt.Println(checker.Check("/build/scripts/app.js"))
	// Output:
	// []
}

func ExampleChecker_Check() {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "/build/scripts/broken.js", []byte("const = ;"), 0644)

	checker := jssyntax.NewChecker(fs)

	errs := checker.Check("/build/scripts/broken.js")

	fmt.Println(len(errs))
	// Output:
	// 1
}
