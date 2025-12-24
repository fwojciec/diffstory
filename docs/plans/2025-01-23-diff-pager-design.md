# Diff Pager Design

Milestone 1: A competent `git diff | diffview` pager with proper domain types.

## Package Structure

```
diffview/
├── diffview.go          # Domain: Diff, Hunk, Line, Parser, Viewer interfaces
├── gitdiff/
│   └── parser.go        # Parser implementation using bluekeyes/go-gitdiff
├── bubbletea/
│   ├── viewer.go        # Viewer implementation, tea.Model
│   ├── keymap.go        # KeyMap, DefaultKeyMap()
│   ├── styles.go        # Theme, Styles, ColorPair
│   └── render.go        # Hunk/line rendering helpers
├── mock/
│   ├── parser.go
│   └── viewer.go
└── cmd/diffview/
    └── main.go          # Wires parser + viewer, reads stdin
```

## Domain Types (root package)

```go
package diffview

import "io"

// Parser parses unified diff input into hunks.
type Parser interface {
    Parse(r io.Reader) (Diff, error)
}

// Viewer displays a diff to the user.
type Viewer interface {
    View(diff Diff) error
}

// Diff is a collection of hunks, the atomic unit for display and navigation.
type Diff []Hunk

// Hunk represents a single change block with its file context.
type Hunk struct {
    FilePath   string
    OldPath    string     // For renames, empty if unchanged
    FileOp     FileOp
    IsBinary   bool       // Binary files have no lines
    Header     HunkHeader
    Lines      []Line
}

type HunkHeader struct {
    OldStart, OldCount int
    NewStart, NewCount int
    Section            string // Function context after @@
}

type Line struct {
    Type      LineType
    Content   string
    OldNum    int  // 0 if Added
    NewNum    int  // 0 if Deleted
    NoNewline bool // "\ No newline at end of file"
}

type LineType int
const (
    Context LineType = iota
    Added
    Deleted
)

type FileOp int
const (
    Modified FileOp = iota
    FileAdded
    FileDeleted
    Renamed
    Copied
)
```

## Parser Implementation (gitdiff/)

Uses `bluekeyes/go-gitdiff` to parse unified diff format.

- Flattens file-centric go-gitdiff output to hunk-centric domain types
- Computes per-line numbers during parse (not deferred)
- Binary files represented as hunks with `IsBinary: true`, no lines

```go
package gitdiff

var _ diffview.Parser = (*Parser)(nil)

type Parser struct{}

func (p *Parser) Parse(r io.Reader) (diffview.Diff, error) {
    files, _, err := gitdiff.Parse(r)
    // Flatten files → hunks, compute line numbers
}
```

## Viewer Implementation (bubbletea/)

Implements `diffview.Viewer` using Bubble Tea.

Construction details (not part of interface):
- `Styles` - computed from `Theme` + light/dark detection
- `KeyMap` - customizable key bindings

```go
package bubbletea

var _ diffview.Viewer = (*Viewer)(nil)

type Viewer struct {
    styles Styles
    keymap KeyMap
}

func NewViewer(styles Styles, keymap KeyMap) *Viewer {
    return &Viewer{styles: styles, keymap: keymap}
}

func (v *Viewer) View(diff diffview.Diff) error {
    m := NewModel(diff, v.styles, v.keymap)
    p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
    _, err := p.Run()
    return err
}
```

### Model

```go
type Model struct {
    // Domain
    diff          diffview.Diff
    hunkPositions []int  // Line offset where each hunk starts

    // View state
    viewport    viewport.Model
    ready       bool
    currentHunk int

    // Input state
    pendingKey string

    // Styling
    styles Styles
    keymap KeyMap
}
```

### Theming

Theme is data, Styles is computed:

```go
type Theme struct {
    Name        string
    Added       ColorPair
    Deleted     ColorPair
    Context     ColorPair
    LineNumber  ColorPair
    HunkHeader  ColorPair
    FileHeader  ColorPair
}

type ColorPair struct {
    Light lipgloss.TerminalColor
    Dark  lipgloss.TerminalColor
}

type Styles struct {
    Added      lipgloss.Style
    Deleted    lipgloss.Style
    Context    lipgloss.Style
    LineNumber lipgloss.Style
    HunkHeader lipgloss.Style
    FileHeader lipgloss.Style
}

func NewStyles(theme Theme, isDark bool) Styles
func DefaultTheme() Theme
```

### Key Bindings

```go
type KeyMap struct {
    Up, Down         key.Binding
    HalfPageUp       key.Binding
    HalfPageDown     key.Binding
    GotoTop          key.Binding  // gg (multi-key)
    GotoBottom       key.Binding  // G
    NextHunk         key.Binding  // n
    PrevHunk         key.Binding  // N
    NextFile         key.Binding  // ]
    PrevFile         key.Binding  // [
    Quit             key.Binding  // q
}

func DefaultKeyMap() KeyMap
```

Multi-key sequences (gg) handled via `pendingKey` state in Model.

## Main (cmd/diffview/)

```go
func main() {
    // Detect stdin pipe
    stat, _ := os.Stdin.Stat()
    if stat.Mode()&os.ModeNamedPipe == 0 {
        fmt.Fprintln(os.Stderr, "Usage: git diff | diffview")
        os.Exit(1)
    }

    // Parse
    parser := &gitdiff.Parser{}
    diff, err := parser.Parse(os.Stdin)
    if err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }

    // View
    theme := bubbletea.DefaultTheme()
    styles := bubbletea.NewStyles(theme, true) // TODO: detect dark mode
    keymap := bubbletea.DefaultKeyMap()
    viewer := bubbletea.NewViewer(styles, keymap)

    if err := viewer.View(diff); err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
}
```

## Navigation

- `j/k` or arrows: scroll lines (viewport handles this)
- `ctrl+d/ctrl+u`: half page
- `gg/G`: top/bottom
- `n/N`: next/prev hunk (uses `hunkPositions` for jump)
- `]/[`: next/prev file (jump to first hunk of next file)
- `q`: quit

## Deferred (not in Milestone 1)

- Syntax highlighting
- Side-by-side view
- Word-level diff highlighting
- Search (`/`)
- Config file loading
- Semantic grouping / AI features
