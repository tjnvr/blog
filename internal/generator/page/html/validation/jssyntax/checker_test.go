package jssyntax

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestChecker_Check_ShouldReturnNilForValidJavaScript(t *testing.T) {
	// setup
	fs := afero.NewMemMapFs()
	assert.NoError(t, afero.WriteFile(fs, "/build/scripts/app.js", []byte("const answer = 42;"), 0644))
	checker := NewChecker(fs)

	// test
	errs := checker.Check("/build/scripts/app.js")

	// expect
	assert.Empty(t, errs)
}

func TestChecker_Check_ShouldReturnSyntaxErrorForInvalidJavaScript(t *testing.T) {
	// setup
	fs := afero.NewMemMapFs()
	assert.NoError(t, afero.WriteFile(fs, "/build/scripts/broken.js", []byte("const = ;"), 0644))
	checker := NewChecker(fs)

	// test
	errs := checker.Check("/build/scripts/broken.js")

	// expect
	assert.Len(t, errs, 1)
	assert.Contains(t, errs[0].Error(), "JavaScript syntax error")
}

func TestChecker_Check_ShouldReturnReadErrorWhenFileIsMissing(t *testing.T) {
	// setup
	checker := NewChecker(afero.NewMemMapFs())

	// test
	errs := checker.Check("/build/scripts/missing.js")

	// expect
	assert.Len(t, errs, 1)
	assert.Contains(t, errs[0].Error(), "failed to read script")
}
