// Package diffview provides domain types for parsing and viewing diffs.
package diffview

import "io/fs"

// Diff represents a complete diff containing one or more file changes.
type Diff struct {
	Files []FileDiff
}

// FileDiff represents changes to a single file.
type FileDiff struct {
	OldPath   string      // "a/file.go" or empty for new files
	NewPath   string      // "b/file.go" or empty for deleted files
	Operation FileOp      // Added, Deleted, Modified, Renamed, Copied
	IsBinary  bool        // Binary files have no hunks
	OldMode   fs.FileMode // 0 if unchanged
	NewMode   fs.FileMode // For permission changes
	Hunks     []Hunk
	Extended  []string // Raw extended headers for passthrough
}

// FileOp represents the type of operation performed on a file.
type FileOp int

// File operation types.
const (
	FileModified FileOp = iota
	FileAdded
	FileDeleted
	FileRenamed
	FileCopied
)

// Hunk represents a contiguous block of changes within a file.
type Hunk struct {
	OldStart int    // From @@ -X,...
	OldCount int    // From @@ -X,Y ...
	NewStart int    // From @@ ...,+X
	NewCount int    // From @@ ...,+X,Y
	Section  string // Optional function name after @@ ... @@
	Lines    []Line
}

// Line represents a single line within a hunk.
type Line struct {
	Type       LineType
	Content    string
	OldLineNum int  // 0 if line is Added
	NewLineNum int  // 0 if line is Deleted
	NoNewline  bool // "\ No newline at end of file" marker
}

// LineType represents the type of a diff line.
type LineType int

// Line types.
const (
	LineContext LineType = iota
	LineAdded
	LineDeleted
)
