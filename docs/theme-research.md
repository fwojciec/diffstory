# Unified TUI theming: patterns from delta, Neovim, and the Charm ecosystem

**The most effective approach for a Go TUI diff viewer is a two-layer architecture that separates syntax highlighting (foreground) from application styling (background), with explicit superimposition at render time.** Delta's proven `superimpose_style_sections` pattern demonstrates this works elegantly—syntax colors from Chroma can overlay diff backgrounds from Lipgloss without runtime color adjustment, provided theme designers choose compatible combinations with sufficient contrast. For the Go stack specifically, a shared color palette struct should generate both Lipgloss styles and Chroma styles, with `CompleteAdaptiveColor` handling graceful degradation across terminal capabilities.

---

## Structural patterns: compose layers, don't merge structs

The clearest architectural insight comes from delta's design: **treat syntax highlighting and application styling as independent layers merged at render time**, rather than forcing them into a single configuration structure.

Delta's configuration separates these concerns explicitly:
- `syntax-theme` selects foreground colors from syntect/bat themes (e.g., "Nord", "Monokai Extended")
- `*-style` options control diff-specific styling: `plus-style = syntax "#001a00"` means "use syntax highlighting colors on a dark green background"

The `syntax` keyword is the key innovation—it tells the renderer to take foreground colors from the syntax highlighting pass while applying the diff background. This superimposition happens in delta's `paint.rs` via `superimpose_style_sections`, which processes two arrays of `(style, substring)` pairs and produces a merged output.

**For Go/Lipgloss/Chroma, the recommended structure:**

```go
type Theme struct {
    // Layer 1: Base palette (shared between layers)
    Palette ColorPalette
    
    // Layer 2: Application UI styles (Lipgloss)
    UI UIStyles
    
    // Layer 3: Syntax highlighting (Chroma)
    Syntax *chroma.Style
    
    // Layer 4: Diff-specific overrides
    Diff DiffStyles
}

type DiffStyles struct {
    AddedBackground    lipgloss.TerminalColor
    DeletedBackground  lipgloss.TerminalColor
    ModifiedBackground lipgloss.TerminalColor
    UseSyntaxOnAdded   bool  // Delta's "syntax" keyword equivalent
    UseSyntaxOnDeleted bool
}
```

Neovim's highlight group linking provides another extensibility pattern: groups can inherit from other groups via `{link = 'TargetGroup'}`, creating cascading fallbacks. Helix uses theme-level inheritance with `inherits = "parent_theme"` in TOML. Both approaches avoid duplication while enabling selective overrides.

---

## Semantic naming conventions that scale

The TextMate scope naming convention, used by syntect/Chroma and adapted by Helix, establishes the dominant pattern: **hierarchical dot-separated names with longest-prefix matching for fallbacks**.

**Syntax token naming (from TextMate/Chroma):**
```
keyword.control.conditional    → keyword.control → keyword
function.builtin.static        → function.builtin → function
string.quoted.double          → string.quoted → string
```

**Chroma's TokenType hierarchy** uses numeric ranges (keywords 1000-1999, names 2000-2999, literals 3000-3999) enabling automatic inheritance when specific tokens aren't styled.

**UI element naming (from Helix):**
```
ui.background
ui.statusline.normal  / ui.statusline.insert / ui.statusline.inactive
ui.cursor.primary     / ui.cursor.match
ui.selection.primary
ui.virtual.inlay-hint.type
```

**Diff-specific naming (from Helix/delta):**
```
diff.plus         / diff.plus.gutter
diff.minus        / diff.minus.gutter  
diff.delta        / diff.delta.moved
```

The critical pattern is **namespace separation**: syntax tokens have no prefix, UI elements use `ui.*`, diagnostics use `diagnostic.*`, and diff elements use `diff.*` or their own category. This prevents collision when extending themes with new components.

**Recommended Go struct organization:**

```go
type SemanticColors struct {
    // Syntax (maps to Chroma TokenTypes)
    Keyword     Color
    String      Color
    Comment     Color
    Function    Color
    Type        Color
    Constant    Color
    
    // UI chrome
    UIBackground    Color
    UIForeground    Color
    UIBorder        Color
    UIStatusline    Color
    
    // Diff-specific
    DiffAdded       Color
    DiffDeleted     Color
    DiffModified    Color
    DiffContext     Color
}
```

---

## Syntax highlighting on diff backgrounds: delta's solution

Delta's approach to the color-on-color problem is pragmatic: **no automatic contrast adjustment—users select compatible theme/background combinations**, with documentation encouraging subtle backgrounds.

**The technical mechanism:**
1. First pass: syntect computes syntax highlighting for each line, producing `[(style, substring)]` pairs with foreground colors
2. Second pass: delta computes diff styling with background colors from `*-style` options
3. Third pass: `superimpose_style_sections` merges these, using diff backgrounds with syntax foregrounds when the `syntax` keyword is present

**The design philosophy (from delta's issues/discussions):**
- Diff backgrounds should be **very subtle tints** (e.g., `#001a00` is nearly black with a slight green cast)
- Dark themes use dark diff backgrounds; light themes use light tints
- Theme authors bear responsibility for choosing readable combinations
- The `dark = true` / `light = true` metadata helps select appropriate defaults

**Practical contrast guidelines:**
- WCAG minimum: **4.5:1** contrast ratio for normal text
- For diff backgrounds, aim for luminance difference <15% from base background
- Test all syntax colors against both normal AND diff backgrounds
- Consider providing a "high contrast" mode with more dramatic background shifts

**Implementation pattern for Go:**

```go
func RenderDiffLine(line string, diffType DiffType, theme *Theme) string {
    // Layer 1: Get syntax highlighting spans from Chroma
    syntaxSpans := highlightWithChroma(line, theme.Syntax)
    
    // Layer 2: Determine diff background
    var bg lipgloss.TerminalColor
    switch diffType {
    case Added:
        bg = theme.Diff.AddedBackground
    case Deleted:
        bg = theme.Diff.DeletedBackground
    }
    
    // Layer 3: Superimpose - apply background while preserving foreground
    var result strings.Builder
    for _, span := range syntaxSpans {
        style := lipgloss.NewStyle().
            Foreground(chromaColorToLipgloss(span.Color)).
            Background(bg)
        result.WriteString(style.Render(span.Text))
    }
    return result.String()
}
```

---

## Graceful degradation across terminal color capabilities

The Lipgloss/termenv stack handles degradation automatically, but Chroma requires explicit formatter selection. **The key is detecting capability once at startup and coordinating both systems.**

**Lipgloss color types for explicit fallbacks:**

```go
// Automatic degradation (termenv picks closest match)
lipgloss.Color("#d3869b")

// Explicit fallback chain
lipgloss.CompleteColor{
    TrueColor: "#d3869b",
    ANSI256:   "175",
    ANSI:      "5",  // magenta
}

// Light/dark adaptation + explicit fallbacks
lipgloss.CompleteAdaptiveColor{
    Light: CompleteColor{TrueColor: "#6c3461", ANSI256: "96", ANSI: "5"},
    Dark:  CompleteColor{TrueColor: "#d3869b", ANSI256: "175", ANSI: "13"},
}
```

**Chroma formatter selection:**

```go
profile := colorprofile.Detect(os.Stdout, os.Environ())

var formatter chroma.Formatter
switch profile {
case colorprofile.TrueColor:
    formatter = formatters.TTY16m
case colorprofile.ANSI256:
    formatter = formatters.TTY256
case colorprofile.ANSI:
    formatter = formatters.TTY16
default:
    formatter = formatters.TTY8
}
```

**The base16 philosophy** offers an alternative: design for 16 ANSI colors only, relying on users to configure their terminal palette. This ensures perfect harmony with terminal theming but limits color precision.

**Recommended hybrid approach:**
- Define themes with TrueColor values as the source of truth
- Auto-generate 256-color and 16-color approximations
- For 16-color mode, map to semantic ANSI names (`red`, `green`, `magenta`) rather than fixed colors
- Never mix terminal-controlled colors (0-15) with fixed palette colors (16-255) in the same theme—this causes broken rendering when terminal palettes differ

---

## Extensibility patterns for new languages and UI components

The most robust extensibility mechanism is **hierarchical fallback with explicit linking**, as implemented across Neovim, Helix, and TextMate.

**Fallback chain pattern:**
```
specific → category → base → default
@function.builtin.lua → @function.builtin → @function → Function → Normal
```

**Neovim's linking mechanism:**
```lua
vim.api.nvim_set_hl(0, "@function", {link = "Function"})
vim.api.nvim_set_hl(0, "@function.builtin", {link = "@function"})
```

New Tree-sitter parsers automatically work if they use standard capture names. The colorscheme only needs to define base groups; everything else inherits.

**Helix's scope resolution** uses longest-prefix matching algorithmically—no explicit links required:
```toml
# Defining just "function" covers all function.* variants
"function" = { fg = "cyan" }
# Override specific subtype if needed
"function.builtin" = { fg = "cyan", modifiers = ["bold"] }
```

**For Go/Chroma extensibility:**

```go
type StyleResolver struct {
    entries map[chroma.TokenType]chroma.StyleEntry
}

func (r *StyleResolver) Get(tokenType chroma.TokenType) chroma.StyleEntry {
    // Exact match
    if entry, ok := r.entries[tokenType]; ok {
        return entry
    }
    // Parent fallback (Chroma token types are hierarchical by numeric range)
    parent := tokenType.Parent()
    if parent != chroma.None {
        return r.Get(parent)
    }
    // Default
    return r.entries[chroma.Text]
}
```

**UI extensibility pattern:**
```go
type UIStyleProvider interface {
    Style(component string) lipgloss.Style
}

type HierarchicalStyles struct {
    styles map[string]lipgloss.Style
}

func (h *HierarchicalStyles) Style(component string) lipgloss.Style {
    // Try exact match: "statusline.insert"
    if style, ok := h.styles[component]; ok {
        return style
    }
    // Try parent: "statusline"
    if dot := strings.LastIndex(component, "."); dot > 0 {
        return h.Style(component[:dot])
    }
    // Default
    return h.styles["default"]
}
```

---

## Architecture comparison across prior art

| System | Config Format | Syntax/UI Separation | Fallback Mechanism | Color Degradation |
|--------|--------------|---------------------|-------------------|-------------------|
| **Delta** | gitconfig | Explicit (`syntax-theme` vs `*-style`) | None (user chooses themes) | Via bat/syntect |
| **Neovim** | Lua | By group category | Explicit `link =` | `guifg`/`ctermfg` pairs |
| **Helix** | TOML | Namespace prefix (`ui.*`) | Longest-prefix match | Terminal's 16-color palette |
| **bat** | .tmTheme | Inherits syntect | TextMate scopes | `ansi`/`base16` themes |
| **Alacritty** | TOML | N/A (terminal only) | N/A | Always TrueColor |
| **Chroma** | Go structs/XML | N/A (syntax only) | TokenType hierarchy | Formatter selection |
| **Lipgloss** | Go structs | N/A (UI only) | Style inheritance | `CompleteColor` type |

**Key architectural takeaways:**
- Delta proves the superimposition pattern works at scale for diff+syntax
- Helix's `ui.*` namespace prefix is the cleanest separation pattern
- Neovim's linking creates the most flexible fallback chains
- Lipgloss's `CompleteAdaptiveColor` is the most explicit degradation handling
- Chroma's TokenType hierarchy enables natural language extensibility

---

## Recommended implementation for Go diff viewer

```go
// Core theme structure following delta's layer separation
type DiffViewerTheme struct {
    Name    string
    Dark    bool
    Palette Palette          // Shared base colors
    Syntax  *chroma.Style    // Generated from Palette
    UI      UIStyles         // Generated from Palette
    Diff    DiffStyles       // Overlay configuration
}

// Palette defines the 10-15 base colors everything derives from
type Palette struct {
    Background CompleteAdaptiveColor
    Foreground CompleteAdaptiveColor
    // Syntax semantic colors
    Keyword   CompleteAdaptiveColor
    String    CompleteAdaptiveColor
    Comment   CompleteAdaptiveColor
    Function  CompleteAdaptiveColor
    Type      CompleteAdaptiveColor
    Constant  CompleteAdaptiveColor
    // Diff semantic colors
    Added     CompleteAdaptiveColor
    Deleted   CompleteAdaptiveColor
    Modified  CompleteAdaptiveColor
    // UI accents
    Selection CompleteAdaptiveColor
    Border    CompleteAdaptiveColor
}

// Generate Chroma style from palette
func (p Palette) ChromaStyle() *chroma.Style {
    return chroma.MustNewStyleBuilder("custom").
        Add(chroma.Background, p.Background.String()).
        Add(chroma.Keyword, p.Keyword.String()).
        Add(chroma.String, p.String.String()).
        Add(chroma.Comment, p.Comment.String()).
        Add(chroma.NameFunction, p.Function.String()).
        Add(chroma.KeywordType, p.Type.String()).
        Add(chroma.LiteralNumber, p.Constant.String()).
        Build()
}
```

This architecture provides: clear layer separation (delta pattern), semantic palette sharing (Helix pattern), automatic fallback via Chroma's TokenType hierarchy (Neovim pattern), and explicit degradation via CompleteAdaptiveColor (Lipgloss best practice).
