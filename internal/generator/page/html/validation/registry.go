package validation

import (
	"errors"
	"time"

	"github.com/spf13/afero"

	"github.com/tjnvr/blog/internal/abspath"
	"github.com/tjnvr/blog/internal/backbone/section"
	"github.com/tjnvr/blog/internal/generator/page/html/validation/access"
	"github.com/tjnvr/blog/internal/generator/page/html/validation/htmlref"
	"github.com/tjnvr/blog/internal/generator/page/html/validation/jssyntax"
	"github.com/tjnvr/blog/internal/generator/page/html/validation/navigation"
	"github.com/tjnvr/blog/internal/generator/page/html/validation/reference"
	"github.com/tjnvr/blog/internal/relpath"
)

// externalTimeout bounds each HTTP request made while validating external URLs.
const externalTimeout = 10 * time.Second

// Registry runs the configured validators over one page's HTML content.
type Registry struct {
	validators []Validator
}

// NewRegistry builds the default validators for the page at htmlPath: image,
// script and link reference validators plus the navigation validator.
//
// The build root (buildDir) and filesystem (fs) locate local targets;
// sectionResolver and pagePathsResolver drive the navigation check; skipURL
// disables external URL probing. All of these are fixed for the page, so the
// returned Registry only needs the page content when Validate is called.
func NewRegistry(
	htmlPath, buildDir string,
	fs afero.Fs,
	sectionResolver section.Resolver,
	pagePathsResolver relpath.Resolver,
	skipURL bool,
) *Registry {
	external := access.NewHTTPChecker(externalTimeout)
	if skipURL {
		external = access.NoopChecker{}
	}

	local := abspath.NewResolver(fs, buildDir, htmlPath)

	imageValidator := reference.NewValidator(
		htmlref.NewTagAttrExtractor("img", "src"),
		htmlref.NewClassifier(),
		external,
		local,
		nil,
		htmlPath, "image",
	)

	scriptValidator := reference.NewValidator(
		htmlref.NewTagAttrExtractor("script", "src"),
		htmlref.NewClassifier(),
		access.NoopChecker{},
		local,
		jssyntax.NewChecker(fs),
		htmlPath, "script",
	)

	linkValidator := reference.NewValidator(
		htmlref.NewTagAttrExtractor("a", "href"),
		htmlref.NewClassifier("#", "mailto:", "tel:", "javascript:"),
		external,
		local,
		nil,
		htmlPath, "link",
	)

	navigationValidator := navigation.NewValidator(sectionResolver, pagePathsResolver, htmlPath)

	return &Registry{
		validators: []Validator{
			imageValidator,
			scriptValidator,
			linkValidator,
			navigationValidator,
		},
	}
}

// NewRegistryWithValidators returns a Registry running exactly validators, used
// mainly to compose custom validator sets in tests.
func NewRegistryWithValidators(validators ...Validator) *Registry {
	return &Registry{validators: validators}
}

// Register adds v to the registry.
func (r *Registry) Register(v Validator) {
	r.validators = append(r.validators, v)
}

// Validate runs every registered validator over content and joins their errors
// into a single error, or returns nil when the content is valid.
func (r *Registry) Validate(content []byte) error {
	var errs []error
	for _, v := range r.validators {
		errs = append(errs, v.Validate(content)...)
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}
