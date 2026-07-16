// Package access checks whether external references found in generated HTML
// are reachable.
//
// An ExternalChecker probes external URLs over HTTP. It is configured once per
// page and holds its fixed state (timeout) so callers pass only the URL being
// checked.
package access

import (
	"fmt"
	"net/http"
	"time"
)

// ExternalChecker verifies that an external URL is reachable.
type ExternalChecker interface {
	// Check returns nil when url answers with a 2xx or 3xx status, and a
	// descriptive error otherwise (a transport failure or a 4xx/5xx status).
	Check(url string) error
}

type httpChecker struct {
	client *http.Client
}

// NewHTTPChecker returns an ExternalChecker that issues an HTTP HEAD request and
// treats any 2xx or 3xx status as reachable. timeout bounds each request.
func NewHTTPChecker(timeout time.Duration) ExternalChecker {
	return httpChecker{client: &http.Client{Timeout: timeout}}
}

func (c httpChecker) Check(url string) error {
	resp, err := c.client.Head(url)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return nil
}

// NoopChecker is an ExternalChecker that never reports an error. It disables
// external checking, either globally (when URL validation is skipped) or for
// reference kinds that must not be probed, such as scripts.
type NoopChecker struct{}

// Check always returns nil.
func (NoopChecker) Check(string) error { return nil }
