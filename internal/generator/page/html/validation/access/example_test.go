package access_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/spf13/afero"

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

func ExampleNewResolver() {
	resolver := access.NewResolver(afero.NewMemMapFs(), "/build", "/build/posts/article.html")

	fmt.Println(resolver.Resolve("../assets/logo.png"))
	// Output:
	// /build/assets/logo.png
}

func ExampleLocalResolver_Exists() {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "/build/assets/logo.png", []byte("x"), 0644)

	resolver := access.NewResolver(fs, "/build", "/build/posts/article.html")

	fmt.Println(resolver.Exists("/assets/logo.png"))
	// Output:
	// true <nil>
}
