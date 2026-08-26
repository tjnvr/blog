package description

import (
	"fmt"
	"html"

	"github.com/tjnvr/blog/internal/backbone/metadata"
)

// substituter resolves the meta description tag from the page's already-extracted
// `<!-- description: ... -->` metadata. Pages without one omit the tag entirely, so
// Google falls back to auto-generating a snippet.
type substituter struct {
	m metadata.Metadata
}

func NewSubstituer(m metadata.Metadata) substituter {
	return substituter{m: m}
}

func (s substituter) Placeholder() string {
	return `<meta name="description" content="{{description}}">`
}

func (s substituter) Resolve(_ string) (string, error) {
	if s.m.Description == "" {
		return "", nil
	}

	return fmt.Sprintf(`<meta name="description" content="%s">`, html.EscapeString(s.m.Description)), nil
}
