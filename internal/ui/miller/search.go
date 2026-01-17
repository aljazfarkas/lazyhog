package miller

import (
	"strings"

	"github.com/aljazfarkas/lazyhog/internal/ui/styles"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// handleSearchKeys handles keyboard input when in search mode
func (m Model) handleSearchKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		// Cancel search and restore full list
		m.searchMode = false
		m.searchQuery = ""
		m.filteredItems = nil
		return m, nil

	case "enter":
		// Apply filter
		if m.searchQuery != "" {
			m.filteredItems = m.applyFilter(m.listItems, m.searchQuery)
			if len(m.filteredItems) > 0 {
				m.listCursor = 0 // Jump to first result
			}
		} else {
			m.filteredItems = nil
		}
		m.searchMode = false
		return m, nil

	case "backspace":
		// Remove last character
		if len(m.searchQuery) > 0 {
			m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
		}
		return m, nil

	case "ctrl+u":
		// Clear entire query
		m.searchQuery = ""
		return m, nil

	default:
		// Add character to query
		// Only add printable characters
		if len(msg.String()) == 1 {
			m.searchQuery += msg.String()
		}
		return m, nil
	}
}

// applyFilter filters list items by substring match
// Searches in the rendered line content (case-insensitive)
func (m Model) applyFilter(items []ListItem, query string) []ListItem {
	if query == "" {
		return items
	}

	query = strings.ToLower(query)
	var filtered []ListItem

	for _, item := range items {
		// Render the item and search in the rendered text
		rendered := item.RenderLine(100, false) // Use a reasonable width
		if strings.Contains(strings.ToLower(rendered), query) {
			filtered = append(filtered, item)
		}
	}

	return filtered
}

// renderSearchInput renders the search input modal overlay
func (m Model) renderSearchInput(query string, width int) string {
	prompt := "Search: "
	cursor := "█" // Block cursor

	var sb strings.Builder
	sb.WriteString(prompt)
	sb.WriteString(query)
	sb.WriteString(cursor)

	// Style the search input
	searchStyle := lipgloss.NewStyle().
		Foreground(styles.ColorPrimary).
		Background(lipgloss.Color("#1a1a1a")).
		Padding(0, 1).
		Bold(true)

	return searchStyle.Render(sb.String())
}

// getEffectiveListItems returns either filtered items or all items
// This helper is used throughout the code to respect active filters
func (m Model) getEffectiveListItems() []ListItem {
	if m.filteredItems != nil {
		return m.filteredItems
	}
	return m.listItems
}
