package miller

import (
	"fmt"
	"io"

	"github.com/aljazfarkas/lazyhog/internal/ui/styles"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SidebarItem represents an item in the sidebar (Phase 3)
type SidebarItem struct {
	resource Resource
	icon     string
	title    string
	isProject bool // Special flag for project selector
}

// Implement list.Item interface
func (i SidebarItem) FilterValue() string { return i.title }

// SidebarItemDelegate renders sidebar items
type SidebarItemDelegate struct {
	collapsed bool
}

func (d SidebarItemDelegate) Height() int                               { return 1 }
func (d SidebarItemDelegate) Spacing() int                              { return 0 }
func (d SidebarItemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }

func (d SidebarItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	item, ok := listItem.(SidebarItem)
	if !ok {
		return
	}

	// Determine if this item is selected
	isSelected := index == m.Index()

	var str string
	if d.collapsed {
		// Icon-only mode
		if isSelected {
			str = styles.SelectedListItemStyle.Render("▶" + item.icon)
		} else {
			str = styles.ListItemStyle.Render(" " + item.icon)
		}
	} else {
		// Full mode with icon and label
		line := fmt.Sprintf("%s %s", item.icon, item.title)
		if isSelected {
			str = styles.SelectedListItemStyle.Render("▶ " + line)
		} else {
			str = styles.ListItemStyle.Render("  " + line)
		}
	}

	fmt.Fprint(w, str)
}

// SidebarModel wraps the bubbles/list for the sidebar (Phase 3)
type SidebarModel struct {
	list      list.Model
	width     int
	height    int
	collapsed bool
}

// NewSidebarModel creates a new sidebar list model
func NewSidebarModel(width, height int, collapsed bool) SidebarModel {
	items := []list.Item{
		// Project selector as first item
		SidebarItem{
			icon:      "📁",
			title:     "Project",
			isProject: true,
			resource:  -1, // Not a resource
		},
		// Resources
		SidebarItem{
			resource: ResourceEvents,
			icon:     ResourceEvents.Icon(),
			title:    ResourceEvents.String(),
		},
		SidebarItem{
			resource: ResourcePersons,
			icon:     ResourcePersons.Icon(),
			title:    ResourcePersons.String(),
		},
		SidebarItem{
			resource: ResourceFlags,
			icon:     ResourceFlags.Icon(),
			title:    ResourceFlags.String(),
		},
	}

	delegate := SidebarItemDelegate{collapsed: collapsed}

	l := list.New(items, delegate, width, height)
	l.SetShowStatusBar(false)
	l.SetShowTitle(false)
	l.SetShowHelp(false)
	l.SetFilteringEnabled(false)

	// Custom styles
	l.Styles.NoItems = lipgloss.NewStyle().Foreground(styles.ColorDim)

	return SidebarModel{
		list:      l,
		width:     width,
		height:    height,
		collapsed: collapsed,
	}
}

// Update handles sidebar updates
func (m SidebarModel) Update(msg tea.Msg) (SidebarModel, tea.Cmd) {
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// View renders the sidebar
func (m SidebarModel) View() string {
	return m.list.View()
}

// SetSize updates the sidebar size
func (m *SidebarModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.list.SetSize(width, height)
}

// SetCollapsed updates the collapsed state
func (m *SidebarModel) SetCollapsed(collapsed bool) {
	m.collapsed = collapsed
	// Update the delegate
	delegate := SidebarItemDelegate{collapsed: collapsed}
	m.list.SetDelegate(delegate)
}

// GetSelectedItem returns the currently selected sidebar item
func (m SidebarModel) GetSelectedItem() (SidebarItem, bool) {
	item := m.list.SelectedItem()
	if item == nil {
		return SidebarItem{}, false
	}
	sidebarItem, ok := item.(SidebarItem)
	return sidebarItem, ok
}

// GetSelectedIndex returns the current selection index
func (m SidebarModel) GetSelectedIndex() int {
	return m.list.Index()
}

// SetSelectedIndex sets the selection index
func (m *SidebarModel) SetSelectedIndex(index int) {
	if index >= 0 && index < len(m.list.Items()) {
		m.list.Select(index)
	}
}

// MoveUp moves selection up
func (m *SidebarModel) MoveUp() {
	currentIndex := m.list.Index()
	if currentIndex > 0 {
		m.list.Select(currentIndex - 1)
	}
}

// MoveDown moves selection down
func (m *SidebarModel) MoveDown() {
	currentIndex := m.list.Index()
	maxIndex := len(m.list.Items()) - 1
	if currentIndex < maxIndex {
		m.list.Select(currentIndex + 1)
	}
}

// UpdateProjectName updates the project selector item's title
func (m *SidebarModel) UpdateProjectName(name string) {
	items := m.list.Items()
	if len(items) > 0 {
		if item, ok := items[0].(SidebarItem); ok && item.isProject {
			item.title = name
			items[0] = item
			m.list.SetItems(items)
		}
	}
}
