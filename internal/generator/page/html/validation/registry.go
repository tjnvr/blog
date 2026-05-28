package validation

import (
	"errors"

	"github.com/tjnvr/blog/internal/generator/page/html/validation/image"
	"github.com/tjnvr/blog/internal/generator/page/html/validation/link"
	"github.com/tjnvr/blog/internal/generator/page/html/validation/navigation"
	"github.com/tjnvr/blog/internal/generator/page/html/validation/script"
	"github.com/tjnvr/blog/internal/generator/section"

	"github.com/spf13/afero"
)

// Registry manages validators and runs them on HTML content
type Registry struct {
	validators []Validator
}

// NewRegistry creates a validation registry with the navigation validator configured for the given sections
func NewRegistry(fs afero.Fs, sections []section.Section, skipURLValidation bool) *Registry {
	lv := link.NewValidator(fs, skipURLValidation)
	iv := image.NewValidator(fs, skipURLValidation)
	return &Registry{
		validators: []Validator{
			lv,
			iv,
			navigation.NewValidator(sections),
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
func NewDefaultRegistry(fs afero.Fs, build string, sections []section.Section, skipURLValidation bool) *Registry {
	return &Registry{
		validators: []Validator{
			image.NewValidator(fs, skipURLValidation),
			script.NewValidator(fs, build),
			link.NewValidator(fs, skipURLValidation),
			navigation.NewValidator(fs, sections),
		},
	}
}

// Register adds a validator to the registry
func (r *Registry) Register(v Validator) {
	r.validators = append(r.validators, v)
}

// Validate runs all registered validators on the given HTML content
func (r *Registry) Validate(htmlPath string) error {
	var errs []error
	for _, v := range r.validators {
		errs = append(errs, v.Validate(htmlPath)...)
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}
