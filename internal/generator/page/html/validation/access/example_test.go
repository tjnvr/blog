package access_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/tjnvr/blog/internal/generator/page/html/validation/access"
)

func ExampleNewHTTPChecker() {
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	defer server.Close()

	checker := access.NewHTTPChecker(time.Second)

	fmt.Println(checker.Check(server.URL))
	// Output:
	// <nil>
}

func ExampleExternalChecker_Check() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	checker := access.NewHTTPChecker(time.Second)

	fmt.Println(checker.Check(server.URL))
	// Output:
	// HTTP 404
}

func ExampleNoopChecker() {
	checker := access.NoopChecker{}

	fmt.Println(checker.Check("http://unreachable.invalid"))
	// Output:
	// <nil>
}
