package components

import (
	"strings"

	"github.com/aljazfarkas/lazyhog/internal/ui/styles"
	"github.com/charmbracelet/lipgloss"
)

// SidebarItem represents a navigation item
type SidebarItem struct {
	Key   string
	Label string
	Icon  string
}

// RenderSidebar renders a navigation sidebar
func RenderSidebar(items []SidebarItem, selectedIdx int, width, height int) string {
	var sb strings.Builder

	// Title
	title := lipgloss.NewStyle().
		Foreground(styles.ColorPrimary).
		Bold(true).
		Padding(1, 2).
		Render("lazyhog 🦔")

	sb.WriteString(title)
	sb.WriteString("\n\n")

	// Navigation items
	for i, item := range items {
		var line string
		icon := item.Icon
		label := item.Label

		if i == selectedIdx {
			// Selected item
			line = styles.SelectedListItemStyle.Render("▶ " + icon + " " + label)
		} else {
			// Unselected item
			line = lipgloss.NewStyle().
				Foreground(styles.ColorDim).
				Padding(0, 2).
				Render("  " + icon + " " + label)
		}

		sb.WriteString(line)
		sb.WriteString("\n")
	}

	// Spacer
	sb.WriteString("\n")

	// Help text at bottom
	helpStyle := lipgloss.NewStyle().
		Foreground(styles.ColorDimmer).
		Padding(0, 2)

	helpLines := []string{
		"",
		"Navigation:",
		"↑/↓ or j/k",
		"  switch view",
		"",
		"Tab or →",
		"  focus panel",
		"",
		"q: quit",
	}

	for _, line := range helpLines {
		sb.WriteString(helpStyle.Render(line))
		sb.WriteString("\n")
	}

	// Render in a bordered container
	sidebarStyle := lipgloss.NewStyle().
		Width(width - 2).
		Height(height - 2).
		Border(lipgloss.NormalBorder()).
		BorderForeground(styles.ColorBorder).
		Padding(0, 0)

	return sidebarStyle.Render(sb.String())
}
