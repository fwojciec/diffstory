package bubbletea

import (
	"github.com/charmbracelet/lipgloss"
	diffview "github.com/fwojciec/diffstory"
)

// NarrativeDiagram returns a visual representation of the story's narrative flow.
// The diagram adapts to the roles present in the sections.
// If renderer is nil, a default renderer is used.
func NarrativeDiagram(narrative string, sections []diffview.Section, renderer *lipgloss.Renderer) string {
	if len(sections) == 0 {
		return ""
	}
	// Use default renderer if nil (same pattern as newStyle in story.go)
	if renderer == nil {
		renderer = lipgloss.DefaultRenderer()
	}

	switch narrative {
	case "cause-effect", "entry-implementation", "before-after", "rule-instances":
		return linearFlowDiagram(sections, renderer)
	case "core-periphery":
		return hubAndSpokeDiagram(sections, renderer)
	default:
		return ""
	}
}

// linearFlowDiagram renders roles as a horizontal flow: role1 → role2 → role3
func linearFlowDiagram(sections []diffview.Section, renderer *lipgloss.Renderer) string {
	roles := extractRoles(sections)
	if len(roles) == 0 {
		return ""
	}

	nodeStyle := renderer.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(0, 1)

	arrow := " → "

	// Pre-allocate: each role gets a node + arrow, minus trailing arrow
	parts := make([]string, 0, len(roles)*2-1)
	for _, role := range roles {
		parts = append(parts, nodeStyle.Render(role))
		parts = append(parts, arrow)
	}
	// Remove trailing arrow
	if len(parts) > 0 {
		parts = parts[:len(parts)-1]
	}

	return lipgloss.JoinHorizontal(lipgloss.Center, parts...)
}

// hubAndSpokeDiagram renders a hub-and-spoke diagram with core in the center.
// For 2-4 peripheral roles, uses cardinal directions (right, left, top, bottom).
// For 1 or 5+ peripheral roles, falls back to right-only layout.
//
// Example output with 4 peripheral (with rounded borders):
//
//	          cleanup
//	             |
//	supporting ──╭──────╮── test
//	             │ core │
//	             ╰──────╯
//	             |
//	           infra
func hubAndSpokeDiagram(sections []diffview.Section, renderer *lipgloss.Renderer) string {
	roles := extractRoles(sections)
	if len(roles) == 0 {
		return ""
	}

	// Find core role - it's the hub
	var hasCore bool
	var peripheralRoles []string
	for _, role := range roles {
		if role == "core" {
			hasCore = true
		} else {
			peripheralRoles = append(peripheralRoles, role)
		}
	}

	// No core = no hub-and-spoke diagram
	if !hasCore {
		return ""
	}

	nodeStyle := renderer.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(0, 1)

	// Core is always centered
	coreNode := nodeStyle.Render("core")

	// If no peripheral roles, just show core
	if len(peripheralRoles) == 0 {
		return coreNode
	}

	// For 2-4 peripheral roles, use cardinal directions
	// For 1 or 5+ peripheral roles, use right-only layout
	if len(peripheralRoles) >= 2 && len(peripheralRoles) <= 4 {
		return cardinalHubAndSpoke(coreNode, peripheralRoles, renderer)
	}

	return rightOnlyHubAndSpoke(coreNode, peripheralRoles)
}

// rightOnlyHubAndSpoke renders all peripheral roles on the right side of core.
func rightOnlyHubAndSpoke(coreNode string, peripheralRoles []string) string {
	spoke := "── "

	rightParts := make([]string, 0, len(peripheralRoles))
	for _, role := range peripheralRoles {
		rightParts = append(rightParts, spoke+role)
	}
	rightColumn := lipgloss.JoinVertical(lipgloss.Left, rightParts...)

	return lipgloss.JoinHorizontal(lipgloss.Center, coreNode, rightColumn)
}

// cardinalHubAndSpoke renders peripheral roles in cardinal directions around core.
// Placement order: right, left, top, bottom.
func cardinalHubAndSpoke(coreNode string, peripheralRoles []string, renderer *lipgloss.Renderer) string {
	// Assign roles to positions: right, left, top, bottom
	var right, left, top, bottom string
	for i, role := range peripheralRoles {
		switch i {
		case 0:
			right = role
		case 1:
			left = role
		case 2:
			top = role
		case 3:
			bottom = role
		}
	}

	hSpoke := " ──"
	vSpoke := "|"

	// Build middle row: [left spoke] + [core] + [right spoke]
	var middleParts []string
	if left != "" {
		middleParts = append(middleParts, left+hSpoke)
	}
	middleParts = append(middleParts, coreNode)
	if right != "" {
		middleParts = append(middleParts, hSpoke+right)
	}
	middleRow := lipgloss.JoinHorizontal(lipgloss.Center, middleParts...)

	// If no top/bottom, just return middle row
	if top == "" && bottom == "" {
		return middleRow
	}

	// Calculate width of middle row for centering top/bottom
	middleWidth := lipgloss.Width(middleRow)

	// Build top section (role + connector) if present
	var topSection string
	if top != "" {
		topLabel := renderer.NewStyle().Width(middleWidth).Align(lipgloss.Center).Render(top)
		topConnector := renderer.NewStyle().Width(middleWidth).Align(lipgloss.Center).Render(vSpoke)
		topSection = lipgloss.JoinVertical(lipgloss.Center, topLabel, topConnector)
	}

	// Build bottom section (connector + role) if present
	var bottomSection string
	if bottom != "" {
		bottomConnector := renderer.NewStyle().Width(middleWidth).Align(lipgloss.Center).Render(vSpoke)
		bottomLabel := renderer.NewStyle().Width(middleWidth).Align(lipgloss.Center).Render(bottom)
		bottomSection = lipgloss.JoinVertical(lipgloss.Center, bottomConnector, bottomLabel)
	}

	// Compose final diagram
	var rows []string
	if topSection != "" {
		rows = append(rows, topSection)
	}
	rows = append(rows, middleRow)
	if bottomSection != "" {
		rows = append(rows, bottomSection)
	}

	return lipgloss.JoinVertical(lipgloss.Center, rows...)
}

// extractRoles returns unique roles from sections in order.
func extractRoles(sections []diffview.Section) []string {
	var roles []string
	seen := make(map[string]bool)
	for _, s := range sections {
		if s.Role != "" && !seen[s.Role] {
			roles = append(roles, s.Role)
			seen[s.Role] = true
		}
	}
	return roles
}
