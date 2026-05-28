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
	SkipExternal bool
	imgRegex     *regexp.Regexp
	buildDir string
}

// NewValidator creates a new image validator with default settings
func NewValidator(fs afero.Fs, skipExternalValidation bool) *Validator {
	return &Validator{
		fs:           fs,
		Timeout:      10 * time.Second,
		SkipExternal: skipExternalValidation,
		imgRegex:     regexp.MustCompile(`<img[^>]+src="([^"]+)"`),
	}
}

// Validate checks all img src attributes in the HTML content
func (v *Validator) Validate(htmlPath string) []error {
	var errs []error
	content, err := afero.ReadFile(v.fs, htmlPath)
	if err != nil {
		return []error{fmt.Errorf("could not read file (%s): %v", htmlPath, err)}
	}

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
				errs = append(errs, fmt.Errorf("%s: external image not accessible: %s (%w)", htmlPath, src, err))
			}
		} else {
			if err := v.validateLocalImage(src, htmlPath, v.buildDir); err != nil {
				errs = append(errs, fmt.Errorf("%s: local image not found: %s", htmlPath, src))
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
func (v *Validator) validateLocalImage(src, htmlPath, buildDir string) error {
	imagePath := shared.ResolveLocalPath(src, htmlPath, buildDir)
	if _, err := v.fs.Stat(imagePath); err != nil {
		return err
	}
	return nil
}
