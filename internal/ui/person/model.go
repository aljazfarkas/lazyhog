package person

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
	columnLeft = iota
	columnRight
)

// Model represents the person view state
type Model struct {
	client      *client.Client
	distinctID  string
	person      *client.Person
	events      []client.Event
	loading     bool
	err         error
	width       int
	height      int
	activeCol   int
	scrollLeft  int
	scrollRight int
}

type personMsg struct {
	person *client.Person
	events []client.Event
}
type errorMsg struct{ err error }

// New creates a new person model
func New(c *client.Client, distinctID string) Model {
	return Model{
		client:     c,
		distinctID: distinctID,
		loading:    true,
		activeCol:  columnLeft,
	}
}

func (m Model) Init() tea.Cmd {
	return fetchPersonData(m.client, m.distinctID)
}

func fetchPersonData(c *client.Client, distinctID string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		person, err := c.GetPerson(ctx, distinctID)
		if err != nil {
			return errorMsg{err: err}
		}

		events, err := c.GetPersonEvents(ctx, distinctID, 20)
		if err != nil {
			// Don't fail if events fail, just return empty
			events = []client.Event{}
		}

		return personMsg{person: person, events: events}
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

		case "tab":
			if m.activeCol == columnLeft {
				m.activeCol = columnRight
			} else {
				m.activeCol = columnLeft
			}

		case "up", "k":
			if m.activeCol == columnLeft && m.scrollLeft > 0 {
				m.scrollLeft--
			} else if m.activeCol == columnRight && m.scrollRight > 0 {
				m.scrollRight--
			}

		case "down", "j":
			if m.activeCol == columnLeft {
				m.scrollLeft++
			} else if m.activeCol == columnRight {
				m.scrollRight++
			}

		case "r":
			m.loading = true
			return m, fetchPersonData(m.client, m.distinctID)
		}

	case personMsg:
		m.person = msg.person
		m.events = msg.events
		m.loading = false
		m.err = nil
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

	// Error
	if m.err != nil {
		sb.WriteString(styles.ErrorTextStyle.Render(fmt.Sprintf("Error: %v", m.err)))
		sb.WriteString("\n\n")
		sb.WriteString(styles.HelpStyle.Render("Press 'r' to retry, 'q' to quit"))
		return sb.String()
	}

	// Loading
	if m.loading {
		sb.WriteString(styles.DimTextStyle.Render("Loading person data..."))
		sb.WriteString("\n")
		return sb.String()
	}

	// Two-column layout
	if m.person != nil {
		leftCol := m.renderLeftColumn()
		rightCol := m.renderRightColumn()

		colWidth := (m.width - 4) / 2

		leftStyled := styles.BorderStyle.
			Width(colWidth).
			Height(m.height - 10).
			Render(leftCol)

		rightStyled := styles.BorderStyle.
			Width(colWidth).
			Height(m.height - 10).
			Render(rightCol)

		columns := lipgloss.JoinHorizontal(lipgloss.Top, leftStyled, rightStyled)
		sb.WriteString(columns)
	}

	sb.WriteString("\n")
	sb.WriteString(m.renderFooter())

	return sb.String()
}

func (m Model) renderHeader() string {
	title := styles.TitleStyle.Render(fmt.Sprintf("👤 Person: %s", m.distinctID))
	return styles.HeaderStyle.Width(m.width - 2).Render(title)
}

func (m Model) renderLeftColumn() string {
	if m.person == nil {
		return ""
	}

	var sb strings.Builder

	// Column title
	titleStyle := styles.HeaderStyle
	if m.activeCol == columnLeft {
		titleStyle = titleStyle.Foreground(styles.ColorPrimary)
	} else {
		titleStyle = titleStyle.Foreground(styles.ColorDim)
	}
	sb.WriteString(titleStyle.Render("Properties"))
	sb.WriteString("\n\n")

	// Basic info
	sb.WriteString(styles.JSONKeyStyle.Render("Name: "))
	if m.person.Name != "" {
		sb.WriteString(m.person.Name)
	} else {
		sb.WriteString(styles.DimTextStyle.Render("(no name)"))
	}
	sb.WriteString("\n\n")

	sb.WriteString(styles.JSONKeyStyle.Render("Distinct IDs:"))
	sb.WriteString("\n")
	for _, id := range m.person.DistinctIDs {
		sb.WriteString(fmt.Sprintf("  • %s\n", id))
	}
	sb.WriteString("\n")

	if m.person.CreatedAt != "" {
		sb.WriteString(styles.JSONKeyStyle.Render("Created: "))
		sb.WriteString(m.person.CreatedAt)
		sb.WriteString("\n\n")
	}

	// Properties
	sb.WriteString(styles.JSONKeyStyle.Render("Properties:"))
	sb.WriteString("\n")

	if len(m.person.Properties) == 0 {
		sb.WriteString(styles.DimTextStyle.Render("  (no properties)"))
		sb.WriteString("\n")
	} else {
		maxLines := m.height - 20
		if maxLines < 5 {
			maxLines = 5
		}
		propertiesJSON := components.FormatJSONWithColors(m.person.Properties, maxLines)
		sb.WriteString(propertiesJSON)
	}

	return sb.String()
}

func (m Model) renderRightColumn() string {
	var sb strings.Builder

	// Column title
	titleStyle := styles.HeaderStyle
	if m.activeCol == columnRight {
		titleStyle = titleStyle.Foreground(styles.ColorPrimary)
	} else {
		titleStyle = titleStyle.Foreground(styles.ColorDim)
	}
	sb.WriteString(titleStyle.Render(fmt.Sprintf("Recent Events (%d)", len(m.events))))
	sb.WriteString("\n\n")

	if len(m.events) == 0 {
		sb.WriteString(styles.DimTextStyle.Render("  (no events)"))
		sb.WriteString("\n")
		return sb.String()
	}

	// Events list
	visibleHeight := m.height - 15
	if visibleHeight < 5 {
		visibleHeight = 5
	}

	start := m.scrollRight
	if start < 0 {
		start = 0
	}
	end := start + visibleHeight
	if end > len(m.events) {
		end = len(m.events)
	}

	for i := start; i < end; i++ {
		event := m.events[i]
		timeStr := client.FormatEventTimeShort(event.Timestamp)
		eventName := event.Event

		maxNameLen := 25
		if len(eventName) > maxNameLen {
			eventName = styles.TruncateString(eventName, maxNameLen)
		}

		sb.WriteString(styles.DimTextStyle.Render(timeStr))
		sb.WriteString("\n")
		sb.WriteString(fmt.Sprintf("  %s", eventName))
		sb.WriteString("\n")

		// Show one interesting property if available
		if len(event.Properties) > 0 {
			for key, val := range event.Properties {
				if key != "$lib" && key != "$lib_version" {
					valStr := fmt.Sprintf("%v", val)
					if len(valStr) > 30 {
						valStr = styles.TruncateString(valStr, 30)
					}
					sb.WriteString(styles.DimTextStyle.Render(fmt.Sprintf("    %s: %s", key, valStr)))
					sb.WriteString("\n")
					break
				}
			}
		}

		sb.WriteString("\n")
	}

	return sb.String()
}

func (m Model) renderFooter() string {
	help := "Tab: switch column • ↑/↓: scroll • r: refresh • q: quit"

	var activeInfo string
	if m.activeCol == columnLeft {
		activeInfo = "Properties"
	} else {
		activeInfo = "Events"
	}

	footer := lipgloss.JoinHorizontal(
		lipgloss.Left,
		styles.DimTextStyle.Render(help),
		"  ",
		styles.DimTextStyle.Render("│"),
		"  ",
		styles.HighlightTextStyle.Render(activeInfo),
	)

	return styles.FooterStyle.Width(m.width - 2).Render(footer)
}
