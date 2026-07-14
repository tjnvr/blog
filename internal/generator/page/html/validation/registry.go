package validation

import (
	"errors"

	"github.com/tjnvr/blog/internal/generator/backbone/section"
	"github.com/tjnvr/blog/internal/generator/page/html/validation/image"
	"github.com/tjnvr/blog/internal/generator/page/html/validation/link"
	"github.com/tjnvr/blog/internal/generator/page/html/validation/navigation"
	"github.com/tjnvr/blog/internal/generator/page/html/validation/script"
	"github.com/tjnvr/blog/internal/relpath"
)

// Registry manages validators and runs them on HTML content
type Registry struct {
	validators []Validator
}

// NewRegistry creates a validation registry with the navigation validator configured for the given sections
func NewRegistry(sectionResolver section.Resolver, pathResolver relpath.Resolver, skipURLValidation bool) *Registry {
	lv := link.NewValidator()
	lv.SkipExternal = skipURLValidation
	iv := image.NewValidator()
	iv.SkipExternal = skipURLValidation
	return &Registry{
		validators: []Validator{
			lv,
			iv,
			navigation.NewValidator(sectionResolver, pathResolver),
		},
	}
}

// NewRegistryWithValidators creates a registry with custom validators
func NewRegistryWithValidators(validators ...Validator) *Registry {
	return &Registry{
		validators: validators,
	}
}

// NewDefaultRegistry creates a validation registry with default validators (image, script, link, navigation)
func NewDefaultRegistry(sectionResolver section.Resolver, pathResolver relpath.Resolver, skipURLValidation bool) *Registry {
	lv := link.NewValidator()
	lv.SkipExternal = skipURLValidation
	iv := image.NewValidator()
	iv.SkipExternal = skipURLValidation
	return &Registry{
		validators: []Validator{
			iv,
			script.NewValidator(),
			lv,
			navigation.NewValidator(sectionResolver, pathResolver),
		},
	}
}

// Register adds a validator to the registry
func (r *Registry) Register(v Validator) {
	r.validators = append(r.validators, v)
}

// Validate runs all registered validators on the given HTML content
func (r *Registry) Validate(htmlPath, buildDir string, content []byte) error {
	var errs []error
	for _, v := range r.validators {
		errs = append(errs, v.Validate(htmlPath, buildDir, content)...)
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}
