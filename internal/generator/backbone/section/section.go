// Package section resolves the top-level site sections that a piece of
// content belongs to.
//
// A section is a first-level directory under the content directory, and its
// display name is taken from the first Markdown "# " heading of the index.md
// file at the root of that directory. Content placed directly under the
// content directory belongs to the root section, whose index.md sits at the
// content root.
//
// Use NewResolver to build a Resolver, then call Resolve to list every
// section or ResolveForFile to find the section owning a given Markdown file.
package section

// Section represents a top-level site section.
type Section struct {
	HomePath    string // Path to the section's index.md, in the same form as contentDirectory.
	DisplayName string // Display name shown in navigation, from the index.md "# " title.
}
