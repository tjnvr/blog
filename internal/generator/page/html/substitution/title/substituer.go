package title

import (
	"fmt"
	"regexp"
)

// substituter resolves {{title}} placeholder
type substituter struct {
}

func NewSubstituer() substituter {
	return substituter{}
}

func (t substituter) Placeholder() string {
	return "{{title}}"
}

func (t substituter) Resolve(content string) (string, error) {
	// Look for <h1> tag in HTML content, skipping any anchor link inside
	re := regexp.MustCompile(`<h1[^>]*>([^<]+)(?:<a[^>]*>[^<]*</a>)?</h1>`)
	match := re.FindSubmatch([]byte(content))
	if len(match) >= 2 {
		return string(match[1]), nil
	}

	return "", fmt.Errorf("could not find a page title")
}
