package miller

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/aymanbagabas/go-osc52/v2"
	"github.com/aljazfarkas/lazyhog/internal/client"
)

// CopyToClipboard copies text to the system clipboard using OSC 52 escape sequences
// This works across local terminals, tmux, SSH sessions, and most modern terminal emulators
func (m *Model) CopyToClipboard(text string) error {
	// Use OSC 52 to copy to clipboard
	// This works in most modern terminals including over SSH
	output := osc52.New(text)
	_, err := output.WriteTo(os.Stderr)
	return err
}

// showClipboardFeedback displays a temporary success message
func (m *Model) showClipboardFeedback(msg string) {
	m.clipboardMsg = msg
	m.clipboardTime = time.Now()
}

// extractCopyableID extracts the distinct_id or event_id from the current inspector data
func (m Model) extractCopyableID() string {
	if m.inspectorData == nil {
		return ""
	}

	// Try to extract ID based on resource type
	switch m.selectedResource {
	case ResourceEvents:
		if event, ok := m.inspectorData.(client.Event); ok {
			// Prefer UUID if available, otherwise use distinct_id
			if event.UUID != "" {
				return event.UUID
			}
			return event.DistinctID
		}

	case ResourcePersons:
		if person, ok := m.inspectorData.(client.Person); ok {
			// Return the first distinct ID
			if len(person.DistinctIDs) > 0 {
				return person.DistinctIDs[0]
			}
			return person.ID
		}

	case ResourceFlags:
		if flag, ok := m.inspectorData.(client.FeatureFlag); ok {
			return flag.Key
		}
	}

	return ""
}

// extractFullJSON extracts the full JSON representation of the current inspector data
func (m Model) extractFullJSON() (string, error) {
	if m.inspectorData == nil {
		return "", fmt.Errorf("no data to copy")
	}

	jsonBytes, err := json.MarshalIndent(m.inspectorData, "", "  ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}
