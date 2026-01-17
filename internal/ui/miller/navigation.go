package miller

import (
	tea "github.com/charmbracelet/bubbletea"
)

// handleKeyPress handles all keyboard input based on current focus
func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Record interaction for polling pause
	m.recordInteraction()

	// Global help toggle
	if msg.String() == "?" {
		m.showHelp = !m.showHelp
		return m, nil
	}

	// Block other input when help is shown
	if m.showHelp {
		if msg.String() == "esc" {
			m.showHelp = false
		}
		return m, nil
	}

	// Check search mode BEFORE global navigation shortcuts
	// This prevents "l", "h", and other keys from triggering navigation while typing search queries
	if m.searchMode {
		return m.handleSearchKeys(msg)
	}

	// Global shortcuts
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit

	case "q":
		// Always quit from any pane
		return m, tea.Quit

	case "esc":
		// In help mode, close help
		if m.showHelp {
			m.showHelp = false
			return m, nil
		}
		// In search mode, cancel search
		if m.searchMode {
			m.searchMode = false
			m.searchQuery = ""
			m.filteredItems = nil
			return m, nil
		}
		// Otherwise, go back (move focus left)
		m.MoveFocusLeft()
		return m, nil

	case "h":
		// Vim-style: move focus left (go back)
		m.MoveFocusLeft()
		return m, nil

	case "l":
		// Vim-style: move focus right (go forward)
		m.MoveFocusRight()
		return m, nil

	case "tab", "right":
		m.MoveFocusRight()
		return m, nil

	case "shift+tab", "left":
		m.MoveFocusLeft()
		return m, nil
	}

	// Pane-specific shortcuts
	switch m.focus {
	case FocusPane1:
		return m.handlePane1Keys(msg)
	case FocusPane2:
		return m.handlePane2Keys(msg)
	case FocusPane3:
		return m.handlePane3Keys(msg)
	}

	return m, nil
}

// handlePane1Keys handles keyboard input for Pane 1 (Resource Selector)
func (m Model) handlePane1Keys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		m.MovePane1CursorUp()
		return m, nil

	case "down", "j":
		m.MovePane1CursorDown()
		return m, nil

	case "enter":
		// If on project (cursor = -1), cycle to next project
		if m.pane1Cursor == -1 {
			return m.handleProjectSwitch()
		}

		// If on resource, select it
		m.selectedResource = Resource(m.pane1Cursor)
		m.loading = true
		m.listCursor = 0
		m.inspectorData = nil
		return m, m.fetchCurrentResource()

	case "1":
		m.pane1Cursor = 0
		m.selectedResource = ResourceEvents
		m.loading = true
		m.listCursor = 0
		m.inspectorData = nil
		return m, m.fetchCurrentResource()

	case "2":
		m.pane1Cursor = 1
		m.selectedResource = ResourcePersons
		m.loading = true
		m.listCursor = 0
		m.inspectorData = nil
		return m, m.fetchCurrentResource()

	case "3":
		m.pane1Cursor = 2
		m.selectedResource = ResourceFlags
		m.loading = true
		m.listCursor = 0
		m.inspectorData = nil
		return m, m.fetchCurrentResource()
	}

	return m, nil
}

// handlePane2Keys handles keyboard input for Pane 2 (List View)
func (m Model) handlePane2Keys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Search mode is now handled globally in handleKeyPress()

	switch msg.String() {
	case "/":
		m.searchMode = true
		m.searchQuery = ""
		return m, nil

	case "G":
		m.enableAutoScroll()
		return m, nil

	case "up", "k":
		m.MoveListCursorUp()
		// Disable auto-scroll if we move away from bottom
		if m.autoScroll && !m.isAtBottomOfList() {
			m.autoScroll = false
		}
		return m, nil

	case "down", "j":
		m.MoveListCursorDown()
		// Re-enable auto-scroll if we reach bottom
		if !m.autoScroll && m.isAtBottomOfList() {
			m.autoScroll = true
			m.newEventCount = 0
		}
		return m, nil

	case "r":
		m.loading = true
		return m, m.fetchCurrentResource()

	case "p":
		// Pivot feature: only available for Events
		if m.selectedResource == ResourceEvents {
			return m.handlePivot()
		}
		return m, nil
	}

	return m, nil
}

// handlePane3Keys handles keyboard input for Pane 3 (Inspector)
func (m Model) handlePane3Keys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "j", "down":
		if m.inspectorScroll < m.inspectorMaxScroll {
			m.inspectorScroll++
		}
		return m, nil

	case "k", "up":
		if m.inspectorScroll > 0 {
			m.inspectorScroll--
		}
		return m, nil

	case " ":
		// Toggle fold at cursor
		m.toggleJSONFoldAtCursor()
		return m, nil

	case "Z":
		// Shift+Z: Fold all top-level keys
		m.jsonFoldAll()
		return m, nil

	case "y":
		// Copy raw JSON
		if m.inspectorData != nil {
			jsonStr, err := m.extractFullJSON()
			if err == nil {
				err = m.CopyToClipboard(jsonStr)
				if err == nil {
					m.showClipboardFeedback("Copied JSON")
				} else {
					m.showClipboardFeedback("Copy failed")
				}
			}
		}
		return m, nil

	case "c":
		// Copy ID
		id := m.extractCopyableID()
		if id != "" {
			err := m.CopyToClipboard(id)
			if err == nil {
				m.showClipboardFeedback("Copied ID: " + id)
			} else {
				m.showClipboardFeedback("Copy failed")
			}
		}
		return m, nil

	case "p":
		// Pivot feature: only available for Events
		if m.selectedResource == ResourceEvents {
			return m.handlePivot()
		}
		return m, nil
	}

	return m, nil
}

// MovePane1CursorUp moves cursor up in Pane 1 (project + resources)
func (m *Model) MovePane1CursorUp() {
	if m.pane1Cursor > -1 {
		m.pane1Cursor--
	}
}

// MovePane1CursorDown moves cursor down in Pane 1 (project + resources)
func (m *Model) MovePane1CursorDown() {
	maxCursor := 2 // ResourceFlags
	if m.pane1Cursor < maxCursor {
		m.pane1Cursor++
	}
}

// handleProjectSwitch handles Enter key on project selector
func (m Model) handleProjectSwitch() (tea.Model, tea.Cmd) {
	if !m.projectsLoaded || len(m.availableProjects) == 0 {
		return m, nil
	}

	// Simple cycling: find current project index and move to next
	currentIndex := -1
	for i, proj := range m.availableProjects {
		if proj.ID == m.selectedProjectID {
			currentIndex = i
			break
		}
	}

	// Cycle to next project
	nextIndex := (currentIndex + 1) % len(m.availableProjects)
	m.selectedProjectID = m.availableProjects[nextIndex].ID

	// Update client project ID
	m.client.SetProjectID(m.selectedProjectID)

	// Refetch current resource with new project
	m.loading = true
	m.listCursor = 0
	m.inspectorData = nil

	return m, m.fetchCurrentResource()
}
