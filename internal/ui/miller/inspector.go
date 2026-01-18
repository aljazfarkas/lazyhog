package miller

import (
	"fmt"
	"strings"
	"time"

	"github.com/aljazfarkas/lazyhog/internal/client"
	"github.com/aljazfarkas/lazyhog/internal/ui/components"
	"github.com/aljazfarkas/lazyhog/internal/ui/styles"
)

// renderInspector renders Pane 3 (inspector)
func (m Model) renderInspector(width, height int) string {
	var sb strings.Builder

	// Title
	title := "Inspector"
	if m.inspectorData != nil {
		title = "Details"
	}

	// Show clipboard feedback if recent (2 second TTL)
	if !m.clipboardTime.IsZero() && time.Since(m.clipboardTime) < 2*time.Second {
		title += " - " + m.clipboardMsg
	}

	titleStyled := styles.TitleStyle.Render(title)
	sb.WriteString(titleStyled)
	sb.WriteString("\n\n")

	// Empty state
	if m.inspectorData == nil {
		emptyMsg := "Select an item to view details"
		sb.WriteString(styles.DimTextStyle.Render(emptyMsg))
		sb.WriteString("\n")
	} else {
		// Render based on resource type with scrolling
		switch m.selectedResource {
		case ResourceEvents:
			sb.WriteString(m.renderEventInspectorScrollable(width, height))
		case ResourcePersons:
			sb.WriteString(m.renderPersonInspectorScrollable(width, height))
		case ResourceFlags:
			sb.WriteString(m.renderFlagInspectorScrollable(width, height))
		}
	}

	// Wrap in styled container
	borderStyle := GetBorderStyle(m.focus, 2)
	content := borderStyle.
		Width(width - 2).
		Height(height - 2).
		Padding(1).
		Render(sb.String())

	return content
}

// renderEventInspector renders event details
func (m Model) renderEventInspector(width, height int) string {
	event, ok := m.inspectorData.(client.Event)
	if !ok {
		return styles.ErrorTextStyle.Render("Error: Invalid event data")
	}

	var sb strings.Builder

	// Event header
	sb.WriteString(styles.JSONKeyStyle.Render("Event: "))
	sb.WriteString(event.Event)
	sb.WriteString("\n\n")

	// Basic info
	sb.WriteString(styles.JSONKeyStyle.Render("Timestamp: "))
	sb.WriteString(client.FormatEventTime(event.Timestamp))
	sb.WriteString("\n")

	sb.WriteString(styles.JSONKeyStyle.Render("Distinct ID: "))
	sb.WriteString(event.DistinctID)
	sb.WriteString("\n")

	if event.UUID != "" {
		sb.WriteString(styles.JSONKeyStyle.Render("Event ID: "))
		sb.WriteString(event.UUID)
		sb.WriteString("\n")
	}

	sb.WriteString("\n")
	sb.WriteString(styles.JSONKeyStyle.Render("Properties:"))
	sb.WriteString("\n")

	// Properties
	maxLines := styles.Max(10, height-15)
	propertiesJSON := components.FormatJSONWithColors(event.Properties, maxLines)
	sb.WriteString(propertiesJSON)

	// Add hint for pivot
	sb.WriteString("\n\n")
	sb.WriteString(styles.DimTextStyle.Render("Press 'p' to view this person"))

	return sb.String()
}

// renderPersonInspector renders person details
func (m Model) renderPersonInspector(width, height int) string {
	person, ok := m.inspectorData.(client.Person)
	if !ok {
		return styles.ErrorTextStyle.Render("Error: Invalid person data")
	}

	var sb strings.Builder

	// Person header
	sb.WriteString(styles.JSONKeyStyle.Render("Name: "))
	if person.Name != "" {
		sb.WriteString(person.Name)
	} else {
		sb.WriteString(styles.DimTextStyle.Render("(no name)"))
	}
	sb.WriteString("\n\n")

	// Distinct IDs
	sb.WriteString(styles.JSONKeyStyle.Render("Distinct IDs:"))
	sb.WriteString("\n")
	for _, id := range person.DistinctIDs {
		sb.WriteString(fmt.Sprintf("  • %s\n", id))
	}
	sb.WriteString("\n")

	if person.CreatedAt != "" {
		sb.WriteString(styles.JSONKeyStyle.Render("Created: "))
		sb.WriteString(person.CreatedAt)
		sb.WriteString("\n\n")
	}

	// Properties
	sb.WriteString(styles.JSONKeyStyle.Render("Properties:"))
	sb.WriteString("\n")

	if len(person.Properties) == 0 {
		sb.WriteString(styles.DimTextStyle.Render("  (no properties)"))
		sb.WriteString("\n")
	} else {
		maxLines := styles.Max(10, height-20)
		propertiesJSON := components.FormatJSONWithColors(person.Properties, maxLines)
		sb.WriteString(propertiesJSON)
	}

	return sb.String()
}

// renderFlagInspector renders feature flag details
func (m Model) renderFlagInspector(width, height int) string {
	flag, ok := m.inspectorData.(client.FeatureFlag)
	if !ok {
		return styles.ErrorTextStyle.Render("Error: Invalid flag data")
	}

	var sb strings.Builder

	// Flag header
	sb.WriteString(styles.JSONKeyStyle.Render("Key: "))
	sb.WriteString(flag.Key)
	sb.WriteString("\n\n")

	sb.WriteString(styles.JSONKeyStyle.Render("Name: "))
	if flag.Name != "" {
		sb.WriteString(flag.Name)
	} else {
		sb.WriteString(styles.DimTextStyle.Render("(no name)"))
	}
	sb.WriteString("\n\n")

	// Status
	sb.WriteString(styles.JSONKeyStyle.Render("Status: "))
	if flag.Active {
		sb.WriteString(styles.SuccessTextStyle.Render("Active"))
	} else {
		sb.WriteString(styles.DimTextStyle.Render("Inactive"))
	}
	sb.WriteString("\n\n")

	// Filters (if available)
	if len(flag.Filters) > 0 {
		sb.WriteString(styles.JSONKeyStyle.Render("Filters:"))
		sb.WriteString("\n")
		maxLines := styles.Max(5, height-20)
		filtersJSON := components.FormatJSONWithColors(flag.Filters, maxLines)
		sb.WriteString(filtersJSON)
		sb.WriteString("\n")
	}

	// Created/modified dates if available
	if flag.CreatedAt != "" {
		sb.WriteString(styles.JSONKeyStyle.Render("Created: "))
		sb.WriteString(flag.CreatedAt)
		sb.WriteString("\n")
	}

	return sb.String()
}

// renderEventInspectorScrollable renders event details with scrolling support
func (m Model) renderEventInspectorScrollable(width, height int) string {
	event, ok := m.inspectorData.(client.Event)
	if !ok {
		return styles.ErrorTextStyle.Render("Error: Invalid event data")
	}

	var lines []string

	// Event header
	lines = append(lines, styles.JSONKeyStyle.Render("Event: ")+event.Event)
	lines = append(lines, "")

	// Basic info
	lines = append(lines, styles.JSONKeyStyle.Render("Timestamp: ")+client.FormatEventTime(event.Timestamp))
	lines = append(lines, styles.JSONKeyStyle.Render("Distinct ID: ")+event.DistinctID)

	if event.UUID != "" {
		lines = append(lines, styles.JSONKeyStyle.Render("Event ID: ")+event.UUID)
	}

	lines = append(lines, "")
	lines = append(lines, styles.JSONKeyStyle.Render("Properties:"))

	// Add JSON properties with folding
	jsonLines := m.renderFoldedJSON(event.Properties, 0)
	lines = append(lines, jsonLines...)

	// Add hint for pivot
	lines = append(lines, "")
	lines = append(lines, styles.DimTextStyle.Render("Press 'p' to view this person"))

	// Build full content and update viewport
	content := strings.Join(lines, "\n")
	m.inspectorViewport.Width = width - 4
	m.inspectorViewport.Height = height - 8
	m.inspectorViewport.SetContent(content)

	return m.inspectorViewport.View()
}

// renderPersonInspectorScrollable renders person details with scrolling support
func (m Model) renderPersonInspectorScrollable(width, height int) string {
	person, ok := m.inspectorData.(client.Person)
	if !ok {
		return styles.ErrorTextStyle.Render("Error: Invalid person data")
	}

	var lines []string

	// Person header
	nameValue := person.Name
	if nameValue == "" {
		nameValue = styles.DimTextStyle.Render("(no name)")
	}
	lines = append(lines, styles.JSONKeyStyle.Render("Name: ")+nameValue)
	lines = append(lines, "")

	// Distinct IDs
	lines = append(lines, styles.JSONKeyStyle.Render("Distinct IDs:"))
	for _, id := range person.DistinctIDs {
		lines = append(lines, fmt.Sprintf("  • %s", id))
	}
	lines = append(lines, "")

	if person.CreatedAt != "" {
		lines = append(lines, styles.JSONKeyStyle.Render("Created: ")+person.CreatedAt)
		lines = append(lines, "")
	}

	// Properties
	lines = append(lines, styles.JSONKeyStyle.Render("Properties:"))

	if len(person.Properties) == 0 {
		lines = append(lines, styles.DimTextStyle.Render("  (no properties)"))
	} else {
		jsonLines := m.renderFoldedJSON(person.Properties, 0)
		lines = append(lines, jsonLines...)
	}

	// Build full content and update viewport
	content := strings.Join(lines, "\n")
	m.inspectorViewport.Width = width - 4
	m.inspectorViewport.Height = height - 8
	m.inspectorViewport.SetContent(content)

	return m.inspectorViewport.View()
}

// renderFlagInspectorScrollable renders feature flag details with scrolling support
func (m Model) renderFlagInspectorScrollable(width, height int) string {
	flag, ok := m.inspectorData.(client.FeatureFlag)
	if !ok {
		return styles.ErrorTextStyle.Render("Error: Invalid flag data")
	}

	var lines []string

	// Flag header
	lines = append(lines, styles.JSONKeyStyle.Render("Key: ")+flag.Key)
	lines = append(lines, "")

	nameValue := flag.Name
	if nameValue == "" {
		nameValue = styles.DimTextStyle.Render("(no name)")
	}
	lines = append(lines, styles.JSONKeyStyle.Render("Name: ")+nameValue)
	lines = append(lines, "")

	// Status
	statusValue := styles.SuccessTextStyle.Render("Active")
	if !flag.Active {
		statusValue = styles.DimTextStyle.Render("Inactive")
	}
	lines = append(lines, styles.JSONKeyStyle.Render("Status: ")+statusValue)
	lines = append(lines, "")

	// Filters (if available)
	if len(flag.Filters) > 0 {
		lines = append(lines, styles.JSONKeyStyle.Render("Filters:"))
		jsonLines := m.renderFoldedJSON(flag.Filters, 0)
		lines = append(lines, jsonLines...)
		lines = append(lines, "")
	}

	// Created/modified dates if available
	if flag.CreatedAt != "" {
		lines = append(lines, styles.JSONKeyStyle.Render("Created: ")+flag.CreatedAt)
	}

	// Build full content and update viewport
	content := strings.Join(lines, "\n")
	m.inspectorViewport.Width = width - 4
	m.inspectorViewport.Height = height - 8
	m.inspectorViewport.SetContent(content)

	return m.inspectorViewport.View()
}


