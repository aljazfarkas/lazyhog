package miller

import (
	"strings"

	"github.com/aljazfarkas/lazyhog/internal/client"
	"github.com/aljazfarkas/lazyhog/internal/ui/styles"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// StreamTableModel wraps bubbles/table for Pane 2 (Phase 4)
type StreamTableModel struct {
	table      table.Model
	resource   Resource
	autoScroll bool
	newCount   int
	items      []ListItem // Keep reference to items for inspector sync
}

// NewStreamTableModel creates a new table model for the stream
func NewStreamTableModel(width, height int, resource Resource) StreamTableModel {
	columns := getColumnsForResource(resource, width)

	t := table.New(
		table.WithColumns(columns),
		table.WithRows([]table.Row{}),
		table.WithFocused(true),
		table.WithHeight(height),
	)

	// Custom styles
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(styles.ColorPrimary).
		BorderBottom(true).
		Bold(true).
		Foreground(styles.ColorPrimary)

	s.Selected = s.Selected.
		Foreground(styles.ColorPrimary).
		Bold(true)

	t.SetStyles(s)

	return StreamTableModel{
		table:      t,
		resource:   resource,
		autoScroll: true,
		newCount:   0,
		items:      []ListItem{},
	}
}

// getColumnsForResource returns table columns based on resource type
func getColumnsForResource(resource Resource, totalWidth int) []table.Column {
	// Adjust for borders and padding
	usableWidth := totalWidth - 10

	switch resource {
	case ResourceEvents:
		timeWidth := 15
		eventWidth := 30
		idWidth := usableWidth - timeWidth - eventWidth
		if idWidth < 15 {
			idWidth = 15
		}
		return []table.Column{
			{Title: "Time", Width: timeWidth},
			{Title: "Event", Width: eventWidth},
			{Title: "Distinct ID", Width: idWidth},
		}

	case ResourcePersons:
		nameWidth := 35
		idsWidth := usableWidth - nameWidth
		if idsWidth < 20 {
			idsWidth = 20
		}
		return []table.Column{
			{Title: "Name", Width: nameWidth},
			{Title: "Distinct IDs", Width: idsWidth},
		}

	case ResourceFlags:
		statusWidth := 8
		keyWidth := usableWidth - statusWidth
		if keyWidth < 30 {
			keyWidth = 30
		}
		return []table.Column{
			{Title: "Status", Width: statusWidth},
			{Title: "Key", Width: keyWidth},
		}

	default:
		return []table.Column{
			{Title: "Data", Width: usableWidth},
		}
	}
}

// itemToRow converts a ListItem to a table row
func itemToRow(item ListItem, resource Resource) table.Row {
	switch resource {
	case ResourceEvents:
		if eventItem, ok := item.(EventListItem); ok {
			timeStr := client.FormatEventTimeShort(eventItem.Event.Timestamp)
			eventName := eventItem.Event.Event
			distinctID := eventItem.Event.DistinctID

			// Truncate if needed
			if len(eventName) > 28 {
				eventName = styles.TruncateString(eventName, 28)
			}
			if len(distinctID) > 25 {
				distinctID = styles.TruncateString(distinctID, 25)
			}

			return table.Row{timeStr, eventName, distinctID}
		}

	case ResourcePersons:
		if personItem, ok := item.(PersonListItem); ok {
			name := personItem.Person.Name
			if name == "" {
				name = "(no name)"
			}

			distinctIDs := ""
			if len(personItem.Person.DistinctIDs) > 0 {
				distinctIDs = strings.Join(personItem.Person.DistinctIDs, ", ")
			}

			// Truncate if needed
			if len(name) > 33 {
				name = styles.TruncateString(name, 33)
			}
			if len(distinctIDs) > 40 {
				distinctIDs = styles.TruncateString(distinctIDs, 40)
			}

			return table.Row{name, distinctIDs}
		}

	case ResourceFlags:
		if flagItem, ok := item.(FlagListItem); ok {
			status := "✗"
			if flagItem.Flag.Active {
				status = "✓"
			}

			key := flagItem.Flag.Key
			if len(key) > 50 {
				key = styles.TruncateString(key, 50)
			}

			return table.Row{status, key}
		}
	}

	// Fallback
	return table.Row{"Unknown item"}
}

// SetItems updates the table with new items
func (m *StreamTableModel) SetItems(items []ListItem, resource Resource) {
	m.items = items
	m.resource = resource

	// Convert items to rows
	rows := make([]table.Row, len(items))
	for i, item := range items {
		rows[i] = itemToRow(item, resource)
	}

	m.table.SetRows(rows)

	// Auto-scroll to bottom if enabled
	if m.autoScroll && len(rows) > 0 {
		m.table.GotoBottom()
		m.newCount = 0
	}
}

// Update handles table updates
func (m StreamTableModel) Update(msg tea.Msg) (StreamTableModel, tea.Cmd) {
	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

// View renders the table
func (m StreamTableModel) View() string {
	return m.table.View()
}

// SetSize updates the table size and columns
func (m *StreamTableModel) SetSize(width, height int) {
	m.table.SetHeight(height)

	// Update columns for new width
	columns := getColumnsForResource(m.resource, width)
	m.table.SetColumns(columns)

	// Refresh rows with new column widths
	if len(m.items) > 0 {
		rows := make([]table.Row, len(m.items))
		for i, item := range m.items {
			rows[i] = itemToRow(item, m.resource)
		}
		m.table.SetRows(rows)
	}
}

// GetCursor returns the current cursor position
func (m StreamTableModel) GetCursor() int {
	return m.table.Cursor()
}

// SetCursor sets the cursor position
func (m *StreamTableModel) SetCursor(index int) {
	m.table.SetCursor(index)
}

// MoveUp moves cursor up
func (m *StreamTableModel) MoveUp() {
	m.table.MoveUp(1)

	// Disable auto-scroll if we move away from bottom
	if m.autoScroll && !m.isAtBottom() {
		m.autoScroll = false
	}
}

// MoveDown moves cursor down
func (m *StreamTableModel) MoveDown() {
	m.table.MoveDown(1)

	// Re-enable auto-scroll if we reach bottom
	if !m.autoScroll && m.isAtBottom() {
		m.autoScroll = true
		m.newCount = 0
	}
}

// GotoBottom moves cursor to the bottom
func (m *StreamTableModel) GotoBottom() {
	m.table.GotoBottom()
	m.autoScroll = true
	m.newCount = 0
}

// isAtBottom checks if cursor is at the last item
func (m StreamTableModel) isAtBottom() bool {
	if len(m.items) == 0 {
		return true
	}
	return m.table.Cursor() >= len(m.items)-1
}

// EnableAutoScroll enables auto-scroll mode
func (m *StreamTableModel) EnableAutoScroll() {
	m.autoScroll = true
	m.newCount = 0
	m.table.GotoBottom()
}

// IsAutoScrollEnabled returns auto-scroll status
func (m StreamTableModel) IsAutoScrollEnabled() bool {
	return m.autoScroll
}

// GetNewCount returns the count of new items since scroll detached
func (m StreamTableModel) GetNewCount() int {
	return m.newCount
}

// IncrementNewCount increments the new items counter
func (m *StreamTableModel) IncrementNewCount(delta int) {
	m.newCount += delta
}

// GetSelectedItem returns the currently selected item
func (m StreamTableModel) GetSelectedItem() ListItem {
	cursor := m.table.Cursor()
	if cursor >= 0 && cursor < len(m.items) {
		return m.items[cursor]
	}
	return nil
}

// GetItems returns all items
func (m StreamTableModel) GetItems() []ListItem {
	return m.items
}
