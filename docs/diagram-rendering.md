# Go libraries for terminal diagram rendering in Bubble Tea

**No dedicated flowchart library exists for Go terminals**, but a combination of Lipgloss primitives and the ntcharts Canvas component can accomplish all four diagram types with clean results. The charmbracelet ecosystem provides the strongest foundation: Lipgloss's border system and composition functions handle linear flows and two-column layouts natively, while ntcharts fills the gap for complex hub-and-spoke diagrams requiring arbitrary character placement.

## Lipgloss handles most diagram needs directly

The official charmbracelet/lipgloss library (**9.3k stars**, actively maintained) provides everything needed for boxed nodes with connectors. Its border system includes seven pre-built styles—`RoundedBorder()` (╭╮╰╯), `DoubleBorder()` (═║), `ThickBorder()` (┏┓), and others—plus a `Border` struct for fully custom character sets. The composition functions `JoinHorizontal()` and `JoinVertical()` let you assemble diagrams from styled blocks:

```go
nodeStyle := lipgloss.NewStyle().
    Border(lipgloss.RoundedBorder()).
    BorderForeground(lipgloss.Color("63")).
    Padding(0, 2).
    Width(15).
    Align(lipgloss.Center)

problem := nodeStyle.Render("problem")
fix := nodeStyle.Render("fix")
test := nodeStyle.Render("test")
arrow := " ───→ "

flow := lipgloss.JoinHorizontal(lipgloss.Center, problem, arrow, fix, arrow, test)
```

Variable-width labels work through Lipgloss's `Width()` constraint combined with internal alignment. For two-column layouts, nest `JoinVertical` calls inside `JoinHorizontal`. Before/after diagrams follow the same pattern with a styled divider between columns.

## ntcharts Canvas enables complex diagram layouts

For hub-and-spoke diagrams requiring arbitrary character placement, **ntcharts** (github.com/NimbleMarkets/ntcharts, **566 stars**, MIT license, actively maintained) provides a Canvas component built specifically for Bubble Tea. The Canvas is a 2D grid accepting arbitrary runes with Lipgloss styling and BubbleZone mouse support:

```go
import "github.com/NimbleMarkets/ntcharts/canvas"

c := canvas.New(50, 20)
// Draw central hub
c.SetCell(25, 10, '●', hubStyle)
// Draw connecting lines to spoke nodes
c.SetCell(20, 10, '─', lineStyle)
c.SetCell(15, 10, '○', spokeStyle)
```

This canvas approach enables the spoke-connection geometry that pure composition can't achieve. The library also includes bar charts, line charts (braille-based), and heatmaps if the diffstory intro needs data visualization.

## Supporting libraries for robust implementation

**go-runewidth** (github.com/mattn/go-runewidth, **600 stars**) is essential for proper text alignment—it calculates display widths of Unicode strings and specifically handles box-drawing characters (U+2500-U+257F) as width-1 regardless of locale. All major table libraries depend on it.

**lipgloss/tree** (official subpackage) renders hierarchical structures if any diagram needs tree branches:
```go
t := tree.Root("Changes").Child("Added", "Modified", "Deleted")
```
Supports `DefaultEnumerator` (├── └──) and `RoundedEnumerator` with full Lipgloss styling.

**xlab/treeprint** (**418 stars**) offers an alternative tree API with metadata support via `AddMetaBranch()`.

## Table libraries provide reference implementations

**olekukonko/tablewriter** (**4.7k stars**, v1.1.2 December 2024) demonstrates robust Unicode border handling with cell merging—useful patterns for complex diagram nodes. Its colorized renderer integrates with fatih/color.

**jedib0t/go-pretty** (**2k stars**) includes a `text` subpackage with alignment utilities, padding, and word-wrapping that handle the text formatting inside diagram boxes.

## Building custom diagrams from primitives

For maximum flexibility, implement a minimal canvas buffer following patterns from **tcg** (github.com/msoap/tcg) or **tcell**:

```go
type Cell struct {
    Rune  rune
    Style lipgloss.Style
}

type Canvas struct {
    cells         [][]Cell
    width, height int
}

func (c *Canvas) DrawBox(x, y, w, h int, border lipgloss.Border) { ... }
func (c *Canvas) DrawHLine(x1, x2, y int, char rune) { ... }
func (c *Canvas) DrawArrow(x1, y1, x2, y2 int) { ... }
func (c *Canvas) Render() string { ... }
```

tcg demonstrates BitBlt-style compositing for overlaying sub-buffers—useful for positioning spoke nodes around a hub.

## Practical approach for each diagram type

| Diagram Type | Recommended Approach | Complexity |
|-------------|---------------------|------------|
| **Linear flow** | Lipgloss `JoinHorizontal` with arrow strings | Simple |
| **Two columns** | Lipgloss `JoinVertical` nested in `JoinHorizontal` | Simple |
| **Before/after** | Same as two columns with styled divider | Simple |
| **Hub-and-spoke** | ntcharts Canvas or custom buffer | Moderate |

For the linear flow and columnar layouts, Lipgloss alone suffices—the library's border characters and composition functions produce clean output with no additional dependencies. Hub-and-spoke diagrams require either ntcharts Canvas or a custom 150-line buffer implementation.

## Integration with Bubble Tea rendering

All recommended libraries integrate cleanly with Bubble Tea's architecture. Lipgloss-styled strings return from your model's `View()` method unchanged. The ntcharts Canvas implements the standard `tea.Model` interface with `Init()`, `Update()`, and `View()` methods, making it a drop-in component. For static intro slides without interactivity, render the diagram once in `View()` and return the string.

The key constraint is that Bubble Tea redraws the entire view on each update—but for static diagrams this has no performance impact. Use `lipgloss.Place()` to position diagrams within the terminal viewport if centering is needed.

## Libraries to avoid for this use case

**Graphviz wrappers** (gographviz, etc.) require the external dot binary and output images, not ASCII. **esimov/diagram** converts ASCII art to PNG—wrong direction. **blampe/goat** outputs SVG from ASCII sources, useful for documentation but not terminal display. **termui** (**13k stars**) has a Canvas widget but uses a different architecture than Bubble Tea with sporadic maintenance.

## Conclusion

Start with **Lipgloss alone** for the linear flow, two-column, and before/after diagrams—its border system and join functions produce publication-quality ASCII output with minimal code. Add **ntcharts Canvas** only if hub-and-spoke complexity demands arbitrary character placement. Keep **go-runewidth** as a dependency for any custom text-centering logic. This combination stays within the charmbracelet ecosystem, ensuring consistent styling and zero architectural conflicts with your existing Bubble Tea application.
