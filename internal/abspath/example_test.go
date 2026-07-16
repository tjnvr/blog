package abspath_test

import (
	"fmt"

	"github.com/spf13/afero"

	"github.com/tjnvr/blog/internal/abspath"
)

func ExampleNewResolver() {
	resolver := abspath.NewResolver(afero.NewMemMapFs(), "/build", "/build/posts/article.html")

	fmt.Println(resolver.Resolve("../assets/logo.png"))
	// Output:
	// /build/assets/logo.png
}

func ExampleResolver_Exists() {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "/build/assets/logo.png", []byte("x"), 0644)

	resolver := abspath.NewResolver(fs, "/build", "/build/posts/article.html")

	fmt.Println(resolver.Exists("/assets/logo.png"))
	// Output:
	// true <nil>
}
