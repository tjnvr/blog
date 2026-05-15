package link

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/tjnvr/blog/internal/generator/page/html/validation/shared"
)

// Validator checks that all links in HTML are accessible
type Validator struct {
	// Timeout for HTTP requests to external links
	Timeout time.Duration
	// SkipExternal skips validation of external URLs
	SkipExternal bool
	linkRegex    *regexp.Regexp
}

// NewValidator creates a new link validator with default settings
func NewValidator() *Validator {
	return &Validator{
		Timeout:      10 * time.Second,
		SkipExternal: false,
		linkRegex:    regexp.MustCompile(`<a[^>]+href="([^"]+)"`),
	}
}

// Validate checks all anchor href attributes in the HTML content
func (v *Validator) Validate(htmlPath, buildDir string, content []byte) []error {
	var errs []error

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
			if err := v.validateLocalLink(href, htmlPath, buildDir); err != nil {
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
	if _, err := os.Stat(linkPath); err == nil {
		return nil
	}

	// If path doesn't have an extension, check for index.html
	if filepath.Ext(linkPath) == "" {
		indexPath := filepath.Join(linkPath, "index.html")
		if _, err := os.Stat(indexPath); err == nil {
			return nil
		}
	}

	return fmt.Errorf("path not found: %s", linkPath)
}
