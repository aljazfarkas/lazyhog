package styles

import "github.com/charmbracelet/lipgloss"

// PostHog brand colors - Enhanced palette
var (
	// Enhanced Primary Colors
	ColorPrimaryBlue      = lipgloss.Color("#1D4AFF") // PostHog blue
	ColorPrimaryBlueDark  = lipgloss.Color("#0F2A9F") // Darker variant
	ColorPrimaryBlueLight = lipgloss.Color("#5B7FFF") // Lighter variant
	ColorSecondaryPurple  = lipgloss.Color("#9B59FF") // PostHog purple
	ColorAccentOrange     = lipgloss.Color("#F96132") // PostHog accent
	ColorAccentYellow     = lipgloss.Color("#F7A501") // PostHog warning

	// Enhanced Semantic Colors with Dark Backgrounds
	ColorSuccess     = lipgloss.Color("#46C66B") // PostHog green
	ColorSuccessDark = lipgloss.Color("#1D5A33") // For toast backgrounds
	ColorError       = lipgloss.Color("#EF4444") // PostHog red
	ColorErrorDark   = lipgloss.Color("#5A1D1D") // For toast backgrounds
	ColorWarning     = lipgloss.Color("#F7A501") // PostHog yellow
	ColorWarningDark = lipgloss.Color("#5A4A1D") // For toast backgrounds
	ColorInfo        = lipgloss.Color("#1DA1F2") // Cyan/blue
	ColorInfoDark    = lipgloss.Color("#1D3D5A") // For toast backgrounds

	// Enhanced Grayscale (PostHog dark theme)
	ColorBG             = lipgloss.Color("#0A0A0A") // Almost black
	ColorBGElevated     = lipgloss.Color("#1A1A1A") // Elevated surfaces
	ColorBGModal        = lipgloss.Color("#262626") // Modals/overlays
	ColorText           = lipgloss.Color("#F5F5F5") // Primary text
	ColorTextSecondary  = lipgloss.Color("#A8A8A8") // Secondary text
	ColorTextTertiary   = lipgloss.Color("#666666") // Dim text
	ColorBorder         = lipgloss.Color("#333333") // Standard border
	ColorBorderSubtle   = lipgloss.Color("#1F1F1F") // Subtle borders
	ColorDim            = lipgloss.Color("#666666") // Legacy dim
	ColorDimmer         = lipgloss.Color("#444444") // Legacy dimmer

	// Enhanced JSON Syntax Colors
	ColorJSONKey     = lipgloss.Color("#9B59FF") // Purple
	ColorJSONString  = lipgloss.Color("#46C66B") // Green
	ColorJSONNumber  = lipgloss.Color("#F7A501") // Orange
	ColorJSONBoolean = lipgloss.Color("#1DA1F2") // Blue
	ColorJSONNull    = lipgloss.Color("#666666") // Dim

	// Legacy compatibility (keep old names pointing to new colors)
	ColorPrimary   = ColorPrimaryBlue
	ColorSecondary = ColorSecondaryPurple
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

	// JSON viewer
	JSONKeyStyle = lipgloss.NewStyle().
			Foreground(ColorJSONKey).
			Bold(true)

	JSONValueStyle = lipgloss.NewStyle().
			Foreground(ColorText)

	JSONStringStyle = lipgloss.NewStyle().
			Foreground(ColorJSONString)

	JSONNumberStyle = lipgloss.NewStyle().
			Foreground(ColorJSONNumber)

	JSONBoolStyle = lipgloss.NewStyle().
			Foreground(ColorJSONBoolean)

	JSONNullStyle = lipgloss.NewStyle().
			Foreground(ColorJSONNull)

	// Search Input Styles
	SearchPromptStyle = lipgloss.NewStyle().
				Foreground(ColorPrimaryBlue).
				Bold(true)

	SearchTextStyle = lipgloss.NewStyle().
			Foreground(ColorText)

	SearchContainerStyle = lipgloss.NewStyle().
				Background(ColorBGElevated).
				Padding(0, 1).
				Bold(true)

	// Toast Styles
	ToastSuccessStyle = lipgloss.NewStyle().
				Foreground(ColorSuccess).
				Background(ColorSuccessDark).
				Padding(0, 2).
				Bold(true)

	ToastErrorStyle = lipgloss.NewStyle().
				Foreground(ColorError).
				Background(ColorErrorDark).
				Padding(0, 2).
				Bold(true)

	ToastWarningStyle = lipgloss.NewStyle().
				Foreground(ColorWarning).
				Background(ColorWarningDark).
				Padding(0, 2).
				Bold(true)

	ToastInfoStyle = lipgloss.NewStyle().
				Foreground(ColorInfo).
				Background(ColorInfoDark).
				Padding(0, 2).
				Bold(true)

	// Login Screen Styles
	LoginTitleStyle = lipgloss.NewStyle().
			Foreground(ColorPrimaryBlue).
			Bold(true).
			MarginBottom(1)

	LoginPromptStyle = lipgloss.NewStyle().
				Foreground(ColorText)

	LoginInputStyle = lipgloss.NewStyle().
			Foreground(ColorInfo)

	LoginErrorStyle = lipgloss.NewStyle().
			Foreground(ColorError)

	LoginHelpStyle = lipgloss.NewStyle().
			Foreground(ColorTextTertiary).
			MarginTop(1)

	LoginSuccessStyle = lipgloss.NewStyle().
				Foreground(ColorSuccess).
				Bold(true)

	// Typography Hierarchy
	H1Style = lipgloss.NewStyle().
		Foreground(ColorText).
		Bold(true).
		Underline(true).
		MarginBottom(1)

	H2Style = lipgloss.NewStyle().
		Foreground(ColorPrimaryBlue).
		Bold(true).
		MarginBottom(1)

	H3Style = lipgloss.NewStyle().
		Foreground(ColorTextSecondary).
		Bold(true)

	BodyStyle = lipgloss.NewStyle().
			Foreground(ColorText)

	CaptionStyle = lipgloss.NewStyle().
			Foreground(ColorTextTertiary).
			Italic(true)

	// Keyboard Shortcut Display
	KeyStyle = lipgloss.NewStyle().
		Foreground(ColorPrimaryBlue).
		Bold(true)

	// Spinner Style
	SpinnerStyle = lipgloss.NewStyle().
			Foreground(ColorPrimaryBlue)
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
