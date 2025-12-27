// Large diff performance tests verify that diffview handles large diffs
// (like the 7.6MB eval-dataset.jsonl) without issues. These tests were
// added for diffview-ohe after confirming the existing implementation
// already meets performance requirements (~50ms render, ~23MB memory for 7.6MB diff).
package bubbletea_test

import (
	"runtime"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fwojciec/diffview"
	"github.com/fwojciec/diffview/bubbletea"
	"github.com/fwojciec/diffview/gitdiff"
	"github.com/fwojciec/diffview/lipgloss"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// generateLargeDiff creates a diff with the specified number of lines.
// Each line is approximately lineLength characters long.
func generateLargeDiff(linesPerFile, lineLength int) *diffview.Diff {
	file := diffview.FileDiff{
		OldPath:   "file0.go",
		NewPath:   "file0.go",
		Operation: diffview.FileModified,
		Hunks:     make([]diffview.Hunk, 1),
	}

	lines := make([]diffview.Line, linesPerFile)
	content := strings.Repeat("x", lineLength) + "\n"

	for j := 0; j < linesPerFile; j++ {
		var lineType diffview.LineType
		switch j % 3 {
		case 0:
			lineType = diffview.LineAdded
		case 1:
			lineType = diffview.LineDeleted
		default:
			lineType = diffview.LineContext
		}

		// Set line numbers according to diff semantics:
		// Added lines have no old line number (0)
		// Deleted lines have no new line number (0)
		oldLineNum := j + 1
		newLineNum := j + 1
		switch lineType {
		case diffview.LineAdded:
			oldLineNum = 0
		case diffview.LineDeleted:
			newLineNum = 0
		}

		lines[j] = diffview.Line{
			Type:       lineType,
			Content:    content,
			OldLineNum: oldLineNum,
			NewLineNum: newLineNum,
		}
	}

	file.Hunks[0] = diffview.Hunk{
		OldStart: 1,
		OldCount: linesPerFile,
		NewStart: 1,
		NewCount: linesPerFile,
		Lines:    lines,
	}

	return &diffview.Diff{
		Files: []diffview.FileDiff{file},
	}
}

func TestLargeDiff_ModelCreation(t *testing.T) {
	t.Parallel()

	// Simulate a 7.6MB diff: ~100 lines of ~76KB each
	diff := generateLargeDiff(100, 76000)

	model := bubbletea.NewModel(diff, bubbletea.WithTheme(lipgloss.DefaultTheme()))

	// Positions should be computed eagerly and correctly
	assert.Len(t, model.FilePositions(), 1)
	assert.Len(t, model.HunkPositions(), 1)
}

func TestLargeDiff_Parse(t *testing.T) {
	t.Parallel()

	// Create a large diff string (~7.25 MB)
	var sb strings.Builder
	sb.WriteString("diff --git a/large.jsonl b/large.jsonl\n")
	sb.WriteString("new file mode 100644\n")
	sb.WriteString("index 0000000..1234567\n")
	sb.WriteString("--- /dev/null\n")
	sb.WriteString("+++ b/large.jsonl\n")
	sb.WriteString("@@ -0,0 +1,100 @@\n")
	lineContent := strings.Repeat("x", 76000)
	for i := 0; i < 100; i++ {
		sb.WriteString("+" + lineContent + "\n")
	}
	diffStr := sb.String()

	start := time.Now()
	parser := gitdiff.NewParser()
	diff, err := parser.Parse(strings.NewReader(diffStr))
	duration := time.Since(start)

	require.NoError(t, err)
	require.Len(t, diff.Files, 1)
	assert.Len(t, diff.Files[0].Hunks[0].Lines, 100)

	// Parse should complete in under 5 seconds
	assert.Less(t, duration, 5*time.Second, "Parse took too long: %v", duration)
}

func TestLargeDiff_RenderAndView(t *testing.T) {
	t.Parallel()

	// Simulate a 7.6MB diff: ~100 lines of ~76KB each
	diff := generateLargeDiff(100, 76000)

	model := bubbletea.NewModel(diff, bubbletea.WithTheme(lipgloss.DefaultTheme()))

	// Trigger rendering via WindowSizeMsg
	msg := tea.WindowSizeMsg{Width: 120, Height: 40}
	updatedModel, _ := model.Update(msg)
	model = updatedModel.(bubbletea.Model)

	// Get view output
	view := model.View()

	// View should produce non-empty output
	assert.NotEmpty(t, view)
}

func TestLargeDiff_PerformanceBounds(t *testing.T) {
	t.Parallel()

	// Simulate viewing eval-dataset.jsonl: 83 lines, each ~90KB (~7.6MB total)
	diff := generateLargeDiff(83, 91000)

	var memBefore runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&memBefore)

	start := time.Now()

	// Create model
	model := bubbletea.NewModel(diff, bubbletea.WithTheme(lipgloss.DefaultTheme()))

	// Render
	msg := tea.WindowSizeMsg{Width: 120, Height: 40}
	updatedModel, _ := model.Update(msg)
	model = updatedModel.(bubbletea.Model)

	// Get view
	view := model.View()

	totalTime := time.Since(start)

	var memAfter runtime.MemStats
	runtime.ReadMemStats(&memAfter)
	memUsed := memAfter.Alloc - memBefore.Alloc

	// Assertions
	assert.NotEmpty(t, view)
	assert.Less(t, totalTime, 2*time.Second, "Total time exceeded 2s")
	// Memory bound is generous (200MB) to account for parallel test noise.
	// Actual usage is typically ~23MB for a 7.6MB diff. Benchmarks provide
	// more precise memory tracking via b.ReportAllocs().
	assert.Less(t, memUsed, uint64(200*1024*1024), "Memory usage exceeded 200MB")
}

// benchResult prevents compiler from optimizing away benchmark results.
var benchResult any

func BenchmarkLargeDiff_Parse(b *testing.B) {
	var sb strings.Builder
	sb.WriteString("diff --git a/large.jsonl b/large.jsonl\n")
	sb.WriteString("new file mode 100644\n")
	sb.WriteString("index 0000000..1234567\n")
	sb.WriteString("--- /dev/null\n")
	sb.WriteString("+++ b/large.jsonl\n")
	sb.WriteString("@@ -0,0 +1,100 @@\n")
	lineContent := strings.Repeat("x", 76000)
	for i := 0; i < 100; i++ {
		sb.WriteString("+" + lineContent + "\n")
	}
	diffStr := sb.String()

	b.ResetTimer()
	b.ReportAllocs()

	var result *diffview.Diff
	for i := 0; i < b.N; i++ {
		parser := gitdiff.NewParser()
		diff, err := parser.Parse(strings.NewReader(diffStr))
		if err != nil {
			b.Fatal(err)
		}
		result = diff
	}
	benchResult = result
}

func BenchmarkLargeDiff_ModelCreate(b *testing.B) {
	diff := generateLargeDiff(100, 76000)

	b.ResetTimer()
	b.ReportAllocs()

	var result bubbletea.Model
	for i := 0; i < b.N; i++ {
		result = bubbletea.NewModel(diff, bubbletea.WithTheme(lipgloss.DefaultTheme()))
	}
	benchResult = result
}

func BenchmarkLargeDiff_Render(b *testing.B) {
	diff := generateLargeDiff(100, 76000)
	msg := tea.WindowSizeMsg{Width: 120, Height: 40}

	b.ResetTimer()
	b.ReportAllocs()

	// Create fresh model each iteration to benchmark cold render path.
	// Model.Update returns new model (value semantics), so state is not mutated.
	var result string
	for i := 0; i < b.N; i++ {
		model := bubbletea.NewModel(diff, bubbletea.WithTheme(lipgloss.DefaultTheme()))
		updatedModel, _ := model.Update(msg)
		result = updatedModel.(bubbletea.Model).View()
	}
	benchResult = result
}
