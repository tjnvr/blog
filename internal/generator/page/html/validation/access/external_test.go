package access

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHTTPChecker_Check_ShouldReturnNilWhenStatusIsSuccessful(t *testing.T) {
	// setup
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	checker := NewHTTPChecker(time.Second)

	// test
	err := checker.Check(server.URL)

	// expect
	assert.NoError(t, err)
}

func TestHTTPChecker_Check_ShouldReturnErrorWhenStatusIsNotFound(t *testing.T) {
	// setup
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()
	checker := NewHTTPChecker(time.Second)

	// test
	err := checker.Check(server.URL)

	// expect
	assert.EqualError(t, err, "HTTP 404")
}

func TestHTTPChecker_Check_ShouldReturnErrorWhenHostIsUnreachable(t *testing.T) {
	// setup
	checker := NewHTTPChecker(time.Second)

	// test
	err := checker.Check("http://127.0.0.1:0")

	// expect
	assert.Error(t, err)
}

func TestNoopChecker_Check_ShouldAlwaysReturnNil(t *testing.T) {
	// setup
	checker := NoopChecker{}

	// test
	err := checker.Check("http://anything.invalid")

	// expect
	assert.NoError(t, err)
}
