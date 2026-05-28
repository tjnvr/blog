package validation

import (
	"errors"
	"fmt"

	"github.com/tjnvr/blog/internal/generator/page/html/validation/image"
	"github.com/tjnvr/blog/internal/generator/page/html/validation/link"
	"github.com/tjnvr/blog/internal/generator/page/html/validation/navigation"
	"github.com/tjnvr/blog/internal/generator/page/html/validation/script"
	"github.com/tjnvr/blog/internal/generator/section"

	"github.com/spf13/afero"
)

// Registry manages validators and runs them on HTML content
type Registry struct {
	fs         afero.Fs
	validators []Validator
}

// NewRegistry creates a validation registry with the navigation validator configured for the given sections
func NewRegistry(fs afero.Fs, HTMLPath, buildDir string, sections []section.Section, skipURLValidation bool) *Registry {
	return &Registry{
		fs: fs,
		validators: []Validator{
			image.NewValidator(fs, HTMLPath, buildDir, skipURLValidation),
			script.NewValidator(fs, HTMLPath, buildDir),
			link.NewValidator(fs, HTMLPath, buildDir, skipURLValidation),
			navigation.NewValidator(fs, HTMLPath, sections),
		},
	}
}

// NewRegistryWithValidators creates a registry with custom validators
func NewRegistryWithValidators(fs afero.Fs, validators ...Validator) *Registry {
	return &Registry{
		fs:         fs,
		validators: validators,
	}
}

// NewDefaultRegistry creates a validation registry with default validators (image, script, link, navigation)
func NewDefaultRegistry(fs afero.Fs, HTMLPath, buildDir string, sections []section.Section, skipURLValidation bool) *Registry {
	return &Registry{
		fs: fs,
		validators: []Validator{
			image.NewValidator(fs, HTMLPath, buildDir, skipURLValidation),
			script.NewValidator(fs, HTMLPath, buildDir),
			link.NewValidator(fs, HTMLPath, buildDir, skipURLValidation),
			navigation.NewValidator(fs, HTMLPath, sections),
		},
	}
}

// Register adds a validator to the registry
func (r *Registry) Register(v Validator) {
	r.validators = append(r.validators, v)
}

// Validate runs all registered validators on the given HTML content
func (r *Registry) Validate(HTMLFilePath string) error {
	if _, err := r.fs.Stat(HTMLFilePath); err != nil {
		return fmt.Errorf("Cannot get file info for file: %v", HTMLFilePath)
	}
	content, err := afero.ReadFile(r.fs, HTMLFilePath)
	if err != nil {
		return fmt.Errorf("Cannot read file for file: %v", HTMLFilePath)
	}
	errs := make([]error, 0)
	for _, v := range r.validators {
		errs = append(errs, v.Validate(content)...)
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}
