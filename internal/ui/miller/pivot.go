package miller

import (
	"context"
	"time"

	"github.com/aljazfarkas/lazyhog/internal/client"
	tea "github.com/charmbracelet/bubbletea"
)

// handlePivot handles the pivot from event to person
func (m Model) handlePivot() (tea.Model, tea.Cmd) {
	if len(m.listItems) == 0 || m.listCursor >= len(m.listItems) {
		return m, nil
	}

	// Get the distinct_id from the current item
	distinctID := m.listItems[m.listCursor].GetDistinctID()
	if distinctID == "" {
		// No distinct ID available, can't pivot
		return m, nil
	}

	// Fetch person and their events
	return m, fetchPersonByDistinctID(m.client, distinctID)
}

// fetchPersonByDistinctID fetches a person and their recent events
func fetchPersonByDistinctID(c client.PostHogClient, distinctID string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		person, err := c.GetPerson(ctx, distinctID)
		if err != nil {
			return errorMsg{err: err}
		}

		events, eventsErr := c.GetPersonEvents(ctx, distinctID, 20)
		if eventsErr != nil {
			// Don't fail completely if events fail, continue with empty events
			events = []client.Event{}
		}

		return pivotMsg{
			person:      person,
			events:      events,
			eventsError: eventsErr,
		}
	}
}
