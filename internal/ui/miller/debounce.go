package miller

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// debounceMsg is sent after the debounce delay
type debounceMsg struct {
	resourceType Resource
	timestamp    time.Time
}

// startDebounce returns a command that sends a debounceMsg after 200ms
func startDebounce(resource Resource) tea.Cmd {
	return tea.Tick(200*time.Millisecond, func(t time.Time) tea.Msg {
		return debounceMsg{
			resourceType: resource,
			timestamp:    t,
		}
	})
}
