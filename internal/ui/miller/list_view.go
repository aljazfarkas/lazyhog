package miller

import (
	"fmt"
	"strings"

	"github.com/aljazfarkas/lazyhog/internal/client"
	"github.com/aljazfarkas/lazyhog/internal/ui/styles"
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
		searchInput := m.renderSearchInput(m.searchQuery, width)
		sb.WriteString(searchInput)
		sb.WriteString("\n")
	}

	// Error state
	if m.err != nil {
		sb.WriteString(styles.ErrorTextStyle.Render(fmt.Sprintf("Error: %v", m.err)))
		sb.WriteString("\n\n")
		sb.WriteString(styles.DimTextStyle.Render("Press 'r' to retry"))
		sb.WriteString("\n")
	} else if len(m.listItems) == 0 {
		// Empty state
		emptyMsg := m.getEmptyStateMessage()
		sb.WriteString(styles.DimTextStyle.Render(emptyMsg))
		sb.WriteString("\n")
	} else {
		// Render list items with viewport management
		visibleHeight := height - 8
		if visibleHeight < 5 {
			visibleHeight = 5
		}

		start := m.listCursor - visibleHeight/2
		if start < 0 {
			start = 0
		}
		end := start + visibleHeight
		if end > len(m.listItems) {
			end = len(m.listItems)
			start = end - visibleHeight
			if start < 0 {
				start = 0
			}
		}

		for i := start; i < end; i++ {
			item := m.listItems[i]
			line := item.RenderLine(width-4, i == m.listCursor)
			sb.WriteString(line)
			if i < end-1 {
				sb.WriteString("\n")
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
	switch m.selectedResource {
	case ResourceEvents:
		if m.loading {
			return "Loading events..."
		}
		return "No events yet. Waiting for new events..."
	case ResourcePersons:
		if m.loading {
			return "Loading persons..."
		}
		return "No persons found."
	case ResourceFlags:
		if m.loading {
			return "Loading flags..."
		}
		return "No feature flags. Create flags in PostHog."
	default:
		return "No data available."
	}
}

// MoveListCursorUp moves the list cursor up
func (m *Model) MoveListCursorUp() {
	if m.listCursor > 0 {
		m.listCursor--
	}
}

// MoveListCursorDown moves the list cursor down
func (m *Model) MoveListCursorDown() {
	if m.listCursor < len(m.listItems)-1 {
		m.listCursor++
	}
}

// SelectCurrentListItem populates Pane 3 with the selected item's data
func (m *Model) SelectCurrentListItem() {
	if len(m.listItems) == 0 || m.listCursor >= len(m.listItems) {
		return
	}

	m.inspectorData = m.listItems[m.listCursor].GetInspectorData()
	m.focus = FocusPane3
}
