package bubbletea

import "github.com/charmbracelet/lipgloss"

// tabWidth is the standard terminal tab stop interval.
const tabWidth = 8

// DisplayWidth calculates the display width of a string, correctly handling
// tab characters which expand to the next 8-column boundary.
// This fixes the issue where lipgloss.Width returns 0 for tabs.
func DisplayWidth(s string) int {
	return displayWidthFrom(s, 0)
}

// displayWidthFrom calculates the display width of a string starting from
// a given column position. This is needed when calculating widths of
// multiple strings that will be concatenated, as tab expansion depends
// on the current column position.
func displayWidthFrom(s string, startCol int) int {
	col := startCol
	for _, r := range s {
		if r == '\t' {
			// Tab advances to next tab stop (multiple of tabWidth)
			col = ((col / tabWidth) + 1) * tabWidth
		} else {
			col += lipgloss.Width(string(r))
		}
	}
	return col
}
