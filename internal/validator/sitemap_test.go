package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckRobots(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		status     int
		wantErrSub string
	}{
		{
			name: "valid robots.txt passes",
			body: "User-agent: *\nSitemap: https://" + uuid.New().String() + "/sitemap.xml",
		},
		{
			name:       "empty body fails",
			body:       "",
			wantErrSub: "empty",
		},
		{
			name:       "unsubstituted placeholder fails",
			body:       "Sitemap: __BASE_URL__/sitemap.xml",
			wantErrSub: "placeholder",
		},
		{
			name:       "non-2xx/3xx fails",
			body:       uuid.New().String(),
			status:     http.StatusNotFound,
			wantErrSub: "HTTP 404",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			status := tt.status
			if status == 0 {
				status = http.StatusOK
			}

			// setup
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(status)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer server.Close()

			// test
			err := checkRobots(newFetcher(), server.URL)

			// expect
			if tt.wantErrSub == "" {
				assert.NoError(t, err)
				return
			}
			assert.ErrorContains(t, err, tt.wantErrSub)
		})
	}
}

func TestFetchSitemap_ShouldReturnEveryPageURL(t *testing.T) {
	// given
	pageA := "https://" + uuid.New().String() + ".example.com/"
	pageB := "https://" + uuid.New().String() + ".example.com/posts"

	// setup
	body := `<?xml version="1.0"?><urlset><url><loc>` + pageA + `</loc></url>` +
		`<url><loc>` + pageB + `</loc></url></urlset>`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(body))
	}))
	defer server.Close()

	// test
	pages, err := fetchSitemap(newFetcher(), server.URL)

	// expect
	require.NoError(t, err)
	assert.Equal(t, []string{pageA, pageB}, pages)
}

func TestFetchSitemap_ShouldFailOnInvalidXML(t *testing.T) {
	// setup
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(uuid.New().String()))
	}))
	defer server.Close()

	// test
	_, err := fetchSitemap(newFetcher(), server.URL)

	// expect
	assert.ErrorContains(t, err, "invalid XML")
}

func TestFetchSitemap_ShouldFailOnUnsubstitutedPlaceholder(t *testing.T) {
	// setup
	body := `<urlset><url><loc>__BASE_URL__/</loc></url></urlset>`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(body))
	}))
	defer server.Close()

	// test
	_, err := fetchSitemap(newFetcher(), server.URL)

	// expect
	assert.ErrorContains(t, err, "placeholder")
}

func TestFetchSitemap_ShouldFailWhenThereAreNoEntries(t *testing.T) {
	// setup
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`<urlset></urlset>`))
	}))
	defer server.Close()

	// test
	_, err := fetchSitemap(newFetcher(), server.URL)

	// expect
	assert.ErrorContains(t, err, "no <url> entries")
}
