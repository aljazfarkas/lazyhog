package miller

import (
	"github.com/aljazfarkas/lazyhog/internal/ui/styles"
	"github.com/charmbracelet/lipgloss"
)

// Focus represents which pane is currently focused
type Focus int

const (
	FocusPane1 Focus = iota // Resource selector
	FocusPane2              // List view
	FocusPane3              // Inspector
)

// String returns a human-readable representation of the focus
func (f Focus) String() string {
	switch f {
	case FocusPane1:
		return "Resource Selector"
	case FocusPane2:
		return "List View"
	case FocusPane3:
		return "Inspector"
	default:
		return "Unknown"
	}
}

// GetBorderStyle returns the border style for a pane based on focus
func GetBorderStyle(currentFocus Focus, paneNumber int) lipgloss.Style {
	focused := (int(currentFocus) == paneNumber)

	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder())

	if focused {
		style = style.BorderForeground(styles.ColorPrimary)
	} else {
		style = style.BorderForeground(styles.ColorBorder)
	}

	return style
}

// MoveFocusRight moves focus to the next pane (right)
func (m *Model) MoveFocusRight() {
	switch m.focus {
	case FocusPane1:
		m.focus = FocusPane2
	case FocusPane2:
		m.focus = FocusPane3
	case FocusPane3:
		// Stay at Pane 3
	}
}

// MoveFocusLeft moves focus to the previous pane (left)
func (m *Model) MoveFocusLeft() {
	switch m.focus {
	case FocusPane3:
		m.focus = FocusPane2
	case FocusPane2:
		m.focus = FocusPane1
	case FocusPane1:
		// Stay at Pane 1
	}
}

// Phase 2 - Navigation enhancements

// TogglePane1Collapse toggles the sidebar between full and icon-only mode
func (m *Model) TogglePane1Collapse() {
	m.pane1Collapsed = !m.pane1Collapsed

	// Phase 3 - Update sidebar collapsed state
	if m.sidebar != nil {
		m.sidebar.SetCollapsed(m.pane1Collapsed)

		// Resize sidebar to match new collapsed state
		pane1Width := 10 // collapsed width
		if !m.pane1Collapsed {
			pane1Width = m.width * 15 / 100
			if pane1Width < 20 {
				pane1Width = 20
			}
		}
		m.sidebar.SetSize(pane1Width-2, m.height-5)
	}
}

// JumpToPane jumps directly to a specific pane
func (m *Model) JumpToPane(pane Focus) {
	if pane >= FocusPane1 && pane <= FocusPane3 {
		m.focus = pane
	}
}
