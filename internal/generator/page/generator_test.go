package page

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockMarkdownSubstituer struct {
	mock.Mock
	MarkdownSubstituer
}

func (m *mockMarkdownSubstituer) Apply(content string) (string, error) {
	args := m.Called(content)
	return args.String(0), args.Error(1)
}

type mockMarkdownConverter struct {
	mock.Mock
	MarkdownConverter
}

func (m *mockMarkdownConverter) Convert(content []byte) (string, error) {
	args := m.Called(content)
	return args.String(0), args.Error(1)
}

type mockHTMLSubstituer struct {
	mock.Mock
	HTMLSubstituer
}

func (m *mockHTMLSubstituer) Apply(template, content string) (string, error) {
	args := m.Called(template, content)
	return args.String(0), args.Error(1)
}

type mockPageValidator struct {
	mock.Mock
	PageValidator
}

func (m *mockPageValidator) Validate(htmlPath string) error {
	args := m.Called(htmlPath)
	return args.Error(0)
}

func TestGeneratorGenerate(t *testing.T) {
	// given
	fs := afero.NewMemMapFs()
	sourcePath := "source.md"
	destPath := "output/dest.html"
	dummyContent := "dummy markdown content"

	// setup
	err := afero.WriteFile(fs, sourcePath, []byte(dummyContent), 0644)
	require.NoError(t, err)
	mdSub := new(mockMarkdownSubstituer)
	mdConv := new(mockMarkdownConverter)
	htmlSub := new(mockHTMLSubstituer)
	validator := new(mockPageValidator)
	mdSub.On("Apply", dummyContent).Return(dummyContent, nil)
	mdConv.On("Convert", []byte(dummyContent)).Return(dummyContent, nil)
	htmlSub.On("Apply", defaultTemplate, dummyContent).Return("<html>dummy</html>", nil)
	g := NewGenerator(fs, mdSub, mdConv, htmlSub, validator)

	// test
	err = g.Generate(sourcePath, destPath)
	require.NoError(t, err)

	// expect
	mdSub.AssertCalled(t, "Apply", dummyContent)
	mdConv.AssertCalled(t, "Convert", []byte(dummyContent))
	htmlSub.AssertCalled(t, "Apply", defaultTemplate, dummyContent)
	mdSub.AssertExpectations(t)
	mdConv.AssertExpectations(t)
	htmlSub.AssertExpectations(t)
	validator.AssertNotCalled(t, "Validate", mock.Anything)
	exists, err := afero.Exists(fs, destPath)
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestGeneratorValidate(t *testing.T) {
	// given
	htmlPath := "some/path.html"

	// setup
	validator := new(mockPageValidator)
	validator.On("Validate", htmlPath).Return(nil)

	// test
	g := NewGenerator(afero.NewMemMapFs(), nil, nil, nil, validator)

	// expect
	require.NoError(t, g.Validate(htmlPath))
	validator.AssertExpectations(t)
}
