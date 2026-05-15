package validation

import (
	"strings"
	"testing"

	"github.com/tjnvr/blog/internal/generator/section"
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
	r := NewRegistry(sections, false)
	if r == nil {
		t.Fatal("NewRegistry() returned nil")
		return
	}
	if len(r.validators) != 3 {
		t.Errorf("NewRegistry() should have 3 validator (link, image, navigation), got %d", len(r.validators))
	}
}

func TestNewRegistryWithValidators(t *testing.T) {
	r := NewRegistryWithValidators()
	if r == nil {
		t.Fatal("NewRegistryWithValidators() returned nil")
		return
	}
	if len(r.validators) != 0 {
		t.Errorf("NewRegistryWithValidators() should have 0 validators, got %d", len(r.validators))
	}
}

func TestNewDefaultRegistry(t *testing.T) {
	r := NewDefaultRegistry([]section.Section{{DirName: "", DisplayName: "Accueil"}, {DirName: "posts", DisplayName: "Posts"}}, false)
	if r == nil {
		t.Fatal("NewDefaultRegistry() returned nil")
		return
	}
	if len(r.validators) != 4 {
		t.Errorf("NewDefaultRegistry() should have 4 validators (image, script, link, navigation), got %d", len(r.validators))
	}
}

func TestRegistry_Register(t *testing.T) {
	r := NewRegistryWithValidators()
	v := fakeValidator{validateFunc: func(string, string, []byte) []error { return nil }}

	r.Register(v)

	if len(r.validators) != 1 {
		t.Errorf("expected 1 validator after Register, got %d", len(r.validators))
	}

	r.Register(v)
	if len(r.validators) != 2 {
		t.Errorf("expected 2 validators after second Register, got %d", len(r.validators))
	}
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
				if err == nil {
					t.Error("Validate() expected error, got nil")
					return
				}
				// Count individual errors in the joined error
				errParts := strings.Split(err.Error(), "\n")
				if len(errParts) != tt.errCount {
					t.Errorf("Validate() expected %d errors, got %d: %v", tt.errCount, len(errParts), err)
				}
			} else {
				if err != nil {
					t.Errorf("Validate() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestNewError(t *testing.T) {
	err := NewError("page.html", "missing image %s", "logo.png")
	if err.File != "page.html" {
		t.Errorf("File = %q, want %q", err.File, "page.html")
	}
	if err.Message != "missing image logo.png" {
		t.Errorf("Message = %q, want %q", err.Message, "missing image logo.png")
	}
	expected := "page.html: missing image logo.png"
	if err.Error() != expected {
		t.Errorf("Error() = %q, want %q", err.Error(), expected)
	}
}
