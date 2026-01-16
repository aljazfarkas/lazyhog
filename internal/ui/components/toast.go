package components

import (
	"time"

	"github.com/aljazfarkas/lazyhog/internal/ui/styles"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ToastType represents the type of toast notification
type ToastType int

const (
	ToastSuccess ToastType = iota
	ToastError
	ToastInfo
	ToastWarning
)

const toastDuration = 3 * time.Second

// Toast represents a toast notification
type Toast struct {
	Message   string
	Type      ToastType
	Visible   bool
	ShowUntil time.Time
}

// ToastHideMsg is sent when the toast should be hidden
type ToastHideMsg struct{}

// NewToast creates a new toast notification
func NewToast(message string, toastType ToastType) Toast {
	return Toast{
		Message:   message,
		Type:      toastType,
		Visible:   true,
		ShowUntil: time.Now().Add(toastDuration),
	}
}

// Show displays the toast
func (t *Toast) Show(message string, toastType ToastType) tea.Cmd {
	t.Message = message
	t.Type = toastType
	t.Visible = true
	t.ShowUntil = time.Now().Add(toastDuration)

	return tea.Tick(toastDuration, func(time.Time) tea.Msg {
		return ToastHideMsg{}
	})
}

// Hide hides the toast
func (t *Toast) Hide() {
	t.Visible = false
}

// Update handles toast updates
func (t *Toast) Update(msg tea.Msg) tea.Cmd {
	switch msg.(type) {
	case ToastHideMsg:
		if time.Now().After(t.ShowUntil) {
			t.Hide()
		}
	}
	return nil
}

// View renders the toast
func (t Toast) View() string {
	if !t.Visible {
		return ""
	}

	var style lipgloss.Style
	var icon string

	switch t.Type {
	case ToastSuccess:
		style = lipgloss.NewStyle().
			Foreground(styles.ColorSuccess).
			Background(lipgloss.Color("#002200")).
			Padding(0, 2).
			Bold(true)
		icon = "✓"
	case ToastError:
		style = lipgloss.NewStyle().
			Foreground(styles.ColorError).
			Background(lipgloss.Color("#220000")).
			Padding(0, 2).
			Bold(true)
		icon = "✗"
	case ToastWarning:
		style = lipgloss.NewStyle().
			Foreground(styles.ColorWarning).
			Background(lipgloss.Color("#222200")).
			Padding(0, 2).
			Bold(true)
		icon = "⚠"
	case ToastInfo:
		style = lipgloss.NewStyle().
			Foreground(styles.ColorInfo).
			Background(lipgloss.Color("#002222")).
			Padding(0, 2).
			Bold(true)
		icon = "ℹ"
	}

	return style.Render(icon + " " + t.Message)
}
