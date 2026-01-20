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

// renderProjectSection renders the project selector at the top of Pane 1
func (m Model) renderProjectSection() string {
	var sb strings.Builder

	// Project header with icon
	header := "📁 Project"
	sb.WriteString(styles.DimTextStyle.Render(header))
	sb.WriteString("\n")

	// Project name (highlight if cursor is on project)
	projectName := "Loading..."
	if m.projectsLoaded {
		if len(m.availableProjects) == 0 {
			projectName = "No projects"
		} else if m.selectedProjectID == 0 {
			projectName = "No selection"
		} else {
			// Find current project name
			for _, proj := range m.availableProjects {
				if proj.ID == m.selectedProjectID {
					projectName = proj.Name
					break
				}
			}
			// If still not found, show ID
			if projectName == "Loading..." {
				projectName = fmt.Sprintf("Project #%d", m.selectedProjectID)
			}
		}
	}

	// Highlight if cursor is on project
	isSelected := (m.focus == FocusPane1 && m.pane1Cursor == pane1CursorProject)

	if isSelected {
		projectLine := styles.SelectedListItemStyle.Render("▶ " + projectName)
		sb.WriteString(projectLine)
	} else {
		projectLine := styles.ListItemStyle.Render("  " + projectName)
		sb.WriteString(projectLine)
	}

	return sb.String()
}

// renderResourceSelector renders Pane 1 (resource selector)
func (m Model) renderResourceSelector(width, height int) string {
	var sb strings.Builder

	// Add project section at top
	sb.WriteString(m.renderProjectSection())
	sb.WriteString("\n\n") // Extra spacing between sections

	// Resource rendering
	resources := []Resource{ResourceEvents, ResourcePersons, ResourceFlags}

	for i, resource := range resources {
		// Check if THIS resource is selected based on cursor position
		selected := (m.focus == FocusPane1 && m.pane1Cursor == i)

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
