package diffview

import "context"

// Viewer displays a diff to the user.
type Viewer interface {
	// View displays the diff and blocks until the user exits.
	View(ctx context.Context, diff *Diff) error
}
