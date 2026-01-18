package miller

import (
	"fmt"
	"strings"

	"github.com/aljazfarkas/lazyhog/internal/client"
	"github.com/aljazfarkas/lazyhog/internal/ui/components"
	"github.com/aljazfarkas/lazyhog/internal/ui/styles"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

// InspectorModel wraps bubbles/viewport for Pane 3 (Phase 5)
type InspectorModel struct {
	viewport   viewport.Model
	data       interface{}
	resource   Resource
	foldState  map[string]bool
	allFolded  bool
	width      int
	height     int
}

// NewInspectorModel creates a new inspector viewport model
func NewInspectorModel(width, height int) InspectorModel {
	vp := viewport.New(width, height)
	vp.Style = styles.ListItemStyle // Base style for content

	return InspectorModel{
		viewport:  vp,
		data:      nil,
		resource:  ResourceEvents,
		foldState: make(map[string]bool),
		allFolded: false,
		width:     width,
		height:    height,
	}
}

// SetContent updates the inspector with new data
func (i *InspectorModel) SetContent(data interface{}, resource Resource) {
	i.data = data
	i.resource = resource

	// Generate content based on resource type
	content := i.generateContent()
	i.viewport.SetContent(content)

	// Reset scroll to top when content changes
	i.viewport.GotoTop()
}

// generateContent creates the formatted content for the viewport
func (i InspectorModel) generateContent() string {
	if i.data == nil {
		return ""
	}

	var sb strings.Builder

	switch i.resource {
	case ResourceEvents:
		if event, ok := i.data.(client.Event); ok {
			sb.WriteString(i.formatEvent(event))
		}
	case ResourcePersons:
		if person, ok := i.data.(client.Person); ok {
			sb.WriteString(i.formatPerson(person))
		}
	case ResourceFlags:
		if flag, ok := i.data.(client.FeatureFlag); ok {
			sb.WriteString(i.formatFlag(flag))
		}
	}

	return sb.String()
}

// formatEvent formats an event for display
func (i InspectorModel) formatEvent(event client.Event) string {
	var sb strings.Builder

	// Event header
	sb.WriteString(styles.JSONKeyStyle.Render("Event: "))
	sb.WriteString(styles.JSONValueStyle.Render(event.Event))
	sb.WriteString("\n\n")

	// Basic info
	sb.WriteString(styles.JSONKeyStyle.Render("Timestamp: "))
	sb.WriteString(styles.JSONValueStyle.Render(client.FormatEventTime(event.Timestamp)))
	sb.WriteString("\n")

	sb.WriteString(styles.JSONKeyStyle.Render("Distinct ID: "))
	sb.WriteString(styles.JSONValueStyle.Render(event.DistinctID))
	sb.WriteString("\n")

	if event.UUID != "" {
		sb.WriteString(styles.JSONKeyStyle.Render("Event ID: "))
		sb.WriteString(styles.JSONValueStyle.Render(event.UUID))
		sb.WriteString("\n")
	}

	sb.WriteString("\n")
	sb.WriteString(styles.JSONKeyStyle.Render("Properties:"))
	sb.WriteString("\n")

	// Properties with folding support
	propertiesJSON := components.FormatJSONWithColors(event.Properties, 0) // No line limit in viewport
	sb.WriteString(propertiesJSON)

	// Add hint for pivot
	sb.WriteString("\n\n")
	sb.WriteString(styles.DimTextStyle.Render("Press 'p' to view this person"))

	return sb.String()
}

// formatPerson formats a person for display
func (i InspectorModel) formatPerson(person client.Person) string {
	var sb strings.Builder

	// Person header
	name := person.Name
	if name == "" {
		name = "(no name)"
	}
	sb.WriteString(styles.JSONKeyStyle.Render("Person: "))
	sb.WriteString(styles.JSONValueStyle.Render(name))
	sb.WriteString("\n\n")

	// Distinct IDs
	if len(person.DistinctIDs) > 0 {
		sb.WriteString(styles.JSONKeyStyle.Render("Distinct IDs:"))
		sb.WriteString("\n")
		for _, id := range person.DistinctIDs {
			sb.WriteString("  • ")
			sb.WriteString(styles.JSONValueStyle.Render(id))
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}

	// Created at
	if person.CreatedAt != "" {
		sb.WriteString(styles.JSONKeyStyle.Render("Created: "))
		sb.WriteString(styles.JSONValueStyle.Render(person.CreatedAt))
		sb.WriteString("\n\n")
	}

	// Properties
	sb.WriteString(styles.JSONKeyStyle.Render("Properties:"))
	sb.WriteString("\n")
	propertiesJSON := components.FormatJSONWithColors(person.Properties, 0)
	sb.WriteString(propertiesJSON)

	return sb.String()
}

// formatFlag formats a feature flag for display
func (i InspectorModel) formatFlag(flag client.FeatureFlag) string {
	var sb strings.Builder

	// Flag header
	sb.WriteString(styles.JSONKeyStyle.Render("Flag: "))
	sb.WriteString(styles.JSONValueStyle.Render(flag.Key))
	sb.WriteString("\n\n")

	// Status
	sb.WriteString(styles.JSONKeyStyle.Render("Status: "))
	if flag.Active {
		sb.WriteString(styles.SuccessTextStyle.Render("✓ Active"))
	} else {
		sb.WriteString(styles.ErrorTextStyle.Render("✗ Inactive"))
	}
	sb.WriteString("\n")

	// ID
	sb.WriteString(styles.JSONKeyStyle.Render("ID: "))
	sb.WriteString(styles.JSONValueStyle.Render(fmt.Sprintf("%d", flag.ID)))
	sb.WriteString("\n")

	// Name
	if flag.Name != "" {
		sb.WriteString(styles.JSONKeyStyle.Render("Name: "))
		sb.WriteString(styles.JSONValueStyle.Render(flag.Name))
		sb.WriteString("\n")
	}

	// Filters
	if len(flag.Filters) > 0 {
		sb.WriteString("\n")
		sb.WriteString(styles.JSONKeyStyle.Render("Filters:"))
		sb.WriteString("\n")
		filtersJSON := components.FormatJSONWithColors(flag.Filters, 0)
		sb.WriteString(filtersJSON)
	}

	sb.WriteString("\n\n")
	sb.WriteString(styles.DimTextStyle.Render("Press Space to toggle this flag"))

	return sb.String()
}

// Update handles viewport updates
func (i InspectorModel) Update(msg tea.Msg) (InspectorModel, tea.Cmd) {
	var cmd tea.Cmd
	i.viewport, cmd = i.viewport.Update(msg)
	return i, cmd
}

// View renders the viewport
func (i InspectorModel) View() string {
	return i.viewport.View()
}

// SetSize updates the viewport size
func (i *InspectorModel) SetSize(width, height int) {
	i.width = width
	i.height = height
	i.viewport.Width = width
	i.viewport.Height = height
}

// ScrollDown scrolls the viewport down
func (i *InspectorModel) ScrollDown() {
	i.viewport.LineDown(1)
}

// ScrollUp scrolls the viewport up
func (i *InspectorModel) ScrollUp() {
	i.viewport.LineUp(1)
}

// PageDown scrolls down by a page
func (i *InspectorModel) PageDown() {
	i.viewport.ViewDown()
}

// PageUp scrolls up by a page
func (i *InspectorModel) PageUp() {
	i.viewport.ViewUp()
}

// GotoTop scrolls to the top
func (i *InspectorModel) GotoTop() {
	i.viewport.GotoTop()
}

// GotoBottom scrolls to the bottom
func (i *InspectorModel) GotoBottom() {
	i.viewport.GotoBottom()
}

// GetScrollPercent returns the current scroll percentage
func (i InspectorModel) GetScrollPercent() float64 {
	return i.viewport.ScrollPercent()
}

// ToggleFoldAtCursor toggles JSON folding at cursor (Phase 5 - Future enhancement)
func (i *InspectorModel) ToggleFoldAtCursor() {
	// TODO: Implement JSON folding with viewport
	// This requires tracking cursor position within JSON structure
}

// FoldAll folds all top-level JSON keys (Phase 5 - Future enhancement)
func (i *InspectorModel) FoldAll() {
	// TODO: Implement fold all with viewport
	i.allFolded = !i.allFolded
	// Regenerate content with folded state
	content := i.generateContent()
	i.viewport.SetContent(content)
}
