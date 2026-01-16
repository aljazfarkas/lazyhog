package unified

import (
	"github.com/aljazfarkas/lazyhog/internal/client"
	"github.com/aljazfarkas/lazyhog/internal/ui/components"
	"github.com/aljazfarkas/lazyhog/internal/ui/flags"
	"github.com/aljazfarkas/lazyhog/internal/ui/live"
	"github.com/aljazfarkas/lazyhog/internal/ui/styles"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	viewLive = iota
	viewFlags
	viewQuery
)

const (
	focusSidebar = iota
	focusPanel
)

// Model represents the unified TUI with sidebar
type Model struct {
	client       *client.Client
	width        int
	height       int
	currentView  int
	focus        int
	sidebarWidth int

	// View models
	liveModel  live.Model
	flagsModel flags.Model

	// Sidebar items
	sidebarItems []components.SidebarItem
}

// New creates a new unified model
func New(c *client.Client) Model {
	sidebarItems := []components.SidebarItem{
		{Key: "live", Label: "Live Events", Icon: "📡"},
		{Key: "flags", Label: "Feature Flags", Icon: "🚩"},
		{Key: "query", Label: "HogQL Query", Icon: "🔍"},
	}

	return Model{
		client:       c,
		currentView:  viewLive,
		focus:        focusPanel,
		sidebarWidth: 20,
		liveModel:    live.New(c),
		flagsModel:   flags.New(c),
		sidebarItems: sidebarItems,
	}
}

func (m Model) Init() tea.Cmd {
	// Initialize the first view
	return m.liveModel.Init()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Forward to active view with adjusted width
		panelWidth := m.width - m.sidebarWidth
		panelMsg := tea.WindowSizeMsg{
			Width:  panelWidth,
			Height: m.height,
		}

		switch m.currentView {
		case viewLive:
			var newModel tea.Model
			newModel, cmd = m.liveModel.Update(panelMsg)
			m.liveModel = newModel.(live.Model)
		case viewFlags:
			var newModel tea.Model
			newModel, cmd = m.flagsModel.Update(panelMsg)
			m.flagsModel = newModel.(flags.Model)
		}

		return m, cmd

	case tea.KeyMsg:
		// Global shortcuts
		switch msg.String() {
		case "ctrl+c", "q":
			if m.focus == focusSidebar {
				return m, tea.Quit
			}
			// If in panel, q goes back to sidebar
			if msg.String() == "q" {
				m.focus = focusSidebar
				return m, nil
			}
			// Ctrl+C always quits
			return m, tea.Quit

		case "tab", "right":
			if m.focus == focusSidebar {
				m.focus = focusPanel
				return m, nil
			}

		case "left", "esc":
			if m.focus == focusPanel {
				m.focus = focusSidebar
				return m, nil
			}

		case "up", "k":
			if m.focus == focusSidebar && m.currentView > 0 {
				m.currentView--
				return m, m.switchView()
			}

		case "down", "j":
			if m.focus == focusSidebar && m.currentView < len(m.sidebarItems)-1 {
				m.currentView++
				return m, m.switchView()
			}

		case "enter":
			if m.focus == focusSidebar {
				m.focus = focusPanel
				return m, nil
			}
		}

		// Forward to active view only if panel is focused
		if m.focus == focusPanel {
			switch m.currentView {
			case viewLive:
				var newModel tea.Model
				newModel, cmd = m.liveModel.Update(msg)
				m.liveModel = newModel.(live.Model)
				return m, cmd

			case viewFlags:
				var newModel tea.Model
				newModel, cmd = m.flagsModel.Update(msg)
				m.flagsModel = newModel.(flags.Model)
				return m, cmd
			}
		}
	}

	// Forward other messages to active view
	if m.focus == focusPanel {
		switch m.currentView {
		case viewLive:
			var newModel tea.Model
			newModel, cmd = m.liveModel.Update(msg)
			m.liveModel = newModel.(live.Model)
			return m, cmd

		case viewFlags:
			var newModel tea.Model
			newModel, cmd = m.flagsModel.Update(msg)
			m.flagsModel = newModel.(flags.Model)
			return m, cmd
		}
	}

	return m, nil
}

func (m Model) switchView() tea.Cmd {
	// Initialize the newly selected view
	switch m.currentView {
	case viewLive:
		return m.liveModel.Init()
	case viewFlags:
		return m.flagsModel.Init()
	}
	return nil
}

func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	// Render sidebar
	sidebar := components.RenderSidebar(m.sidebarItems, m.currentView, m.sidebarWidth, m.height)

	// Add focus indicator to sidebar
	if m.focus == focusSidebar {
		sidebar = lipgloss.NewStyle().
			BorderStyle(lipgloss.ThickBorder()).
			BorderForeground(styles.ColorPrimary).
			Render(sidebar)
	}

	// Render active panel
	panelWidth := m.width - m.sidebarWidth - 4
	panelHeight := m.height

	var panel string
	switch m.currentView {
	case viewLive:
		panel = m.liveModel.View()

	case viewFlags:
		panel = m.flagsModel.View()

	case viewQuery:
		panel = renderQueryPlaceholder(panelWidth, panelHeight)
	}

	// Add focus indicator to panel
	panelStyle := lipgloss.NewStyle().
		Width(panelWidth).
		Height(panelHeight - 2).
		Border(lipgloss.NormalBorder()).
		BorderForeground(styles.ColorBorder)

	if m.focus == focusPanel {
		panelStyle = panelStyle.
			BorderStyle(lipgloss.ThickBorder()).
			BorderForeground(styles.ColorPrimary)
	}

	styledPanel := panelStyle.Render(panel)

	// Combine sidebar and panel
	return lipgloss.JoinHorizontal(lipgloss.Top, sidebar, styledPanel)
}

func renderQueryPlaceholder(width, height int) string {
	msg := styles.DimTextStyle.Render("HogQL Query console\n\nUse 'lazyhog query' to access the query console")
	return lipgloss.NewStyle().
		Width(width - 4).
		Height(height - 4).
		Padding(2, 2).
		Render(msg)
}
