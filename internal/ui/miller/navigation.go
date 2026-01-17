package miller

import (
	tea "github.com/charmbracelet/bubbletea"
)

// handleKeyPress handles all keyboard input based on current focus
func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Record interaction for polling pause
	m.recordInteraction()

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

	case "tab", "l":
		m.MoveFocusRight()
		return m, nil

	case "shift+tab", "h":
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
	switch msg.String() {
	case "up", "k":
		m.MoveListCursorUp()
		return m, nil

	case "down", "j":
		m.MoveListCursorDown()
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
	case "up", "k", "down", "j":
		// For future scrolling within inspector
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
