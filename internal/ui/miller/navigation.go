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

	// Phase 2 - New navigation shortcuts
	case "b":
		// Toggle sidebar collapse (icons-only mode)
		m.TogglePane1Collapse()
		return m, nil

	case "1":
		// Jump to Pane 1 (unless already in Pane 1, where it selects Events resource)
		if m.focus != FocusPane1 {
			m.JumpToPane(FocusPane1)
			return m, nil
		}
		// Fall through to pane-specific handling for resource selection

	case "2":
		// Jump to Pane 2 (unless in Pane 1, where it selects Persons resource)
		if m.focus != FocusPane1 {
			m.JumpToPane(FocusPane2)
			return m, nil
		}
		// Fall through to pane-specific handling for resource selection

	case "3":
		// Jump to Pane 3 (unless in Pane 1, where it selects Flags resource)
		if m.focus != FocusPane1 {
			m.JumpToPane(FocusPane3)
			return m, nil
		}
		// Fall through to pane-specific handling for resource selection
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
	case "enter":
		// Only handle project cycling
		if item, ok := m.sidebar.GetSelectedItem(); ok && item.isProject {
			return m.handleProjectSwitch()
		}
		return m, nil

	case "1":
		m.selectedResource = ResourceEvents
		if m.sidebar != nil {
			m.sidebar.SetSelectedIndex(1) // Events
		}
		m.pendingResourceFetch = nil
		m.loading = true
		m.listCursor = 0
		m.inspectorData = nil
		return m, m.fetchCurrentResource()

	case "2":
		m.selectedResource = ResourcePersons
		if m.sidebar != nil {
			m.sidebar.SetSelectedIndex(2) // Persons
		}
		m.pendingResourceFetch = nil
		m.loading = true
		m.listCursor = 0
		m.inspectorData = nil
		return m, m.fetchCurrentResource()

	case "3":
		m.selectedResource = ResourceFlags
		if m.sidebar != nil {
			m.sidebar.SetSelectedIndex(3) // Flags
		}
		m.pendingResourceFetch = nil
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
		// Phase 4 - Use stream table for auto-scroll
		if m.streamTable != nil {
			m.streamTable.EnableAutoScroll()
			m.autoScroll = true
			m.newEventCount = 0
		} else {
			m.enableAutoScroll()
		}
		return m, nil

	case "up", "k":
		// Phase 4 - Use stream table for cursor movement
		if m.streamTable != nil {
			m.streamTable.MoveUp()
			m.listCursor = m.streamTable.GetCursor()
			m.autoScroll = m.streamTable.IsAutoScrollEnabled()
		} else {
			m.MoveListCursorUp()
			// Disable auto-scroll if we move away from bottom
			if m.autoScroll && !m.isAtBottomOfList() {
				m.autoScroll = false
			}
		}
		return m, nil

	case "down", "j":
		// Phase 4 - Use stream table for cursor movement
		if m.streamTable != nil {
			m.streamTable.MoveDown()
			m.listCursor = m.streamTable.GetCursor()
			m.autoScroll = m.streamTable.IsAutoScrollEnabled()
			m.newEventCount = m.streamTable.GetNewCount()
		} else {
			m.MoveListCursorDown()
			// Re-enable auto-scroll if we reach bottom
			if !m.autoScroll && m.isAtBottomOfList() {
				m.autoScroll = true
				m.newEventCount = 0
			}
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
		// Phase 5 - Use inspector viewport for scrolling
		if m.inspector != nil {
			m.inspector.ScrollDown()
		} else {
			// Fallback to old scrolling
			if m.inspectorScroll < m.inspectorMaxScroll {
				m.inspectorScroll++
			}
		}
		return m, nil

	case "k", "up":
		// Phase 5 - Use inspector viewport for scrolling
		if m.inspector != nil {
			m.inspector.ScrollUp()
		} else {
			// Fallback to old scrolling
			if m.inspectorScroll > 0 {
				m.inspectorScroll--
			}
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
