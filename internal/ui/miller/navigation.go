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
		if m.focus == FocusPane1 {
			return m, tea.Quit
		}
		// From other panes, q moves focus left
		m.MoveFocusLeft()
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
		m.MoveResourceSelectorUp()
		return m, nil

	case "down", "j":
		m.MoveResourceSelectorDown()
		return m, nil

	case "enter":
		// Switch resource and fetch data
		m.loading = true
		m.listCursor = 0
		m.inspectorData = nil
		return m, m.fetchCurrentResource()

	case "1":
		m.selectedResource = ResourceEvents
		m.loading = true
		m.listCursor = 0
		m.inspectorData = nil
		return m, m.fetchCurrentResource()

	case "2":
		m.selectedResource = ResourcePersons
		m.loading = true
		m.listCursor = 0
		m.inspectorData = nil
		return m, m.fetchCurrentResource()

	case "3":
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

	case "enter":
		m.SelectCurrentListItem()
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
	case "esc", "h":
		m.MoveFocusLeft()
		return m, nil

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
