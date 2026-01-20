package miller

import (
	"strings"

	"github.com/aljazfarkas/lazyhog/internal/ui/styles"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// enterSearchMode initializes and activates search mode
func (m *Model) enterSearchMode() {
	m.searchMode = true
	m.searchInput = textinput.New()
	m.searchInput.Placeholder = "Search..."
	m.searchInput.Prompt = "🔍 "
	m.searchInput.PromptStyle = styles.SearchPromptStyle
	m.searchInput.TextStyle = styles.SearchTextStyle
	m.searchInput.Focus()
}

// handleSearchKeys handles keyboard input when in search mode
func (m Model) handleSearchKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.String() {
	case "esc":
		// Cancel search and restore full list
		m.searchMode = false
		m.searchInput.Blur()
		m.filteredItems = nil
		return m, nil

	case "enter":
		// Apply filter
		query := m.searchInput.Value()
		if query != "" {
			m.filteredItems = m.applyFilter(m.listItems, query)
			if len(m.filteredItems) > 0 {
				m.listCursor = 0 // Jump to first result
			}
		} else {
			m.filteredItems = nil
		}
		m.searchMode = false
		m.searchInput.Blur()
		return m, nil
	}

	// Let textinput handle all other keys
	m.searchInput, cmd = m.searchInput.Update(msg)
	return m, cmd
}

// applyFilter filters list items by substring match
// Searches in the searchable text (case-insensitive, no ANSI codes)
func (m Model) applyFilter(items []ListItem, query string) []ListItem {
	if query == "" {
		return items
	}

	query = strings.ToLower(query)
	var filtered []ListItem

	for _, item := range items {
		// Search in the plain text content (no ANSI codes)
		searchText := item.GetSearchableText()
		if strings.Contains(strings.ToLower(searchText), query) {
			filtered = append(filtered, item)
		}
	}

	return filtered
}

// renderSearchInput renders the search input modal overlay
func (m Model) renderSearchInput(width int) string {
	container := styles.SearchContainerStyle.Render(m.searchInput.View())
	return lipgloss.Place(width, 1, lipgloss.Center, lipgloss.Center, container)
}

// getEffectiveListItems returns either filtered items or all items
// This helper is used throughout the code to respect active filters
func (m Model) getEffectiveListItems() []ListItem {
	if m.filteredItems != nil {
		return m.filteredItems
	}
	return m.listItems
}
