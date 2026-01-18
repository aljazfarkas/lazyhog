package miller

import (
	"fmt"
	"strings"

	"github.com/aljazfarkas/lazyhog/internal/client"
	"github.com/aljazfarkas/lazyhog/internal/ui/styles"
	"github.com/charmbracelet/lipgloss"
)

// ListItem represents an item that can be displayed in Pane 2
type ListItem interface {
	RenderLine(width int, selected bool) string
	GetID() string
	GetInspectorData() interface{}
	GetDistinctID() string // For pivot feature
}

// EventListItem wraps a client.Event for list display
type EventListItem struct {
	Event client.Event
}

func (e EventListItem) RenderLine(width int, selected bool) string {
	timeStr := client.FormatEventTimeShort(e.Event.Timestamp)
	eventName := e.Event.Event
	distinctID := e.Event.DistinctID

	// Truncate if needed
	maxEventNameLen := 25
	maxDistinctIDLen := 20

	if len(eventName) > maxEventNameLen {
		eventName = styles.TruncateString(eventName, maxEventNameLen)
	}
	if len(distinctID) > maxDistinctIDLen {
		distinctID = styles.TruncateString(distinctID, maxDistinctIDLen)
	}

	timeStyled := styles.DimTextStyle.Render(timeStr)
	eventStyled := eventName
	idStyled := styles.DimTextStyle.Render(distinctID)

	line := fmt.Sprintf("%s %s %s", timeStyled, eventStyled, idStyled)

	if selected {
		line = styles.SelectedListItemStyle.Render("▶ " + line)
	} else {
		line = styles.ListItemStyle.Render("  " + line)
	}

	return line
}

func (e EventListItem) GetID() string {
	if e.Event.UUID != "" {
		return e.Event.UUID
	}
	return e.Event.ID
}

func (e EventListItem) GetInspectorData() interface{} {
	return e.Event
}

func (e EventListItem) GetDistinctID() string {
	return e.Event.DistinctID
}

// PersonListItem wraps a client.Person for list display
type PersonListItem struct {
	Person client.Person
}

func (p PersonListItem) RenderLine(width int, selected bool) string {
	name := p.Person.Name
	if name == "" {
		name = "(no name)"
	}

	distinctID := ""
	if len(p.Person.DistinctIDs) > 0 {
		distinctID = p.Person.DistinctIDs[0]
	}

	// Truncate if needed
	maxNameLen := 25
	maxDistinctIDLen := 20

	if len(name) > maxNameLen {
		name = styles.TruncateString(name, maxNameLen)
	}
	if len(distinctID) > maxDistinctIDLen {
		distinctID = styles.TruncateString(distinctID, maxDistinctIDLen)
	}

	line := fmt.Sprintf("%s %s", name, styles.DimTextStyle.Render(distinctID))

	if selected {
		line = styles.SelectedListItemStyle.Render("▶ " + line)
	} else {
		line = styles.ListItemStyle.Render("  " + line)
	}

	return line
}

func (p PersonListItem) GetID() string {
	if len(p.Person.DistinctIDs) > 0 {
		return p.Person.DistinctIDs[0]
	}
	return p.Person.ID
}

func (p PersonListItem) GetInspectorData() interface{} {
	return p.Person
}

func (p PersonListItem) GetDistinctID() string {
	if len(p.Person.DistinctIDs) > 0 {
		return p.Person.DistinctIDs[0]
	}
	return ""
}

// FlagListItem wraps a client.FeatureFlag for list display
type FlagListItem struct {
	Flag client.FeatureFlag
}

func (f FlagListItem) RenderLine(width int, selected bool) string {
	key := f.Flag.Key

	// Status indicator
	status := "○"
	if f.Flag.Active {
		status = "●"
	}

	// Truncate if needed
	maxKeyLen := 35

	if len(key) > maxKeyLen {
		key = styles.TruncateString(key, maxKeyLen)
	}

	line := fmt.Sprintf("%s %s", status, key)

	if selected {
		line = styles.SelectedListItemStyle.Render("▶ " + line)
	} else {
		line = styles.ListItemStyle.Render("  " + line)
	}

	return line
}

func (f FlagListItem) GetID() string {
	return fmt.Sprintf("%d", f.Flag.ID)
}

func (f FlagListItem) GetInspectorData() interface{} {
	return f.Flag
}

func (f FlagListItem) GetDistinctID() string {
	return "" // Flags don't have distinct IDs
}

// renderListView renders Pane 2 (list view)
func (m Model) renderListView(width, height int) string {
	var sb strings.Builder

	// Title based on resource type with auto-scroll indicator
	title := m.selectedResource.String()
	if m.selectedResource == ResourceEvents {
		indicator := m.getAutoScrollIndicator()
		title += " " + indicator
	}
	titleStyled := styles.TitleStyle.Render(title)
	sb.WriteString(titleStyled)
	sb.WriteString("\n\n")

	// Search input overlay if active
	if m.searchMode {
		searchInput := m.renderSearchInput(width)
		sb.WriteString(searchInput)
		sb.WriteString("\n")
	}

	// Error state
	if m.err != nil {
		sb.WriteString(styles.ErrorTextStyle.Render(fmt.Sprintf("Error: %v", m.err)))
		sb.WriteString("\n\n")
		sb.WriteString(styles.DimTextStyle.Render("Press 'r' to retry"))
		sb.WriteString("\n")
	} else {
		// Use effective list items (filtered or all)
		effectiveItems := m.getEffectiveListItems()

		if len(effectiveItems) == 0 {
			// Empty state or no search results
			if m.filteredItems != nil {
				sb.WriteString(styles.DimTextStyle.Render("No matches found"))
			} else {
				emptyMsg := m.getEmptyStateMessage()
				sb.WriteString(styles.DimTextStyle.Render(emptyMsg))
			}
			sb.WriteString("\n")
		} else {
			// Render list items with viewport management
			visibleHeight := height - 8
			if m.searchMode {
				visibleHeight -= 2 // Account for search input overlay
			}
			if visibleHeight < 5 {
				visibleHeight = 5
			}

			start := m.listCursor - visibleHeight/2
			if start < 0 {
				start = 0
			}
			end := start + visibleHeight
			if end > len(effectiveItems) {
				end = len(effectiveItems)
				start = end - visibleHeight
				if start < 0 {
					start = 0
				}
			}

			for i := start; i < end; i++ {
				item := effectiveItems[i]
				line := item.RenderLine(width-4, i == m.listCursor)
				sb.WriteString(line)
				if i < end-1 {
					sb.WriteString("\n")
				}
			}
		}
	}

	// Wrap in styled container
	borderStyle := GetBorderStyle(m.focus, 1)
	content := borderStyle.
		Width(width - 2).
		Height(height - 2).
		Padding(1).
		Render(sb.String())

	return content
}

// getEmptyStateMessage returns the appropriate empty state message
func (m Model) getEmptyStateMessage() string {
	var icon, message, hint string

	switch m.selectedResource {
	case ResourceEvents:
		icon = "📡"
		message = "No events yet"
		hint = "Events will appear here as they're captured by PostHog"
		if m.loading {
			return lipgloss.JoinVertical(lipgloss.Left,
				styles.H1Style.Render(icon),
				m.spinner.View()+" Loading events...",
			)
		}
	case ResourcePersons:
		icon = "👤"
		message = "No persons found"
		hint = "Persons are created when events are sent with distinct IDs"
		if m.loading {
			return lipgloss.JoinVertical(lipgloss.Left,
				styles.H1Style.Render(icon),
				m.spinner.View()+" Loading persons...",
			)
		}
	case ResourceFlags:
		icon = "🚩"
		message = "No feature flags"
		hint = "Create flags in PostHog to manage feature rollouts"
		if m.loading {
			return lipgloss.JoinVertical(lipgloss.Left,
				styles.H1Style.Render(icon),
				m.spinner.View()+" Loading flags...",
			)
		}
	default:
		return "No data available."
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		styles.H1Style.Render(icon),
		styles.H2Style.Render(message),
		"",
		styles.CaptionStyle.Render(hint),
	)
}

// MoveListCursorUp moves the list cursor up and updates inspector
func (m *Model) MoveListCursorUp() {
	if m.listCursor > 0 {
		m.listCursor--
		m.updateInspectorFromCursor()
	}
}

// MoveListCursorDown moves the list cursor down and updates inspector
func (m *Model) MoveListCursorDown() {
	effectiveItems := m.getEffectiveListItems()
	if m.listCursor < len(effectiveItems)-1 {
		m.listCursor++
		m.updateInspectorFromCursor()
	}
}

// SelectCurrentListItem populates Pane 3 with the selected item's data
func (m *Model) SelectCurrentListItem() {
	effectiveItems := m.getEffectiveListItems()
	if len(effectiveItems) == 0 || m.listCursor >= len(effectiveItems) {
		return
	}

	m.inspectorData = effectiveItems[m.listCursor].GetInspectorData()
	m.focus = FocusPane3
	// Reset scroll when selecting new item
	m.inspectorViewport.GotoTop()
}

// updateInspectorFromCursor updates inspector data based on current cursor position
func (m *Model) updateInspectorFromCursor() {
	effectiveItems := m.getEffectiveListItems()
	if len(effectiveItems) == 0 || m.listCursor >= len(effectiveItems) {
		return
	}

	m.inspectorData = effectiveItems[m.listCursor].GetInspectorData()
	// Reset scroll when updating item
	m.inspectorViewport.GotoTop()
}
