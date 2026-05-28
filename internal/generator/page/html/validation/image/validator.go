package image

import (
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/tjnvr/blog/internal/generator/page/html/validation/shared"

	"github.com/spf13/afero"
)

// Validator checks that all images in HTML are accessible
type Validator struct {
	fs afero.Fs
	// Timeout for HTTP requests to external images
	Timeout time.Duration
	// SkipExternal skips validation of external URLs
	SkipExternal       bool
	imgRegex           *regexp.Regexp
	HTMLPath, buildDir string
}

// NewValidator creates a new image validator with default settings
func NewValidator(fs afero.Fs, HTMLPath, buildDir string, skipExternalValidation bool) *Validator {
	return &Validator{
		fs:           fs,
		Timeout:      10 * time.Second,
		SkipExternal: skipExternalValidation,
		imgRegex:     regexp.MustCompile(`<img[^>]+src="([^"]+)"`),
		HTMLPath:     HTMLPath,
		buildDir:     buildDir,
	}
}

// Validate checks all img src attributes in the HTML content
func (v *Validator) Validate(content []byte) []error {
	var errs []error

	// Find all img src attributes
	matches := v.imgRegex.FindAllSubmatch(content, -1)
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}

		src := string(match[1])
		if shared.IsExternalURL(src) {
			if v.SkipExternal {
				continue
			}
			if err := v.validateExternalImage(src); err != nil {
				errs = append(errs, fmt.Errorf("%s: external image not accessible: %s (%w)", v.HTMLPath, src, err))
			}
		} else {
			if err := v.validateLocalImage(src); err != nil {
				errs = append(errs, fmt.Errorf("%s: local image not found: %s", v.HTMLPath, src))
			}
		}
	}

	return errs
}

// validateExternalImage checks if an external image URL is accessible
func (v *Validator) validateExternalImage(url string) error {
	client := &http.Client{
		Timeout: v.Timeout,
	}

	resp, err := client.Head(url)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	return nil
}

// validateLocalImage checks if a local image file exists
func (v *Validator) validateLocalImage(src string) error {
	if _, err := v.fs.Stat(shared.ResolveLocalPath(src, v.HTMLPath, v.buildDir)); err != nil {
		return err
	}
	return nil
}
