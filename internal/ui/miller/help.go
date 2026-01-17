package miller

import (
	"strings"

	"github.com/aljazfarkas/lazyhog/internal/ui/styles"
	"github.com/charmbracelet/lipgloss"
)

// renderHelpOverlay renders a full-screen help cheat sheet
func (m Model) renderHelpOverlay(width, height int) string {
	var sb strings.Builder

	// Title
	title := styles.TitleStyle.Render("⌨️  Keyboard Shortcuts")
	sb.WriteString(title)
	sb.WriteString("\n\n")

	// Create sections
	sections := []struct {
		title string
		items [][]string // [key, description]
	}{
		{
			title: "Global",
			items: [][]string{
				{"?", "Toggle this help overlay"},
				{"q", "Quit app"},
				{"Ctrl+C", "Force quit"},
				{"Tab / → / l", "Move focus right"},
				{"Shift+Tab / ← / h / Esc", "Move focus left / Go back"},
			},
		},
		{
			title: "Resource Selector (Pane 1)",
			items: [][]string{
				{"↑/↓ or j/k", "Navigate projects and resources"},
				{"Enter", "Select resource (or cycle projects)"},
				{"1", "Quick select Events"},
				{"2", "Quick select Persons"},
				{"3", "Quick select Flags"},
			},
		},
		{
			title: "List View (Pane 2)",
			items: [][]string{
				{"↑/↓ or j/k", "Navigate list (auto-updates Inspector)"},
				{"G", "Jump to bottom (resume auto-scroll)"},
				{"/", "Search/filter (modal)"},
				{"r", "Refresh current resource"},
				{"p", "Pivot to person (Events only)"},
			},
		},
		{
			title: "Inspector (Pane 3)",
			items: [][]string{
				{"j/k or ↑/↓", "Scroll content"},
				{"Space", "Fold/expand JSON object at cursor"},
				{"Shift+Z", "Fold/expand all top-level keys"},
				{"y", "Copy full JSON to clipboard"},
				{"c", "Copy ID to clipboard"},
				{"p", "Pivot to person (Events only)"},
			},
		},
		{
			title: "Search Mode",
			items: [][]string{
				{"Type", "Enter search query"},
				{"Enter", "Apply filter and jump to first result"},
				{"Esc", "Cancel and restore full list"},
			},
		},
	}

	// Render each section
	for i, section := range sections {
		// Section title
		sectionTitle := styles.JSONKeyStyle.Render(section.title)
		sb.WriteString(sectionTitle)
		sb.WriteString("\n")

		// Section items
		for _, item := range section.items {
			key := item[0]
			desc := item[1]

			// Format: "  key........description"
			keyStyled := styles.HighlightTextStyle.Render(key)
			padding := strings.Repeat(" ", 20-len(key))
			line := "  " + keyStyled + padding + desc
			sb.WriteString(line)
			sb.WriteString("\n")
		}

		// Add spacing between sections
		if i < len(sections)-1 {
			sb.WriteString("\n")
		}
	}

	// Footer
	sb.WriteString("\n")
	footer := styles.DimTextStyle.Render("Press ? or Esc to close this help")
	sb.WriteString(footer)

	// Wrap in a bordered box
	helpStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.ColorPrimary).
		Padding(2, 4).
		Width(width - 4).
		Height(height - 4)

	return helpStyle.Render(sb.String())
}
