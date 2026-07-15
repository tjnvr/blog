package validation

// Validator validates generated HTML content.
//
// The page path, the build output root and the filesystem are injected when the
// validator is constructed (see NewRegistry), so Validate receives only the
// generated content of the page being checked.
type Validator interface {
	// Validate checks content and returns one error per validation failure, or
	// an empty slice when the content is valid.
	Validate(content []byte) []error
}
