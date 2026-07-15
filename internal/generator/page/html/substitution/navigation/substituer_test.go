package navigation

import (
	"errors"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/tjnvr/blog/internal/backbone/section"
)

type mockPathResolver struct {
	ResolveFunc func(oldPath, fromPath string) (string, error)
}

func (m *mockPathResolver) Resolve(oldPath, fromPath string) (string, error) {
	if m.ResolveFunc != nil {
		return m.ResolveFunc(oldPath, fromPath)
	}
	return "", errors.New("not implemented")
}

type mockSectionResolver struct {
	ResolveFunc        func() ([]section.Section, error)
	ResolveForFileFunc func(markdownSourceFile string) (section.Section, error)
}

func (m *mockSectionResolver) Resolve() ([]section.Section, error) {
	if m.ResolveFunc != nil {
		return m.ResolveFunc()
	}
	return nil, errors.New("not implemented")
}

func (m *mockSectionResolver) ResolveForFile(markdownSourceFile string) (section.Section, error) {
	if m.ResolveForFileFunc != nil {
		return m.ResolveForFileFunc(markdownSourceFile)
	}
	return section.Section{}, errors.New("not implemented")
}

// navAnchor mirrors the actual rendering of template.html's "node" definition.
func navAnchor(href, text string, focused bool) string {
	class := "nav-link font-semibold"
	if focused {
		class = "nav-link font-semibold is-active"
	}
	return fmt.Sprintf(`<a href="%s" class="%s">%s</a>`, href, class, text)
}

func TestSubstituter_Placeholder_ShouldReturnCorrectPlaceholder(t *testing.T) {
	// setup
	substituter := NewSubstituer(&mockPathResolver{}, &mockSectionResolver{}, "", "")

	// test
	result := substituter.Placeholder()

	// expect
	assert.Equal(t, "{{navigation}}", result)
}

func TestNewSubstituer_ShouldCreateValidSubstituter(t *testing.T) {
	// given
	markdownSourcePath := uuid.New().String()
	htmlPath := uuid.New().String()

	// test
	substituter := NewSubstituer(&mockPathResolver{}, &mockSectionResolver{}, markdownSourcePath, htmlPath)

	// expect
	assert.NotNil(t, substituter.template)
}

func TestSubstituter_Resolve_ShouldCallCollaboratorsWithExpectedArguments(t *testing.T) {
	// given
	markdownSourcePath := uuid.New().String()
	htmlPath := uuid.New().String()
	sect := section.Section{HomePath: uuid.New().String(), DisplayName: uuid.New().String()}
	href := uuid.New().String()

	// setup
	var capturedMarkdownSourcePath, capturedOldPath, capturedFromPath string
	pathResolver := &mockPathResolver{
		ResolveFunc: func(oldPath, fromPath string) (string, error) {
			capturedOldPath = oldPath
			capturedFromPath = fromPath
			return href, nil
		},
	}
	sectionsResolver := &mockSectionResolver{
		ResolveFunc: func() ([]section.Section, error) {
			return []section.Section{sect}, nil
		},
		ResolveForFileFunc: func(markdownSourceFile string) (section.Section, error) {
			capturedMarkdownSourcePath = markdownSourceFile
			return sect, nil
		},
	}
	substituter := NewSubstituer(pathResolver, sectionsResolver, markdownSourcePath, htmlPath)

	// test
	_, err := substituter.Resolve("")

	// expect
	assert.NoError(t, err)
	assert.Equal(t, markdownSourcePath, capturedMarkdownSourcePath)
	assert.Equal(t, sect.HomePath, capturedOldPath)
	assert.Equal(t, htmlPath, capturedFromPath)
}

func TestSubstituter_Resolve(t *testing.T) {
	// given
	sectionA := section.Section{HomePath: uuid.New().String(), DisplayName: uuid.New().String()}
	sectionB := section.Section{HomePath: uuid.New().String(), DisplayName: uuid.New().String()}
	hrefA := uuid.New().String()
	hrefB := uuid.New().String()
	hrefs := map[string]string{sectionA.HomePath: hrefA, sectionB.HomePath: hrefB}
	notPresentSection := section.Section{HomePath: uuid.New().String(), DisplayName: uuid.New().String()}

	tests := []struct {
		name              string
		sections          []section.Section
		sectionsErr       error
		currentSection    section.Section
		resolveForFileErr error
		hrefErr           error
		wantOutput        string
		wantErrSubstr     string
	}{
		{
			name:       "renders empty nav when there are no sections",
			sections:   []section.Section{},
			wantOutput: `<nav class="not-prose flex flex-col sm:flex-row sm:items-baseline gap-4"></nav>`,
		},
		{
			name:           "renders each section with the current one focused",
			sections:       []section.Section{sectionA, sectionB},
			currentSection: sectionB,
			wantOutput: fmt.Sprintf(
				`<nav class="not-prose flex flex-col sm:flex-row sm:items-baseline gap-4">%s%s</nav>`,
				navAnchor(hrefA, sectionA.DisplayName, false),
				navAnchor(hrefB, sectionB.DisplayName, true),
			),
		},
		{
			name:           "focuses nothing when the current section is not in the list",
			sections:       []section.Section{sectionA, sectionB},
			currentSection: notPresentSection,
			wantOutput: fmt.Sprintf(
				`<nav class="not-prose flex flex-col sm:flex-row sm:items-baseline gap-4">%s%s</nav>`,
				navAnchor(hrefA, sectionA.DisplayName, false),
				navAnchor(hrefB, sectionB.DisplayName, false),
			),
		},
		{
			name:          "wraps the error when the sections resolver fails",
			sectionsErr:   errors.New(uuid.New().String()),
			wantErrSubstr: "Resolve err",
		},
		{
			name:              "wraps the error when resolving the current section fails",
			sections:          []section.Section{sectionA},
			resolveForFileErr: errors.New(uuid.New().String()),
			wantErrSubstr:     "ResolveForFile",
		},
		{
			name:           "wraps the error when resolving a section href fails",
			sections:       []section.Section{sectionA},
			currentSection: sectionA,
			hrefErr:        errors.New(uuid.New().String()),
			wantErrSubstr:  "GetHRef err",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// setup
			pathResolver := &mockPathResolver{
				ResolveFunc: func(oldPath, _ string) (string, error) {
					if tt.hrefErr != nil {
						return "", tt.hrefErr
					}
					return hrefs[oldPath], nil
				},
			}
			sectionsResolver := &mockSectionResolver{
				ResolveFunc: func() ([]section.Section, error) {
					return tt.sections, tt.sectionsErr
				},
				ResolveForFileFunc: func(string) (section.Section, error) {
					return tt.currentSection, tt.resolveForFileErr
				},
			}
			substituter := NewSubstituer(pathResolver, sectionsResolver, "", "")

			// test
			result, err := substituter.Resolve("")

			// expect
			if tt.wantErrSubstr != "" {
				assert.ErrorContains(t, err, tt.wantErrSubstr)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantOutput, result)
		})
	}
}
