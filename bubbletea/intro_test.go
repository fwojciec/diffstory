package bubbletea_test

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	diffview "github.com/fwojciec/diffstory"
	"github.com/fwojciec/diffstory/bubbletea"
	"github.com/muesli/termenv"
	"github.com/stretchr/testify/assert"
)

func TestNarrativeDiagram_CauseEffect(t *testing.T) {
	t.Parallel()

	// cause-effect narrative with problem, fix, test roles
	sections := []diffview.Section{
		{Role: "problem", Title: "The Bug"},
		{Role: "fix", Title: "The Solution"},
		{Role: "test", Title: "Verification"},
	}

	renderer := lipgloss.NewRenderer(nil, termenv.WithProfile(termenv.Ascii))
	diagram := bubbletea.NarrativeDiagram("cause-effect", sections, renderer)

	// Should show linear flow with arrows between roles
	assert.Contains(t, diagram, "problem")
	assert.Contains(t, diagram, "fix")
	assert.Contains(t, diagram, "test")
	assert.Contains(t, diagram, "→")
}

func TestNarrativeDiagram_EntryImplementation(t *testing.T) {
	t.Parallel()

	// entry-implementation narrative with entry and implementation roles
	sections := []diffview.Section{
		{Role: "entry", Title: "API Contract"},
		{Role: "implementation", Title: "Core Logic"},
	}

	renderer := lipgloss.NewRenderer(nil, termenv.WithProfile(termenv.Ascii))
	diagram := bubbletea.NarrativeDiagram("entry-implementation", sections, renderer)

	// Should show linear flow with entry and implementation roles
	assert.Contains(t, diagram, "entry")
	assert.Contains(t, diagram, "implementation")
	assert.Contains(t, diagram, "→")
}

func TestNarrativeDiagram_BeforeAfter(t *testing.T) {
	t.Parallel()

	// before-after narrative with cleanup and core roles
	sections := []diffview.Section{
		{Role: "cleanup", Title: "Remove old code"},
		{Role: "core", Title: "Add new code"},
	}

	renderer := lipgloss.NewRenderer(nil, termenv.WithProfile(termenv.Ascii))
	diagram := bubbletea.NarrativeDiagram("before-after", sections, renderer)

	// Should show transformation flow
	assert.Contains(t, diagram, "cleanup")
	assert.Contains(t, diagram, "core")
	assert.Contains(t, diagram, "→")
}

func TestNarrativeDiagram_EmptySections(t *testing.T) {
	t.Parallel()

	renderer := lipgloss.NewRenderer(nil, termenv.WithProfile(termenv.Ascii))
	diagram := bubbletea.NarrativeDiagram("cause-effect", nil, renderer)

	assert.Empty(t, diagram)
}

func TestNarrativeDiagram_NoRoles(t *testing.T) {
	t.Parallel()

	// Sections without roles
	sections := []diffview.Section{
		{Title: "First"},
		{Title: "Second"},
	}

	renderer := lipgloss.NewRenderer(nil, termenv.WithProfile(termenv.Ascii))
	diagram := bubbletea.NarrativeDiagram("cause-effect", sections, renderer)

	assert.Empty(t, diagram)
}

func TestNarrativeDiagram_UnknownNarrative(t *testing.T) {
	t.Parallel()

	sections := []diffview.Section{
		{Role: "core", Title: "Changes"},
	}

	renderer := lipgloss.NewRenderer(nil, termenv.WithProfile(termenv.Ascii))
	diagram := bubbletea.NarrativeDiagram("unknown-narrative", sections, renderer)

	assert.Empty(t, diagram)
}

func TestNarrativeDiagram_DeduplicatesRoles(t *testing.T) {
	t.Parallel()

	// Multiple sections with same role should show role only once
	sections := []diffview.Section{
		{Role: "fix", Title: "First fix"},
		{Role: "fix", Title: "Second fix"},
		{Role: "test", Title: "Tests"},
	}

	renderer := lipgloss.NewRenderer(nil, termenv.WithProfile(termenv.Ascii))
	diagram := bubbletea.NarrativeDiagram("cause-effect", sections, renderer)

	// Count occurrences of "fix" - should appear only once in diagram
	// The diagram contains borders, so we check the role appears in a box
	assert.Contains(t, diagram, "fix")
	assert.Contains(t, diagram, "test")
	// There should be exactly one arrow (between fix and test)
	count := strings.Count(diagram, "→")
	assert.Equal(t, 1, count, "should have exactly one arrow between two unique roles")
}

func TestNarrativeDiagram_RuleInstances(t *testing.T) {
	t.Parallel()

	// rule-instances narrative with pattern and instance roles
	sections := []diffview.Section{
		{Role: "pattern", Title: "The Pattern"},
		{Role: "instance", Title: "First Application"},
		{Role: "instance", Title: "Second Application"},
	}

	renderer := lipgloss.NewRenderer(nil, termenv.WithProfile(termenv.Ascii))
	diagram := bubbletea.NarrativeDiagram("rule-instances", sections, renderer)

	// Should show flow with pattern and instance
	assert.Contains(t, diagram, "pattern")
	assert.Contains(t, diagram, "instance")
	assert.Contains(t, diagram, "→")
}

func TestNarrativeDiagram_CorePeriphery(t *testing.T) {
	t.Parallel()

	// core-periphery narrative should produce a hub-and-spoke diagram
	// with core in the center and other roles radiating from it
	sections := []diffview.Section{
		{Role: "core", Title: "Main Change"},
		{Role: "supporting", Title: "Ripple Effect"},
		{Role: "test", Title: "Verification"},
	}

	renderer := lipgloss.NewRenderer(nil, termenv.WithProfile(termenv.Ascii))
	diagram := bubbletea.NarrativeDiagram("core-periphery", sections, renderer)

	// Hub-and-spoke should show core with connections to other roles
	assert.Contains(t, diagram, "core")
	assert.Contains(t, diagram, "supporting")
	assert.Contains(t, diagram, "test")
	// Should NOT have linear arrows (that's for linear narratives)
	assert.NotContains(t, diagram, "→")
}

func TestNarrativeDiagram_CorePeriphery_SingleRole(t *testing.T) {
	t.Parallel()

	// With only core role, should still render (even if no spokes)
	sections := []diffview.Section{
		{Role: "core", Title: "Only Core"},
	}

	renderer := lipgloss.NewRenderer(nil, termenv.WithProfile(termenv.Ascii))
	diagram := bubbletea.NarrativeDiagram("core-periphery", sections, renderer)

	assert.Contains(t, diagram, "core")
}

func TestNarrativeDiagram_CorePeriphery_NoCore(t *testing.T) {
	t.Parallel()

	// Without core role, should return empty (can't have hub without hub)
	sections := []diffview.Section{
		{Role: "supporting", Title: "Support Only"},
		{Role: "test", Title: "Test Only"},
	}

	renderer := lipgloss.NewRenderer(nil, termenv.WithProfile(termenv.Ascii))
	diagram := bubbletea.NarrativeDiagram("core-periphery", sections, renderer)

	// No core = no hub-and-spoke diagram
	assert.Empty(t, diagram)
}

func TestNarrativeDiagram_CorePeriphery_ManyRoles(t *testing.T) {
	t.Parallel()

	// With multiple peripheral roles, all should be shown
	sections := []diffview.Section{
		{Role: "core", Title: "Main"},
		{Role: "supporting", Title: "Support 1"},
		{Role: "supporting", Title: "Support 2"},
		{Role: "test", Title: "Tests"},
		{Role: "cleanup", Title: "Cleanup"},
	}

	renderer := lipgloss.NewRenderer(nil, termenv.WithProfile(termenv.Ascii))
	diagram := bubbletea.NarrativeDiagram("core-periphery", sections, renderer)

	assert.Contains(t, diagram, "core")
	assert.Contains(t, diagram, "supporting")
	assert.Contains(t, diagram, "test")
	assert.Contains(t, diagram, "cleanup")
	// Roles should be deduplicated
	count := strings.Count(diagram, "supporting")
	assert.Equal(t, 1, count, "supporting should appear only once")
}

func TestNarrativeDiagram_NilRenderer(t *testing.T) {
	t.Parallel()

	sections := []diffview.Section{
		{Role: "fix", Title: "The Fix"},
	}

	diagram := bubbletea.NarrativeDiagram("cause-effect", sections, nil)

	// Should use default renderer when nil is passed
	assert.Contains(t, diagram, "fix")
}

func TestNarrativeDiagram_CorePeriphery_OnePeripheral_CoreAppearsFirst(t *testing.T) {
	t.Parallel()

	// With only 1 peripheral role, use right-only layout
	// Core should appear before the peripheral role (left-to-right reading order)
	sections := []diffview.Section{
		{Role: "core", Title: "Main Change"},
		{Role: "test", Title: "Verification"},
	}

	renderer := lipgloss.NewRenderer(nil, termenv.WithProfile(termenv.Ascii))
	diagram := bubbletea.NarrativeDiagram("core-periphery", sections, renderer)

	// In right-only layout, core appears before peripheral roles
	corePos := strings.Index(diagram, "core")
	testPos := strings.Index(diagram, "test")

	assert.Less(t, corePos, testPos, "core should appear before test in right-only layout")
}

func TestNarrativeDiagram_CorePeriphery_TwoPeripheral_LeftAndRight(t *testing.T) {
	t.Parallel()

	// With 2 peripheral roles, they should appear on left and right of core
	// Order: right first, then left (so test=right, supporting=left)
	sections := []diffview.Section{
		{Role: "core", Title: "Main Change"},
		{Role: "test", Title: "Verification"},
		{Role: "supporting", Title: "Support"},
	}

	renderer := lipgloss.NewRenderer(nil, termenv.WithProfile(termenv.Ascii))
	diagram := bubbletea.NarrativeDiagram("core-periphery", sections, renderer)

	// All roles should be present
	assert.Contains(t, diagram, "core")
	assert.Contains(t, diagram, "test")
	assert.Contains(t, diagram, "supporting")

	// Layout should be: supporting --[core]-- test
	// So supporting appears before core, and test appears after core
	lines := strings.Split(diagram, "\n")
	var coreLineIdx int
	for i, line := range lines {
		if strings.Contains(line, "core") {
			coreLineIdx = i
			break
		}
	}

	// The core line should have both left and right spokes
	coreLine := lines[coreLineIdx]
	corePos := strings.Index(coreLine, "core")
	supportingPos := strings.Index(coreLine, "supporting")
	testPos := strings.Index(coreLine, "test")

	// supporting should be on the left (before core)
	assert.Greater(t, corePos, supportingPos, "supporting should appear left of core")
	// test should be on the right (after core)
	assert.Less(t, corePos, testPos, "test should appear right of core")
}

func TestNarrativeDiagram_CorePeriphery_ThreePeripheral_LeftRightTop(t *testing.T) {
	t.Parallel()

	// With 3 peripheral roles: right, left, top
	// Order: test=right, supporting=left, cleanup=top
	sections := []diffview.Section{
		{Role: "core", Title: "Main Change"},
		{Role: "test", Title: "Verification"},
		{Role: "supporting", Title: "Support"},
		{Role: "cleanup", Title: "Cleanup"},
	}

	renderer := lipgloss.NewRenderer(nil, termenv.WithProfile(termenv.Ascii))
	diagram := bubbletea.NarrativeDiagram("core-periphery", sections, renderer)

	// All roles should be present
	assert.Contains(t, diagram, "core")
	assert.Contains(t, diagram, "test")
	assert.Contains(t, diagram, "supporting")
	assert.Contains(t, diagram, "cleanup")

	// Should have vertical connector for top spoke
	assert.Contains(t, diagram, "|", "should have vertical connector")

	// cleanup (top) should appear before core in the output (higher line number)
	cleanupPos := strings.Index(diagram, "cleanup")
	corePos := strings.Index(diagram, "core")
	assert.Less(t, cleanupPos, corePos, "cleanup (top) should appear before core in output")
}

func TestNarrativeDiagram_CorePeriphery_FourPeripheral_AllCardinal(t *testing.T) {
	t.Parallel()

	// With 4 peripheral roles: right, left, top, bottom
	sections := []diffview.Section{
		{Role: "core", Title: "Main Change"},
		{Role: "test", Title: "Verification"},
		{Role: "supporting", Title: "Support"},
		{Role: "cleanup", Title: "Cleanup"},
		{Role: "infra", Title: "Infrastructure"},
	}

	renderer := lipgloss.NewRenderer(nil, termenv.WithProfile(termenv.Ascii))
	diagram := bubbletea.NarrativeDiagram("core-periphery", sections, renderer)

	// All roles should be present
	assert.Contains(t, diagram, "core")
	assert.Contains(t, diagram, "test")
	assert.Contains(t, diagram, "supporting")
	assert.Contains(t, diagram, "cleanup")
	assert.Contains(t, diagram, "infra")

	// Should have vertical connectors for top and bottom spokes
	assert.Contains(t, diagram, "|", "should have vertical connectors")

	// Verify vertical positions:
	// - cleanup (top) should appear before core
	// - infra (bottom) should appear after core
	cleanupPos := strings.Index(diagram, "cleanup")
	corePos := strings.Index(diagram, "core")
	infraPos := strings.Index(diagram, "infra")

	assert.Less(t, cleanupPos, corePos, "cleanup (top) should appear before core")
	assert.Greater(t, infraPos, corePos, "infra (bottom) should appear after core")
}

func TestNarrativeDiagram_CorePeriphery_FivePeripheral_FallsBackToRightOnly(t *testing.T) {
	t.Parallel()

	// With 5+ peripheral roles, fall back to simple right-only layout
	sections := []diffview.Section{
		{Role: "core", Title: "Main Change"},
		{Role: "test", Title: "Test"},
		{Role: "supporting", Title: "Support"},
		{Role: "cleanup", Title: "Cleanup"},
		{Role: "infra", Title: "Infrastructure"},
		{Role: "docs", Title: "Documentation"},
	}

	renderer := lipgloss.NewRenderer(nil, termenv.WithProfile(termenv.Ascii))
	diagram := bubbletea.NarrativeDiagram("core-periphery", sections, renderer)

	// All roles should be present
	assert.Contains(t, diagram, "core")
	assert.Contains(t, diagram, "test")
	assert.Contains(t, diagram, "supporting")
	assert.Contains(t, diagram, "cleanup")
	assert.Contains(t, diagram, "infra")
	assert.Contains(t, diagram, "docs")

	// In right-only layout, horizontal spoke connector appears (pointing right)
	// and no vertical connector (no top/bottom)
	assert.Contains(t, diagram, "──", "should have horizontal spoke connectors")

	// Find the line with core - peripheral roles should be on right side
	lines := strings.Split(diagram, "\n")
	for _, line := range lines {
		if strings.Contains(line, "core") {
			// Core line shouldn't have peripheral roles to its left
			// (no "role ──" pattern before core)
			coreIdx := strings.Index(line, "core")
			leftPart := line[:coreIdx]
			// Left part should only contain box drawing characters and spaces
			assert.NotContains(t, leftPart, "test", "test should not be left of core")
			assert.NotContains(t, leftPart, "supporting", "supporting should not be left of core")
			break
		}
	}
}
