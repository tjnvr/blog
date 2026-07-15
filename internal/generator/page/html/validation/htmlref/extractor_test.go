package htmlref

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestTagAttrExtractor_Extract_ShouldReturnEveryAttributeValueInOrder(t *testing.T) {
	// given
	first := uuid.New().String()
	second := uuid.New().String()

	// setup
	extractor := NewTagAttrExtractor("img", "src")
	content := []byte(`<img src="` + first + `"><img alt="x" src="` + second + `">`)

	// test
	refs := extractor.Extract(content)

	// expect
	assert.Equal(t, []string{first, second}, refs)
}

func TestTagAttrExtractor_Extract_ShouldReturnEmptyWhenTagDoesNotMatch(t *testing.T) {
	// given
	value := uuid.New().String()

	// setup
	extractor := NewTagAttrExtractor("img", "src")
	content := []byte(`<a href="` + value + `">link</a>`)

	// test
	refs := extractor.Extract(content)

	// expect
	assert.Empty(t, refs)
}

func TestTagAttrExtractor_Extract_ShouldMatchTargetAttributeOnly(t *testing.T) {
	// given
	src := uuid.New().String()

	// setup
	extractor := NewTagAttrExtractor("script", "src")
	content := []byte(`<script type="text/javascript" src="` + src + `"></script>`)

	// test
	refs := extractor.Extract(content)

	// expect
	assert.Equal(t, []string{src}, refs)
}
