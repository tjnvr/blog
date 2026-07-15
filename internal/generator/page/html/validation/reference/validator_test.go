package reference

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/tjnvr/blog/internal/generator/page/html/validation/htmlref"
)

type mockExtractor struct {
	refs []string
}

func (m mockExtractor) Extract([]byte) []string { return m.refs }

type mockClassifier struct {
	kind htmlref.Kind
}

func (m mockClassifier) Classify(string) htmlref.Kind { return m.kind }

type mockExternal struct {
	err error
}

func (m mockExternal) Check(string) error { return m.err }

type mockLocal struct {
	ok       bool
	err      error
	resolved string
}

func (m mockLocal) Resolve(string) string { return m.resolved }

func (m mockLocal) Exists(string) (bool, error) { return m.ok, m.err }

type mockPost struct {
	errs []error
}

func (m mockPost) Check(string) []error { return m.errs }

func TestValidator_Validate_ShouldReportUnreachableExternalReference(t *testing.T) {
	// given
	ref := "https://" + uuid.New().String() + ".example.com"

	// setup
	validator := NewValidator(
		mockExtractor{refs: []string{ref}},
		mockClassifier{kind: htmlref.KindExternal},
		mockExternal{err: errors.New("HTTP 404")},
		mockLocal{},
		nil,
		"page.html", "link",
	)

	// test
	errs := validator.Validate(nil)

	// expect
	assert.Len(t, errs, 1)
	assert.Contains(t, errs[0].Error(), "external link not accessible")
}

func TestValidator_Validate_ShouldReportMissingLocalReference(t *testing.T) {
	// given
	ref := "/assets/" + uuid.New().String() + ".png"

	// setup
	validator := NewValidator(
		mockExtractor{refs: []string{ref}},
		mockClassifier{kind: htmlref.KindLocal},
		mockExternal{},
		mockLocal{ok: false},
		nil,
		"page.html", "image",
	)

	// test
	errs := validator.Validate(nil)

	// expect
	assert.Len(t, errs, 1)
	assert.Contains(t, errs[0].Error(), "local image not found")
}

func TestValidator_Validate_ShouldRunPostCheckOnExistingLocalReference(t *testing.T) {
	// given
	postErr := errors.New("JavaScript syntax error")

	// setup
	validator := NewValidator(
		mockExtractor{refs: []string{"/scripts/app.js"}},
		mockClassifier{kind: htmlref.KindLocal},
		mockExternal{},
		mockLocal{ok: true, resolved: "/build/scripts/app.js"},
		mockPost{errs: []error{postErr}},
		"page.html", "script",
	)

	// test
	errs := validator.Validate(nil)

	// expect
	assert.Equal(t, []error{postErr}, errs)
}

func TestValidator_Validate_ShouldSkipReferencesClassifiedAsSkip(t *testing.T) {
	// setup
	validator := NewValidator(
		mockExtractor{refs: []string{"#section", "mailto:a@b.com"}},
		mockClassifier{kind: htmlref.KindSkip},
		mockExternal{err: errors.New("should not be called")},
		mockLocal{ok: false},
		nil,
		"page.html", "link",
	)

	// test
	errs := validator.Validate(nil)

	// expect
	assert.Empty(t, errs)
}

func TestValidator_Validate_ShouldReturnNoErrorWhenReferencesAreValid(t *testing.T) {
	// setup
	validator := NewValidator(
		mockExtractor{refs: []string{"/assets/logo.png"}},
		mockClassifier{kind: htmlref.KindLocal},
		mockExternal{},
		mockLocal{ok: true, resolved: "/build/assets/logo.png"},
		nil,
		"page.html", "image",
	)

	// test
	errs := validator.Validate(nil)

	// expect
	assert.Empty(t, errs)
}

func TestValidator_Validate_ShouldReportFilesystemFailureOnLocalReference(t *testing.T) {
	// setup
	validator := NewValidator(
		mockExtractor{refs: []string{"/assets/logo.png"}},
		mockClassifier{kind: htmlref.KindLocal},
		mockExternal{},
		mockLocal{err: errors.New("permission denied")},
		nil,
		"page.html", "image",
	)

	// test
	errs := validator.Validate(nil)

	// expect
	assert.Len(t, errs, 1)
	assert.Contains(t, errs[0].Error(), "permission denied")
}
