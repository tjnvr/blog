package title

import (
	"fmt"

	"github.com/tjnvr/blog/internal/backbone/metadata"
)

// substituter resolves {{title}} placeholder from the page's already-extracted metadata.
type substituter struct {
	m metadata.Metadata
}

func NewSubstituer(m metadata.Metadata) substituter {
	return substituter{m: m}
}

func (t substituter) Placeholder() string {
	return "{{title}}"
}

func (t substituter) Resolve(_ string) (string, error) {
	if t.m.Title == "" {
		return "", fmt.Errorf("could not find a page title")
	}

	return t.m.Title, nil
}
