package main

import "fmt"

func checkPage(f *fetcher, pageURL string) (errs, warnings []error) {
	body, err := f.get(pageURL)
	if err != nil {
		return []error{fmt.Errorf("%s: %w", pageURL, err)}, nil
	}

	imgErrs, imgWarnings := checkImages(f, pageURL, body)
	scriptErrs, scriptWarnings := checkScripts(f, pageURL, body)
	linkErrs, linkWarnings := checkLinks(f, pageURL, body)

	errs = append(errs, imgErrs...)
	errs = append(errs, scriptErrs...)
	errs = append(errs, linkErrs...)
	if err := checkNavbar(pageURL, body); err != nil {
		errs = append(errs, err)
	}

	warnings = append(warnings, imgWarnings...)
	warnings = append(warnings, scriptWarnings...)
	warnings = append(warnings, linkWarnings...)

	return errs, warnings
}
