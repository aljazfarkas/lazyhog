package miller

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aljazfarkas/lazyhog/internal/client"
	"github.com/aljazfarkas/lazyhog/internal/ui/styles"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	maxEvents         = 50
	maxPersons        = 50
	pollInterval      = 2 * time.Second
	pausePollDuration = 30 * time.Second

	// Pane width ratios (percentages)
	pane1WidthPercent = 15
	pane2WidthPercent = 35
	// pane3 gets the remainder

	// Minimum pane widths
	minPane1Width = 20
	minPane2Width = 30
	minPane3Width = 40

	// Narrow terminal threshold
	narrowTerminalWidth = 100
)

// Model represents the Miller Columns TUI state
type Model struct {
	client client.PostHogClient
	width  int
	height int

	// --- Focus and Navigation ---
	focus            Focus
	selectedResource Resource
	pane1Cursor      int // -1 = project, 0 = Events, 1 = Persons, 2 = Flags

	// --- Project State ---
	availableProjects []client.Project
	selectedProjectID int
	projectsLoaded    bool

	// --- List State (Pane 2) ---
	listItems     []ListItem
	listCursor    int
	filteredItems []ListItem // nil means no filter active

	// --- Inspector State (Pane 3) ---
	inspectorData     interface{}
	inspectorViewport viewport.Model
	jsonFoldState     map[string]bool // JSON path -> folded status
	allFolded         bool

	// --- Auto-scroll State ---
	autoScroll      bool
	newEventCount   int
	lastSeenEventID string

	// --- Search State ---
	searchMode  bool
	searchInput textinput.Model

	// --- Polling State ---
	isPolling       bool
	lastInteraction time.Time
	lastPoll        time.Time

	// --- Clipboard State ---
	clipboardMsg  string
	clipboardTime time.Time

	// --- Debounce State ---
	pendingResourceFetch *Resource
	lastDebounceTime     time.Time

	// --- Loading and Error State ---
	loading bool
	err     error

	// --- UI Mode State ---
	showHelp bool

	// --- UI Components ---
	spinner spinner.Model
}

// Messages
type tickMsg time.Time
type eventsMsg []client.Event
type personsMsg []client.Person
type flagsMsg []client.FeatureFlag
type projectsMsg []client.Project
type errorMsg struct{ err error }
type pivotMsg struct {
	person      *client.Person
	events      []client.Event
	eventsError error // Non-fatal error when fetching events
}

// New creates a new Miller Columns model
func New(c client.PostHogClient) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = styles.SpinnerStyle

	return Model{
		client:               c,
		focus:                FocusPane1,
		selectedResource:     ResourceEvents,
		pane1Cursor:          0, // Start on Events
		availableProjects:    []client.Project{},
		selectedProjectID:    0,
		projectsLoaded:       false,
		listItems:            []ListItem{},
		listCursor:           0,
		filteredItems:        nil,
		inspectorData:        nil,
		inspectorViewport:    viewport.New(0, 0), // Will be sized on WindowSizeMsg
		jsonFoldState:        make(map[string]bool),
		allFolded:            false,
		autoScroll:           true,
		newEventCount:        0,
		lastSeenEventID:      "",
		searchMode:           false,
		searchInput:          textinput.Model{},
		isPolling:            true,
		lastInteraction:      time.Now(),
		lastPoll:             time.Now(),
		clipboardMsg:         "",
		clipboardTime:        time.Time{},
		pendingResourceFetch: nil,
		lastDebounceTime:     time.Time{},
		loading:              true,
		showHelp:             false,
		spinner:              s,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		tickCmd(),
		m.fetchCurrentResource(),
		fetchProjects(m.client),
		m.spinner.Tick,
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

func fetchEvents(c client.PostHogClient) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		events, err := c.ListRecentEvents(ctx, maxEvents)
		if err != nil {
			return errorMsg{err: err}
		}
		return eventsMsg(events)
	}
}

func fetchPersons(c client.PostHogClient) tea.Cmd {
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

func fetchFlags(c client.PostHogClient) tea.Cmd {
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

func fetchProjects(c client.PostHogClient) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		projects, err := c.FetchProjects(ctx)
		if err != nil {
			return errorMsg{err: err}
		}
		return projectsMsg(projects)
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
		// Detect new events if paused
		if !m.autoScroll && len(m.listItems) > 0 {
			newCount := m.detectNewEvents(len(msg))
			m.newEventCount += newCount
		}

		m.listItems = make([]ListItem, len(msg))
		for i, event := range msg {
			m.listItems[i] = EventListItem{Event: event}
		}
		m.loading = false
		m.err = nil

		// Auto-scroll: stay at bottom if enabled
		if m.autoScroll && len(m.listItems) > 0 {
			m.listCursor = len(m.listItems) - 1
			m.newEventCount = 0
		} else {
			// Adjust cursor if out of bounds
			if m.listCursor >= len(m.listItems) && len(m.listItems) > 0 {
				m.listCursor = len(m.listItems) - 1
			}
			if m.listCursor < 0 {
				m.listCursor = 0
			}
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

	case projectsMsg:
		m.availableProjects = msg
		m.projectsLoaded = true

		// Set initial selected project
		if len(msg) > 0 && m.selectedProjectID == 0 {
			// Try to use client's project ID first
			clientProjectID := m.client.GetProjectID()
			if clientProjectID != 0 {
				m.selectedProjectID = clientProjectID
			} else {
				// Fallback: use first available project
				m.selectedProjectID = msg[0].ID
				m.client.SetProjectID(msg[0].ID)
			}
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

	case debounceMsg:
		// Only process if this debounce is still pending and matches current resource
		if m.pendingResourceFetch != nil &&
			*m.pendingResourceFetch == msg.resourceType &&
			(msg.timestamp.Equal(m.lastDebounceTime) || msg.timestamp.After(m.lastDebounceTime)) {

			// Clear pending state
			m.pendingResourceFetch = nil
			m.lastDebounceTime = msg.timestamp

			// Trigger fetch
			m.selectedResource = msg.resourceType
			m.loading = true
			m.listCursor = 0
			m.inspectorData = nil
			return m, m.fetchCurrentResource()
		}
		// Stale debounce, ignore
		return m, nil
	}

	// Update spinner
	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	// Overlay help if active
	if m.showHelp {
		return m.renderHelpOverlay(m.width, m.height)
	}

	var content string

	// Check for narrow terminal
	if m.width < narrowTerminalWidth {
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
	if totalWidth < narrowTerminalWidth {
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
	pane1Width := totalWidth * pane1WidthPercent / 100
	pane2Width := totalWidth * pane2WidthPercent / 100
	pane3Width := totalWidth - pane1Width - pane2Width

	// Ensure minimum widths
	if pane1Width < minPane1Width {
		pane1Width = minPane1Width
	}
	if pane2Width < minPane2Width {
		pane2Width = minPane2Width
	}
	if pane3Width < minPane3Width {
		pane3Width = minPane3Width
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
	var shortcuts []string

	if m.showHelp {
		shortcuts = []string{
			styles.KeyStyle.Render("?") + " close help",
			styles.KeyStyle.Render("Esc") + " close help",
		}
	} else if m.searchMode {
		shortcuts = []string{
			"🔍 " + m.searchInput.Value(),
			styles.KeyStyle.Render("Enter") + " apply",
			styles.KeyStyle.Render("Esc") + " cancel",
		}
	} else {
		// Common shortcuts
		shortcuts = []string{
			styles.KeyStyle.Render("?") + " help",
			styles.KeyStyle.Render("q") + " quit",
		}

		// Context-specific shortcuts based on focus
		switch m.focus {
		case FocusPane1:
			if m.pane1Cursor == pane1CursorProject {
				// On project selector
				shortcuts = append([]string{
					styles.KeyStyle.Render("j/k") + " navigate",
					styles.KeyStyle.Render("Enter") + " select",
					styles.KeyStyle.Render("Tab") + " next",
				}, shortcuts...)
			} else {
				// On resource selector
				shortcuts = append([]string{
					styles.KeyStyle.Render("j/k") + " navigate",
					styles.KeyStyle.Render("1/2/3") + " quick select",
					styles.KeyStyle.Render("Tab") + " next",
				}, shortcuts...)
			}
		case FocusPane2:
			if m.selectedResource == ResourceEvents {
				if m.autoScroll {
					shortcuts = append([]string{
						styles.KeyStyle.Render("j/k") + " navigate",
						styles.KeyStyle.Render("G") + " jump bottom",
						styles.KeyStyle.Render("/") + " search",
						styles.KeyStyle.Render("Tab") + " details",
					}, shortcuts...)
				} else {
					shortcuts = append([]string{
						styles.KeyStyle.Render("j/k") + " navigate",
						styles.KeyStyle.Render("G") + fmt.Sprintf(" resume (%d new)", m.newEventCount),
						styles.KeyStyle.Render("/") + " search",
					}, shortcuts...)
				}
			} else {
				shortcuts = append([]string{
					styles.KeyStyle.Render("j/k") + " navigate",
					styles.KeyStyle.Render("/") + " search",
					styles.KeyStyle.Render("Tab") + " next",
					styles.KeyStyle.Render("Esc") + " back",
				}, shortcuts...)
			}
		case FocusPane3:
			shortcuts = append([]string{
				styles.KeyStyle.Render("j/k") + " scroll",
				styles.KeyStyle.Render("Space") + " fold",
				styles.KeyStyle.Render("y") + " copy",
				styles.KeyStyle.Render("Esc") + " back",
			}, shortcuts...)
		}
	}

	joined := strings.Join(shortcuts, " • ")
	return styles.FooterStyle.Width(m.width - 2).Render(joined)
}
