package link

import (
	"fmt"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/tjnvr/blog/internal/generator/page/html/validation/shared"

	"github.com/spf13/afero"
)

// Validator checks that all links in HTML are accessible
type Validator struct {
	fs afero.Fs
	// Timeout for HTTP requests to external links
	Timeout time.Duration
	// SkipExternal skips validation of external URLs
	SkipExternal bool
	linkRegex    *regexp.Regexp
	buildDir string
}

// NewValidator creates a new link validator with default settings
func NewValidator(fs afero.Fs, skipExternalValidation bool) *Validator {
	return &Validator{
		fs:           fs,
		Timeout:      10 * time.Second,
		SkipExternal: skipExternalValidation,
		linkRegex:    regexp.MustCompile(`<a[^>]+href="([^"]+)"`),
	}
}

// Validate checks all anchor href attributes in the HTML content
func (v *Validator) Validate(htmlPath string) []error {
	var errs []error
	content, err := afero.ReadFile(v.fs, htmlPath)
	if err != nil {
		return []error{fmt.Errorf("could not read file (%s): %v", htmlPath, err)}
	}

	// Find all anchor href attributes
	matches := v.linkRegex.FindAllSubmatch(content, -1)

	for _, match := range matches {
		if len(match) < 2 {
			continue
		}

		href := string(match[1])

		// Skip fragment-only links (e.g., #section)
		if strings.HasPrefix(href, "#") {
			continue
		}

		// Skip mailto and tel links
		if strings.HasPrefix(href, "mailto:") || strings.HasPrefix(href, "tel:") {
			continue
		}

		// Skip javascript links
		if strings.HasPrefix(href, "javascript:") {
			continue
		}

		if shared.IsExternalURL(href) {
			if v.SkipExternal {
				continue
			}
			if err := v.validateExternalLink(href); err != nil {
				errs = append(errs, fmt.Errorf("%s: external link not accessible: %s (%w)", htmlPath, href, err))
			}
		} else {
			if err := v.validateLocalLink(href, htmlPath, v.buildDir); err != nil {
				errs = append(errs, fmt.Errorf("%s: local link not found: %s", htmlPath, href))
			}
		}
	}

	return errs
}

// validateExternalLink checks if an external URL is accessible
func (v *Validator) validateExternalLink(url string) error {
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

// validateLocalLink checks if a local link target exists
func (v *Validator) validateLocalLink(href, htmlPath, buildDir string) error {
	// Remove fragment identifier if present
	href = strings.Split(href, "#")[0]

	// Empty href after removing fragment means same-page link
	if href == "" {
		return nil
	}

	linkPath := shared.ResolveLocalPath(href, htmlPath, buildDir)

	// Check if path exists as-is (could be a file or directory)
	if _, err := v.fs.Stat(linkPath); err == nil {
		return nil
	}

	// If path doesn't have an extension, check for index.html
	if filepath.Ext(linkPath) == "" {
		indexPath := filepath.Join(linkPath, "index.html")
		if _, err := v.fs.Stat(indexPath); err == nil {
			return nil
		}
	}

	return fmt.Errorf("path not found: %s", linkPath)
}
