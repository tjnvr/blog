package description

import (
	"fmt"
	"html"

	"github.com/spf13/afero"

	"github.com/tjnvr/blog/internal/backbone/metadata"
)

// substituter resolves the meta description tag from an author-supplied
// `<!-- description: ... -->` markdown comment. Pages without one omit the
// tag entirely, so Google falls back to auto-generating a snippet.
type substituter struct {
	fs           afero.Fs
	markdownPath string
}

func NewSubstituer(fs afero.Fs, markdownPath string) substituter {
	return substituter{
		fs:           fs,
		markdownPath: markdownPath,
	}
}

func (s substituter) Placeholder() string {
	return `<meta name="description" content="{{description}}">`
}

func (s substituter) Resolve(_ string) (string, error) {
	data, err := afero.ReadFile(s.fs, s.markdownPath)
	if err != nil {
		return "", err
	}

	description := metadata.Extract(data).Description
	if description == "" {
		return "", nil
	}

	return fmt.Sprintf(`<meta name="description" content="%s">`, html.EscapeString(description)), nil
}
