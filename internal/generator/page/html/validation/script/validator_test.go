package script

import (
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestValidator_ValidateLocalScript(t *testing.T) {
	fs := afero.NewMemMapFs()
	buildDir := "/build"

	_ = fs.MkdirAll("/build/scripts", 0755)
	validJS := `(function() { console.log("hello"); })();`
	_ = afero.WriteFile(fs, "/build/scripts/valid.js", []byte(validJS), 0644)

	htmlPath := "/build/test.html"

	tests := []struct {
		name      string
		html      string
		wantError bool
	}{
		{
			name:      "valid script exists",
			html:      `<script src="/scripts/valid.js"></script>`,
			wantError: false,
		},
		{
			name:      "missing script",
			html:      `<script src="/scripts/missing.js"></script>`,
			wantError: true,
		},
		{
			name:      "no scripts",
			html:      `<p>No scripts here</p>`,
			wantError: false,
		},
		{
			name:      "external script ignored",
			html:      `<script src="https://cdn.example.com/lib.js"></script>`,
			wantError: false,
		},
	}

	v := NewValidator(fs, htmlPath, buildDir)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := v.Validate([]byte(tt.html))
			if tt.wantError {
				assert.NotEmpty(t, errs)
			} else {
				assert.Empty(t, errs)
			}
		})
	}
}

func TestValidator_ValidateJSSyntax(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = fs.MkdirAll("/build/scripts", 0755)

	tests := []struct {
		name      string
		jsContent string
		wantError bool
	}{
		{
			name:      "valid IIFE",
			jsContent: `(function() { var x = 1; })();`,
			wantError: false,
		},
		{
			name:      "valid arrow function",
			jsContent: `const fn = () => { return 42; };`,
			wantError: false,
		},
		{
			name:      "valid ES6 class",
			jsContent: `class Foo { constructor() { this.x = 1; } }`,
			wantError: false,
		},
		{
			name:      "syntax error - missing bracket",
			jsContent: `function foo( { return 1; }`,
			wantError: true,
		},
		{
			name:      "syntax error - invalid token",
			jsContent: `var x = @invalid;`,
			wantError: true,
		},
		{
			name:      "syntax error - unclosed string",
			jsContent: `var x = "unclosed;`,
			wantError: true,
		},
	}

	v := NewValidator(fs, "/build/test.html", "/build/")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = afero.WriteFile(fs, "/build/scripts/test.js", []byte(tt.jsContent), 0644)

			html := `<script src="/scripts/test.js"></script>`
			errs := v.Validate([]byte(html))
			if tt.wantError {
				assert.NotEmpty(t, errs)
			} else {
				assert.Empty(t, errs)
			}
		})
	}
}

func TestValidator_RelativePath(t *testing.T) {
	fs := afero.NewMemMapFs()
	buildDir := "/build"

	_ = fs.MkdirAll("/build/post", 0755)
	_ = fs.MkdirAll("/build/scripts", 0755)

	validJS := `console.log("test");`
	_ = afero.WriteFile(fs, "/build/scripts/app.js", []byte(validJS), 0644)

	htmlPath := filepath.Join("/build/post", "article.html")

	tests := []struct {
		name      string
		html      string
		wantError bool
	}{
		{
			name:      "relative path from nested dir",
			html:      `<script src="../scripts/app.js"></script>`,
			wantError: false,
		},
		{
			name:      "absolute path from nested dir",
			html:      `<script src="/scripts/app.js"></script>`,
			wantError: false,
		},
	}

	v := NewValidator(fs, htmlPath, buildDir)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := v.Validate([]byte(tt.html))
			if tt.wantError {
				assert.NotEmpty(t, errs)
			} else {
				assert.Empty(t, errs)
			}
		})
	}
}
