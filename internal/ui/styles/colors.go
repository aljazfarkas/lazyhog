package styles

import "github.com/charmbracelet/lipgloss"

// New PostHog color palette (Phase 1)
var (
	// Primary colors - PostHog Orange
	ColorOrange      = lipgloss.Color("#f54e00") // PostHog brand orange
	ColorDeepCharcoal = lipgloss.Color("#1d1d26") // Deep charcoal background

	// Legacy colors (kept for rollback)
	ColorPrimary_LEGACY   = lipgloss.Color("#1D4AFF") // PostHog blue
	ColorSecondary_LEGACY = lipgloss.Color("#9B59FF") // PostHog purple

	// Syntax highlighting colors (Monokai-inspired)
	ColorBlue        = lipgloss.Color("#66d9ef") // Cyan-blue for JSON keys
	ColorGreen       = lipgloss.Color("#a6e22e") // Lime green for JSON values
	ColorYellow      = lipgloss.Color("#e6db74") // Yellow for strings
	ColorPurple      = lipgloss.Color("#ae81ff") // Purple for numbers

	// Semantic colors
	ColorSuccess     = lipgloss.Color("#00FF00")
	ColorError       = lipgloss.Color("#FF0000")
	ColorWarning     = lipgloss.Color("#FFA500")
	ColorInfo        = lipgloss.Color("#00FFFF")

	// Grayscale
	ColorDim         = lipgloss.Color("#666666")
	ColorDimmer      = lipgloss.Color("#444444")
	ColorText        = lipgloss.Color("#FFFFFF")
	ColorBorder      = lipgloss.Color("#333333")
)

// Theme represents a color theme
type Theme string

const (
	ThemeOrange Theme = "orange"
	ThemeBlue   Theme = "blue"
)

// GetThemeColors returns the appropriate primary and secondary colors for a theme
func GetThemeColors(theme Theme) (primary, secondary lipgloss.Color) {
	switch theme {
	case ThemeOrange:
		return ColorOrange, ColorDeepCharcoal
	case ThemeBlue:
		return ColorPrimary_LEGACY, ColorSecondary_LEGACY
	default:
		return ColorOrange, ColorDeepCharcoal
	}
}
