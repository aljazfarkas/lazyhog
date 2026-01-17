package miller

import (
	"fmt"
	"strings"

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
	titleStyled := styles.TitleStyle.Render(title)
	sb.WriteString(titleStyled)
	sb.WriteString("\n\n")

	// Empty state
	if m.inspectorData == nil {
		emptyMsg := "Select an item to view details"
		sb.WriteString(styles.DimTextStyle.Render(emptyMsg))
		sb.WriteString("\n")
	} else {
		// Render based on resource type
		switch m.selectedResource {
		case ResourceEvents:
			sb.WriteString(m.renderEventInspector(width, height))
		case ResourcePersons:
			sb.WriteString(m.renderPersonInspector(width, height))
		case ResourceFlags:
			sb.WriteString(m.renderFlagInspector(width, height))
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
