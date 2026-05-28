package script

import (
	"fmt"
	"regexp"

	"github.com/tjnvr/blog/internal/generator/page/html/validation/shared"

	jsparser "github.com/dop251/goja/parser"
	"github.com/spf13/afero"
)

// Validator checks that all scripts in HTML are accessible
type Validator struct {
	fs                 afero.Fs
	scriptRegex        *regexp.Regexp
	HTMLPath, buildDir string
}

// NewValidator creates a new script validator
func NewValidator(fs afero.Fs, HTMLPath, buildDir string) *Validator {
	return &Validator{
		fs:          fs,
		scriptRegex: regexp.MustCompile(`<script[^>]+src="([^"]+)"`),
		HTMLPath:    HTMLPath,
		buildDir:    buildDir,
	}
}

// Validate checks all script src attributes in the HTML content
func (v *Validator) Validate(content []byte) []error {
	// Find all script src attributes
	matches := v.scriptRegex.FindAllSubmatch(content, -1)
	errs := make([]error, 0)
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}

		src := string(match[1])

		// Skip external scripts
		if shared.IsExternalURL(src) {
			continue
		}

		scriptPath := shared.ResolveLocalPath(src, v.HTMLPath, v.buildDir)
		if _, err := v.fs.Stat(scriptPath); err != nil {
			errs = append(errs, fmt.Errorf("%s: local script not found: %s", v.HTMLPath, src))
			continue
		}

		// Validate JavaScript syntax
		if syntaxErrs := v.validateJSSyntax(scriptPath); len(syntaxErrs) > 0 {
			errs = append(errs, syntaxErrs...)
		}
	}

	return errs
}

// validateJSSyntax parses the JavaScript file and returns any syntax errors
func (v *Validator) validateJSSyntax(scriptPath string) []error {
	content, err := afero.ReadFile(v.fs, scriptPath)
	if err != nil {
		return []error{fmt.Errorf("%s: failed to read script: %w", scriptPath, err)}
	}

	_, parseErr := jsparser.ParseFile(nil, scriptPath, string(content), 0)
	if parseErr != nil {
		return []error{fmt.Errorf("%s: JavaScript syntax error: %w", scriptPath, parseErr)}
	}

	return nil
}
