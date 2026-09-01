package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestExtractAttr(t *testing.T) {
	tests := []struct {
		name string
		tag  string
		attr string
	}{
		{name: "img src", tag: "img", attr: "src"},
		{name: "a href", tag: "a", attr: "href"},
		{name: "script src", tag: "script", attr: "src"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			ref := "/" + uuid.New().String()

			// setup
			body := fmt.Appendf(nil, `<%s %s="%s">`, tt.tag, tt.attr, ref)

			// test
			refs := extractAttr(body, tt.tag, tt.attr)

			// expect
			assert.Equal(t, []string{ref}, refs)
		})
	}
}

func TestSkippable(t *testing.T) {
	tests := []struct {
		name string
		ref  string
		want bool
	}{
		{name: "fragment is skippable", ref: "#" + uuid.New().String(), want: true},
		{name: "mailto is skippable", ref: "mailto:" + uuid.New().String() + "@example.com", want: true},
		{name: "tel is skippable", ref: "tel:+" + uuid.New().String(), want: true},
		{name: "javascript is skippable", ref: "javascript:" + uuid.New().String(), want: true},
		{name: "data URI is skippable", ref: "data:" + uuid.New().String(), want: true},
		{name: "root-relative path is not skippable", ref: "/" + uuid.New().String(), want: false},
		{name: "external URL is not skippable", ref: "https://" + uuid.New().String() + ".example.com", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// test
			got := skippable(tt.ref)

			// expect
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCheckRefsReachable_ShouldReportOnlyUnreachableNonSkippedReferences(t *testing.T) {
	// given
	okRef := "/" + uuid.New().String()
	missingRef := "/" + uuid.New().String()

	// setup
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == missingRef {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()
	body := []byte(fmt.Sprintf(
		`<img src="%s"><img src="%s"><img src="#skip"><img src="data:image/png;base64,AA">`,
		okRef, missingRef,
	))

	// test
	errs, warnings := checkRefsReachable(newFetcher(), server.URL+"/page", body, "img", "src")

	// expect
	assert.Empty(t, warnings)
	if assert.Len(t, errs, 1) {
		assert.ErrorContains(t, errs[0], missingRef)
	}
}

func TestCheckRefsReachable_ShouldTreatExternalFailuresAsWarnings(t *testing.T) {
	// given
	externalServer := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {}))
	externalURL := externalServer.URL + "/" + uuid.New().String()
	externalServer.Close() // guarantees a refused connection, on a host distinct from pageServer

	// setup
	pageServer := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {}))
	defer pageServer.Close()
	body := []byte(fmt.Sprintf(`<a href="%s">external</a>`, externalURL))

	// test
	errs, warnings := checkRefsReachable(newFetcher(), pageServer.URL+"/page", body, "a", "href")

	// expect
	assert.Empty(t, errs)
	if assert.Len(t, warnings, 1) {
		assert.ErrorContains(t, warnings[0], externalURL)
	}
}

func TestCheckScripts(t *testing.T) {
	tests := []struct {
		name       string
		respBody   string
		respStatus int
		wantErrSub string
	}{
		{
			name:       "valid script passes",
			respBody:   "console.log('hi');",
			respStatus: http.StatusOK,
		},
		{
			name:       "unreachable script fails",
			respStatus: http.StatusNotFound,
			wantErrSub: "not accessible",
		},
		{
			name:       "invalid syntax fails",
			respBody:   "function( {",
			respStatus: http.StatusOK,
			wantErrSub: "syntax error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			path := "/" + uuid.New().String() + ".js"

			// setup
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tt.respStatus)
				_, _ = w.Write([]byte(tt.respBody))
			}))
			defer server.Close()
			body := []byte(fmt.Sprintf(`<script src="%s"></script>`, path))

			// test
			errs, warnings := checkScripts(newFetcher(), server.URL+"/page", body)

			// expect
			assert.Empty(t, warnings)
			if tt.wantErrSub == "" {
				assert.Empty(t, errs)
				return
			}
			if assert.Len(t, errs, 1) {
				assert.ErrorContains(t, errs[0], tt.wantErrSub)
			}
		})
	}
}

func TestCheckScripts_ShouldTreatExternalFailuresAsWarnings(t *testing.T) {
	// given
	externalServer := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {}))
	externalURL := externalServer.URL + "/" + uuid.New().String() + ".js"
	externalServer.Close() // guarantees a refused connection, on a host distinct from pageServer

	// setup
	pageServer := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {}))
	defer pageServer.Close()
	body := []byte(fmt.Sprintf(`<script src="%s"></script>`, externalURL))

	// test
	errs, warnings := checkScripts(newFetcher(), pageServer.URL+"/page", body)

	// expect
	assert.Empty(t, errs)
	if assert.Len(t, warnings, 1) {
		assert.ErrorContains(t, warnings[0], externalURL)
	}
}
