package page

import (
	"crypto/rand"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGeneratorGenerate(t *testing.T) {
	// given
	sourcePath := rand.Text() + ".md"
	destPath := rand.Text() + "/" + rand.Text() + ".html"
	mdRaw := rand.Text()
	mdSubstituted := rand.Text()
	htmlConverted := rand.Text()
	htmlFinal := rand.Text()

	// setup
	fs := afero.NewMemMapFs()
	err := afero.WriteFile(fs, sourcePath, []byte(mdRaw), 0644)
	require.NoError(t, err)
	mdSub := new(mockMarkdownSubstituer)
	mdConv := new(mockMarkdownConverter)
	htmlSub := new(mockHTMLSubstituer)
	validator := new(mockPageValidator)
	mdSub.On("Apply", mdRaw).Return(mdSubstituted, nil)
	mdConv.On("Convert", []byte(mdSubstituted)).Return(htmlConverted, nil)
	htmlSub.On("Apply", defaultTemplate, htmlConverted).Return(htmlFinal, nil)
	g := NewGenerator(fs, mdSub, mdConv, htmlSub, validator)

	// test
	err = g.Generate(sourcePath, destPath)
	require.NoError(t, err)

	// expect
	mdSub.AssertExpectations(t)
	mdConv.AssertExpectations(t)
	htmlSub.AssertExpectations(t)
	validator.AssertNotCalled(t, "Validate", mock.Anything)
	exists, err := afero.Exists(fs, destPath)
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestGeneratorGenerate_ReadFileFails(t *testing.T) {
	// given
	sourcePath := rand.Text() + ".md"
	destPath := rand.Text() + "/" + rand.Text() + ".html"

	// setup
	fs := afero.NewMemMapFs() // empty file system so the read file will fail
	mdSub := new(mockMarkdownSubstituer)
	mdConv := new(mockMarkdownConverter)
	htmlSub := new(mockHTMLSubstituer)
	validator := new(mockPageValidator)
	g := NewGenerator(fs, mdSub, mdConv, htmlSub, validator)

	// test
	err := g.Generate(sourcePath, destPath)

	// expect
	require.ErrorContains(t, err, sourcePath)
	mdSub.AssertNotCalled(t, "Apply", mock.Anything)
	mdConv.AssertNotCalled(t, "Convert", mock.Anything)
	htmlSub.AssertNotCalled(t, "Apply", mock.Anything, mock.Anything)
	validator.AssertNotCalled(t, "Validate", mock.Anything)
	exists, err := afero.Exists(fs, destPath)
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestGeneratorGenerate_MarkdownSubstitutionFails(t *testing.T) {
	// given
	sourcePath := rand.Text() + ".md"
	destPath := rand.Text() + "/" + rand.Text() + ".html"
	mdRaw := rand.Text()

	// setup
	fs := afero.NewMemMapFs()
	err := afero.WriteFile(fs, sourcePath, []byte(mdRaw), 0644)
	require.NoError(t, err)
	mdSub := new(mockMarkdownSubstituer)
	mdConv := new(mockMarkdownConverter)
	htmlSub := new(mockHTMLSubstituer)
	validator := new(mockPageValidator)
	mdSub.On("Apply", mdRaw).Return("", assert.AnError)
	g := NewGenerator(fs, mdSub, mdConv, htmlSub, validator)

	// test
	err = g.Generate(sourcePath, destPath)

	// expect
	require.ErrorIs(t, err, assert.AnError)
	mdSub.AssertExpectations(t)
	mdConv.AssertNotCalled(t, "Convert", mock.Anything)
	htmlSub.AssertNotCalled(t, "Apply", mock.Anything, mock.Anything)
	validator.AssertNotCalled(t, "Validate", mock.Anything)
}

func TestGeneratorGenerate_MarkdownConversionFails(t *testing.T) {
	// given
	sourcePath := rand.Text() + ".md"
	destPath := rand.Text() + "/" + rand.Text() + ".html"
	mdRaw := rand.Text()
	mdSubstituted := rand.Text()

	// setup
	fs := afero.NewMemMapFs()
	err := afero.WriteFile(fs, sourcePath, []byte(mdRaw), 0644)
	require.NoError(t, err)
	mdSub := new(mockMarkdownSubstituer)
	mdConv := new(mockMarkdownConverter)
	htmlSub := new(mockHTMLSubstituer)
	validator := new(mockPageValidator)
	mdSub.On("Apply", mdRaw).Return(mdSubstituted, nil)
	mdConv.On("Convert", []byte(mdSubstituted)).Return("", assert.AnError)
	g := NewGenerator(fs, mdSub, mdConv, htmlSub, validator)

	// test
	err = g.Generate(sourcePath, destPath)

	// expect
	require.ErrorIs(t, err, assert.AnError)
	mdSub.AssertExpectations(t)
	mdConv.AssertExpectations(t)
	htmlSub.AssertNotCalled(t, "Apply", mock.Anything, mock.Anything)
	validator.AssertNotCalled(t, "Validate", mock.Anything)
}

func TestGeneratorGenerate_HTMLSubstitutionFails(t *testing.T) {
	// given
	sourcePath := rand.Text() + ".md"
	destPath := rand.Text() + "/" + rand.Text() + ".html"
	mdRaw := rand.Text()
	mdSubstituted := rand.Text()
	htmlConverted := rand.Text()

	// setup
	fs := afero.NewMemMapFs()
	err := afero.WriteFile(fs, sourcePath, []byte(mdRaw), 0644)
	require.NoError(t, err)
	mdSub := new(mockMarkdownSubstituer)
	mdConv := new(mockMarkdownConverter)
	htmlSub := new(mockHTMLSubstituer)
	validator := new(mockPageValidator)
	mdSub.On("Apply", mdRaw).Return(mdSubstituted, nil)
	mdConv.On("Convert", []byte(mdSubstituted)).Return(htmlConverted, nil)
	htmlSub.On("Apply", defaultTemplate, htmlConverted).Return("", assert.AnError)
	g := NewGenerator(fs, mdSub, mdConv, htmlSub, validator)

	// test
	err = g.Generate(sourcePath, destPath)

	// expect
	require.ErrorIs(t, err, assert.AnError)
	mdSub.AssertExpectations(t)
	mdConv.AssertExpectations(t)
	htmlSub.AssertExpectations(t)
	validator.AssertNotCalled(t, "Validate", mock.Anything)
}

func TestGeneratorValidate(t *testing.T) {
	// given
	htmlPath := rand.Text() + "/" + rand.Text() + ".html"

	// setup
	validator := new(mockPageValidator)
	validator.On("Validate", htmlPath).Return(nil)

	// test
	g := NewGenerator(afero.NewMemMapFs(), nil, nil, nil, validator)

	// expect
	require.NoError(t, g.Validate(htmlPath))
	validator.AssertExpectations(t)
}

func TestGeneratorValidate_Error(t *testing.T) {
	// given
	htmlPath := rand.Text() + "/" + rand.Text() + ".html"

	// setup
	validator := new(mockPageValidator)
	validator.On("Validate", htmlPath).Return(assert.AnError)

	// test
	g := NewGenerator(afero.NewMemMapFs(), nil, nil, nil, validator)

	// expect
	require.ErrorIs(t, g.Validate(htmlPath), assert.AnError)
	validator.AssertExpectations(t)
}

type mockMarkdownSubstituer struct {
	mock.Mock
}

func (m *mockMarkdownSubstituer) Apply(content string) (string, error) {
	args := m.Called(content)
	return args.String(0), args.Error(1)
}

type mockMarkdownConverter struct {
	mock.Mock
}

func (m *mockMarkdownConverter) Convert(content []byte) (string, error) {
	args := m.Called(content)
	return args.String(0), args.Error(1)
}

type mockHTMLSubstituer struct {
	mock.Mock
}

func (m *mockHTMLSubstituer) Apply(template, content string) (string, error) {
	args := m.Called(template, content)
	return args.String(0), args.Error(1)
}

type mockPageValidator struct {
	mock.Mock
}

func (m *mockPageValidator) Validate(htmlPath string) error {
	args := m.Called(htmlPath)
	return args.Error(0)
}
