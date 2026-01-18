package miller

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aljazfarkas/lazyhog/internal/client"
	"github.com/aljazfarkas/lazyhog/internal/config"
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

	// Project selection state
	availableProjects []client.Project
	selectedProjectID int
	projectsLoaded    bool

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

	// UI Mode State
	showHelp    bool
	searchMode  bool
	searchQuery string

	// Auto-scroll state (Pane2)
	autoScroll      bool   // Whether auto-scroll is active
	newEventCount   int    // Count of new events since scroll up
	lastSeenEventID string // ID of last event when scroll detached

	// Search/filter state
	filteredItems []ListItem // Filtered list items (nil means no filter active)

	// Inspector scroll state (Pane 3)
	inspectorScroll    int // Vertical scroll offset
	inspectorMaxScroll int // Maximum scroll value

	// JSON folding state (Pane 3)
	jsonFoldState map[string]bool // JSON path -> folded status
	allFolded     bool            // All top-level keys folded

	// Clipboard feedback
	clipboardMsg  string    // Temporary message (2 second TTL)
	clipboardTime time.Time // When clipboard message was set

	// Debounce state for Pane 1 resource selection
	pendingResourceFetch *Resource // nil if no pending fetch
	lastDebounceTime     time.Time // timestamp of last debounce

	// Phase 2 - Navigation enhancements
	pane1Collapsed bool // Whether Pane 1 is in icon-only mode

	// Phase 3 - Sidebar with bubbles/list
	sidebar *SidebarModel // Sidebar list model

	// Phase 4 - Stream table with bubbles/table
	streamTable *StreamTableModel // Stream table model for Pane 2

	// Phase 5 - Inspector with bubbles/viewport
	inspector *InspectorModel // Inspector viewport model for Pane 3

	// Phase 6 - Help system with bubbles/help
	helpBubbles *HelpModel // Context-aware help model

	// Phase 7 - Flag forms with huh
	flagForm     *FlagFormModel // Flag toggle confirmation form
	showFlagForm bool           // Whether flag form is visible
	config       *config.Config // Config for environment detection
}

// Messages
type tickMsg time.Time
type eventsMsg []client.Event
type personsMsg []client.Person
type flagsMsg []client.FeatureFlag
type projectsMsg []client.Project
type errorMsg struct{ err error }
type pivotMsg struct {
	person *client.Person
	events []client.Event
}

// New creates a new Miller Columns model
func New(c *client.Client, cfg *config.Config) Model {
	// Phase 3 - Create sidebar
	sidebar := NewSidebarModel(20, 20, false) // Will be resized on first render
	sidebar.SetSelectedIndex(1) // Start on Events (index 1, since project is index 0)

	// Phase 4 - Create stream table
	streamTable := NewStreamTableModel(50, 20, ResourceEvents) // Will be resized on first render

	// Phase 5 - Create inspector viewport
	inspector := NewInspectorModel(50, 20) // Will be resized on first render

	// Phase 6 - Create help model
	helpBubbles := NewHelpModel()

	return Model{
		client:               c,
		focus:                FocusPane1,
		selectedResource:     ResourceEvents,
		availableProjects:    []client.Project{},
		projectsLoaded:       false,
		listItems:            []ListItem{},
		listCursor:           0,
		isPolling:            true,
		lastInteraction:      time.Now(),
		lastPoll:             time.Now(),
		loading:              true,
		autoScroll:           true,
		newEventCount:        0,
		searchMode:           false,
		searchQuery:          "",
		jsonFoldState:        make(map[string]bool),
		allFolded:            false,
		filteredItems:        nil,
		pendingResourceFetch: nil,
		lastDebounceTime:     time.Time{},
		pane1Collapsed:       false, // Phase 2 - Start expanded
		sidebar:              &sidebar, // Phase 3 - Sidebar
		streamTable:          &streamTable, // Phase 4 - Stream table
		inspector:            &inspector, // Phase 5 - Inspector viewport
		helpBubbles:          &helpBubbles, // Phase 6 - Help system
		flagForm:             nil, // Phase 7 - Flag form (created on demand)
		showFlagForm:         false, // Phase 7 - Flag form visibility
		config:               cfg, // Phase 7 - Config for environment detection
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		tickCmd(),
		m.fetchCurrentResource(),
		fetchProjects(m.client),
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

func fetchProjects(c *client.Client) tea.Cmd {
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
	// Phase 3 - Update sidebar if focused on Pane 1
	if m.focus == FocusPane1 && m.sidebar != nil {
		var cmd tea.Cmd
		*m.sidebar, cmd = m.sidebar.Update(msg)

		// Derive selectedResource from sidebar
		if item, ok := m.sidebar.GetSelectedItem(); ok && !item.isProject {
			m.selectedResource = item.resource
		}

		// Trigger debounce if arrow key was pressed
		debounceCmd := m.deriveResourceAndDebounce(msg)

		if cmd != nil && debounceCmd != nil {
			return m, tea.Batch(cmd, debounceCmd)
		} else if cmd != nil {
			return m, cmd
		} else if debounceCmd != nil {
			return m, debounceCmd
		}
	}

	// Phase 4 - Update stream table if focused on Pane 2
	if m.focus == FocusPane2 && m.streamTable != nil {
		var cmd tea.Cmd
		*m.streamTable, cmd = m.streamTable.Update(msg)

		// Sync table cursor with model state
		m.listCursor = m.streamTable.GetCursor()

		// Update inspector from table selection
		if item := m.streamTable.GetSelectedItem(); item != nil {
			m.inspectorData = item.GetInspectorData()
			// Phase 5 - Update inspector viewport with new data
			if m.inspector != nil {
				m.inspector.SetContent(m.inspectorData, m.selectedResource)
			}
		}

		if cmd != nil {
			return m, cmd
		}
	}

	// Phase 5 - Update inspector viewport if focused on Pane 3
	if m.focus == FocusPane3 && m.inspector != nil {
		var cmd tea.Cmd
		*m.inspector, cmd = m.inspector.Update(msg)

		if cmd != nil {
			return m, cmd
		}
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Phase 3 - Resize sidebar
		if m.sidebar != nil {
			pane1Width, _, _ := m.calculatePaneWidths()
			m.sidebar.SetSize(pane1Width-2, m.height-5) // Adjust for borders and padding
		}

		// Phase 4 - Resize stream table
		if m.streamTable != nil {
			_, pane2Width, _ := m.calculatePaneWidths()
			m.streamTable.SetSize(pane2Width-4, m.height-8) // Adjust for borders and padding
		}

		// Phase 5 - Resize inspector viewport
		if m.inspector != nil {
			_, _, pane3Width := m.calculatePaneWidths()
			m.inspector.SetSize(pane3Width-4, m.height-8) // Adjust for borders and padding
		}

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

		// Phase 4 - Update stream table with items
		if m.streamTable != nil {
			m.streamTable.SetItems(m.listItems, ResourceEvents)
			m.listCursor = m.streamTable.GetCursor()

			// Sync auto-scroll state
			m.autoScroll = m.streamTable.IsAutoScrollEnabled()
			m.newEventCount = m.streamTable.GetNewCount()
		} else {
			// Fallback to old auto-scroll logic
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
		}

		return m, nil

	case personsMsg:
		m.listItems = make([]ListItem, len(msg))
		for i, person := range msg {
			m.listItems[i] = PersonListItem{Person: person}
		}
		m.loading = false
		m.err = nil

		// Phase 4 - Update stream table with items
		if m.streamTable != nil {
			m.streamTable.SetItems(m.listItems, ResourcePersons)
			m.listCursor = m.streamTable.GetCursor()
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

	case flagsMsg:
		m.listItems = make([]ListItem, len(msg))
		for i, flag := range msg {
			m.listItems[i] = FlagListItem{Flag: flag}
		}
		m.loading = false
		m.err = nil

		// Phase 4 - Update stream table with items
		if m.streamTable != nil {
			m.streamTable.SetItems(m.listItems, ResourceFlags)
			m.listCursor = m.streamTable.GetCursor()
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

		// Phase 3 - Update sidebar project name
		if m.sidebar != nil && len(msg) > 0 {
			for _, proj := range msg {
				if proj.ID == m.selectedProjectID {
					m.sidebar.UpdateProjectName(proj.Name)
					break
				}
			}
		}

		return m, nil

	case pivotMsg:
		m.selectedResource = ResourcePersons

		// Set sidebar to Persons (index 2)
		if m.sidebar != nil {
			m.sidebar.SetSelectedIndex(2)
		}

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
		if m.pendingResourceFetch != nil &&
			*m.pendingResourceFetch == msg.resourceType &&
			(msg.timestamp.Equal(m.lastDebounceTime) || msg.timestamp.After(m.lastDebounceTime)) {

			m.pendingResourceFetch = nil
			m.selectedResource = msg.resourceType
			m.loading = true
			m.listCursor = 0
			m.inspectorData = nil
			return m, m.fetchCurrentResource()
		}
		return m, nil
	}

	return m, nil
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

	// Phase 2 - Adjust widths if Pane 1 is collapsed
	var pane1Width int
	if m.pane1Collapsed {
		// Icon-only mode: minimal width for icons
		pane1Width = 10
	} else {
		// Standard width: 15%
		pane1Width = totalWidth * 15 / 100
		// Ensure minimum width
		if pane1Width < 20 {
			pane1Width = 20
		}
	}

	// Distribute remaining width between Pane 2 and Pane 3
	remainingWidth := totalWidth - pane1Width
	pane2Width := remainingWidth * 40 / 100 // 40% of remaining
	pane3Width := totalWidth - pane1Width - pane2Width

	// Ensure minimum widths for Pane 2 and Pane 3
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

	if m.showHelp {
		help = "Press ? or Esc to close help"
	} else if m.searchMode {
		help = fmt.Sprintf("Search: %s | Enter: apply • Esc: cancel", m.searchQuery)
	} else {
		switch m.focus {
		case FocusPane1:
			if item, ok := m.sidebar.GetSelectedItem(); ok && item.isProject {
				help = "↑/↓/j/k: navigate • Enter: cycle project • Tab/→/l: next • ?: help • q: quit"
			} else {
				help = "↑/↓/j/k: select resource • 1/2/3: quick select • Tab/→/l: next • ?: help • q: quit"
			}
		case FocusPane2:
			if m.selectedResource == ResourceEvents {
				if m.autoScroll {
					help = "↑/↓/j/k: navigate • G: jump bottom • /: search • Tab/→/l: details • ?: help • q: quit"
				} else {
					help = fmt.Sprintf("↑/↓/j/k: navigate • G: resume auto-scroll (%d new) • /: search • ?: help • q: quit", m.newEventCount)
				}
			} else {
				help = "↑/↓/j/k: navigate • /: search • Tab/→/l: next • ←/h/Esc: back • ?: help • q: quit"
			}
		case FocusPane3:
			help = "j/k: scroll • Space: fold • Shift+Z: fold all • y: copy JSON • c: copy ID • ←/h/Esc: back • ?: help • q: quit"
		}
	}

	return styles.FooterStyle.Width(m.width - 2).Render(styles.DimTextStyle.Render(help))
}

// deriveResourceAndDebounce triggers debounce after arrow key navigation
func (m *Model) deriveResourceAndDebounce(msg tea.Msg) tea.Cmd {
	if m.sidebar == nil {
		return nil
	}

	// Check if this was a navigation key
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "up", "k", "down", "j":
			// Sidebar just moved, derive resource and debounce
			if item, ok := m.sidebar.GetSelectedItem(); ok && !item.isProject {
				resource := item.resource
				m.pendingResourceFetch = &resource
				m.lastDebounceTime = time.Now()
				return startDebounce(resource)
			} else {
				// On project row, clear pending fetch
				m.pendingResourceFetch = nil
			}
		}
	}

	return nil
}
