package bubbletea_test

import (
	"bytes"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"
	diffview "github.com/fwojciec/diffstory"
	"github.com/fwojciec/diffstory/bubbletea"
	dv "github.com/fwojciec/diffstory/lipgloss"
	"github.com/stretchr/testify/assert"
)

func TestModel_WordDiffHighlighting(t *testing.T) {
	t.Parallel()

	// Create a diff with a paired delete/add line (a "replace" operation).
	// Word diff should highlight the changed portions within the lines.
	// "hello world" -> "hello universe"
	// - "world" should be highlighted in deleted line
	// - "universe" should be highlighted in added line
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
							{Type: diffview.LineDeleted, Content: "hello world", OldLineNum: 1, NewLineNum: 0},
							{Type: diffview.LineAdded, Content: "hello universe", OldLineNum: 0, NewLineNum: 1},
						},
					},
				},
			},
		},
	}

	// Create a mock word differ that returns segments
	wordDiffer := &mockWordDiffer{
		DiffFn: func(old, new string) (oldSegs, newSegs []diffview.Segment) {
			if old == "hello world" && new == "hello universe" {
				oldSegs = []diffview.Segment{
					{Text: "hello ", Changed: false},
					{Text: "world", Changed: true},
				}
				newSegs = []diffview.Segment{
					{Text: "hello ", Changed: false},
					{Text: "universe", Changed: true},
				}
			}
			return oldSegs, newSegs
		},
	}

	// TestTheme uses (GitHub-style - same foreground, gutter-intensity background):
	// AddedHighlight: gutter-intensity green background (35% blend) -> "48;2;0;89;0"
	// DeletedHighlight: gutter-intensity red background (35% blend) -> "48;2;89;0;0"
	// Added line (unchanged parts): dimmed green (15% blend) -> "48;2;0;38;0"
	// Deleted line (unchanged parts): dimmed red (15% blend) -> "48;2;38;0;0"
	theme := dv.TestTheme()
	m := bubbletea.NewModel(diff,
		bubbletea.WithTheme(theme),
		bubbletea.WithRenderer(trueColorRenderer()),
		bubbletea.WithWordDiffer(wordDiffer),
	)
	tm := teatest.NewTestModel(t, m,
		teatest.WithInitialTermSize(80, 24),
	)

	// Check that word-level highlighting is applied:
	// - Changed text should have gutter-intensity background (35% blend)
	// - Unchanged text should have dimmed background (15% blend)
	teatest.WaitFor(t, tm.Output(), func(out []byte) bool {
		hasDeletedLine := bytes.Contains(out, []byte("-hello"))
		hasAddedLine := bytes.Contains(out, []byte("+hello"))
		// Check for gutter-intensity highlight backgrounds (35% blend)
		hasDeletedHighlight := bytes.Contains(out, []byte("48;2;89;0;0")) // DeletedHighlight (gutter intensity)
		hasAddedHighlight := bytes.Contains(out, []byte("48;2;0;89;0"))   // AddedHighlight (gutter intensity)
		// Check for dimmed line backgrounds (15% blend)
		hasDimmedBackground := bytes.Contains(out, []byte("48;2;0;38;0")) || // Added dimmed
			bytes.Contains(out, []byte("48;2;38;0;0")) // Deleted dimmed
		return hasDeletedLine && hasAddedLine && hasDeletedHighlight && hasAddedHighlight && hasDimmedBackground
	})

	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	tm.WaitFinished(t, teatest.WithFinalTimeout(0))
}

func TestModel_WordDiffHighlighting_NonPairedLinesNoHighlight(t *testing.T) {
	t.Parallel()

	// Non-paired lines (add without preceding delete) should NOT get word-level highlighting
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
							{Type: diffview.LineContext, Content: "unchanged", OldLineNum: 1, NewLineNum: 1},
							{Type: diffview.LineAdded, Content: "newly added", OldLineNum: 0, NewLineNum: 2},
						},
					},
				},
			},
		},
	}

	// Create a mock word differ that should NOT be called for non-paired lines
	wordDifferCalled := false
	wordDiffer := &mockWordDiffer{
		DiffFn: func(old, new string) (oldSegs, newSegs []diffview.Segment) {
			wordDifferCalled = true
			return nil, nil
		},
	}

	theme := dv.TestTheme()
	m := bubbletea.NewModel(diff,
		bubbletea.WithTheme(theme),
		bubbletea.WithRenderer(trueColorRenderer()),
		bubbletea.WithWordDiffer(wordDiffer),
	)
	tm := teatest.NewTestModel(t, m,
		teatest.WithInitialTermSize(80, 24),
	)

	// Wait for render - the added line should be present
	teatest.WaitFor(t, tm.Output(), func(out []byte) bool {
		return bytes.Contains(out, []byte("+newly added"))
	})

	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	tm.WaitFinished(t, teatest.WithFinalTimeout(0))

	// Word differ should NOT have been called since there's no paired delete/add
	assert.False(t, wordDifferCalled, "WordDiffer should not be called for non-paired lines")
}

func TestModel_WordDiffHighlighting_MultiplePairs(t *testing.T) {
	t.Parallel()

	// Test multiple pairs in sequence
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
							// First pair
							{Type: diffview.LineDeleted, Content: "old line 1", OldLineNum: 1, NewLineNum: 0},
							{Type: diffview.LineAdded, Content: "new line 1", OldLineNum: 0, NewLineNum: 1},
							// Second pair
							{Type: diffview.LineDeleted, Content: "old line 2", OldLineNum: 2, NewLineNum: 0},
							{Type: diffview.LineAdded, Content: "new line 2", OldLineNum: 0, NewLineNum: 2},
						},
					},
				},
			},
		},
	}

	diffCallCount := 0
	wordDiffer := &mockWordDiffer{
		DiffFn: func(old, new string) (oldSegs, newSegs []diffview.Segment) {
			diffCallCount++
			// Return segments that mark the first word as unchanged, second as changed
			oldSegs = []diffview.Segment{
				{Text: "old ", Changed: false},
				{Text: "line " + old[len(old)-1:], Changed: true},
			}
			newSegs = []diffview.Segment{
				{Text: "new ", Changed: false},
				{Text: "line " + new[len(new)-1:], Changed: true},
			}
			return oldSegs, newSegs
		},
	}

	theme := dv.TestTheme()
	m := bubbletea.NewModel(diff,
		bubbletea.WithTheme(theme),
		bubbletea.WithRenderer(trueColorRenderer()),
		bubbletea.WithWordDiffer(wordDiffer),
	)
	tm := teatest.NewTestModel(t, m,
		teatest.WithInitialTermSize(80, 24),
	)

	// Wait for render - both pairs should be present
	teatest.WaitFor(t, tm.Output(), func(out []byte) bool {
		hasFirstDelete := bytes.Contains(out, []byte("-old"))
		hasFirstAdd := bytes.Contains(out, []byte("+new"))
		return hasFirstDelete && hasFirstAdd
	})

	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	tm.WaitFinished(t, teatest.WithFinalTimeout(0))

	// Word differ should be called twice, once for each pair
	assert.Equal(t, 2, diffCallCount, "WordDiffer should be called once per pair")
}

func TestModel_WordDiffHighlighting_ConsecutiveDeletesAndAdds(t *testing.T) {
	t.Parallel()

	// Test that consecutive deletes followed by consecutive adds are paired 1:1
	// This is common when changing multiple lines in a block
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
							// Two consecutive deletes
							{Type: diffview.LineDeleted, Content: `Foreground: "#1e1e2e",`, OldLineNum: 1, NewLineNum: 0},
							{Type: diffview.LineDeleted, Content: `Background: "#a6e3a1",`, OldLineNum: 2, NewLineNum: 0},
							// Two consecutive adds
							{Type: diffview.LineAdded, Content: `Foreground: "#cdd6f4",`, OldLineNum: 0, NewLineNum: 1},
							{Type: diffview.LineAdded, Content: `Background: "#3d5a3d",`, OldLineNum: 0, NewLineNum: 2},
						},
					},
				},
			},
		},
	}

	// Mock word differ that returns segments with shared structure
	pairsProcessed := make(map[string]bool)
	wordDiffer := &mockWordDiffer{
		DiffFn: func(old, new string) (oldSegs, newSegs []diffview.Segment) {
			pairsProcessed[old+"->"+new] = true
			// Simulate word diff where the color code is different but structure is shared
			// "Foreground: " is unchanged, the color code is changed
			if strings.HasPrefix(old, "Foreground") && strings.HasPrefix(new, "Foreground") {
				oldSegs = []diffview.Segment{
					{Text: `Foreground: "`, Changed: false},
					{Text: `#1e1e2e`, Changed: true},
					{Text: `",`, Changed: false},
				}
				newSegs = []diffview.Segment{
					{Text: `Foreground: "`, Changed: false},
					{Text: `#cdd6f4`, Changed: true},
					{Text: `",`, Changed: false},
				}
			} else if strings.HasPrefix(old, "Background") && strings.HasPrefix(new, "Background") {
				oldSegs = []diffview.Segment{
					{Text: `Background: "`, Changed: false},
					{Text: `#a6e3a1`, Changed: true},
					{Text: `",`, Changed: false},
				}
				newSegs = []diffview.Segment{
					{Text: `Background: "`, Changed: false},
					{Text: `#3d5a3d`, Changed: true},
					{Text: `",`, Changed: false},
				}
			}
			return oldSegs, newSegs
		},
	}

	theme := dv.TestTheme()
	m := bubbletea.NewModel(diff,
		bubbletea.WithTheme(theme),
		bubbletea.WithRenderer(trueColorRenderer()),
		bubbletea.WithWordDiffer(wordDiffer),
	)
	tm := teatest.NewTestModel(t, m,
		teatest.WithInitialTermSize(80, 24),
	)

	teatest.WaitFor(t, tm.Output(), func(out []byte) bool {
		hasDeletedLine := bytes.Contains(out, []byte("-Foreground"))
		hasAddedLine := bytes.Contains(out, []byte("+Foreground"))
		// Check for gutter-intensity highlight backgrounds (word diff applied)
		hasHighlight := bytes.Contains(out, []byte("48;2;89;0;0")) || // DeletedHighlight
			bytes.Contains(out, []byte("48;2;0;89;0")) // AddedHighlight
		return hasDeletedLine && hasAddedLine && hasHighlight
	})

	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	tm.WaitFinished(t, teatest.WithFinalTimeout(0))

	// Verify correct pairing: 1st delete with 1st add, 2nd delete with 2nd add
	assert.True(t, pairsProcessed[`Foreground: "#1e1e2e",->`+`Foreground: "#cdd6f4",`],
		"1st delete should pair with 1st add")
	assert.True(t, pairsProcessed[`Background: "#a6e3a1",->`+`Background: "#3d5a3d",`],
		"2nd delete should pair with 2nd add")
}

func TestModel_WordDiffHighlighting_SkipsWhenLinesTooDifferent(t *testing.T) {
	t.Parallel()

	// When lines are too different (< 30% shared content), word-level diff should be skipped
	// to avoid highlighting everything as "changed" which is just noise
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
							{Type: diffview.LineDeleted, Content: "completely different old line", OldLineNum: 1, NewLineNum: 0},
							{Type: diffview.LineAdded, Content: "totally new content here", OldLineNum: 0, NewLineNum: 1},
						},
					},
				},
			},
		},
	}

	// Mock word differ that returns everything as changed (simulating very different lines)
	wordDifferCalled := false
	wordDiffer := &mockWordDiffer{
		DiffFn: func(old, new string) (oldSegs, newSegs []diffview.Segment) {
			wordDifferCalled = true
			// Return segments where everything is changed (no shared content)
			oldSegs = []diffview.Segment{
				{Text: old, Changed: true}, // 100% changed
			}
			newSegs = []diffview.Segment{
				{Text: new, Changed: true}, // 100% changed
			}
			return oldSegs, newSegs
		},
	}

	theme := dv.TestTheme()
	m := bubbletea.NewModel(diff,
		bubbletea.WithTheme(theme),
		bubbletea.WithRenderer(trueColorRenderer()),
		bubbletea.WithWordDiffer(wordDiffer),
	)
	tm := teatest.NewTestModel(t, m,
		teatest.WithInitialTermSize(80, 24),
	)

	// Wait for render
	teatest.WaitFor(t, tm.Output(), func(out []byte) bool {
		hasDeletedLine := bytes.Contains(out, []byte("-completely"))
		hasAddedLine := bytes.Contains(out, []byte("+totally"))
		// Should have dimmed backgrounds (no word-level highlighting applied)
		// because lines are too different
		hasDimmedBackground := bytes.Contains(out, []byte("48;2;0;38;0")) || // Added dimmed
			bytes.Contains(out, []byte("48;2;38;0;0")) // Deleted dimmed
		return hasDeletedLine && hasAddedLine && hasDimmedBackground
	})

	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	tm.WaitFinished(t, teatest.WithFinalTimeout(0))

	// Word differ should have been called (to compute segments)
	assert.True(t, wordDifferCalled, "WordDiffer should be called to compute segments")
}

func TestModel_WordDiffHighlighting_NoWordDiffer(t *testing.T) {
	t.Parallel()

	// When no WordDiffer is provided, rendering should work (graceful degradation)
	// Lines render with uniform line-level styling, no word-level segments
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
							{Type: diffview.LineDeleted, Content: "hello world", OldLineNum: 1, NewLineNum: 0},
							{Type: diffview.LineAdded, Content: "hello universe", OldLineNum: 0, NewLineNum: 1},
						},
					},
				},
			},
		},
	}

	// No WordDiffer provided - should render without crashing
	theme := dv.TestTheme()
	m := bubbletea.NewModel(diff,
		bubbletea.WithTheme(theme),
		bubbletea.WithRenderer(trueColorRenderer()),
		// Intentionally NOT setting WithWordDiffer
	)
	tm := teatest.NewTestModel(t, m,
		teatest.WithInitialTermSize(80, 24),
	)

	// Verify lines render correctly without WordDiffer
	teatest.WaitFor(t, tm.Output(), func(out []byte) bool {
		hasDeletedLine := bytes.Contains(out, []byte("-hello"))
		hasAddedLine := bytes.Contains(out, []byte("+hello"))
		// Lines should have dimmed background (15% blend), not gutter-intensity (35%)
		// since without word diff, entire line is uniformly styled
		hasDimmedAddedBackground := bytes.Contains(out, []byte("48;2;0;38;0"))
		hasDimmedDeletedBackground := bytes.Contains(out, []byte("48;2;38;0;0"))
		return hasDeletedLine && hasAddedLine && hasDimmedAddedBackground && hasDimmedDeletedBackground
	})

	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	tm.WaitFinished(t, teatest.WithFinalTimeout(0))
}
