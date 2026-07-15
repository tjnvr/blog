package section_test

import (
	"fmt"

	"github.com/spf13/afero"
	"github.com/tjnvr/blog/internal/backbone/section"
)

func ExampleNewResolver() {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "content/index.md", []byte("# Home"), 0644)

	resolver := section.NewResolver(fs, "content")
	sections, _ := resolver.Resolve()

	fmt.Println(len(sections))
	// Output:
	// 1
}

// Sections are returned in the order given by their "seq" metadata.
func ExampleResolver_Resolve() {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "content/index.md", []byte("<!-- seq: 1 -->\n# Home"), 0644)
	_ = afero.WriteFile(fs, "content/blog/index.md", []byte("<!-- seq: 2 -->\n# Blog"), 0644)
	_ = afero.WriteFile(fs, "content/about/index.md", []byte("<!-- seq: 3 -->\n# About"), 0644)

	resolver := section.NewResolver(fs, "content")
	sections, _ := resolver.Resolve()

	for _, s := range sections {
		fmt.Println(s.DisplayName)
	}
	// Output:
	// Home
	// Blog
	// About
}

// A section without a "seq" comment is listed after every sequenced one,
// ordered by display name.
func ExampleResolver_Resolve_unsequenced() {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "content/index.md", []byte("<!-- seq: 1 -->\n# Home"), 0644)
	_ = afero.WriteFile(fs, "content/blog/index.md", []byte("# Blog"), 0644)
	_ = afero.WriteFile(fs, "content/about/index.md", []byte("# About"), 0644)

	resolver := section.NewResolver(fs, "content")
	sections, _ := resolver.Resolve()

	for _, s := range sections {
		fmt.Println(s.DisplayName)
	}
	// Output:
	// Home
	// About
	// Blog
}

func ExampleResolver_ResolveForFile() {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "content/index.md", []byte("# Home"), 0644)
	_ = afero.WriteFile(fs, "content/blog/index.md", []byte("# Blog"), 0644)
	_ = afero.WriteFile(fs, "content/blog/my-post.md", []byte("# My post"), 0644)

	resolver := section.NewResolver(fs, "content")
	s, _ := resolver.ResolveForFile("content/blog/my-post.md")

	fmt.Println(s.DisplayName)
	fmt.Println(s.HomePath)
	// Output:
	// Blog
	// content/blog/index.md
}
