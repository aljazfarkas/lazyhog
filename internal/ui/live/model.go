package live

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aljazfarkas/lazyhog/internal/client"
	"github.com/aljazfarkas/lazyhog/internal/ui/components"
	"github.com/aljazfarkas/lazyhog/internal/ui/styles"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	maxEvents      = 50
	pollInterval   = 2 * time.Second
	maxEventHeight = 30
)

// Model represents the live events view state
type Model struct {
	client        *client.Client
	events        []client.Event
	cursor        int
	expanded      bool
	selectedEvent *client.Event
	loading       bool
	err           error
	width         int
	height        int
	lastUpdate    time.Time
}

type tickMsg time.Time
type eventsMsg []client.Event
type errorMsg struct{ err error }

// New creates a new live events model
func New(c *client.Client) Model {
	return Model{
		client:     c,
		events:     []client.Event{},
		cursor:     0,
		expanded:   false,
		loading:    true,
		lastUpdate: time.Now(),
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		tickCmd(),
		fetchEvents(m.client),
	)
}

func tickCmd() tea.Cmd {
	return tea.Tick(pollInterval, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func fetchEvents(c *client.Client) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		events, err := c.GetRecentEvents(ctx, maxEvents)
		if err != nil {
			return errorMsg{err: err}
		}
		return eventsMsg(events)
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "up", "k":
			if !m.expanded && m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if !m.expanded && m.cursor < len(m.events)-1 {
				m.cursor++
			}

		case "enter", "space":
			if len(m.events) > 0 && m.cursor < len(m.events) {
				m.expanded = !m.expanded
				if m.expanded {
					m.selectedEvent = &m.events[m.cursor]
				} else {
					m.selectedEvent = nil
				}
			}

		case "esc":
			m.expanded = false
			m.selectedEvent = nil

		case "r":
			m.loading = true
			return m, fetchEvents(m.client)
		}

	case tickMsg:
		if !m.expanded {
			return m, tea.Batch(
				tickCmd(),
				fetchEvents(m.client),
			)
		}
		return m, tickCmd()

	case eventsMsg:
		m.events = []client.Event(msg)
		m.loading = false
		m.lastUpdate = time.Now()
		m.err = nil

		// Adjust cursor if out of bounds
		if m.cursor >= len(m.events) && len(m.events) > 0 {
			m.cursor = len(m.events) - 1
		}
		return m, nil

	case errorMsg:
		m.err = msg.err
		m.loading = false
		return m, nil
	}

	return m, nil
}

func (m Model) View() string {
	if m.width == 0 {
		return ""
	}

	var sb strings.Builder

	// Header
	header := m.renderHeader()
	sb.WriteString(header)
	sb.WriteString("\n\n")

	// Content
	if m.err != nil {
		sb.WriteString(styles.ErrorTextStyle.Render(fmt.Sprintf("Error: %v", m.err)))
		sb.WriteString("\n\n")
		sb.WriteString(styles.HelpStyle.Render("Press 'r' to retry, 'q' to quit"))
		return sb.String()
	}

	if m.loading && len(m.events) == 0 {
		sb.WriteString(styles.DimTextStyle.Render("Loading events..."))
		sb.WriteString("\n")
		return sb.String()
	}

	if len(m.events) == 0 {
		sb.WriteString(styles.DimTextStyle.Render("No events yet. Waiting for new events..."))
		sb.WriteString("\n")
		return sb.String()
	}

	if m.expanded && m.selectedEvent != nil {
		// Expanded view
		sb.WriteString(m.renderExpandedEvent())
	} else {
		// List view
		sb.WriteString(m.renderEventsList())
	}

	sb.WriteString("\n")
	sb.WriteString(m.renderFooter())

	return sb.String()
}

func (m Model) renderHeader() string {
	title := styles.TitleStyle.Render("📡 PostHog Live Events")

	status := ""
	if m.loading {
		status = styles.DimTextStyle.Render("⟳ Loading...")
	} else {
		status = styles.SuccessTextStyle.Render("● Live")
	}

	timestamp := styles.DimTextStyle.Render(fmt.Sprintf("Last update: %s", m.lastUpdate.Format("15:04:05")))

	// Create a flexible layout
	titleLine := lipgloss.JoinHorizontal(
		lipgloss.Left,
		title,
		"  ",
		status,
		"  ",
		timestamp,
	)

	return styles.HeaderStyle.Width(m.width - 2).Render(titleLine)
}

func (m Model) renderEventsList() string {
	var sb strings.Builder

	visibleHeight := m.height - 10
	if visibleHeight < 5 {
		visibleHeight = 5
	}

	start := m.cursor - visibleHeight/2
	if start < 0 {
		start = 0
	}
	end := start + visibleHeight
	if end > len(m.events) {
		end = len(m.events)
		start = end - visibleHeight
		if start < 0 {
			start = 0
		}
	}

	for i := start; i < end; i++ {
		event := m.events[i]
		line := m.renderEventLine(event, i == m.cursor)
		sb.WriteString(line)
		sb.WriteString("\n")
	}

	return sb.String()
}

func (m Model) renderEventLine(event client.Event, selected bool) string {
	timeStr := client.FormatEventTimeShort(event.Timestamp)
	eventName := event.Event
	distinctID := event.DistinctID

	// Truncate if needed
	maxEventNameLen := 30
	maxDistinctIDLen := 25

	if len(eventName) > maxEventNameLen {
		eventName = styles.TruncateString(eventName, maxEventNameLen)
	}
	if len(distinctID) > maxDistinctIDLen {
		distinctID = styles.TruncateString(distinctID, maxDistinctIDLen)
	}

	timeStyled := styles.DimTextStyle.Render(timeStr)
	eventStyled := eventName
	idStyled := styles.DimTextStyle.Render(distinctID)

	line := fmt.Sprintf("  %s  %s  %s", timeStyled, eventStyled, idStyled)

	if selected {
		line = styles.SelectedListItemStyle.Render("▶ " + line)
	} else {
		line = styles.ListItemStyle.Render("  " + line)
	}

	return line
}

func (m Model) renderExpandedEvent() string {
	if m.selectedEvent == nil {
		return ""
	}

	event := m.selectedEvent

	var sb strings.Builder

	// Event header
	sb.WriteString(styles.HeaderStyle.Render(fmt.Sprintf("Event: %s", event.Event)))
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
	maxLines := styles.Max(10, m.height-15)
	propertiesJSON := components.FormatJSONWithColors(event.Properties, maxLines)
	sb.WriteString(propertiesJSON)

	return sb.String()
}

func (m Model) renderFooter() string {
	help := ""
	if m.expanded {
		help = "↑/↓: scroll • Enter/Esc: collapse • q: quit"
	} else {
		help = "↑/↓: navigate • Enter: expand • r: refresh • q: quit"
	}

	count := fmt.Sprintf("%d events", len(m.events))

	footer := lipgloss.JoinHorizontal(
		lipgloss.Left,
		styles.DimTextStyle.Render(help),
		"  ",
		styles.DimTextStyle.Render("│"),
		"  ",
		styles.DimTextStyle.Render(count),
	)

	return styles.FooterStyle.Width(m.width - 2).Render(footer)
}
