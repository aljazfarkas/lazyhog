package query

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aljazfarkas/lazyhog/internal/client"
	"github.com/aljazfarkas/lazyhog/internal/ui/components"
	"github.com/aljazfarkas/lazyhog/internal/ui/styles"
	"github.com/aljazfarkas/lazyhog/internal/utils"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	modeInput = iota
	modeResult
	modeExport
)

// Model represents the query console state
type Model struct {
	client       *client.Client
	query        string
	cursorPos    int
	result       *client.QueryResult
	queryHistory []string
	historyIdx   int
	mode         int
	loading      bool
	err          error
	width        int
	height       int
	scrollX      int
	scrollY      int
	toast        components.Toast
	exportFile   string
}

type queryResultMsg struct {
	result *client.QueryResult
}
type errorMsg struct{ err error }
type exportSuccessMsg struct{}

// New creates a new query console model
func New(c *client.Client) Model {
	return Model{
		client:       c,
		query:        "",
		queryHistory: []string{},
		historyIdx:   -1,
		mode:         modeInput,
		loading:      false,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func executeQuery(c *client.Client, query string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		result, err := c.ExecuteQuery(ctx, query)
		if err != nil {
			return errorMsg{err: err}
		}
		return queryResultMsg{result: result}
	}
}

func exportToCSV(result *client.QueryResult, filename string) tea.Cmd {
	return func() tea.Msg {
		if err := utils.ExportToCSV(result, filename); err != nil {
			return errorMsg{err: err}
		}
		return exportSuccessMsg{}
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		if m.mode == modeExport {
			return m.handleExportMode(msg)
		}

		switch msg.String() {
		case "ctrl+c", "esc":
			if m.mode == modeResult {
				m.mode = modeInput
				return m, nil
			}
			return m, tea.Quit

		case "ctrl+enter":
			if m.mode == modeInput && strings.TrimSpace(m.query) != "" {
				m.loading = true
				m.mode = modeResult
				m.addToHistory(m.query)
				return m, executeQuery(m.client, m.query)
			}

		case "ctrl+s":
			if m.mode == modeResult && m.result != nil {
				m.mode = modeExport
				m.exportFile = "query_result.csv"
				return m, nil
			}

		case "up":
			if m.mode == modeInput {
				if len(m.queryHistory) > 0 {
					if m.historyIdx == -1 {
						m.historyIdx = len(m.queryHistory) - 1
					} else if m.historyIdx > 0 {
						m.historyIdx--
					}
					if m.historyIdx >= 0 && m.historyIdx < len(m.queryHistory) {
						m.query = m.queryHistory[m.historyIdx]
						m.cursorPos = len(m.query)
					}
				}
			} else if m.mode == modeResult {
				if m.scrollY > 0 {
					m.scrollY--
				}
			}

		case "down":
			if m.mode == modeInput {
				if m.historyIdx != -1 {
					if m.historyIdx < len(m.queryHistory)-1 {
						m.historyIdx++
						m.query = m.queryHistory[m.historyIdx]
					} else {
						m.historyIdx = -1
						m.query = ""
					}
					m.cursorPos = len(m.query)
				}
			} else if m.mode == modeResult {
				if m.result != nil && m.scrollY < len(m.result.Results)-1 {
					m.scrollY++
				}
			}

		case "left":
			if m.mode == modeResult && m.scrollX > 0 {
				m.scrollX--
			}

		case "right":
			if m.mode == modeResult && m.result != nil && m.scrollX < len(m.result.Columns)-1 {
				m.scrollX++
			}

		case "backspace":
			if m.mode == modeInput && len(m.query) > 0 {
				m.query = m.query[:len(m.query)-1]
				m.cursorPos = len(m.query)
			}

		case "enter":
			if m.mode == modeInput {
				m.query += "\n"
				m.cursorPos = len(m.query)
			}

		default:
			if m.mode == modeInput && len(msg.String()) == 1 {
				m.query += msg.String()
				m.cursorPos = len(m.query)
			}
		}

	case queryResultMsg:
		m.result = msg.result
		m.loading = false
		m.err = nil
		m.scrollX = 0
		m.scrollY = 0
		return m, nil

	case errorMsg:
		m.err = msg.err
		m.loading = false
		cmd := m.toast.Show(fmt.Sprintf("Error: %v", msg.err), components.ToastError)
		return m, cmd

	case exportSuccessMsg:
		m.mode = modeResult
		cmd := m.toast.Show(fmt.Sprintf("Exported to %s", m.exportFile), components.ToastSuccess)
		return m, cmd

	case components.ToastHideMsg:
		m.toast.Hide()
		return m, nil
	}

	return m, nil
}

func (m *Model) addToHistory(query string) {
	query = strings.TrimSpace(query)
	if query == "" {
		return
	}

	// Don't add duplicate if it's the last entry
	if len(m.queryHistory) > 0 && m.queryHistory[len(m.queryHistory)-1] == query {
		return
	}

	m.queryHistory = append(m.queryHistory, query)

	// Keep only last 10 queries
	if len(m.queryHistory) > 10 {
		m.queryHistory = m.queryHistory[1:]
	}

	m.historyIdx = -1
}

func (m Model) handleExportMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		if m.exportFile != "" && m.result != nil {
			return m, exportToCSV(m.result, m.exportFile)
		}
		m.mode = modeResult
		return m, nil

	case "esc":
		m.mode = modeResult
		return m, nil

	case "backspace":
		if len(m.exportFile) > 0 {
			m.exportFile = m.exportFile[:len(m.exportFile)-1]
		}

	default:
		if len(msg.String()) == 1 {
			m.exportFile += msg.String()
		}
	}

	return m, nil
}

func (m Model) View() string {
	if m.width == 0 {
		return ""
	}

	var sb strings.Builder

	// Header
	header := m.renderHeader()
	sb.WriteString(header)
	sb.WriteString("\n\n")

	// Toast
	if m.toast.Visible {
		sb.WriteString(m.toast.View())
		sb.WriteString("\n\n")
	}

	if m.mode == modeExport {
		sb.WriteString(m.renderExportPrompt())
	} else if m.mode == modeInput {
		sb.WriteString(m.renderInputMode())
	} else {
		sb.WriteString(m.renderResultMode())
	}

	sb.WriteString("\n")
	sb.WriteString(m.renderFooter())

	return sb.String()
}

func (m Model) renderHeader() string {
	title := styles.TitleStyle.Render("🔍 HogQL Query Console")

	var status string
	if m.loading {
		status = styles.DimTextStyle.Render("⟳ Executing...")
	} else if m.result != nil {
		status = styles.SuccessTextStyle.Render(fmt.Sprintf("✓ %d rows", len(m.result.Results)))
	}

	titleLine := lipgloss.JoinHorizontal(
		lipgloss.Left,
		title,
		"  ",
		status,
	)

	return styles.HeaderStyle.Width(m.width - 2).Render(titleLine)
}

func (m Model) renderInputMode() string {
	var sb strings.Builder

	sb.WriteString(styles.JSONKeyStyle.Render("Query:"))
	sb.WriteString("\n\n")

	// Query input
	queryDisplay := m.query + "█"
	sb.WriteString(styles.BorderStyle.Width(m.width - 4).Render(queryDisplay))
	sb.WriteString("\n\n")

	// Examples
	sb.WriteString(styles.DimTextStyle.Render("Examples:"))
	sb.WriteString("\n")
	sb.WriteString(styles.DimTextStyle.Render("  SELECT event, count() FROM events WHERE timestamp > now() - INTERVAL 1 DAY GROUP BY event"))
	sb.WriteString("\n")
	sb.WriteString(styles.DimTextStyle.Render("  SELECT distinct_id, count() as event_count FROM events GROUP BY distinct_id LIMIT 10"))

	return sb.String()
}

func (m Model) renderResultMode() string {
	if m.loading {
		return styles.DimTextStyle.Render("Executing query...")
	}

	if m.result == nil {
		return styles.DimTextStyle.Render("No results yet")
	}

	tableHeight := m.height - 12
	if tableHeight < 5 {
		tableHeight = 5
	}

	table := components.RenderTable(m.result, m.width-4, tableHeight, m.scrollX, m.scrollY)
	return table
}

func (m Model) renderExportPrompt() string {
	var sb strings.Builder

	sb.WriteString(styles.JSONKeyStyle.Render("Export to CSV"))
	sb.WriteString("\n\n")

	sb.WriteString(styles.DimTextStyle.Render("Filename: "))
	sb.WriteString(m.exportFile + "█")
	sb.WriteString("\n\n")

	sb.WriteString(styles.HelpStyle.Render("Press Enter to export, Esc to cancel"))

	return sb.String()
}

func (m Model) renderFooter() string {
	var help string

	if m.mode == modeInput {
		help = "Ctrl+Enter: execute • ↑/↓: history • Enter: new line • Esc: quit"
	} else if m.mode == modeResult {
		help = "Arrow keys: scroll • Ctrl+S: export • Esc: back to query • Ctrl+C: quit"
	} else {
		help = "Enter: save • Esc: cancel"
	}

	footer := styles.DimTextStyle.Render(help)

	// Add query count info
	if len(m.queryHistory) > 0 {
		footer = lipgloss.JoinHorizontal(
			lipgloss.Left,
			footer,
			"  ",
			styles.DimTextStyle.Render("│"),
			"  ",
			styles.DimTextStyle.Render(fmt.Sprintf("%d queries in history", len(m.queryHistory))),
		)
	}

	return styles.FooterStyle.Width(m.width - 2).Render(footer)
}
