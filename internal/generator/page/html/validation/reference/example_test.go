package reference_test

import (
	"fmt"

	"github.com/spf13/afero"

	"github.com/tjnvr/blog/internal/generator/page/html/validation/access"
	"github.com/tjnvr/blog/internal/generator/page/html/validation/htmlref"
	"github.com/tjnvr/blog/internal/generator/page/html/validation/reference"
)

func ExampleValidator_Validate() {
	// An image validator, wired the same way the registry wires it.
	validator := reference.NewValidator(
		htmlref.NewTagAttrExtractor("img", "src"),
		htmlref.NewClassifier(),
		access.NoopChecker{},
		access.NewResolver(afero.NewMemMapFs(), "/build", "/build/index.html"),
		nil,
		"/build/index.html", "image",
	)

	errs := validator.Validate([]byte(`<img src="/assets/missing.png">`))

	fmt.Println(errs[0])
	// Output:
	// /build/index.html: local image not found: /assets/missing.png
}
