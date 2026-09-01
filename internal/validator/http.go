package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const httpTimeout = 10 * time.Second

type fetchResult struct {
	body []byte
	err  error
}

type fetcher struct {
	client    *http.Client
	getCache  map[string]fetchResult
	headCache map[string]error
}

func newFetcher() *fetcher {
	return &fetcher{
		client:    &http.Client{Timeout: httpTimeout},
		getCache:  make(map[string]fetchResult),
		headCache: make(map[string]error),
	}
}

func resolve(pageURL, ref string) (string, error) {
	base, err := url.Parse(pageURL)
	if err != nil {
		return "", fmt.Errorf("parsing %q: %w", pageURL, err)
	}
	target, err := base.Parse(ref)
	if err != nil {
		return "", fmt.Errorf("resolving %q against %q: %w", ref, pageURL, err)
	}
	return target.String(), nil
}

// isExternal reports whether target points to a different host than pageURL.
// Third-party sites often block automated HTTP clients outright (bot
// detection, rate limiting) regardless of whether the link is actually
// broken, so failures against external hosts are treated as warnings rather
// than hard failures — see checkRefsReachable and checkScripts.
func isExternal(pageURL, target string) bool {
	p, err := url.Parse(pageURL)
	if err != nil {
		return false
	}
	t, err := url.Parse(target)
	if err != nil {
		return false
	}
	return t.Host != p.Host
}

// get fetches target via HTTP GET, treating any 2xx/3xx response as
// reachable and returning its body.
func (f *fetcher) get(target string) ([]byte, error) {
	if cached, ok := f.getCache[target]; ok {
		return cached.body, cached.err
	}
	body, err := f.doGet(target)
	f.getCache[target] = fetchResult{body: body, err: err}
	return body, err
}

func (f *fetcher) doGet(target string) ([]byte, error) {
	resp, err := f.client.Get(target)
	if err != nil {
		return nil, fmt.Errorf("GET %s: %w", target, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return nil, fmt.Errorf("GET %s: HTTP %d", target, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("GET %s: reading body: %w", target, err)
	}
	return body, nil
}

// head checks that target is reachable via HTTP HEAD, treating any 2xx/3xx
// response as reachable.
func (f *fetcher) head(target string) error {
	if cached, ok := f.headCache[target]; ok {
		return cached
	}
	err := f.doHead(target)
	f.headCache[target] = err
	return err
}

func (f *fetcher) doHead(target string) error {
	resp, err := f.client.Head(target)
	if err != nil {
		return fmt.Errorf("HEAD %s: %w", target, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return fmt.Errorf("HEAD %s: HTTP %d", target, resp.StatusCode)
	}
	return nil
}
