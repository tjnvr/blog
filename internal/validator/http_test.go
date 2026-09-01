package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolve(t *testing.T) {
	tests := []struct {
		name    string
		pageURL string
		ref     string
		want    string
		wantErr bool
	}{
		{
			name:    "root-relative reference resolves against the host",
			pageURL: "https://example.com/posts/foo",
			ref:     "/scripts/dark-mode.js",
			want:    "https://example.com/scripts/dark-mode.js",
		},
		{
			name:    "page-relative reference resolves against the page directory",
			pageURL: "https://example.com/posts/",
			ref:     "bar",
			want:    "https://example.com/posts/bar",
		},
		{
			name:    "absolute reference is left untouched",
			pageURL: "https://example.com/",
			ref:     "https://other.com/x.png",
			want:    "https://other.com/x.png",
		},
		{
			name:    "invalid page URL fails",
			pageURL: "://bad",
			ref:     "x",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// test
			got, err := resolve(tt.pageURL, tt.ref)

			// expect
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFetcher_Get_ShouldCacheRequests(t *testing.T) {
	// given
	wantBody := uuid.New().String()

	// setup
	var requestCount int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requestCount++
		_, _ = w.Write([]byte(wantBody))
	}))
	defer server.Close()
	f := newFetcher()

	// test
	body1, err1 := f.get(server.URL)
	body2, err2 := f.get(server.URL)

	// expect
	require.NoError(t, err1)
	require.NoError(t, err2)
	assert.Equal(t, wantBody, string(body1))
	assert.Equal(t, wantBody, string(body2))
	assert.Equal(t, 1, requestCount)
}

func TestFetcher_Get_ShouldReturnErrorOnNon2xxOr3xxStatus(t *testing.T) {
	// setup
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	// test
	_, err := newFetcher().get(server.URL)

	// expect
	assert.ErrorContains(t, err, "HTTP 404")
}

func TestFetcher_Head_ShouldCacheRequests(t *testing.T) {
	// setup
	var requestCount int
	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		requestCount++
	}))
	defer server.Close()
	f := newFetcher()

	// test
	err1 := f.head(server.URL)
	err2 := f.head(server.URL)

	// expect
	require.NoError(t, err1)
	require.NoError(t, err2)
	assert.Equal(t, 1, requestCount)
}

func TestFetcher_Head_ShouldReturnErrorOnNon2xxOr3xxStatus(t *testing.T) {
	// setup
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	// test
	err := newFetcher().head(server.URL)

	// expect
	assert.ErrorContains(t, err, "HTTP 500")
}
