package miller

import (
	"fmt"

	"github.com/aljazfarkas/lazyhog/internal/ui/styles"
)

// isAtBottomOfList checks if the cursor is within the last 3 items
func (m Model) isAtBottomOfList() bool {
	if len(m.listItems) == 0 {
		return true
	}

	// Consider "at bottom" if within last 3 items
	return m.listCursor >= len(m.listItems)-3
}

// detectNewEvents counts new events by comparing the current list with the incoming message
// This is called when auto-scroll is paused to track how many new events have arrived
func (m Model) detectNewEvents(newItemCount int) int {
	if len(m.listItems) == 0 {
		return newItemCount
	}

	// If we have more items now than before, the difference is new events
	if newItemCount > len(m.listItems) {
		return newItemCount - len(m.listItems)
	}

	return 0
}

// enableAutoScroll jumps to the bottom of the list and enables auto-scroll mode
func (m *Model) enableAutoScroll() {
	if len(m.listItems) > 0 {
		m.listCursor = len(m.listItems) - 1
	}
	m.autoScroll = true
	m.newEventCount = 0
}

// getAutoScrollIndicator returns a status indicator for the list view
// Shows [LIVE] when auto-scrolling, or [PAUSED: X new] when paused with new events
func (m Model) getAutoScrollIndicator() string {
	if m.autoScroll {
		return styles.SuccessTextStyle.Render("[LIVE]")
	}

	if m.newEventCount > 0 {
		return styles.DimTextStyle.Render(fmt.Sprintf("[PAUSED: %d new]", m.newEventCount))
	}

	return styles.DimTextStyle.Render("[PAUSED]")
}
