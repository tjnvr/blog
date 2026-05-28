package validation

import (
	"strings"
	"testing"

	"github.com/tjnvr/blog/internal/generator/section"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeValidator is a test double implementing Validator
type fakeValidator struct {
	validateFunc func(htmlPath, buildDir string, content []byte) []error
}

func (f fakeValidator) Validate(htmlPath, buildDir string, content []byte) []error {
	return f.validateFunc(htmlPath, buildDir, content)
}

func TestNewRegistry(t *testing.T) {
	sections := []section.Section{
		{DirName: "", DisplayName: "Accueil"},
		{DirName: "posts", DisplayName: "Posts"},
		{DirName: "about", DisplayName: "About"},
	}
	r := NewRegistry(afero.NewMemMapFs(), sections, false)
	require.NotNil(t, r)
	assert.Len(t, r.validators, 3)
}

func TestNewRegistryWithValidators(t *testing.T) {
	r := NewRegistryWithValidators()
	require.NotNil(t, r)
	assert.Len(t, r.validators, 0)
}

func TestNewDefaultRegistry(t *testing.T) {
	r := NewDefaultRegistry(afero.NewMemMapFs(), []section.Section{{DirName: "", DisplayName: "Accueil"}, {DirName: "posts", DisplayName: "Posts"}}, false)
	require.NotNil(t, r)
	assert.Len(t, r.validators, 4)
}

func TestRegistry_Register(t *testing.T) {
	r := NewRegistryWithValidators()
	v := fakeValidator{validateFunc: func(string, string, []byte) []error { return nil }}

	r.Register(v)
	assert.Len(t, r.validators, 1)

	r.Register(v)
	assert.Len(t, r.validators, 2)
}

func TestRegistry_Validate(t *testing.T) {
	tests := []struct {
		name       string
		validators []Validator
		wantErr    bool
		errCount   int
	}{
		{
			name:       "no validators returns nil",
			validators: []Validator{},
			wantErr:    false,
		},
		{
			name: "all validators pass returns nil",
			validators: []Validator{
				fakeValidator{validateFunc: func(string, string, []byte) []error { return nil }},
				fakeValidator{validateFunc: func(string, string, []byte) []error { return nil }},
			},
			wantErr: false,
		},
		{
			name: "validator returning empty slice returns nil",
			validators: []Validator{
				fakeValidator{validateFunc: func(string, string, []byte) []error { return []error{} }},
			},
			wantErr: false,
		},
		{
			name: "single validator error",
			validators: []Validator{
				fakeValidator{validateFunc: func(string, string, []byte) []error {
					return []error{NewError("file.html", "broken link")}
				}},
			},
			wantErr:  true,
			errCount: 1,
		},
		{
			name: "multiple validators with errors",
			validators: []Validator{
				fakeValidator{validateFunc: func(string, string, []byte) []error {
					return []error{NewError("file.html", "error 1")}
				}},
				fakeValidator{validateFunc: func(string, string, []byte) []error {
					return []error{
						NewError("file.html", "error 2"),
						NewError("file.html", "error 3"),
					}
				}},
			},
			wantErr:  true,
			errCount: 3,
		},
		{
			name: "mix of passing and failing validators",
			validators: []Validator{
				fakeValidator{validateFunc: func(string, string, []byte) []error { return nil }},
				fakeValidator{validateFunc: func(string, string, []byte) []error {
					return []error{NewError("file.html", "one error")}
				}},
			},
			wantErr:  true,
			errCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRegistryWithValidators(tt.validators...)
			err := r.Validate("test.html", "/build", []byte("<html></html>"))
			if tt.wantErr {
				require.Error(t, err)
				errParts := strings.Split(err.Error(), "\n")
				assert.Len(t, errParts, tt.errCount)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNewError(t *testing.T) {
	err := NewError("page.html", "missing image %s", "logo.png")
	assert.Equal(t, "page.html", err.File)
	assert.Equal(t, "missing image logo.png", err.Message)
	assert.Equal(t, "page.html: missing image logo.png", err.Error())
}
