package miller

import (
	"fmt"

	"github.com/aljazfarkas/lazyhog/internal/client"
	"github.com/aljazfarkas/lazyhog/internal/config"
	"github.com/aljazfarkas/lazyhog/internal/ui/styles"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

// FlagFormModel wraps a huh.Form for flag toggle confirmation (Phase 7)
type FlagFormModel struct {
	form      *huh.Form
	flag      client.FeatureFlag
	action    string // "enable" or "disable"
	confirmed *bool  // Pointer to confirmation value
	env       string // "dev" or "prod"
}

// NewFlagToggleForm creates a new flag toggle confirmation form
func NewFlagToggleForm(flag client.FeatureFlag, cfg *config.Config) *FlagFormModel {
	// Determine environment
	env := cfg.DetectEnvironment()

	// Determine action
	action := "enable"
	if flag.Active {
		action = "disable"
	}

	// Create warning message based on environment
	warningMsg := fmt.Sprintf("Are you sure you want to %s flag '%s'?", action, flag.Key)

	var envWarning string
	if env == "prod" {
		envWarning = styles.ErrorTextStyle.Render("⚠️  WARNING: This is a PRODUCTION environment!")
	} else {
		envWarning = styles.SuccessTextStyle.Render("ℹ️  This is a DEV environment")
	}

	// Create confirmation field
	confirmed := false

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewNote().
				Title(fmt.Sprintf("Toggle Feature Flag: %s", flag.Key)).
				Description(fmt.Sprintf("%s\n\n%s\n\nCurrent status: %s\nNew status: %s",
					warningMsg,
					envWarning,
					formatFlagStatus(flag.Active),
					formatFlagStatus(!flag.Active),
				)),

			huh.NewConfirm().
				Title(fmt.Sprintf("Confirm %s?", action)).
				Value(&confirmed).
				Affirmative("Yes, proceed").
				Negative("No, cancel"),
		),
	)

	// Apply environment-specific styling
	if env == "prod" {
		form = form.WithTheme(huh.ThemeBase()).
			WithShowHelp(false)
	}

	return &FlagFormModel{
		form:      form,
		flag:      flag,
		action:    action,
		confirmed: &confirmed,
		env:       env,
	}
}

// formatFlagStatus returns a formatted status string
func formatFlagStatus(active bool) string {
	if active {
		return styles.SuccessTextStyle.Render("✓ Active")
	}
	return styles.ErrorTextStyle.Render("✗ Inactive")
}

// Run runs the form and returns whether it was confirmed
func (f *FlagFormModel) Run() error {
	return f.form.Run()
}

// IsConfirmed returns whether the user confirmed the action
func (f *FlagFormModel) IsConfirmed() bool {
	if f.confirmed != nil {
		return *f.confirmed
	}
	return false
}

// GetFlag returns the flag being toggled
func (f *FlagFormModel) GetFlag() client.FeatureFlag {
	return f.flag
}

// GetAction returns the action (enable/disable)
func (f *FlagFormModel) GetAction() string {
	return f.action
}

// View renders the form
func (f *FlagFormModel) View() string {
	// Render form with environment-specific border
	var borderColor lipgloss.Color
	if f.env == "prod" {
		borderColor = styles.ColorError
	} else {
		borderColor = styles.ColorPrimary
	}

	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(1, 2)

	return style.Render(f.form.View())
}
