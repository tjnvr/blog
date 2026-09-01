package main

import (
	"errors"
	"fmt"
)

// Every failure is collected instead of stopping at the first, so a single
// run reports the full picture.
func run(baseURL string) ([]error, error) {
	f := newFetcher()

	if _, err := f.get(baseURL); err != nil {
		return nil, fmt.Errorf("base URL not reachable: %w", err)
	}

	var errs, warnings []error

	if err := checkRobots(f, baseURL); err != nil {
		errs = append(errs, err)
	}

	pages, err := fetchSitemap(f, baseURL)
	if err != nil {
		errs = append(errs, err)
		return warnings, errors.Join(errs...)
	}

	for _, page := range pages {
		pageErrs, pageWarnings := checkPage(f, page)
		errs = append(errs, pageErrs...)
		warnings = append(warnings, pageWarnings...)
	}

	return warnings, errors.Join(errs...)
}
