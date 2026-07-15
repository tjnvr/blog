package htmlref

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClassifier_Classify(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  Kind
	}{
		{name: "http url is external", value: "http://example.com", want: KindExternal},
		{name: "https url is external", value: "https://example.com/a", want: KindExternal},
		{name: "fragment is skipped", value: "#section", want: KindSkip},
		{name: "mailto is skipped", value: "mailto:user@example.com", want: KindSkip},
		{name: "tel is skipped", value: "tel:+33123456789", want: KindSkip},
		{name: "javascript is skipped", value: "javascript:void(0)", want: KindSkip},
		{name: "relative path is local", value: "../assets/logo.png", want: KindLocal},
		{name: "absolute path is local", value: "/assets/logo.png", want: KindLocal},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// setup
			classifier := NewClassifier("#", "mailto:", "tel:", "javascript:")

			// test
			got := classifier.Classify(tt.value)

			// expect
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestClassifier_Classify_ShouldTreatUnconfiguredPrefixAsLocal(t *testing.T) {
	// setup
	classifier := NewClassifier()

	// test
	got := classifier.Classify("mailto:user@example.com")

	// expect
	assert.Equal(t, KindLocal, got)
}
