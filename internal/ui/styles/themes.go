package styles

import "github.com/charmbracelet/lipgloss"

// Environment represents the application environment
type Environment string

const (
	EnvDev  Environment = "dev"
	EnvProd Environment = "prod"
)

// EnvironmentConfig holds environment-specific styling
type EnvironmentConfig struct {
	Environment Environment
	BorderColor lipgloss.Color
	LabelText   string
	LabelStyle  lipgloss.Style
}

// GetEnvironmentConfig returns the styling configuration for an environment
func GetEnvironmentConfig(env Environment) EnvironmentConfig {
	switch env {
	case EnvDev:
		return EnvironmentConfig{
			Environment: EnvDev,
			BorderColor: ColorWarning, // Yellow border for dev
			LabelText:   "DEV",
			LabelStyle: lipgloss.NewStyle().
				Foreground(ColorWarning).
				Bold(true).
				Padding(0, 1),
		}
	case EnvProd:
		return EnvironmentConfig{
			Environment: EnvProd,
			BorderColor: ColorError, // Red border for production
			LabelText:   "⚠️  PRODUCTION",
			LabelStyle: lipgloss.NewStyle().
				Foreground(ColorError).
				Bold(true).
				Padding(0, 1).
				Background(lipgloss.Color("#330000")), // Dark red background
		}
	default:
		// Default to dev for safety
		return GetEnvironmentConfig(EnvDev)
	}
}

// ApplyEnvironmentBorder applies environment-specific border styling to a style
func ApplyEnvironmentBorder(baseStyle lipgloss.Style, env Environment) lipgloss.Style {
	config := GetEnvironmentConfig(env)
	return baseStyle.BorderForeground(config.BorderColor)
}
