package image

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

func TestValidator_ValidateLocalImage(t *testing.T) {
	fs := afero.NewMemMapFs()
	buildDir := "/build"

	_ = fs.MkdirAll("/build/assets/images", 0755)
	_ = afero.WriteFile(fs, "/build/assets/images/test.png", []byte("fake image"), 0644)

	htmlPath := "/build/post/test.html"

	tests := []struct {
		name      string
		html      string
		wantError bool
	}{
		{
			name:      "valid relative image",
			html:      `<img src="../assets/images/test.png" alt="Test">`,
			wantError: false,
		},
		{
			name:      "missing relative image",
			html:      `<img src="../assets/images/missing.png" alt="Missing">`,
			wantError: true,
		},
		{
			name:      "valid absolute image",
			html:      `<img src="/assets/images/test.png" alt="Test">`,
			wantError: false,
		},
		{
			name:      "missing absolute image",
			html:      `<img src="/assets/images/missing.png" alt="Missing">`,
			wantError: true,
		},
		{
			name:      "no images",
			html:      `<p>No images here</p>`,
			wantError: false,
		},
		{
			name:      "multiple images mixed",
			html:      `<img src="../assets/images/test.png"><img src="../assets/images/missing.png">`,
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

func TestValidator_ValidateExternalImage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/valid.png" {
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
			name:      "valid external image",
			html:      `<img src="` + server.URL + `/valid.png" alt="Valid">`,
			wantError: false,
		},
		{
			name:      "missing external image",
			html:      `<img src="` + server.URL + `/missing.png" alt="Missing">`,
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
	html := `<img src="http://invalid.invalid/image.png" alt="Invalid">`
	v := NewValidator(afero.NewMemMapFs(), true)

	errs := v.Validate("/build/test.html", "/build", []byte(html))
	assert.Empty(t, errs)
}
