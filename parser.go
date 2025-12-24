package diffview

import "io"

// Parser parses diff content into domain types.
type Parser interface {
	// Parse reads diff content and returns the parsed result.
	Parse(r io.Reader) (*Diff, error)
}
