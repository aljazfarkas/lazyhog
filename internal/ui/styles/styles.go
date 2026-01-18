package styles

import "github.com/charmbracelet/lipgloss"

// PostHog brand colors (now using Orange theme - Phase 1)
// All color definitions are now in colors.go
var (
	ColorPrimary   = ColorOrange       // PostHog orange (new default)
	ColorSecondary = ColorDeepCharcoal // Deep charcoal (new default)
)

// Common styles
var (
	// Title styles
	TitleStyle = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true).
			Padding(0, 1)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(ColorDim).
			Italic(true)

	// Border styles
	BorderStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(ColorBorder).
			Padding(1, 2)

	ThinBorderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorder).
			Padding(0, 1)

	// List styles
	ListItemStyle = lipgloss.NewStyle().
			Padding(0, 2)

	SelectedListItemStyle = lipgloss.NewStyle().
				Foreground(ColorPrimary).
				Bold(true).
				Padding(0, 2)

	// Text styles
	DimTextStyle = lipgloss.NewStyle().
			Foreground(ColorDim)

	HighlightTextStyle = lipgloss.NewStyle().
				Foreground(ColorInfo).
				Bold(true)

	ErrorTextStyle = lipgloss.NewStyle().
			Foreground(ColorError).
			Bold(true)

	SuccessTextStyle = lipgloss.NewStyle().
				Foreground(ColorSuccess).
				Bold(true)

	// Help text
	HelpStyle = lipgloss.NewStyle().
			Foreground(ColorDim).
			Padding(1, 0)

	// Header styles
	HeaderStyle = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(ColorBorder).
			Padding(0, 1).
			MarginBottom(1)

	// Footer styles
	FooterStyle = lipgloss.NewStyle().
			Foreground(ColorDim).
			BorderStyle(lipgloss.NormalBorder()).
			BorderTop(true).
			BorderForeground(ColorBorder).
			Padding(0, 1).
			MarginTop(1)

	// Status indicators
	StatusActiveStyle = lipgloss.NewStyle().
				Foreground(ColorSuccess).
				Bold(true)

	StatusInactiveStyle = lipgloss.NewStyle().
				Foreground(ColorDim)

	// JSON viewer (Phase 1 - Monokai-inspired syntax highlighting)
	JSONKeyStyle = lipgloss.NewStyle().
			Foreground(ColorBlue). // Cyan-blue for keys
			Bold(true)

	JSONValueStyle = lipgloss.NewStyle().
			Foreground(ColorGreen) // Lime green for values

	JSONStringStyle = lipgloss.NewStyle().
			Foreground(ColorYellow) // Yellow for strings

	JSONNumberStyle = lipgloss.NewStyle().
			Foreground(ColorPurple) // Purple for numbers

	JSONBoolStyle = lipgloss.NewStyle().
			Foreground(ColorWarning) // Orange for booleans

	JSONNullStyle = lipgloss.NewStyle().
			Foreground(ColorDim) // Dim gray for null
)

// Helper functions

// Max returns the maximum of two integers
func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Min returns the minimum of two integers
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// TruncateString truncates a string to the specified length with ellipsis
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}
