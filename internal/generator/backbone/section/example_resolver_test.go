package section_test

import (
	"fmt"
	"sort"

	"github.com/spf13/afero"
	"github.com/tjnvr/blog/internal/generator/backbone/section"
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

func ExampleResolver_Resolve() {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "content/index.md", []byte("# Home"), 0644)
	_ = afero.WriteFile(fs, "content/blog/index.md", []byte("# Blog"), 0644)
	_ = afero.WriteFile(fs, "content/about/index.md", []byte("# About"), 0644)

	resolver := section.NewResolver(fs, "content")
	sections, _ := resolver.Resolve()

	names := make([]string, 0, len(sections))
	for _, s := range sections {
		names = append(names, s.DisplayName)
	}
	sort.Strings(names)

	for _, name := range names {
		fmt.Println(name)
	}
	// Output:
	// About
	// Blog
	// Home
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
