package navigation

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"

	"github.com/tjnvr/blog/internal/generator/backbone/section"
	"github.com/tjnvr/blog/internal/relpath"
)

var (
	// The HTML template used to render the summary
	//
	// https://pkg.go.dev/html/template
	//go:embed template.html
	navTemplate string
)

type (
	// substituter resolves {{navigation}} placeholder with an auto-generated nav bar
	substituter struct {
		template         *template.Template
		pathsResolver    relpath.Resolver
		sectionsResolver section.Resolver
		markdownPath     string
		HTMLPath         string
	}

	navItem struct {
		HRef      string
		Text      string
		IsFocused bool // To display differently currently activated section
	}
)

func NewSubstituer(pathsResolver relpath.Resolver, sectionsResolver section.Resolver, markdownPath, HTMLPath string) substituter {
	return substituter{
		template:         template.Must(template.New("nav").Parse(navTemplate)),
		pathsResolver:    pathsResolver,
		sectionsResolver: sectionsResolver,
		markdownPath:     markdownPath,
		HTMLPath:         HTMLPath,
	}
}

func (n substituter) Placeholder() string {
	return "{{navigation}}"
}

func (n substituter) Resolve(_ string) (string, error) {
	sections, err := n.sectionsResolver.Resolve()
	if err != nil {
		return "", fmt.Errorf("Resolve err: %v", err)
	}

	currentSection, err := n.sectionsResolver.ResolveForFile(n.markdownPath)
	if err != nil {
		return "", fmt.Errorf("ResolveForFile(%s) err: %v", n.markdownPath, err)
	}
	var navItems = make([]navItem, len(sections))
	for i, s := range sections {
		hRef, err := n.pathsResolver.Resolve(s.HomePath, n.HTMLPath)
		if err != nil {
			return "", fmt.Errorf("GetHRef err for '%s' from '%s', %v", s.HomePath, n.HTMLPath, err)
		}
		navItem := navItem{
			HRef: hRef,
			Text: s.DisplayName,
		}

		if s == currentSection {
			navItem.IsFocused = true
		}
		navItems[i] = navItem
	}

	var buf bytes.Buffer
	err = n.template.ExecuteTemplate(&buf, "toc", navItems)
	if err != nil {
		return "", fmt.Errorf("execute err: %v", err)
	}

	return buf.String(), nil
}
