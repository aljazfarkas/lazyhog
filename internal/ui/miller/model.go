package miller

import (
	"context"
	"strings"
	"time"

	"github.com/aljazfarkas/lazyhog/internal/client"
	"github.com/aljazfarkas/lazyhog/internal/ui/styles"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	maxEvents         = 50
	maxPersons        = 50
	pollInterval      = 2 * time.Second
	pausePollDuration = 30 * time.Second
)

// Model represents the Miller Columns TUI state
type Model struct {
	client *client.Client
	width  int
	height int

	// Focus and resource selection
	focus            Focus
	selectedResource Resource

	// List view (Pane 2)
	listItems  []ListItem
	listCursor int

	// Inspector (Pane 3)
	inspectorData interface{}

	// Polling state
	isPolling         bool
	lastInteraction   time.Time
	lastPoll          time.Time

	// Loading and error state
	loading bool
	err     error
}

// Messages
type tickMsg time.Time
type eventsMsg []client.Event
type personsMsg []client.Person
type flagsMsg []client.FeatureFlag
type errorMsg struct{ err error }
type pivotMsg struct {
	person *client.Person
	events []client.Event
}

// New creates a new Miller Columns model
func New(c *client.Client) Model {
	return Model{
		client:           c,
		focus:            FocusPane1,
		selectedResource: ResourceEvents,
		listItems:        []ListItem{},
		listCursor:       0,
		isPolling:        true,
		lastInteraction:  time.Now(),
		lastPoll:         time.Now(),
		loading:          true,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		tickCmd(),
		m.fetchCurrentResource(),
	)
}

func tickCmd() tea.Cmd {
	return tea.Tick(pollInterval, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m Model) fetchCurrentResource() tea.Cmd {
	switch m.selectedResource {
	case ResourceEvents:
		return fetchEvents(m.client)
	case ResourcePersons:
		return fetchPersons(m.client)
	case ResourceFlags:
		return fetchFlags(m.client)
	default:
		return nil
	}
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

func fetchPersons(c *client.Client) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		persons, err := c.ListPersons(ctx, maxPersons)
		if err != nil {
			return errorMsg{err: err}
		}
		return personsMsg(persons)
	}
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

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case tickMsg:
		// Smart polling: only poll Events when not focused on Pane 3 or after interaction timeout
		if m.shouldPoll() {
			m.lastPoll = time.Now()
			return m, tea.Batch(
				tickCmd(),
				fetchEvents(m.client),
			)
		}
		return m, tickCmd()

	case eventsMsg:
		m.listItems = make([]ListItem, len(msg))
		for i, event := range msg {
			m.listItems[i] = EventListItem{Event: event}
		}
		m.loading = false
		m.err = nil

		// Adjust cursor if out of bounds
		if m.listCursor >= len(m.listItems) && len(m.listItems) > 0 {
			m.listCursor = len(m.listItems) - 1
		}
		if m.listCursor < 0 {
			m.listCursor = 0
		}

		return m, nil

	case personsMsg:
		m.listItems = make([]ListItem, len(msg))
		for i, person := range msg {
			m.listItems[i] = PersonListItem{Person: person}
		}
		m.loading = false
		m.err = nil

		// Adjust cursor if out of bounds
		if m.listCursor >= len(m.listItems) && len(m.listItems) > 0 {
			m.listCursor = len(m.listItems) - 1
		}
		if m.listCursor < 0 {
			m.listCursor = 0
		}

		return m, nil

	case flagsMsg:
		m.listItems = make([]ListItem, len(msg))
		for i, flag := range msg {
			m.listItems[i] = FlagListItem{Flag: flag}
		}
		m.loading = false
		m.err = nil

		// Adjust cursor if out of bounds
		if m.listCursor >= len(m.listItems) && len(m.listItems) > 0 {
			m.listCursor = len(m.listItems) - 1
		}
		if m.listCursor < 0 {
			m.listCursor = 0
		}

		return m, nil

	case pivotMsg:
		// Switch to Persons view after pivot
		m.selectedResource = ResourcePersons
		m.listItems = []ListItem{PersonListItem{Person: *msg.person}}
		m.listCursor = 0
		m.inspectorData = *msg.person
		m.focus = FocusPane3
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
		return "Loading..."
	}

	var content string

	// Check for narrow terminal
	if m.width < 100 {
		// Single pane mode with breadcrumb
		content = m.renderNarrowView()
	} else {
		// Calculate responsive pane widths
		pane1Width, pane2Width, pane3Width := m.calculatePaneWidths()

		// Render each pane
		pane1 := m.renderResourceSelector(pane1Width, m.height-3)
		pane2 := m.renderListView(pane2Width, m.height-3)
		pane3 := m.renderInspector(pane3Width, m.height-3)

		// Combine panes horizontally
		content = lipgloss.JoinHorizontal(lipgloss.Top, pane1, pane2, pane3)
	}

	// Add footer with help
	footer := m.renderFooter()
	return content + "\n" + footer
}

// calculatePaneWidths calculates responsive pane widths based on terminal width
func (m Model) calculatePaneWidths() (int, int, int) {
	totalWidth := m.width

	// Handle narrow terminals
	if totalWidth < 100 {
		// Single pane mode (simplified for now - will enhance later)
		switch m.focus {
		case FocusPane1:
			return totalWidth, 0, 0
		case FocusPane2:
			return 0, totalWidth, 0
		case FocusPane3:
			return 0, 0, totalWidth
		}
	}

	// Standard 3-pane layout
	// Pane 1: 15% | Pane 2: 35% | Pane 3: 50%
	pane1Width := totalWidth * 15 / 100
	pane2Width := totalWidth * 35 / 100
	pane3Width := totalWidth - pane1Width - pane2Width

	// Ensure minimum widths
	if pane1Width < 20 {
		pane1Width = 20
	}
	if pane2Width < 30 {
		pane2Width = 30
	}
	if pane3Width < 40 {
		pane3Width = 40
	}

	return pane1Width, pane2Width, pane3Width
}

// shouldPoll determines if we should poll for new events
func (m Model) shouldPoll() bool {
	// Only poll for Events resource
	if m.selectedResource != ResourceEvents {
		return false
	}

	// Don't poll if focused on Pane 3 (inspector) unless interaction timeout passed
	if m.focus == FocusPane3 {
		timeSinceInteraction := time.Since(m.lastInteraction)
		return timeSinceInteraction > pausePollDuration
	}

	// Poll if not focused on Pane 3
	return true
}

// recordInteraction records user interaction for polling pause
func (m *Model) recordInteraction() {
	m.lastInteraction = time.Now()
}

// renderNarrowView renders a single pane for narrow terminals
func (m Model) renderNarrowView() string {
	// Show breadcrumb
	breadcrumb := m.renderBreadcrumb()

	// Render only the active pane
	var pane string
	switch m.focus {
	case FocusPane1:
		pane = m.renderResourceSelector(m.width, m.height-5)
	case FocusPane2:
		pane = m.renderListView(m.width, m.height-5)
	case FocusPane3:
		pane = m.renderInspector(m.width, m.height-5)
	}

	return breadcrumb + "\n" + pane
}

// renderBreadcrumb renders breadcrumb navigation for narrow terminals
func (m Model) renderBreadcrumb() string {
	parts := []string{}

	// Resource type
	parts = append(parts, m.selectedResource.String())

	// Current pane
	if m.focus == FocusPane2 {
		parts = append(parts, "List")
	} else if m.focus == FocusPane3 {
		parts = append(parts, "Details")
	}

	breadcrumb := strings.Join(parts, " > ")
	return styles.DimTextStyle.Render(breadcrumb)
}

// renderFooter renders help footer with keyboard shortcuts
func (m Model) renderFooter() string {
	var help string

	switch m.focus {
	case FocusPane1:
		help = "↑/↓: navigate • Enter: select • 1/2/3: quick select • Tab/l: next • q: quit"
	case FocusPane2:
		help = "↑/↓: navigate • Enter: details • p: pivot to person • r: refresh • Tab/l: next • h: back • q: quit"
	case FocusPane3:
		help = "p: pivot to person • h/Tab: navigate panes • q: back"
	}

	return styles.FooterStyle.Width(m.width - 2).Render(styles.DimTextStyle.Render(help))
}
