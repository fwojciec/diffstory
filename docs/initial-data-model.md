# Diff Pager Architecture: Data Model and Display Patterns for Bubble Tea

Your proposed data model is a solid foundation but requires key additions for edge cases. **bluekeyes/go-gitdiff** emerges as the recommended parser for Git-specific diffs, with a file-centric parse → hunk-centric display transformation. The Bubble Tea architecture should use the standard `viewport.Model` with pre-rendered content for PR-sized diffs, falling back to hunk pagination for massive files.

## Validated data model with edge case coverage

Your proposed model captures the essentials but misses several edge cases that Git diffs commonly produce. Analysis of sourcegraph/go-diff and bluekeyes/go-gitdiff reveals these gaps:

```go
// Recommended enhanced model
type Diff struct {
    Files []FileDiff
}

type FileDiff struct {
    OldPath    string        // "a/file.go" or empty for new files
    NewPath    string        // "b/file.go" or empty for deleted files
    Operation  FileOp        // Added, Deleted, Modified, Renamed, Copied
    IsBinary   bool          // Binary files have no hunks
    OldMode    os.FileMode   // 0 if unchanged
    NewMode    os.FileMode   // For permission changes
    Hunks      []Hunk
    Extended   []string      // Raw extended headers for passthrough
}

type FileOp int
const (
    FileModified FileOp = iota
    FileAdded
    FileDeleted
    FileRenamed
    FileCopied
)

type Hunk struct {
    OldStart  int           // From @@ -X,...
    OldCount  int           // From @@ -X,Y ...
    NewStart  int           // From @@ ...,+X
    NewCount  int           // From @@ ...,+X,Y
    Section   string        // Optional function name after @@ ... @@
    Lines     []Line
}

type Line struct {
    Type       LineType
    Content    string
    OldLineNum int          // 0 if line is Added
    NewLineNum int          // 0 if line is Deleted
    NoNewline  bool         // "\ No newline at end of file" marker
}

type LineType int
const (
    LineContext LineType = iota
    LineAdded
    LineDeleted
)
```

**Critical additions to your proposed model:**
- **`IsBinary` flag** — Binary files appear in diffs but have no parseable hunks
- **`FileOp` enum** — Distinguishes new/deleted/renamed/copied files (affects header rendering)
- **`OldMode`/`NewMode`** — Mode-only changes are valid diffs with empty hunks
- **`NoNewline` on Line** — The `\ No newline` marker must be tracked, not discarded
- **`Section` on Hunk** — The `@@ ... @@ function_name()` portion aids navigation
- **`Extended` array** — Preserves unparsed extended headers for display

## Parser selection: go-gitdiff for Git, sourcegraph for general

| Library | Purpose | Line Parsing | Best For |
|---------|---------|--------------|----------|
| **bluekeyes/go-gitdiff** | Parse + apply Git patches | Pre-parsed `[]Line` | Git-specific pager (recommended) |
| **sourcegraph/go-diff** | Parse unified diffs | Raw `[]byte` body | General unified diff, streaming |
| **sergi/go-diff** | Compute diffs (Myers) | N/A | **Not a parser — do not use** |

**go-gitdiff advantages:** Typed `IsRename`/`IsCopy`/`IsNew`/`IsDelete` booleans, proper binary patch support, pre-parsed lines with `OpAdd`/`OpDelete`/`OpContext` operations. **Neither library stores per-line numbers** — you must compute them yourself from hunk headers.

## File-centric parsing with hunk-centric display

All major parsers use file-centric structures (hunks nested under files), matching the Git diff format. Your flat `[]Hunk` proposal offers navigation advantages but denormalizes file metadata.

**Recommended pattern:** Parse file-centric, flatten to hunk-centric for display:

```go
// Parse using go-gitdiff (file-centric)
files, _, err := gitdiff.Parse(reader)

// Flatten for pager navigation
type DisplayHunk struct {
    File     *gitdiff.File      // Back-reference to parent
    Fragment *gitdiff.TextFragment
    FileIdx  int                // For "reviewed" tracking
    HunkIdx  int                // Stable identity
}

var hunks []DisplayHunk
for fi, f := range files {
    for hi, h := range f.TextFragments {
        hunks = append(hunks, DisplayHunk{f, h, fi, hi})
    }
}
```

This gives you flat iteration for scrolling and keyboard navigation while preserving file relationships for header rendering. The `(FileIdx, HunkIdx)` pair provides **stable hunk identity** for future "reviewed" state tracking.

## Line number computation requires manual implementation

Hunk headers only specify starting positions (`@@ -82,7 +82,9 @@`). Per-line numbers must be computed:

```go
func ComputeLineNumbers(fragment *gitdiff.TextFragment) []LineWithNums {
    oldNum := int(fragment.OldPosition)
    newNum := int(fragment.NewPosition)
    result := make([]LineWithNums, 0, len(fragment.Lines))
    
    for _, line := range fragment.Lines {
        switch line.Op {
        case gitdiff.OpContext:
            result = append(result, LineWithNums{oldNum, newNum, line})
            oldNum++
            newNum++
        case gitdiff.OpDelete:
            result = append(result, LineWithNums{oldNum, 0, line})
            oldNum++
        case gitdiff.OpAdd:
            result = append(result, LineWithNums{0, newNum, line})
            newNum++
        }
    }
    return result
}
```

**Gaps between hunks** (lines 15-49 when hunks cover 10-14 and 50-55) require visual separators. Line numbers restart from each hunk's header values — there's no continuity to compute.

## Bubble Tea architecture: viewport with lazy initialization

The standard `viewport.Model` handles thousands of lines efficiently by slicing visible lines during `View()`. For PR-sized diffs, **pre-render all content once** and let the viewport handle windowing.

```go
type Model struct {
    // Domain model
    diff          *ParsedDiff
    hunks         []DisplayHunk
    hunkPositions []int         // Line number where each hunk starts
    
    // View model
    viewport      viewport.Model
    ready         bool          // Two-phase initialization flag
    width, height int
    
    // Navigation state
    currentHunk   int
    
    // UI state
    mode          Mode          // Normal, Search
    searchInput   textinput.Model
    styles        Styles
}
```

**Critical pattern — lazy viewport initialization:**

```go
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        if !m.ready {
            // First init after terminal dimensions known
            m.viewport = viewport.New(msg.Width, msg.Height-2)
            m.viewport.SetContent(m.renderFullDiff())
            m.ready = true
        } else {
            m.viewport.Width = msg.Width
            m.viewport.Height = msg.Height - 2
        }
    }
    // Delegate scrolling to viewport
    var cmd tea.Cmd
    m.viewport, cmd = m.viewport.Update(msg)
    return m, cmd
}
```

**Graceful degradation for massive diffs (>10,000 lines):** Switch to hunk-at-a-time pagination mode where only the current hunk is rendered, and `n`/`N` load adjacent hunks.

## Hunk jump navigation requires position tracking

Store line positions when rendering to enable jump-to-hunk:

```go
func (m *Model) renderFullDiff() string {
    var buf strings.Builder
    lineNum := 0
    
    for i, h := range m.hunks {
        m.hunkPositions = append(m.hunkPositions, lineNum)
        
        rendered := m.renderHunk(h)
        buf.WriteString(rendered)
        lineNum += strings.Count(rendered, "\n")
    }
    return buf.String()
}

func (m *Model) jumpToHunk(idx int) {
    if idx >= 0 && idx < len(m.hunkPositions) {
        m.currentHunk = idx
        m.viewport.SetYOffset(m.hunkPositions[idx])
    }
}
```

## Keyboard handling: bubbles/key with manual multi-key state

Use the `bubbles/key` package for customizable keymaps with auto-generated help. Multi-key sequences like `gg` require manual state tracking.

```go
type KeyMap struct {
    Up           key.Binding
    Down         key.Binding
    HalfPageUp   key.Binding
    HalfPageDown key.Binding
    GotoTop      key.Binding  // "gg" - needs custom handling
    GotoBottom   key.Binding
    NextHunk     key.Binding
    PrevHunk     key.Binding
    Search       key.Binding
    Quit         key.Binding
}

func DefaultKeyMap() KeyMap {
    return KeyMap{
        Up:           key.NewBinding(key.WithKeys("k", "up"), key.WithHelp("k", "up")),
        Down:         key.NewBinding(key.WithKeys("j", "down"), key.WithHelp("j", "down")),
        HalfPageUp:   key.NewBinding(key.WithKeys("ctrl+u"), key.WithHelp("^u", "½ pg up")),
        HalfPageDown: key.NewBinding(key.WithKeys("ctrl+d"), key.WithHelp("^d", "½ pg down")),
        GotoBottom:   key.NewBinding(key.WithKeys("G"), key.WithHelp("G", "bottom")),
        NextHunk:     key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "next hunk")),
        PrevHunk:     key.NewBinding(key.WithKeys("N"), key.WithHelp("N", "prev hunk")),
        Search:       key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "search")),
        Quit:         key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
    }
}
```

**Multi-key sequence handling for "gg":**

```go
type Model struct {
    pendingKey string
    keyMap     KeyMap
}

func (m Model) handleNormalMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
    keyStr := msg.String()
    
    // Complete "gg" sequence
    if m.pendingKey == "g" && keyStr == "g" {
        m.pendingKey = ""
        m.viewport.SetYOffset(0)
        return m, nil
    }
    
    // Start potential sequence
    if keyStr == "g" {
        m.pendingKey = "g"
        return m, nil
    }
    
    // Clear pending on any other key
    m.pendingKey = ""
    
    switch {
    case key.Matches(msg, m.keyMap.Down):
        m.viewport.ScrollDown(1)
    case key.Matches(msg, m.keyMap.NextHunk):
        m.jumpToHunk(m.currentHunk + 1)
    case key.Matches(msg, m.keyMap.Search):
        m.mode = ModeSearch
        m.searchInput.Focus()
        return m, textinput.Blink
    }
    return m, nil
}
```

**Mode switching** for search uses a `Mode` enum field, routing key handling through separate functions for `ModeNormal` vs `ModeSearch`.

## Theme architecture: centralized styles with adaptive colors

Structure themes as data, compute styles once from theme. Use `lipgloss.AdaptiveColor` for automatic light/dark detection.

```go
// ColorPair enables light/dark mode switching
type ColorPair struct {
    Light lipgloss.TerminalColor
    Dark  lipgloss.TerminalColor
}

func (cp ColorPair) Resolve(isDark bool) lipgloss.TerminalColor {
    if isDark { return cp.Dark }
    return cp.Light
}

// Theme defines all themeable elements as data
type Theme struct {
    Name    string
    Added   ColorPair  // Background
    Removed ColorPair
    Context ColorPair
    AddedText   ColorPair  // Foreground
    RemovedText ColorPair
    LineNumber  ColorPair
    FileHeader  ColorPair
    HunkHeader  ColorPair
}

// Styles holds computed lipgloss.Style instances
type Styles struct {
    AddedLine    lipgloss.Style
    RemovedLine  lipgloss.Style
    ContextLine  lipgloss.Style
    LineNumber   lipgloss.Style
    FileHeader   lipgloss.Style
    HunkHeader   lipgloss.Style
}

func NewStyles(theme Theme, isDark bool) Styles {
    return Styles{
        AddedLine: lipgloss.NewStyle().
            Background(theme.Added.Resolve(isDark)).
            Foreground(theme.AddedText.Resolve(isDark)),
        RemovedLine: lipgloss.NewStyle().
            Background(theme.Removed.Resolve(isDark)).
            Foreground(theme.RemovedText.Resolve(isDark)),
        // ... other styles
    }
}
```

**Runtime theme detection** in Bubble Tea v2:

```go
func (m Model) Init() tea.Cmd {
    return tea.RequestBackgroundColor
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.BackgroundColorMsg:
        m.isDark = msg.IsDark()
        m.styles = NewStyles(m.theme, m.isDark)
    }
}
```

**Future syntax highlighting compatibility:** Use a layered approach — syntax colors apply to foreground, diff colors apply to background. The `chroma` library provides Go syntax highlighting with TextMate-compatible themes.

## Conclusion

The validated architecture combines **go-gitdiff** for parsing (best Git edge case coverage), file-centric parse with hunk-centric display transformation (preserving stable `(FileIdx, HunkIdx)` identity), standard **viewport.Model** for efficient scrolling (with hunk-pagination fallback), **bubbles/key** for customizable keymaps with manual multi-key state, and a **centralized Theme → Styles pattern** with `ColorPair` for trivial light/dark switching.

Key implementation priorities:
1. Add `IsBinary`, `FileOp`, `NoNewline`, `Section` to your data model
2. Compute line numbers manually during parsing
3. Store hunk line positions during rendering for jump navigation
4. Use two-phase viewport initialization (wait for `WindowSizeMsg`)
5. Track `pendingKey` state for multi-key sequences like "gg"
