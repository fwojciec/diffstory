package bubbletea_test

import (
	"bytes"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"
	diffview "github.com/fwojciec/diffstory"
	"github.com/fwojciec/diffstory/bubbletea"
	dv "github.com/fwojciec/diffstory/lipgloss"
	"github.com/stretchr/testify/assert"
)

func TestModel_RendersFileHeaders(t *testing.T) {
	t.Parallel()

	diff := &diffview.Diff{
		Files: []diffview.FileDiff{
			{
				OldPath:   "a/test.go",
				NewPath:   "b/test.go",
				Operation: diffview.FileModified,
				Hunks: []diffview.Hunk{
					{
						OldStart: 1,
						OldCount: 1,
						NewStart: 1,
						NewCount: 1,
						Lines: []diffview.Line{
							{Type: diffview.LineContext, Content: "context line"},
						},
					},
				},
			},
		},
	}

	m := bubbletea.NewModel(diff)
	tm := teatest.NewTestModel(t, m,
		teatest.WithInitialTermSize(80, 24),
	)

	// Should render enhanced file header with box-drawing chars
	teatest.WaitFor(t, tm.Output(), func(out []byte) bool {
		return bytes.Contains(out, []byte("── ")) &&
			bytes.Contains(out, []byte("test.go"))
	})

	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	tm.WaitFinished(t, teatest.WithFinalTimeout(0))
}

func TestModel_RendersHunkHeaders(t *testing.T) {
	t.Parallel()

	diff := &diffview.Diff{
		Files: []diffview.FileDiff{
			{
				OldPath:   "a/test.go",
				NewPath:   "b/test.go",
				Operation: diffview.FileModified,
				Hunks: []diffview.Hunk{
					{
						OldStart: 10,
						OldCount: 3,
						NewStart: 10,
						NewCount: 5,
						Section:  "func Example",
						Lines: []diffview.Line{
							{Type: diffview.LineContext, Content: "context line"},
						},
					},
				},
			},
		},
	}

	m := bubbletea.NewModel(diff)
	tm := teatest.NewTestModel(t, m,
		teatest.WithInitialTermSize(80, 24),
	)

	// Should render hunk header with @@ markers
	teatest.WaitFor(t, tm.Output(), func(out []byte) bool {
		return bytes.Contains(out, []byte("@@ -10,3 +10,5 @@ func Example"))
	})

	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	tm.WaitFinished(t, teatest.WithFinalTimeout(0))
}

func TestModel_RendersLinePrefixes(t *testing.T) {
	t.Parallel()

	diff := &diffview.Diff{
		Files: []diffview.FileDiff{
			{
				OldPath:   "a/test.go",
				NewPath:   "b/test.go",
				Operation: diffview.FileModified,
				Hunks: []diffview.Hunk{
					{
						OldStart: 1,
						OldCount: 2,
						NewStart: 1,
						NewCount: 2,
						Lines: []diffview.Line{
							{Type: diffview.LineContext, Content: "unchanged"},
							{Type: diffview.LineDeleted, Content: "removed"},
							{Type: diffview.LineAdded, Content: "added"},
						},
					},
				},
			},
		},
	}

	m := bubbletea.NewModel(diff)
	tm := teatest.NewTestModel(t, m,
		teatest.WithInitialTermSize(80, 24),
	)

	// Should render lines with prefixes
	teatest.WaitFor(t, tm.Output(), func(out []byte) bool {
		hasContext := bytes.Contains(out, []byte(" unchanged"))
		hasDeleted := bytes.Contains(out, []byte("-removed"))
		hasAdded := bytes.Contains(out, []byte("+added"))
		return hasContext && hasDeleted && hasAdded
	})

	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	tm.WaitFinished(t, teatest.WithFinalTimeout(0))
}

func TestModel_AppliesColors(t *testing.T) {
	t.Parallel()

	// Diff with all line types for comprehensive color testing
	diff := &diffview.Diff{
		Files: []diffview.FileDiff{
			{
				OldPath:   "a/test.go",
				NewPath:   "b/test.go",
				Operation: diffview.FileModified,
				Hunks: []diffview.Hunk{
					{
						OldStart: 1,
						OldCount: 2,
						NewStart: 1,
						NewCount: 2,
						Lines: []diffview.Line{
							{Type: diffview.LineContext, Content: "context"},
							{Type: diffview.LineDeleted, Content: "deleted"},
							{Type: diffview.LineAdded, Content: "added"},
						},
					},
				},
			},
		},
	}

	m := bubbletea.NewModel(diff, bubbletea.WithRenderer(trueColorRenderer()))
	tm := teatest.NewTestModel(t, m,
		teatest.WithInitialTermSize(80, 24),
	)

	// Wait for output with both foreground and background colors
	// True color uses 38;2;R;G;B for foreground, 48;2;R;G;B for background
	teatest.WaitFor(t, tm.Output(), func(out []byte) bool {
		hasForegroundColor := bytes.Contains(out, []byte("38;2;"))
		hasBackgroundColor := bytes.Contains(out, []byte("48;2;"))
		hasAddedLine := bytes.Contains(out, []byte("+added"))
		hasDeletedLine := bytes.Contains(out, []byte("-deleted"))
		return hasForegroundColor && hasBackgroundColor && hasAddedLine && hasDeletedLine
	})

	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	tm.WaitFinished(t, teatest.WithFinalTimeout(0))
}

func TestModel_BackgroundExtendsFullWidth(t *testing.T) {
	t.Parallel()

	diff := &diffview.Diff{
		Files: []diffview.FileDiff{
			{
				OldPath:   "a/test.go",
				NewPath:   "b/test.go",
				Operation: diffview.FileModified,
				Hunks: []diffview.Hunk{
					{
						OldStart: 1,
						OldCount: 1,
						NewStart: 1,
						NewCount: 2,
						Lines: []diffview.Line{
							{Type: diffview.LineAdded, Content: "short"},
						},
					},
				},
			},
		},
	}

	m := bubbletea.NewModel(diff, bubbletea.WithRenderer(trueColorRenderer()))
	tm := teatest.NewTestModel(t, m,
		teatest.WithInitialTermSize(80, 24),
	)

	// Background should extend beyond just the text "+short"
	// The styled content should include padding spaces within the style
	teatest.WaitFor(t, tm.Output(), func(out []byte) bool {
		hasAddedLine := bytes.Contains(out, []byte("+short"))
		// Check for padding spaces within styled region (spaces before reset code)
		// Pattern: spaces followed by ESC[0m (reset)
		hasStyledPadding := bytes.Contains(out, []byte("   \x1b[0m")) ||
			bytes.Contains(out, []byte("  \x1b[0m"))
		return hasAddedLine && hasStyledPadding
	})

	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	tm.WaitFinished(t, teatest.WithFinalTimeout(0))
}

func TestModel_BackgroundExtendsFullWidthWithUnicode(t *testing.T) {
	t.Parallel()

	// Test with multi-byte Unicode characters to ensure padding uses display width
	// "日本語" is 3 characters, 9 bytes, but 6 display cells (CJK are double-width)
	diff := &diffview.Diff{
		Files: []diffview.FileDiff{
			{
				OldPath:   "a/test.go",
				NewPath:   "b/test.go",
				Operation: diffview.FileModified,
				Hunks: []diffview.Hunk{
					{
						OldStart: 1,
						OldCount: 1,
						NewStart: 1,
						NewCount: 2,
						Lines: []diffview.Line{
							{Type: diffview.LineAdded, Content: "日本語"},
						},
					},
				},
			},
		},
	}

	m := bubbletea.NewModel(diff, bubbletea.WithRenderer(trueColorRenderer()))
	tm := teatest.NewTestModel(t, m,
		teatest.WithInitialTermSize(80, 24),
	)

	// Background should extend full width even with Unicode content
	// The line "+日本語" should be padded with spaces within the styled region
	teatest.WaitFor(t, tm.Output(), func(out []byte) bool {
		hasUnicodeLine := bytes.Contains(out, []byte("+日本語"))
		// Check for padding spaces within styled region (spaces before reset code)
		hasStyledPadding := bytes.Contains(out, []byte("   \x1b[0m")) ||
			bytes.Contains(out, []byte("  \x1b[0m"))
		return hasUnicodeLine && hasStyledPadding
	})

	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	tm.WaitFinished(t, teatest.WithFinalTimeout(0))
}

func TestModel_StatusBarShowsFilePosition(t *testing.T) {
	t.Parallel()

	// Create diff with 3 files
	diff := &diffview.Diff{
		Files: []diffview.FileDiff{
			{
				OldPath: "a/first.go",
				NewPath: "b/first.go",
				Hunks: []diffview.Hunk{
					{Lines: []diffview.Line{{Type: diffview.LineContext, Content: "first file"}}},
				},
			},
			{
				OldPath: "a/second.go",
				NewPath: "b/second.go",
				Hunks: []diffview.Hunk{
					{Lines: []diffview.Line{{Type: diffview.LineContext, Content: "second file"}}},
				},
			},
			{
				OldPath: "a/third.go",
				NewPath: "b/third.go",
				Hunks: []diffview.Hunk{
					{Lines: []diffview.Line{{Type: diffview.LineContext, Content: "third file"}}},
				},
			},
		},
	}

	m := bubbletea.NewModel(diff)
	tm := teatest.NewTestModel(t, m,
		teatest.WithInitialTermSize(80, 24),
	)

	// Status bar should show file 1/3 when at top
	teatest.WaitFor(t, tm.Output(), func(out []byte) bool {
		return bytes.Contains(out, []byte("file 1/3"))
	})

	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	tm.WaitFinished(t, teatest.WithFinalTimeout(0))
}

func TestModel_StatusBarShowsHunkPosition(t *testing.T) {
	t.Parallel()

	// Create diff with one file containing 3 hunks
	diff := &diffview.Diff{
		Files: []diffview.FileDiff{
			{
				OldPath: "a/file.go",
				NewPath: "b/file.go",
				Hunks: []diffview.Hunk{
					{Lines: []diffview.Line{{Type: diffview.LineContext, Content: "hunk1"}}},
					{Lines: []diffview.Line{{Type: diffview.LineContext, Content: "hunk2"}}},
					{Lines: []diffview.Line{{Type: diffview.LineContext, Content: "hunk3"}}},
				},
			},
		},
	}

	m := bubbletea.NewModel(diff)
	tm := teatest.NewTestModel(t, m,
		teatest.WithInitialTermSize(80, 24),
	)

	// Status bar should show hunk 1/3 when at top
	teatest.WaitFor(t, tm.Output(), func(out []byte) bool {
		return bytes.Contains(out, []byte("hunk 1/3"))
	})

	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	tm.WaitFinished(t, teatest.WithFinalTimeout(0))
}

func TestModel_StatusBarShowsScrollPosition(t *testing.T) {
	t.Parallel()

	// Create diff with many lines to enable scrolling
	lines := make([]diffview.Line, 100)
	for i := range lines {
		lines[i] = diffview.Line{Type: diffview.LineContext, Content: "content line"}
	}

	diff := &diffview.Diff{
		Files: []diffview.FileDiff{
			{
				OldPath: "a/file.go",
				NewPath: "b/file.go",
				Hunks: []diffview.Hunk{
					{Lines: lines},
				},
			},
		},
	}

	m := bubbletea.NewModel(diff)
	tm := teatest.NewTestModel(t, m,
		teatest.WithInitialTermSize(80, 10), // Small height to enable scrolling
	)

	// At top, should show "Top"
	teatest.WaitFor(t, tm.Output(), func(out []byte) bool {
		return bytes.Contains(out, []byte("Top"))
	})

	// Scroll down half page to get percentage display
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlD})

	// Should show a percentage (contains %)
	teatest.WaitFor(t, tm.Output(), func(out []byte) bool {
		return bytes.Contains(out, []byte("%"))
	})

	// Scroll to bottom
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'G'}})

	// At bottom, should show "Bot"
	teatest.WaitFor(t, tm.Output(), func(out []byte) bool {
		return bytes.Contains(out, []byte("Bot"))
	})

	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	tm.WaitFinished(t, teatest.WithFinalTimeout(0))
}

func TestModel_StatusBarShowsKeyHints(t *testing.T) {
	t.Parallel()

	diff := &diffview.Diff{
		Files: []diffview.FileDiff{
			{
				OldPath: "a/file.go",
				NewPath: "b/file.go",
				Hunks: []diffview.Hunk{
					{Lines: []diffview.Line{{Type: diffview.LineContext, Content: "content"}}},
				},
			},
		},
	}

	m := bubbletea.NewModel(diff)
	tm := teatest.NewTestModel(t, m,
		teatest.WithInitialTermSize(80, 24),
	)

	// Status bar should show key hints
	teatest.WaitFor(t, tm.Output(), func(out []byte) bool {
		hasScroll := bytes.Contains(out, []byte("j/k"))
		hasHunk := bytes.Contains(out, []byte("n/N"))
		hasFile := bytes.Contains(out, []byte("]/["))
		hasQuit := bytes.Contains(out, []byte("q"))
		return hasScroll && hasHunk && hasFile && hasQuit
	})

	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	tm.WaitFinished(t, teatest.WithFinalTimeout(0))
}

func TestModel_RendersLineNumbersInGutter(t *testing.T) {
	t.Parallel()

	// Create diff with known line numbers
	// Context line at old:10, new:10
	// Deleted line at old:11, new:-
	// Added line at old:-, new:11
	diff := &diffview.Diff{
		Files: []diffview.FileDiff{
			{
				OldPath:   "a/test.go",
				NewPath:   "b/test.go",
				Operation: diffview.FileModified,
				Hunks: []diffview.Hunk{
					{
						OldStart: 10,
						OldCount: 2,
						NewStart: 10,
						NewCount: 2,
						Lines: []diffview.Line{
							{Type: diffview.LineContext, Content: "context", OldLineNum: 10, NewLineNum: 10},
							{Type: diffview.LineDeleted, Content: "deleted", OldLineNum: 11, NewLineNum: 0},
							{Type: diffview.LineAdded, Content: "added", OldLineNum: 0, NewLineNum: 11},
						},
					},
				},
			},
		},
	}

	m := bubbletea.NewModel(diff)
	tm := teatest.NewTestModel(t, m,
		teatest.WithInitialTermSize(80, 24),
	)

	// Should render line numbers in gutter
	// Format: "  10    10 │" for context line
	// Format: "  11     - │" for deleted line (no new line number)
	// Format: "   -    11 │" for added line (no old line number)
	teatest.WaitFor(t, tm.Output(), func(out []byte) bool {
		// Check for context line with both numbers
		hasContext := bytes.Contains(out, []byte("10")) && bytes.Contains(out, []byte("context"))
		// Check for deleted line with old number and prefix
		hasDeleted := bytes.Contains(out, []byte("11")) && bytes.Contains(out, []byte("-deleted"))
		// Check for added line with new number and prefix
		hasAdded := bytes.Contains(out, []byte("+added"))
		return hasContext && hasDeleted && hasAdded
	})

	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	tm.WaitFinished(t, teatest.WithFinalTimeout(0))
}

func TestModel_GutterUsesEmptySpaceForMissingLineNumbers(t *testing.T) {
	t.Parallel()

	diff := &diffview.Diff{
		Files: []diffview.FileDiff{
			{
				OldPath:   "a/test.go",
				NewPath:   "b/test.go",
				Operation: diffview.FileModified,
				Hunks: []diffview.Hunk{
					{
						OldStart: 1,
						OldCount: 1,
						NewStart: 1,
						NewCount: 2,
						Lines: []diffview.Line{
							{Type: diffview.LineContext, Content: "context", OldLineNum: 1, NewLineNum: 1},
							{Type: diffview.LineAdded, Content: "new line", OldLineNum: 0, NewLineNum: 2},
						},
					},
				},
			},
		},
	}

	m := bubbletea.NewModel(diff)
	tm := teatest.NewTestModel(t, m,
		teatest.WithInitialTermSize(80, 24),
	)

	// For added lines, old line number should be empty space (not "-")
	// Gutter has no divider - color transition provides separation
	// The gutter directly precedes the line content: "    2 +new line"
	teatest.WaitFor(t, tm.Output(), func(out []byte) bool {
		// The gutter should NOT have divider character before the + prefix
		// (status bar uses │ as separator, so we check specifically for gutter)
		hasOldGutterFormat := bytes.Contains(out, []byte("│+new line"))
		hasContent := bytes.Contains(out, []byte("+new line"))
		// Also verify "-" placeholder is replaced with empty space
		hasDashPlaceholder := bytes.Contains(out, []byte("-    2"))
		return !hasOldGutterFormat && hasContent && !hasDashPlaceholder
	})

	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	tm.WaitFinished(t, teatest.WithFinalTimeout(0))
}

func TestModel_GutterHasColoredBackgroundForAddedLines(t *testing.T) {
	t.Parallel()

	diff := &diffview.Diff{
		Files: []diffview.FileDiff{
			{
				OldPath:   "a/test.go",
				NewPath:   "b/test.go",
				Operation: diffview.FileModified,
				Hunks: []diffview.Hunk{
					{
						OldStart: 1,
						OldCount: 1,
						NewStart: 1,
						NewCount: 2,
						Lines: []diffview.Line{
							{Type: diffview.LineContext, Content: "context", OldLineNum: 1, NewLineNum: 1},
							{Type: diffview.LineAdded, Content: "added", OldLineNum: 0, NewLineNum: 2},
						},
					},
				},
			},
		},
	}

	// TestTheme has AddedGutter with background from blending #00ff00 with #000000 at 35%
	// Result: RGB(0, 89, 0) -> "48;2;0;89;0"
	theme := dv.TestTheme()
	m := bubbletea.NewModel(diff,
		bubbletea.WithTheme(theme),
		bubbletea.WithRenderer(trueColorRenderer()),
	)
	tm := teatest.NewTestModel(t, m,
		teatest.WithInitialTermSize(80, 24),
	)

	// The gutter for added lines should have the AddedGutter background color
	teatest.WaitFor(t, tm.Output(), func(out []byte) bool {
		hasContent := bytes.Contains(out, []byte("+added"))
		// Check for the gutter background color (stronger green)
		hasGutterBackground := bytes.Contains(out, []byte("48;2;0;89;0"))
		return hasContent && hasGutterBackground
	})

	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	tm.WaitFinished(t, teatest.WithFinalTimeout(0))
}

func TestModel_GutterHasColoredBackgroundForDeletedLines(t *testing.T) {
	t.Parallel()

	diff := &diffview.Diff{
		Files: []diffview.FileDiff{
			{
				OldPath:   "a/test.go",
				NewPath:   "b/test.go",
				Operation: diffview.FileModified,
				Hunks: []diffview.Hunk{
					{
						OldStart: 1,
						OldCount: 2,
						NewStart: 1,
						NewCount: 1,
						Lines: []diffview.Line{
							{Type: diffview.LineContext, Content: "context", OldLineNum: 1, NewLineNum: 1},
							{Type: diffview.LineDeleted, Content: "deleted", OldLineNum: 2, NewLineNum: 0},
						},
					},
				},
			},
		},
	}

	// TestTheme has DeletedGutter with background from blending #ff0000 with #000000 at 35%
	// Result: RGB(89, 0, 0) -> "48;2;89;0;0"
	theme := dv.TestTheme()
	m := bubbletea.NewModel(diff,
		bubbletea.WithTheme(theme),
		bubbletea.WithRenderer(trueColorRenderer()),
	)
	tm := teatest.NewTestModel(t, m,
		teatest.WithInitialTermSize(80, 24),
	)

	// The gutter for deleted lines should have the DeletedGutter background color
	teatest.WaitFor(t, tm.Output(), func(out []byte) bool {
		hasContent := bytes.Contains(out, []byte("-deleted"))
		// Check for the gutter background color (stronger red)
		hasGutterBackground := bytes.Contains(out, []byte("48;2;89;0;0"))
		return hasContent && hasGutterBackground
	})

	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	tm.WaitFinished(t, teatest.WithFinalTimeout(0))
}

func TestModel_RendersFileHeaderWithStats(t *testing.T) {
	t.Parallel()

	diff := &diffview.Diff{
		Files: []diffview.FileDiff{
			{
				OldPath:   "a/handler.go",
				NewPath:   "b/handler.go",
				Operation: diffview.FileModified,
				Hunks: []diffview.Hunk{
					{
						OldStart: 1,
						OldCount: 5,
						NewStart: 1,
						NewCount: 7,
						Lines: []diffview.Line{
							{Type: diffview.LineContext, Content: "context"},
							{Type: diffview.LineDeleted, Content: "old1"},
							{Type: diffview.LineDeleted, Content: "old2"},
							{Type: diffview.LineAdded, Content: "new1"},
							{Type: diffview.LineAdded, Content: "new2"},
							{Type: diffview.LineAdded, Content: "new3"},
							{Type: diffview.LineAdded, Content: "new4"},
							{Type: diffview.LineContext, Content: "context"},
						},
					},
				},
			},
		},
	}

	m := bubbletea.NewModel(diff, bubbletea.WithTheme(dv.TestTheme()))
	tm := teatest.NewTestModel(t, m,
		teatest.WithInitialTermSize(80, 24),
	)

	// File header should be enhanced with box-drawing and stats: ── file ─── +N -M ──
	teatest.WaitFor(t, tm.Output(), func(out []byte) bool {
		// Should have box-drawing prefix, filename, and stats
		return bytes.Contains(out, []byte("── ")) &&
			bytes.Contains(out, []byte("handler.go")) &&
			bytes.Contains(out, []byte("+4 -2"))
	})

	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	tm.WaitFinished(t, teatest.WithFinalTimeout(0))
}

func TestModel_WithTheme(t *testing.T) {
	t.Parallel()

	diff := &diffview.Diff{
		Files: []diffview.FileDiff{
			{
				OldPath:   "a/test.go",
				NewPath:   "b/test.go",
				Operation: diffview.FileModified,
				Hunks: []diffview.Hunk{
					{
						OldStart: 1,
						OldCount: 1,
						NewStart: 1,
						NewCount: 2,
						Lines: []diffview.Line{
							{Type: diffview.LineContext, Content: "context"},
							{Type: diffview.LineAdded, Content: "added"},
						},
					},
				},
			},
		},
	}

	// TestTheme uses neutral foreground (#ffffff) with green-tinted background for added lines
	theme := dv.TestTheme()
	m := bubbletea.NewModel(diff,
		bubbletea.WithTheme(theme),
		bubbletea.WithRenderer(trueColorRenderer()),
	)
	tm := teatest.NewTestModel(t, m,
		teatest.WithInitialTermSize(80, 24),
	)

	// TestTheme uses neutral foreground (#ffffff) with green-tinted background
	// Should see background color code with green tint (48;2;...)
	teatest.WaitFor(t, tm.Output(), func(out []byte) bool {
		hasContent := bytes.Contains(out, []byte("added"))
		// Check for any background color on the added line (48;2; prefix)
		hasBackground := bytes.Contains(out, []byte("48;2;"))
		return hasContent && hasBackground
	})

	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	tm.WaitFinished(t, teatest.WithFinalTimeout(0))
}

func TestModel_StatusBarUsesThemeUIColors(t *testing.T) {
	t.Parallel()

	diff := &diffview.Diff{
		Files: []diffview.FileDiff{
			{
				OldPath: "a/file.go",
				NewPath: "b/file.go",
				Hunks: []diffview.Hunk{
					{Lines: []diffview.Line{{Type: diffview.LineContext, Content: "content"}}},
				},
			},
		},
	}

	// TestTheme has UIBackground=#333333 = RGB(51, 51, 51)
	// The status bar text "file 1/1" should have this background color
	theme := dv.TestTheme()
	m := bubbletea.NewModel(diff,
		bubbletea.WithTheme(theme),
		bubbletea.WithRenderer(trueColorRenderer()),
	)
	tm := teatest.NewTestModel(t, m,
		teatest.WithInitialTermSize(80, 24),
	)

	// Wait for the model to render and collect output
	var finalOutput []byte
	teatest.WaitFor(t, tm.Output(), func(out []byte) bool {
		if bytes.Contains(out, []byte("file 1/1")) {
			finalOutput = out
			return true
		}
		return false
	})

	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	tm.WaitFinished(t, teatest.WithFinalTimeout(0))

	// The status bar should use themed colors, which means there should be
	// color codes immediately before "file 1/1". Previously it used
	// lipgloss.NewStyle(), which ignored the renderer so the status bar
	// rendered without colors in tests; this test verifies it now has colors.
	//
	// Look for the pattern: background color code followed by "file 1/1"
	// TestTheme UIBackground is #333333 = RGB(51, 51, 51) -> "48;2;51;51;51"
	statusBarLine := extractLastLine(string(finalOutput))
	assert.Contains(t, statusBarLine, "48;2;51;51;51", "status bar should use TestTheme UIBackground color")
}

func TestModel_AppliesSyntaxHighlighting(t *testing.T) {
	t.Parallel()

	// Create a diff with Go code that will get syntax highlighted
	diff := &diffview.Diff{
		Files: []diffview.FileDiff{
			{
				OldPath:   "a/main.go",
				NewPath:   "b/main.go",
				Operation: diffview.FileModified,
				Hunks: []diffview.Hunk{
					{
						OldStart: 1,
						OldCount: 1,
						NewStart: 1,
						NewCount: 2,
						Lines: []diffview.Line{
							{Type: diffview.LineContext, Content: "package main"},
							{Type: diffview.LineAdded, Content: "func main() {}", OldLineNum: 0, NewLineNum: 2},
						},
					},
				},
			},
		},
	}

	// Use TestTheme which has predictable colors:
	// Keyword: #ff00ff (magenta) = RGB(255, 0, 255) -> "38;2;255;0;255"
	theme := dv.TestTheme()

	// Create a mock tokenizer that returns tokens with keyword style
	tokenizer := &mockTokenizer{
		TokenizeLinesFn: func(language, source string) [][]diffview.Token {
			if language != "Go" {
				return nil
			}
			// For "package main\nfunc main() {}" return tokens for both lines
			if source == "package main\nfunc main() {}" {
				return [][]diffview.Token{
					{
						{Text: "package", Style: diffview.Style{Foreground: "#ff00ff", Bold: true}},
						{Text: " ", Style: diffview.Style{}},
						{Text: "main", Style: diffview.Style{}},
					},
					{
						{Text: "func", Style: diffview.Style{Foreground: "#ff00ff", Bold: true}},
						{Text: " ", Style: diffview.Style{}},
						{Text: "main", Style: diffview.Style{Foreground: "#0000ff"}},
						{Text: "()", Style: diffview.Style{}},
						{Text: " {}", Style: diffview.Style{}},
					},
				}
			}
			return nil
		},
	}

	// Create a mock detector that returns "Go" for .go files
	detector := &mockLanguageDetector{
		DetectFromPathFn: func(path string) string {
			if len(path) >= 3 && path[len(path)-3:] == ".go" {
				return "Go"
			}
			return ""
		},
	}

	m := bubbletea.NewModel(diff,
		bubbletea.WithTheme(theme),
		bubbletea.WithRenderer(trueColorRenderer()),
		bubbletea.WithLanguageDetector(detector),
		bubbletea.WithTokenizer(tokenizer),
	)
	tm := teatest.NewTestModel(t, m,
		teatest.WithInitialTermSize(80, 24),
	)

	// Wait for output with syntax highlighting
	// The keyword "package" or "func" should have magenta foreground
	// RGB(255, 0, 255) -> "38;2;255;0;255"
	teatest.WaitFor(t, tm.Output(), func(out []byte) bool {
		hasContent := bytes.Contains(out, []byte("package"))
		hasMagentaKeyword := bytes.Contains(out, []byte("38;2;255;0;255"))
		return hasContent && hasMagentaKeyword
	})

	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	tm.WaitFinished(t, teatest.WithFinalTimeout(0))
}

func TestModel_PaddingBetweenGutterAndCodePrefix(t *testing.T) {
	t.Parallel()

	// Create a diff with added, deleted, and context lines
	diff := &diffview.Diff{
		Files: []diffview.FileDiff{
			{
				OldPath:   "a/test.go",
				NewPath:   "b/test.go",
				Operation: diffview.FileModified,
				Hunks: []diffview.Hunk{
					{
						OldStart: 1,
						OldCount: 2,
						NewStart: 1,
						NewCount: 2,
						Lines: []diffview.Line{
							{Type: diffview.LineContext, Content: "context", OldLineNum: 1, NewLineNum: 1},
							{Type: diffview.LineDeleted, Content: "deleted", OldLineNum: 2, NewLineNum: 0},
							{Type: diffview.LineAdded, Content: "added", OldLineNum: 0, NewLineNum: 2},
						},
					},
				},
			},
		},
	}

	m := bubbletea.NewModel(diff, bubbletea.WithRenderer(trueColorRenderer()))
	tm := teatest.NewTestModel(t, m,
		teatest.WithInitialTermSize(80, 24),
	)

	// Wait for output with padding space between gutter and line prefix
	// The padding space appears between the gutter and the prefix character (+/-/space)
	// Due to ANSI color codes, the padding space may be separated from the prefix by escape sequences
	// We verify by checking that the rendered text shows " +added", " -deleted", "  context"
	// (padding space + prefix + content for each line type)
	teatest.WaitFor(t, tm.Output(), func(out []byte) bool {
		// After the gutter styling ends (reset code), we should see the padding space
		// followed by the prefix character. Check for space-prefix-content patterns.
		hasAddedWithPadding := bytes.Contains(out, []byte(" +added"))
		hasDeletedWithPadding := bytes.Contains(out, []byte(" -deleted"))
		// For context lines, the prefix is a space, so we get "  context" (padding + prefix + content)
		hasContextWithPadding := bytes.Contains(out, []byte("  context"))
		return hasAddedWithPadding && hasDeletedWithPadding && hasContextWithPadding
	})

	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	tm.WaitFinished(t, teatest.WithFinalTimeout(0))
}

func TestModel_PaddingUsesCodeLineBackgroundColor(t *testing.T) {
	t.Parallel()

	// Create a diff with an added line to test padding background color
	diff := &diffview.Diff{
		Files: []diffview.FileDiff{
			{
				OldPath:   "a/test.go",
				NewPath:   "b/test.go",
				Operation: diffview.FileModified,
				Hunks: []diffview.Hunk{
					{
						OldStart: 1,
						OldCount: 1,
						NewStart: 1,
						NewCount: 2,
						Lines: []diffview.Line{
							{Type: diffview.LineContext, Content: "context", OldLineNum: 1, NewLineNum: 1},
							{Type: diffview.LineAdded, Content: "added", OldLineNum: 0, NewLineNum: 2},
						},
					},
				},
			},
		},
	}

	// TestTheme has different colors for gutter vs line background:
	// AddedGutter background: RGB(0, 89, 0) -> "48;2;0;89;0" (stronger green)
	// Added line background: RGB(0, 38, 0) -> "48;2;0;38;0" (subtler green)
	theme := dv.TestTheme()
	m := bubbletea.NewModel(diff,
		bubbletea.WithTheme(theme),
		bubbletea.WithRenderer(trueColorRenderer()),
	)
	tm := teatest.NewTestModel(t, m,
		teatest.WithInitialTermSize(80, 24),
	)

	// The padding space should use the line background color (0, 38, 0), not gutter (0, 89, 0)
	// The padding immediately follows the gutter, so we look for the pattern:
	// gutter ends with gutter-background -> padding has line-background
	teatest.WaitFor(t, tm.Output(), func(out []byte) bool {
		hasContent := bytes.Contains(out, []byte("+added"))
		// The padding space should have the line background color
		// Check that the output contains both the gutter background and line background colors
		hasGutterBackground := bytes.Contains(out, []byte("48;2;0;89;0"))
		hasLineBackground := bytes.Contains(out, []byte("48;2;0;38;0"))
		return hasContent && hasGutterBackground && hasLineBackground
	})

	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	tm.WaitFinished(t, teatest.WithFinalTimeout(0))
}

func TestModel_ShowsEmptyFileCreation(t *testing.T) {
	t.Parallel()

	// Create a diff with an empty file creation (no hunks, but Operation=FileAdded)
	diff := &diffview.Diff{
		Files: []diffview.FileDiff{
			{
				NewPath:   "empty.txt",
				Operation: diffview.FileAdded,
				// No hunks - this is an empty file
			},
		},
	}

	m := bubbletea.NewModel(diff)
	tm := teatest.NewTestModel(t, m,
		teatest.WithInitialTermSize(80, 24),
	)

	// Empty file should appear with filename and "(empty)" indicator
	teatest.WaitFor(t, tm.Output(), func(out []byte) bool {
		hasFilename := bytes.Contains(out, []byte("empty.txt"))
		hasEmptyIndicator := bytes.Contains(out, []byte("(empty)"))
		return hasFilename && hasEmptyIndicator
	}, teatest.WithDuration(2*time.Second))

	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	tm.WaitFinished(t, teatest.WithFinalTimeout(0))
}

func TestModel_ShowsEmptyFileDeletion(t *testing.T) {
	t.Parallel()

	// Create a diff with an empty file deletion (no hunks, but Operation=FileDeleted)
	diff := &diffview.Diff{
		Files: []diffview.FileDiff{
			{
				OldPath:   "deleted.txt",
				Operation: diffview.FileDeleted,
				// No hunks - this was an empty file that got deleted
			},
		},
	}

	m := bubbletea.NewModel(diff)
	tm := teatest.NewTestModel(t, m,
		teatest.WithInitialTermSize(80, 24),
	)

	// Empty deleted file should appear with filename and "(empty)" indicator
	teatest.WaitFor(t, tm.Output(), func(out []byte) bool {
		hasFilename := bytes.Contains(out, []byte("deleted.txt"))
		hasEmptyIndicator := bytes.Contains(out, []byte("(empty)"))
		return hasFilename && hasEmptyIndicator
	}, teatest.WithDuration(2*time.Second))

	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	tm.WaitFinished(t, teatest.WithFinalTimeout(0))
}
