// Package jssyntax validates the syntax of local JavaScript files referenced by
// generated HTML.
//
// It provides the optional local post-check consumed by the reference
// validator: once a script reference has been resolved to an existing file,
// the Checker parses it and reports any syntax error.
package jssyntax

import (
	"fmt"

	jsparser "github.com/dop251/goja/parser"
	"github.com/spf13/afero"
)

// Checker parses JavaScript files read from an injected filesystem.
type Checker struct {
	fs afero.Fs
}

// NewChecker returns a Checker that reads script files from fs.
func NewChecker(fs afero.Fs) Checker {
	return Checker{fs: fs}
}

// Check reads the file at path and parses it as JavaScript. It returns a single
// error for a read or syntax failure, or nil when the file parses cleanly. The
// slice return matches the reference validator's local post-check contract.
func (c Checker) Check(path string) []error {
	content, err := afero.ReadFile(c.fs, path)
	if err != nil {
		return []error{fmt.Errorf("%s: failed to read script: %w", path, err)}
	}

	if _, err := jsparser.ParseFile(nil, path, string(content), 0); err != nil {
		return []error{fmt.Errorf("%s: JavaScript syntax error: %w", path, err)}
	}
	return nil
}
