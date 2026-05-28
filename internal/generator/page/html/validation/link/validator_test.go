package link

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewValidator(t *testing.T) {
	v := NewValidator(afero.NewMemMapFs(), false)
	require.NotNil(t, v)
	assert.Equal(t, 10*time.Second, v.Timeout)
	assert.False(t, v.SkipExternal)
}

func TestValidator_ValidateLocalLink(t *testing.T) {
	fs := afero.NewMemMapFs()
	buildDir := "/build"

	_ = fs.MkdirAll("/build/pages", 0755)
	_ = afero.WriteFile(fs, "/build/pages/about.html", []byte("<html></html>"), 0644)

	_ = fs.MkdirAll("/build/posts", 0755)
	_ = afero.WriteFile(fs, "/build/posts/index.html", []byte("<html></html>"), 0644)

	htmlPath := "/build/blog/test.html"

	tests := []struct {
		name      string
		html      string
		wantError bool
	}{
		{
			name:      "valid relative link",
			html:      `<a href="../pages/about.html">About</a>`,
			wantError: false,
		},
		{
			name:      "missing relative link",
			html:      `<a href="../pages/missing.html">Missing</a>`,
			wantError: true,
		},
		{
			name:      "valid absolute link",
			html:      `<a href="/pages/about.html">About</a>`,
			wantError: false,
		},
		{
			name:      "missing absolute link",
			html:      `<a href="/pages/missing.html">Missing</a>`,
			wantError: true,
		},
		{
			name:      "valid directory link with index",
			html:      `<a href="/posts">Posts</a>`,
			wantError: false,
		},
		{
			name:      "valid directory link with trailing slash",
			html:      `<a href="/posts/">Posts</a>`,
			wantError: false,
		},
		{
			name:      "fragment only link",
			html:      `<a href="#section">Section</a>`,
			wantError: false,
		},
		{
			name:      "link with fragment",
			html:      `<a href="/pages/about.html#section">About Section</a>`,
			wantError: false,
		},
		{
			name:      "mailto link",
			html:      `<a href="mailto:test@example.com">Email</a>`,
			wantError: false,
		},
		{
			name:      "tel link",
			html:      `<a href="tel:+1234567890">Call</a>`,
			wantError: false,
		},
		{
			name:      "no links",
			html:      `<p>No links here</p>`,
			wantError: false,
		},
		{
			name:      "multiple links mixed",
			html:      `<a href="/pages/about.html">Valid</a><a href="/missing.html">Missing</a>`,
			wantError: true,
		},
	}

	v := NewValidator(fs, false)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := v.Validate(htmlPath, buildDir, []byte(tt.html))
			if tt.wantError {
				assert.NotEmpty(t, errs)
			} else {
				assert.Empty(t, errs)
			}
		})
	}
}

func TestValidator_ValidateExternalLink(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/valid" {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	buildDir := "/build"
	htmlPath := filepath.Join(buildDir, "test.html")

	tests := []struct {
		name      string
		html      string
		wantError bool
	}{
		{
			name:      "valid external link",
			html:      `<a href="` + server.URL + `/valid">Valid</a>`,
			wantError: false,
		},
		{
			name:      "missing external link",
			html:      `<a href="` + server.URL + `/missing">Missing</a>`,
			wantError: true,
		},
	}

	v := NewValidator(afero.NewMemMapFs(), false)
	v.Timeout = 5 * time.Second

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := v.Validate(htmlPath, buildDir, []byte(tt.html))
			if tt.wantError {
				assert.NotEmpty(t, errs)
			} else {
				assert.Empty(t, errs)
			}
		})
	}
}

func TestValidator_SkipExternal(t *testing.T) {
	html := `<a href="http://invalid.invalid/page">Invalid</a>`
	v := NewValidator(afero.NewMemMapFs(),true)
	errs := v.Validate("/build/test.html", "/build", []byte(html))
	assert.Empty(t, errs)
}

func TestValidator_SkipsJavascriptLinks(t *testing.T) {
	html := `<a href="javascript:void(0)">Click</a>`
	v := NewValidator(afero.NewMemMapFs(), false)

	errs := v.Validate("/build/test.html", "/build", []byte(html))
	assert.Empty(t, errs)
}
