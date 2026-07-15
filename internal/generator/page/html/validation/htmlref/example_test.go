package htmlref_test

import (
	"fmt"

	"github.com/tjnvr/blog/internal/generator/page/html/validation/htmlref"
)

func ExampleNewTagAttrExtractor() {
	extractor := htmlref.NewTagAttrExtractor("img", "src")

	refs := extractor.Extract([]byte(`<img src="/assets/logo.png" alt="logo">`))

	fmt.Println(refs[0])
	// Output:
	// /assets/logo.png
}

func ExampleExtractor_Extract() {
	extractor := htmlref.NewTagAttrExtractor("a", "href")

	refs := extractor.Extract([]byte(`<a href="/">Home</a><a href="/about/">About</a>`))

	fmt.Println(refs)
	// Output:
	// [/ /about/]
}

func ExampleNewClassifier() {
	classifier := htmlref.NewClassifier("#", "mailto:")

	fmt.Println(classifier.Classify("mailto:hi@example.com"))
	// Output:
	// skip
}

func ExampleClassifier_Classify() {
	classifier := htmlref.NewClassifier("#", "mailto:", "tel:", "javascript:")

	fmt.Println(classifier.Classify("https://example.com"))
	fmt.Println(classifier.Classify("#top"))
	fmt.Println(classifier.Classify("/posts/index.html"))
	// Output:
	// external
	// skip
	// local
}

func ExampleKind_String() {
	fmt.Println(htmlref.KindExternal)
	// Output:
	// external
}
