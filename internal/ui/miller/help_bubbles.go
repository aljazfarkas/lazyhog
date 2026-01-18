package miller

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
)

// KeyMap defines all keybindings for the application (Phase 6)
type KeyMap struct {
	// Global
	Help     key.Binding
	Quit     key.Binding
	Tab      key.Binding
	Back     key.Binding
	Collapse key.Binding
	Jump1    key.Binding
	Jump2    key.Binding
	Jump3    key.Binding

	// Navigation
	NavUp   key.Binding
	NavDown key.Binding
	NavLeft key.Binding
	NavRight key.Binding

	// Pane 1 (Sidebar)
	SelectResource key.Binding
	CycleProject   key.Binding

	// Pane 2 (Stream/List)
	Search      key.Binding
	Refresh     key.Binding
	Pivot       key.Binding
	GotoBottom  key.Binding

	// Pane 3 (Inspector)
	Scroll    key.Binding
	Fold      key.Binding
	FoldAll   key.Binding
	CopyJSON  key.Binding
	CopyID    key.Binding
}

// ShortHelp returns keybindings to be shown in the mini help view (context-aware)
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Quit, k.Tab, k.Back}
}

// FullHelp returns keybindings for the expanded help view
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Help, k.Quit, k.Tab, k.Back},
		{k.Collapse, k.Jump1, k.Jump2, k.Jump3},
		{k.NavUp, k.NavDown, k.Search, k.Refresh},
		{k.Pivot, k.GotoBottom, k.Fold, k.FoldAll},
		{k.CopyJSON, k.CopyID},
	}
}

// DefaultKeyMap returns the default keybindings
func DefaultKeyMap() KeyMap {
	return KeyMap{
		// Global
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "toggle help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		Tab: key.NewBinding(
			key.WithKeys("tab", "right", "l"),
			key.WithHelp("tab/l/→", "next pane"),
		),
		Back: key.NewBinding(
			key.WithKeys("shift+tab", "left", "h", "esc"),
			key.WithHelp("esc/h/←", "prev pane"),
		),
		Collapse: key.NewBinding(
			key.WithKeys("b"),
			key.WithHelp("b", "collapse sidebar"),
		),
		Jump1: key.NewBinding(
			key.WithKeys("1"),
			key.WithHelp("1", "jump to pane 1"),
		),
		Jump2: key.NewBinding(
			key.WithKeys("2"),
			key.WithHelp("2", "jump to pane 2"),
		),
		Jump3: key.NewBinding(
			key.WithKeys("3"),
			key.WithHelp("3", "jump to pane 3"),
		),

		// Navigation
		NavUp: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		NavDown: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		NavLeft: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("←/h", "left"),
		),
		NavRight: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("→/l", "right"),
		),

		// Pane 1
		SelectResource: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select"),
		),
		CycleProject: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "cycle project"),
		),

		// Pane 2
		Search: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "search"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh"),
		),
		Pivot: key.NewBinding(
			key.WithKeys("p"),
			key.WithHelp("p", "pivot to person"),
		),
		GotoBottom: key.NewBinding(
			key.WithKeys("G"),
			key.WithHelp("G", "go to bottom"),
		),

		// Pane 3
		Scroll: key.NewBinding(
			key.WithKeys("j", "k", "up", "down"),
			key.WithHelp("j/k", "scroll"),
		),
		Fold: key.NewBinding(
			key.WithKeys(" "),
			key.WithHelp("space", "fold/unfold"),
		),
		FoldAll: key.NewBinding(
			key.WithKeys("Z"),
			key.WithHelp("Z", "fold all"),
		),
		CopyJSON: key.NewBinding(
			key.WithKeys("y"),
			key.WithHelp("y", "copy JSON"),
		),
		CopyID: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "copy ID"),
		),
	}
}

// GetContextualHelp returns context-aware keybindings based on focus
func GetContextualHelp(focus Focus, resource Resource) []key.Binding {
	km := DefaultKeyMap()

	switch focus {
	case FocusPane1:
		// Sidebar
		return []key.Binding{
			km.NavUp,
			km.NavDown,
			km.CycleProject,
			km.Collapse,
			km.Tab,
		}

	case FocusPane2:
		// Stream/List
		bindings := []key.Binding{
			km.NavUp,
			km.NavDown,
			km.Search,
			km.Refresh,
		}
		if resource == ResourceEvents {
			bindings = append(bindings, km.GotoBottom, km.Pivot)
		}
		return bindings

	case FocusPane3:
		// Inspector
		bindings := []key.Binding{
			km.Scroll,
			km.Fold,
			km.FoldAll,
			km.CopyJSON,
			km.CopyID,
		}
		if resource == ResourceEvents {
			bindings = append(bindings, km.Pivot)
		}
		return bindings

	default:
		return km.ShortHelp()
	}
}

// HelpModel wraps bubbles/help for dynamic footer
type HelpModel struct {
	help   help.Model
	keymap KeyMap
	width  int
}

// NewHelpModel creates a new help model
func NewHelpModel() HelpModel {
	h := help.New()
	h.ShowAll = false // Start with short help

	return HelpModel{
		help:   h,
		keymap: DefaultKeyMap(),
		width:  80,
	}
}

// SetWidth updates the help width
func (h *HelpModel) SetWidth(width int) {
	h.width = width
	h.help.Width = width
}

// ToggleShowAll toggles between short and full help
func (h *HelpModel) ToggleShowAll() {
	h.help.ShowAll = !h.help.ShowAll
}

// View renders the help view
func (h HelpModel) View(focus Focus, resource Resource) string {
	if h.help.ShowAll {
		return h.help.View(h.keymap)
	}

	// Context-aware short help
	contextKeys := GetContextualHelp(focus, resource)
	return h.help.ShortHelpView(contextKeys)
}
