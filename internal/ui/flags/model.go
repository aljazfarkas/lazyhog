package flags

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

// Model represents the feature flags view state
type Model struct {
	client         *client.Client
	flags          []client.FeatureFlag
	filteredFlags  []client.FeatureFlag
	cursor         int
	searchQuery    string
	searchMode     bool
	loading        bool
	err            error
	width          int
	height         int
	toast          components.Toast
	togglingFlagID int
}

type flagsMsg []client.FeatureFlag
type errorMsg struct{ err error }
type toggleSuccessMsg struct {
	flagID int
	active bool
}

// New creates a new flags model
func New(c *client.Client) Model {
	return Model{
		client:         c,
		flags:          []client.FeatureFlag{},
		filteredFlags:  []client.FeatureFlag{},
		cursor:         0,
		searchQuery:    "",
		searchMode:     false,
		loading:        true,
		togglingFlagID: -1,
	}
}

func (m Model) Init() tea.Cmd {
	return fetchFlags(m.client)
}

func fetchFlags(c *client.Client) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		flags, err := c.ListFlags(ctx)
		if err != nil {
			return errorMsg{err: err}
		}
		return flagsMsg(flags)
	}
}

func toggleFlag(c *client.Client, flagID int, active bool) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := c.ToggleFlag(ctx, flagID, active); err != nil {
			return errorMsg{err: err}
		}
		return toggleSuccessMsg{flagID: flagID, active: active}
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.searchMode {
				m.searchMode = false
				m.searchQuery = ""
				m.filterFlags()
				return m, nil
			}
			return m, tea.Quit

		case "esc":
			if m.searchMode {
				m.searchMode = false
				m.searchQuery = ""
				m.filterFlags()
			}

		case "/":
			m.searchMode = true

		case "up", "k":
			if !m.searchMode && m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if !m.searchMode && m.cursor < len(m.filteredFlags)-1 {
				m.cursor++
			}

		case "space":
			if !m.searchMode && len(m.filteredFlags) > 0 && m.cursor < len(m.filteredFlags) {
				flag := m.filteredFlags[m.cursor]
				m.togglingFlagID = flag.ID
				newActive := !flag.Active
				return m, toggleFlag(m.client, flag.ID, newActive)
			}

		case "r":
			if !m.searchMode {
				m.loading = true
				return m, fetchFlags(m.client)
			}

		case "enter":
			if m.searchMode {
				m.searchMode = false
			}

		case "backspace":
			if m.searchMode && len(m.searchQuery) > 0 {
				m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
				m.filterFlags()
			}

		default:
			if m.searchMode && len(msg.String()) == 1 {
				m.searchQuery += msg.String()
				m.filterFlags()
			}
		}

	case flagsMsg:
		m.flags = []client.FeatureFlag(msg)
		m.loading = false
		m.err = nil
		m.filterFlags()

		if m.cursor >= len(m.filteredFlags) && len(m.filteredFlags) > 0 {
			m.cursor = len(m.filteredFlags) - 1
		}
		return m, nil

	case toggleSuccessMsg:
		// Update the flag in our list
		for i := range m.flags {
			if m.flags[i].ID == msg.flagID {
				m.flags[i].Active = msg.active
				break
			}
		}
		m.filterFlags()
		m.togglingFlagID = -1

		status := "activated"
		if !msg.active {
			status = "deactivated"
		}
		cmd = m.toast.Show(fmt.Sprintf("Flag %s", status), components.ToastSuccess)
		return m, cmd

	case errorMsg:
		m.err = msg.err
		m.loading = false
		m.togglingFlagID = -1
		cmd = m.toast.Show(fmt.Sprintf("Error: %v", msg.err), components.ToastError)
		return m, cmd

	case components.ToastHideMsg:
		m.toast.Hide()
		return m, nil
	}

	return m, nil
}

func (m *Model) filterFlags() {
	if m.searchQuery == "" {
		m.filteredFlags = m.flags
		return
	}

	query := strings.ToLower(m.searchQuery)
	m.filteredFlags = []client.FeatureFlag{}

	for _, flag := range m.flags {
		if strings.Contains(strings.ToLower(flag.Key), query) ||
			strings.Contains(strings.ToLower(flag.Name), query) {
			m.filteredFlags = append(m.filteredFlags, flag)
		}
	}

	if m.cursor >= len(m.filteredFlags) {
		m.cursor = len(m.filteredFlags) - 1
		if m.cursor < 0 {
			m.cursor = 0
		}
	}
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

	// Toast
	if m.toast.Visible {
		sb.WriteString(m.toast.View())
		sb.WriteString("\n\n")
	}

	// Error
	if m.err != nil && !m.toast.Visible {
		sb.WriteString(styles.ErrorTextStyle.Render(fmt.Sprintf("Error: %v", m.err)))
		sb.WriteString("\n\n")
		sb.WriteString(styles.HelpStyle.Render("Press 'r' to retry, 'q' to quit"))
		return sb.String()
	}

	// Loading
	if m.loading && len(m.flags) == 0 {
		sb.WriteString(styles.DimTextStyle.Render("Loading feature flags..."))
		sb.WriteString("\n")
		return sb.String()
	}

	// Search bar
	if m.searchMode {
		searchBar := m.renderSearchBar()
		sb.WriteString(searchBar)
		sb.WriteString("\n\n")
	}

	// Flags list
	if len(m.filteredFlags) == 0 {
		if m.searchQuery != "" {
			sb.WriteString(styles.DimTextStyle.Render(fmt.Sprintf("No flags matching '%s'", m.searchQuery)))
		} else {
			sb.WriteString(styles.DimTextStyle.Render("No feature flags found"))
		}
		sb.WriteString("\n")
	} else {
		sb.WriteString(m.renderFlagsList())
	}

	sb.WriteString("\n")
	sb.WriteString(m.renderFooter())

	return sb.String()
}

func (m Model) renderHeader() string {
	title := styles.TitleStyle.Render("🚩 Feature Flags")

	count := fmt.Sprintf("%d flags", len(m.filteredFlags))
	if len(m.filteredFlags) != len(m.flags) {
		count = fmt.Sprintf("%d / %d flags", len(m.filteredFlags), len(m.flags))
	}

	titleLine := lipgloss.JoinHorizontal(
		lipgloss.Left,
		title,
		"  ",
		styles.DimTextStyle.Render(count),
	)

	return styles.HeaderStyle.Width(m.width - 2).Render(titleLine)
}

func (m Model) renderSearchBar() string {
	prompt := styles.HighlightTextStyle.Render("Search: ")
	query := m.searchQuery + "█"

	return prompt + query
}

func (m Model) renderFlagsList() string {
	var sb strings.Builder

	visibleHeight := m.height - 12
	if visibleHeight < 5 {
		visibleHeight = 5
	}

	start := m.cursor - visibleHeight/2
	if start < 0 {
		start = 0
	}
	end := start + visibleHeight
	if end > len(m.filteredFlags) {
		end = len(m.filteredFlags)
		start = end - visibleHeight
		if start < 0 {
			start = 0
		}
	}

	for i := start; i < end; i++ {
		flag := m.filteredFlags[i]
		line := m.renderFlagLine(flag, i == m.cursor)
		sb.WriteString(line)
		sb.WriteString("\n")
	}

	return sb.String()
}

func (m Model) renderFlagLine(flag client.FeatureFlag, selected bool) string {
	// Status indicator
	var status string
	if m.togglingFlagID == flag.ID {
		status = styles.DimTextStyle.Render("⟳")
	} else if flag.Active {
		status = styles.StatusActiveStyle.Render("●")
	} else {
		status = styles.StatusInactiveStyle.Render("○")
	}

	// Flag key
	key := flag.Key
	maxKeyLen := 40
	if len(key) > maxKeyLen {
		key = styles.TruncateString(key, maxKeyLen)
	}

	// Flag name
	name := flag.Name
	if name == "" {
		name = styles.DimTextStyle.Render("(no name)")
	} else {
		maxNameLen := 30
		if len(name) > maxNameLen {
			name = styles.TruncateString(name, maxNameLen)
		}
		name = styles.DimTextStyle.Render(name)
	}

	line := fmt.Sprintf("  %s  %s  %s", status, key, name)

	if selected {
		line = styles.SelectedListItemStyle.Render("▶ " + line)
	} else {
		line = styles.ListItemStyle.Render("  " + line)
	}

	return line
}

func (m Model) renderFooter() string {
	help := ""
	if m.searchMode {
		help = "Type to search • Enter: done • Esc: cancel"
	} else {
		help = "↑/↓: navigate • Space: toggle • /: search • r: refresh • q: quit"
	}

	return styles.FooterStyle.Width(m.width - 2).Render(styles.DimTextStyle.Render(help))
}
