package htmlref

import "strings"

// Classifier decides how a reference must be handled. It folds together the
// external-URL test and the skip rules (fragments, mailto, tel, javascript)
// that were previously duplicated across the validators.
type Classifier interface {
	// Classify reports how value must be treated. It never returns an error;
	// an unrecognised value defaults to KindLocal so existence is still
	// checked rather than silently ignored.
	Classify(value string) Kind
}

type classifier struct {
	skipPrefixes []string
}

// NewClassifier returns a Classifier that reports KindSkip for any value
// starting with one of skipPrefixes, KindExternal for http:// and https://
// URLs, and KindLocal otherwise. Typical skip prefixes are "#", "mailto:",
// "tel:" and "javascript:"; pass none to classify only external versus local.
func NewClassifier(skipPrefixes ...string) Classifier {
	return classifier{skipPrefixes: skipPrefixes}
}

func (c classifier) Classify(value string) Kind {
	for _, p := range c.skipPrefixes {
		if strings.HasPrefix(value, p) {
			return KindSkip
		}
	}
	if strings.HasPrefix(value, "http://") || strings.HasPrefix(value, "https://") {
		return KindExternal
	}
	return KindLocal
}
