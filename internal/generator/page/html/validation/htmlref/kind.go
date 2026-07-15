// Package htmlref locates and classifies the references (the values of src and
// href attributes) found in generated HTML.
//
// It is the "read the HTML" half of validation: an Extractor pulls raw
// reference strings out of the markup, and a Classifier decides how each one
// must be handled (skipped, checked over the network, or resolved on the local
// filesystem). Both are configured once and reused for every page.
package htmlref

// Kind classifies a reference so a validator can route it to the right check.
type Kind int

const (
	// KindSkip marks references that must not be validated: fragment-only
	// links ("#..."), and the mailto:, tel: and javascript: schemes.
	KindSkip Kind = iota
	// KindExternal marks http:// and https:// references, checked over the
	// network.
	KindExternal
	// KindLocal marks references resolved against the local build output and
	// checked for existence on the filesystem.
	KindLocal
)

// String returns the lowercase name of the kind ("skip", "external" or
// "local"), or "unknown" for an out-of-range value.
func (k Kind) String() string {
	switch k {
	case KindSkip:
		return "skip"
	case KindExternal:
		return "external"
	case KindLocal:
		return "local"
	default:
		return "unknown"
	}
}
