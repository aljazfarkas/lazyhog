package miller

import (
	"fmt"
	"strings"

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
	var sb strings.Builder

	resources := []Resource{ResourceEvents, ResourcePersons, ResourceFlags}

	for i, resource := range resources {
		selected := (resource == m.selectedResource)

		icon := resource.Icon()
		label := resource.String()

		line := fmt.Sprintf("%s %s", icon, label)

		if selected {
			line = styles.SelectedListItemStyle.Render("▶ " + line)
		} else {
			line = styles.ListItemStyle.Render("  " + line)
		}

		sb.WriteString(line)
		if i < len(resources)-1 {
			sb.WriteString("\n")
		}
	}

	// Add polling indicator if on Events
	if m.selectedResource == ResourceEvents && m.isPolling {
		sb.WriteString("\n\n")
		indicator := styles.SuccessTextStyle.Render("● Live")
		sb.WriteString(indicator)
	} else if m.selectedResource == ResourceEvents && !m.isPolling {
		sb.WriteString("\n\n")
		indicator := styles.DimTextStyle.Render("⏸ Paused")
		sb.WriteString(indicator)
	}

	// Wrap in styled container
	borderStyle := GetBorderStyle(m.focus, 0)
	content := borderStyle.
		Width(width - 2).
		Height(height - 2).
		Padding(1).
		Render(sb.String())

	return content
}

// MoveResourceSelectorUp moves the resource selection up
func (m *Model) MoveResourceSelectorUp() {
	if m.selectedResource > ResourceEvents {
		m.selectedResource--
	}
}

// MoveResourceSelectorDown moves the resource selection down
func (m *Model) MoveResourceSelectorDown() {
	if m.selectedResource < ResourceFlags {
		m.selectedResource++
	}
}
