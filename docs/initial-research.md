# Diff Viewer Design for AI-Assisted Coding Sessions: Problem Space Map

A terminal diff viewer for Claude Code review faces a fundamentally different challenge than traditional git pagers: **understanding AI intent across scattered changes matters more than line-by-line accuracy**. This research synthesizes prior art, cognitive research, and AI-specific review challenges into actionable design decisions for a Go-based viewer optimized for macOS Ghostty.

## Core problems a diff viewer solves

The problem taxonomy breaks into four interconnected domains, each with distinct priorities for AI code review:

**Readability** addresses the visual parsing problem—how quickly a developer can identify what changed. Traditional diffs fail here because they treat code as text, not structure. Word-level highlighting (delta's Levenshtein approach), syntax-aware coloring, and clean typography dramatically reduce the cognitive load of scanning changes. For AI review, readability must extend beyond individual lines to convey **change significance**—formatting noise versus logic modifications.

**Navigation** determines how efficiently developers move through multi-file diffs. Human commits rarely touch more than 10 files; AI agents routinely modify 50-100+. Standard hunk-jumping (n/N keys) becomes insufficient when changes scatter across a codebase. The missing pattern in current tools: **semantic navigation**—jumping between logically related changes regardless of file location.

**Context** provides the surrounding code necessary for comprehension. The default 3 lines of context works for focused human commits but fails for AI refactors where understanding requires seeing the function signature, class definition, or import statements. Variable context depth and the ability to see "the code as it will exist" (not just the diff) addresses this gap.

**Comprehension** is the synthesis layer: understanding not just what changed but why. For human code, commit messages provide intent. For AI code, the original prompt, the agent's reasoning, and the task context become essential metadata. No current diff viewer surfaces this information.

## Pager mode versus interactive TUI requirements

The distinction profoundly affects architecture. A pager receives stdin, displays content, and exits—like `less` with syntax highlighting. An interactive TUI maintains persistent state, supports editing or staging, and manages complex widget hierarchies.

For a read-only viewer focused on feedback, **pager architecture wins**. Key differences:

| Requirement | Pager Mode | Interactive TUI |
|-------------|------------|-----------------|
| stdin piping | Essential (`git diff \| viewer`) | Optional |
| Session state | Minimal—scroll position only | Complex—selections, edits |
| Memory model | Stream-process, forget | Maintain full document |
| Rendering | Framerate-throttled scrolling | Event-driven updates |
| Exit behavior | Returns to shell | May persist or spawn |

Delta succeeds by embracing pager philosophy: pipe styled output to a child `less` process, let `less` handle navigation. This avoids reimplementing scrollback buffers and keyboard handling. The limitation: no custom navigation beyond regex search.

A Bubble Tea viewport approach offers a middle ground—custom key handling with pager-like simplicity. The `viewport.Model` component was designed specifically for this use case, rendering only visible lines while supporting both keyboard and mouse scrolling.

## What makes AI-generated code review different

AI agents produce fundamentally different change patterns than human developers. Research from ArXiv shows AI-generated code has **distinct vulnerability distributions** from human code, and Microsoft's internal deployment found that **routine issues** suit AI review while **architectural concerns** require human judgment.

**Scale of changes**: AI agents regularly produce refactors spanning 100+ files versus focused human commits touching 5-10. Claude Code's codebase-wide modifications and Cursor's Composer mode create cross-cutting changes that overwhelm traditional file-by-file review.

**Scattered versus concentrated**: Human commits tend toward concentrated changes—one feature, one module. AI agents often make many small, scattered modifications—renaming a variable across 40 files, updating import statements everywhere, reformatting code to match style guidelines. The viewer must convey "these 40 changes are semantically one rename" versus "these 5 changes are 5 distinct modifications."

**Intent verification gap**: Human commits carry implicit intent through commit messages and PR descriptions. AI code "looks right" but may miss architectural context, business logic edge cases, or project conventions. The critical question shifts from "is this syntactically correct?" to "does this solve the right problem?"

**Feedback loop optimization**: The viewer serves a specific workflow—human reviews diff, provides feedback to AI agent, agent iterates. This requires easy copying of file:line references with surrounding context, structured feedback formats, and potentially integration with the agent's explanation of its changes.

## Go TUI framework decision: Bubble Tea

**Bubble Tea is the clear choice** for a diff viewer/pager. The framework comparison reveals several decisive factors:

**Architecture fit**: Bubble Tea's Elm-style Model→Update→View pattern maps perfectly to a pager's simple state machine (content + scroll position + mode). tview's widget hierarchy adds complexity without benefit for a read-only viewer.

**Stdin support**: Bubble Tea supports both inline and fullscreen modes, enabling `git diff | mydiffviewer` patterns. tview is fullscreen-only, problematic for pager use cases. This is described by the framework authors as "the biggest difference."

**Viewport component**: The `bubbles/viewport` component is purpose-built for pagers—handles PageUp/Down, vim keys, mouse wheel, and only renders visible lines. This provides the scrolling behavior needed without reimplementing basic pager functionality.

**Performance**: Bubble Tea's framerate-based renderer (throttled to ~60fps) handles large diffs efficiently. tview's documentation explicitly warns about performance degradation for large content, recommending `SetMaxLines()` limits.

**Rendering flexibility**: View() returns a string with complete rendering control. Lipgloss provides CSS-like styling (borders, padding, alignment). Chroma's ANSI output integrates directly. tview's widget constraints would complicate custom diff rendering.

**Proven pattern**: diffnav (dlvhdr/diffnav) demonstrates exactly this use case—a Bubble Tea diff pager wrapping delta with file tree navigation. The architecture works.

**Terminal features**: Modern Ghostty features like true color, mouse tracking, and OSC 52 clipboard are supported through termenv (Bubble Tea's terminal abstraction layer). Adaptive colors handle light/dark themes automatically.

## Diff representation trade-offs

### Unified versus side-by-side
Neither format universally wins. Matklad's analysis identifies the core limitation: both show diffs in isolation rather than **code in context with changes highlighted**. His ideal: "On the left, the current state of the code with changes subtly highlighted. On the right, the unified diff for the portion currently visible."

**Side-by-side works when**: terminal width exceeds 160 characters (80 per side minimum), changes span many lines, comparing structural flow matters. git-split-diffs automatically falls back to unified below this threshold.

**Unified works when**: terminal is narrow, scanning many hunks quickly, changes are interleaved (old line immediately followed by new), preserving vertical density matters.

**Design decision**: Support both modes with automatic selection based on terminal width, defaulting to side-by-side above 160 chars. Consider Ghostty's common wide terminal usage.

### Word-level and character-level highlighting
This is non-negotiable for readability. Showing which characters within a line changed (versus highlighting entire lines) dramatically reduces cognitive load.

**Delta's approach**: Uses Levenshtein edit distance to compute within-line changes. The `--max-line-distance` parameter (default 0.6) controls when lines are considered homologous enough to compare. Configurable word-diff regex for language-appropriate tokenization.

**Git's built-in `--word-diff`**: Simpler but less sophisticated. Words delimited by whitespace, limited prose handling.

**Implementation complexity**: Levenshtein on line pairs is O(n×m) per line pair—fast enough for typical hunks but potentially slow for pathologically long lines.

### Structural/semantic diffs
Difftastic proves that AST-aware diffing filters formatting noise and shows semantically meaningful changes. However, **complexity and performance costs are significant**:

- Tree-sitter grammar required per language
- Dijkstra-based tree alignment is compute-intensive  
- "Scales relatively poorly on files with large number of changes"
- Falls back to line-based diff on parse errors

**Practical middle ground**: Whitespace-insensitive line diff with word-level highlighting captures 80% of the benefit. Consider AST-based diffing as an optional enhancement for specific languages, not the default.

### Syntax highlighting integration
Delta uses syntect (Sublime Text grammars) for highlighting. Key challenge: **diff colors (red/green for removed/added) must coexist with syntax colors (blue for keywords, etc.)**.

Delta's solution: separate style layers—line background color for diff status, syntax colors overlaid on top, emphasis styling for within-line changes. Themes must ensure contrast remains readable across combinations.

**For Ghostty's true color support**: Full 24-bit color enables richer palettes than terminal256 constraints required historically.

## Prior art: what to adopt, what to avoid

### Delta: the current gold standard
**Adopt**: Levenshtein word-level diffing, syntax highlighting via syntect/bat, removal of +/- markers for clean copy/paste, n/N navigation between hunks, side-by-side mode with line numbers, configurable style strings using git color syntax, hyperlinks for commit hashes.

**Avoid**: Over-configuration complexity (>20 stylable elements creates analysis paralysis). Delta's approach of spawning a child `less` process limits custom navigation—a native Bubble Tea viewport could offer more control.

### diff-so-fancy: simplicity matters
**Adopt**: Clean file headers with Unicode line-drawing, simplified hunk headers, minimal configuration surface. Proves that readability improvements don't require syntax highlighting—sometimes less is more.

**Key insight**: The tool's popularity despite lacking delta's features suggests many developers prioritize simplicity over power.

### Difftastic: structural diffing future
**Adopt conceptually**: The insight that reformatting, indentation, and line wrapping shouldn't appear as changes. Understanding that "this function moved" is more useful than "40 lines deleted, 40 lines added elsewhere."

**Avoid**: Building AST parsing as core dependency. Performance issues with large changesets. Side-by-side output that can confuse. Better as optional integration than core feature.

### diffnav: the Go precedent
**Study closely**: Demonstrates Bubble Tea + delta integration for file tree navigation. Proves the pattern works. Uses delta for actual diff rendering, adds navigation layer on top.

**Consider**: Whether to shell out to delta (simpler) versus implement rendering directly (more control but more work).

## Copy/selection: a UX-critical detail

Terminal diff viewers struggle with copy/paste. Selecting code grabs line numbers, +/- markers, or content from both columns in side-by-side view. This friction matters significantly for the AI feedback workflow.

**OSC 52 clipboard escapes**: Write directly to system clipboard via ANSI escape sequence. Works over SSH, in tmux, supported by Ghostty. Implementation: `\033]52;c;$(base64 text)\007`. Enables "copy this hunk" commands without mouse selection.

**Design principles for clean copying**:
- Render line numbers in gutter area outside selectable region
- Strip diff markers from copied content (delta's default behavior)
- In side-by-side mode, only activate selection in the focused column
- Provide explicit "copy hunk," "copy file path," "copy file:line reference" commands

The feedback loop use case benefits from structured copy: "Copy this change as file:line with 3 lines of context" formats output for pasting directly into AI agent conversation.

## Design decisions for AI-specific challenges

### Intent layer
No current diff viewer surfaces AI reasoning. Design should include:
- Display original prompt/instruction that generated changes
- Show commit message prominently (AI agents typically generate detailed messages)
- Link to planning documents or task context when available
- Panel or overlay for "why this changed" alongside "what changed"

### Attention prioritization  
AI produces changes of varying significance. Help humans focus:
- Auto-collapse trivial changes (import reordering, formatting)
- Heatmap or annotation for high-significance modifications
- Semantic classification: "logic change" versus "style change" versus "refactoring move"
- Progressive disclosure—overview first, details on demand

### Multi-file coherence
When AI renames a function, show all call sites as one logical change:
- Dependency visualization between changed files
- Symbol tracking across files (function definition → usages)
- Grouped navigation by logical change unit, not file order
- Cross-file search: "show all places this symbol appears"

### Feedback optimization
The viewer serves an iterative workflow:
- One-click copy of file:line + context in AI-friendly format
- Structured comment templates: "[file:line] Issue: X, Suggestion: Y"
- Quick actions: "Accept," "Reject with reason," "Request specific change"
- Track which hunks have been reviewed versus pending

## Key architectural recommendations

**Framework**: Bubble Tea with viewport component for core scrolling, Lipgloss for styling, Chroma for syntax highlighting. This stack provides Ghostty-optimized terminal features, efficient large-diff handling, and proven pager patterns.

**Diff computation**: Consider go-diff (port of Google's diff-match-patch) for core algorithm, with Levenshtein-based word-level highlighting layered on top. Fall back gracefully for binary or huge files.

**Rendering pipeline**: Separate concerns—parse unified diff → compute word-level changes → apply syntax highlighting → merge diff colors with syntax colors → render to terminal. This mirrors delta's architecture.

**Navigation model**: Hunk-jumping (n/N), file-jumping (]/[), search (/), and optionally semantic navigation for logically grouped changes. Keyboard shortcuts should follow vim conventions where applicable.

**Configuration philosophy**: Opinionated defaults that work well for the AI review use case. Limited configuration surface—avoid delta's 50+ options. Ship with Ghostty-optimized color themes.

**Extension point**: Consider shelling out to delta for initial implementation, adding custom navigation and AI-specific features on top. diffnav proves this pattern works. Full custom rendering is a future optimization.

The unique opportunity: build the first diff viewer designed specifically for AI-agent code review workflows—not just a prettier `git diff`, but a tool optimized for understanding scattered, intent-opaque, large-scale AI modifications and providing structured feedback for iteration.
