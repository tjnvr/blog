package summary

import (
	"bytes"
	_ "embed"
	"html/template"

	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var (
	// The HTML template used to render the summary
	//
	// https://pkg.go.dev/html/template
	//go:embed template.html
	summaryTemplate string

	// The summary include all headings from h2 to h6 (h1 being the page title is never listed in the summary)
	headingsRegex = regexp.MustCompile(`<h([2-6])[^>]*id="([^"]+)"[^>]*>([^<]+)(?:<a[^>]*>[^<]*</a>)?</h[2-6]>`)
)

type (
	Heading struct {
		ID    string
		Level int
		Text  string
	}

	headingNode struct {
		Heading
		Children []*headingNode
	}
)

// Substituter resolves the {{summary}} placeholder with a generated table of contents.
type Substituter struct {
	template *template.Template
}

func NewSubstituer() Substituter {
	return Substituter{
		template: template.Must(template.New("summary").Parse(summaryTemplate)),
	}
}

func (s Substituter) Placeholder() string {
	return "<p>{{summary}}</p>"
}

func (s Substituter) Resolve(content string) (string, error) {
	headings := listHeadings(content)
	if len(headings) == 0 {
		return "", nil
	}

	headingTrees := buildHeadingTrees(headings)

	var buf bytes.Buffer
	err := s.template.ExecuteTemplate(&buf, "toc", headingTrees)
	if err != nil {
		return "", fmt.Errorf("Execute err: %v", err)
	}

	return buf.String(), nil
}

func listHeadings(content string) []Heading {
	matches := headingsRegex.FindAllStringSubmatch(content, -1)
	if len(matches) == 0 {
		return nil
	}

	headings := make([]Heading, len(matches))
	for i, m := range matches {
		// Ensured by the [headingsRegex]
		level, _ := strconv.Atoi(m[1])
		headings[i] = Heading{Level: level, ID: m[2], Text: strings.TrimSpace(m[3])}
	}

	return headings
}

func buildHeadingTrees(headings []Heading) []*headingNode {
	if len(headings) == 0 {
		return nil
	}

	var roots []*headingNode
	stack := make([]*headingNode, 0, 8)

	for _, h := range headings {
		n := &headingNode{Heading: h}

		// Pop until we find a proper parent
		for len(stack) > 0 && stack[len(stack)-1].Level >= h.Level {
			stack = stack[:len(stack)-1]
		}

		if len(stack) == 0 {
			roots = append(roots, n)
		} else {
			stack[len(stack)-1].Children = append(stack[len(stack)-1].Children, n)
		}

		stack = append(stack, n)
	}

	return roots
}
