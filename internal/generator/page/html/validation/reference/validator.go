// Package reference validates the references (images, links, scripts) embedded
// in generated HTML.
//
// A single Validator covers any reference kind, configured entirely by
// injection: the Extractor selects which markup to scan, the Classifier decides
// how each reference is handled, the ExternalChecker probes external URLs, the
// LocalResolver checks local targets, and an optional LocalCheck runs an extra
// check (such as JavaScript syntax) on each existing local file. Image, link
// and script validators differ only by the collaborators passed to NewValidator.
package reference

import (
	"fmt"

	"github.com/tjnvr/blog/internal/generator/page/html/validation/access"
	"github.com/tjnvr/blog/internal/generator/page/html/validation/htmlref"
)

// LocalCheck is an optional extra validation run on a local reference that has
// been found to exist, keyed by its resolved filesystem path. A nil LocalCheck
// is ignored by the validator.
type LocalCheck interface {
	Check(resolvedPath string) []error
}

// Validator checks every reference of one kind in generated HTML.
type Validator struct {
	extractor  htmlref.Extractor
	classifier htmlref.Classifier
	external   access.ExternalChecker
	local      access.LocalResolver
	post       LocalCheck
	htmlPath   string
	label      string
}

// NewValidator returns a Validator. label names the reference kind in error
// messages (for example "image", "link" or "script"); htmlPath is the page
// being validated, used both by local and in messages; post may be nil.
func NewValidator(
	extractor htmlref.Extractor,
	classifier htmlref.Classifier,
	external access.ExternalChecker,
	local access.LocalResolver,
	post LocalCheck,
	htmlPath, label string,
) *Validator {
	return &Validator{
		extractor:  extractor,
		classifier: classifier,
		external:   external,
		local:      local,
		post:       post,
		htmlPath:   htmlPath,
		label:      label,
	}
}

// Validate checks every reference found in content and returns one error per
// unreachable external URL, missing local target, or failed post-check.
func (v *Validator) Validate(content []byte) []error {
	var errs []error
	for _, ref := range v.extractor.Extract(content) {
		switch v.classifier.Classify(ref) {
		case htmlref.KindSkip:
			continue
		case htmlref.KindExternal:
			if err := v.external.Check(ref); err != nil {
				errs = append(errs, fmt.Errorf("%s: external %s not accessible: %s (%w)", v.htmlPath, v.label, ref, err))
			}
		case htmlref.KindLocal:
			errs = append(errs, v.validateLocal(ref)...)
		}
	}
	return errs
}

func (v *Validator) validateLocal(ref string) []error {
	ok, err := v.local.Exists(ref)
	if err != nil {
		return []error{fmt.Errorf("%s: checking local %s %s: %w", v.htmlPath, v.label, ref, err)}
	}
	if !ok {
		return []error{fmt.Errorf("%s: local %s not found: %s", v.htmlPath, v.label, ref)}
	}
	if v.post != nil {
		return v.post.Check(v.local.Resolve(ref))
	}
	return nil
}
