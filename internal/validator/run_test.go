package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRun_ShouldSucceedWhenSiteIsValid(t *testing.T) {
	// setup
	var baseURL string
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()
	baseURL = server.URL

	page := `<html><body><nav class="x"><a href="/">Home</a></nav>` +
		`<img src="/logo.png"><a href="/posts">Posts</a>` +
		`<script src="/scripts/dark-mode.js"></script></body></html>`

	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(page))
	})
	mux.HandleFunc("/posts", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(page))
	})
	mux.HandleFunc("/logo.png", func(_ http.ResponseWriter, _ *http.Request) {})
	mux.HandleFunc("/scripts/dark-mode.js", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("console.log('ok');"))
	})
	mux.HandleFunc("/robots.txt", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("User-agent: *\nSitemap: " + baseURL + "/sitemap.xml"))
	})
	mux.HandleFunc("/sitemap.xml", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`<urlset><url><loc>` + baseURL + `/</loc></url>` +
			`<url><loc>` + baseURL + `/posts</loc></url></urlset>`))
	})

	// test
	warnings, err := run(baseURL)

	// expect
	assert.NoError(t, err)
	assert.Empty(t, warnings)
}

func TestRun_ShouldNotFailOnExternalReferenceFailures(t *testing.T) {
	// given
	externalServer := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {}))
	externalURL := externalServer.URL + "/" + uuid.New().String()
	externalServer.Close() // guarantees a refused connection, on a host distinct from the site under test

	// setup
	var baseURL string
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()
	baseURL = server.URL

	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprintf(
			w,
			`<html><body><nav class="x"><a href="/">Home</a></nav><a href="%s">ext</a></body></html>`,
			externalURL,
		)
	})
	mux.HandleFunc("/robots.txt", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("User-agent: *"))
	})
	mux.HandleFunc("/sitemap.xml", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`<urlset><url><loc>` + baseURL + `/</loc></url></urlset>`))
	})

	// test
	warnings, err := run(baseURL)

	// expect
	assert.NoError(t, err)
	if assert.Len(t, warnings, 1) {
		assert.ErrorContains(t, warnings[0], externalURL)
	}
}

func TestRun_ShouldReportAllFailuresAtOnce(t *testing.T) {
	// setup
	var baseURL string
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()
	baseURL = server.URL

	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`<html><body><img src="/missing.png"><script src="/broken.js"></script></body></html>`))
	})
	mux.HandleFunc("/broken.js", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("function( {"))
	})
	mux.HandleFunc("/missing.png", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	mux.HandleFunc("/robots.txt", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("User-agent: *"))
	})
	mux.HandleFunc("/sitemap.xml", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`<urlset><url><loc>` + baseURL + `/</loc></url></urlset>`))
	})

	// test
	_, err := run(baseURL)

	// expect
	require.Error(t, err)
	assert.ErrorContains(t, err, "missing <nav>")
	assert.ErrorContains(t, err, "missing.png")
	assert.ErrorContains(t, err, "syntax error")
}

func TestRun_ShouldFailFastWhenBaseURLUnreachable(t *testing.T) {
	// setup
	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {}))
	badURL := server.URL
	server.Close()

	// test
	_, err := run(badURL)

	// expect
	assert.ErrorContains(t, err, "base URL not reachable")
}
