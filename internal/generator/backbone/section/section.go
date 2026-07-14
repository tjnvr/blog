package section

// Section represents a top-level site section.
type Section struct {
	HomePath    string // Absolute path of the section from the root
	DisplayName string // display name shown in navigation (from # title in index.md)
}
