package validation

import "fmt"

// Error represents a validation failure
type Error struct {
	File    string
	Message string
}

func (e Error) Error() string {
	return e.File + ": " + e.Message
}

// NewError creates a new validation error
func NewError(file, format string, args ...any) Error {
	return Error{
		File:    file,
		Message: fmt.Sprintf(format, args...),
	}
}

// Validator validates generated HTML content
type Validator interface {
	// Validate checks the HTML content and returns any validation errors
	// htmlPath is the path to the generated HTML file
	// buildDir is the root build directory for resolving relative paths
	Validate(content []byte) []error
}
