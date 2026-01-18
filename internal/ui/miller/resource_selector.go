package miller

import (
	"github.com/aljazfarkas/lazyhog/internal/ui/styles"
)

// Resource represents a type of data to display
type Resource int

const (
	ResourceEvents Resource = iota
	ResourcePersons
	ResourceFlags
)

// String returns a human-readable representation of the resource
func (r Resource) String() string {
	switch r {
	case ResourceEvents:
		return "Events"
	case ResourcePersons:
		return "Persons"
	case ResourceFlags:
		return "Flags"
	default:
		return "Unknown"
	}
}

// Icon returns the emoji icon for the resource
func (r Resource) Icon() string {
	switch r {
	case ResourceEvents:
		return "📡"
	case ResourcePersons:
		return "👤"
	case ResourceFlags:
		return "🚩"
	default:
		return "❓"
	}
}


// renderResourceSelector renders Pane 1 (resource selector)
func (m Model) renderResourceSelector(width, height int) string {
	// Phase 3 - Use bubbles/list sidebar if available
	if m.sidebar != nil {
		// Calculate available height for sidebar
		// Account for polling indicator (2 lines with spacing)
		extraLines := 0
		if m.selectedResource == ResourceEvents {
			extraLines = 2
		}

		// height - 2 for border, - 2 for padding, - extraLines for indicator
		sidebarHeight := height - 4 - extraLines
		if sidebarHeight < 5 {
			sidebarHeight = 5
		}

		// Ensure sidebar is properly sized
		m.sidebar.SetSize(width-4, sidebarHeight)
		m.sidebar.SetCollapsed(m.pane1Collapsed)

		// Get sidebar view
		sidebarContent := m.sidebar.View()

		// Add polling indicator if on Events
		if m.selectedResource == ResourceEvents {
			var indicator string
			if m.isPolling {
				indicator = styles.SuccessTextStyle.Render("● Live")
			} else {
				indicator = styles.DimTextStyle.Render("⏸ Paused")
			}
			sidebarContent += "\n\n" + indicator
		}

		// Wrap in border
		borderStyle := GetBorderStyle(m.focus, 0)
		content := borderStyle.
			Width(width - 2).
			Height(height - 2).
			Padding(1).
			Render(sidebarContent)

		return content
	}

	// Sidebar should always be initialized, but as a safety measure
	return ""
}

