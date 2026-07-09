package relpath_test

import (
	"fmt"

	"github.com/tjnvr/blog/internal/relpath"
)

func ExampleResolver_Resolve() {
	// Assets were copied from content/assets directory to target/assets directory
	contentAssets := "content/assets"
	buildDirAssets := "target/assets"

	resolver := relpath.NewResolver(contentAssets, buildDirAssets)

	// Asset path in the contentAssets directory
	oldPath := "content/assets/picture.png"

	// We want to resolve new asset path relatively to fromPath
	fromPath := "target/posts/index.html"

	resolvedPath, _ := resolver.Resolve(oldPath, fromPath)
	fmt.Println(resolvedPath)

	// Output:
	// ../assets/picture.png
}
