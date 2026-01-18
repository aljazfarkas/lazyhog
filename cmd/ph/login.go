package main

import (
	"fmt"
	"strings"

	"github.com/aljazfarkas/lazyhog/internal/config"
	"github.com/aljazfarkas/lazyhog/internal/ui/styles"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var (
	apiKeyFlag     string
	instanceURLFlag string
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with your PostHog instance",
	Long: `Configure your PostHog API credentials.

You need a Personal API Key (starts with phx_) from PostHog → Settings → Personal API Keys.
Note: Project API Keys (phc_) are for sending events and won't work with this tool.
For PostHog Cloud, use the default instance URL (https://app.posthog.com).
For self-hosted instances, provide your custom URL.`,
	RunE: runLogin,
}

func init() {
	rootCmd.AddCommand(loginCmd)
	loginCmd.Flags().StringVar(&apiKeyFlag, "api-key", "", "PostHog Personal API key (starts with phx_)")
	loginCmd.Flags().StringVar(&instanceURLFlag, "instance-url", "https://app.posthog.com", "PostHog instance URL")
}

type loginModel struct {
	apiKey      string
	instanceURL string
	step        int // 0: api key, 1: instance url, 2: done
	err         error
	width       int
	height      int
}

type loginSuccessMsg struct{}
type loginErrorMsg struct{ err error }

func (m loginModel) Init() tea.Cmd {
	return tea.EnterAltScreen
}

func (m loginModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit

		case "enter":
			if m.step == 0 {
				// Validate API key
				if err := config.ValidateAPIKey(m.apiKey); err != nil {
					m.err = err
					return m, nil
				}
				m.step = 1
				m.err = nil
				return m, nil
			} else if m.step == 1 {
				// Validate instance URL
				if err := config.ValidateInstanceURL(m.instanceURL); err != nil {
					m.err = err
					return m, nil
				}
				// Save config
				cfg := &config.Config{
					ProjectAPIKey: strings.TrimSpace(m.apiKey),
					InstanceURL:   config.NormalizeInstanceURL(m.instanceURL),
				}
				if err := config.Save(cfg); err != nil {
					return m, func() tea.Msg {
						return loginErrorMsg{err: err}
					}
				}
				m.step = 2
				return m, func() tea.Msg {
					return loginSuccessMsg{}
				}
			}

		case "backspace":
			if m.step == 0 && len(m.apiKey) > 0 {
				m.apiKey = m.apiKey[:len(m.apiKey)-1]
			} else if m.step == 1 && len(m.instanceURL) > 0 {
				m.instanceURL = m.instanceURL[:len(m.instanceURL)-1]
			}

		default:
			// Type characters
			if m.step == 0 {
				m.apiKey += msg.String()
			} else if m.step == 1 {
				m.instanceURL += msg.String()
			}
		}

	case loginSuccessMsg:
		return m, tea.Quit

	case loginErrorMsg:
		m.err = msg.err
		return m, tea.Quit
	}

	return m, nil
}

func (m loginModel) View() string {
	if m.step == 2 {
		configPath, _ := config.GetConfigPath()

		return fmt.Sprintf("\n%s\n\n%s\n\n%s\n\n",
			styles.LoginSuccessStyle.Render("✓ Authentication configured successfully!"),
			fmt.Sprintf("Configuration saved to: %s", styles.CaptionStyle.Render(configPath)),
			"Run 'lazyhog live' to start streaming events.",
		)
	}

	var sb strings.Builder

	sb.WriteString(styles.LoginTitleStyle.Render("PostHog Authentication Setup"))
	sb.WriteString("\n\n")

	if m.step == 0 {
		sb.WriteString(styles.LoginPromptStyle.Render("Enter your PostHog Personal API Key:"))
		sb.WriteString("\n")
		// Mask API key display
		maskedKey := strings.Repeat("*", len(m.apiKey))
		if len(m.apiKey) < 10 {
			maskedKey = m.apiKey
		}
		sb.WriteString(styles.LoginInputStyle.Render("> " + maskedKey))
		sb.WriteString("\n")
		if m.err != nil {
			sb.WriteString("\n")
			sb.WriteString(styles.LoginErrorStyle.Render("✗ " + m.err.Error()))
		}
		sb.WriteString("\n")
		sb.WriteString(styles.LoginHelpStyle.Render("Find your API key in PostHog → Settings → Personal API Keys"))
		sb.WriteString("\n")
		sb.WriteString(styles.LoginHelpStyle.Render("Must start with phx_ (not phc_) • Press Enter to continue • Esc to cancel"))
	} else if m.step == 1 {
		sb.WriteString(styles.LoginPromptStyle.Render("Enter your PostHog Instance URL:"))
		sb.WriteString("\n")
		sb.WriteString(styles.LoginInputStyle.Render("> " + m.instanceURL))
		sb.WriteString("\n")
		if m.err != nil {
			sb.WriteString("\n")
			sb.WriteString(styles.LoginErrorStyle.Render("✗ " + m.err.Error()))
		}
		sb.WriteString("\n")
		sb.WriteString(styles.LoginHelpStyle.Render("Default: https://app.posthog.com (for PostHog Cloud)"))
		sb.WriteString("\n")
		sb.WriteString(styles.LoginHelpStyle.Render("Press Enter to save • Esc to cancel"))
	}

	return "\n" + sb.String() + "\n"
}

func runLogin(cmd *cobra.Command, args []string) error {
	// Non-interactive mode if flags provided
	if apiKeyFlag != "" {
		if err := config.ValidateAPIKey(apiKeyFlag); err != nil {
			return fmt.Errorf("invalid API key: %w", err)
		}

		if instanceURLFlag == "" {
			instanceURLFlag = "https://app.posthog.com"
		}

		if err := config.ValidateInstanceURL(instanceURLFlag); err != nil {
			return fmt.Errorf("invalid instance URL: %w", err)
		}

		cfg := &config.Config{
			ProjectAPIKey: strings.TrimSpace(apiKeyFlag),
			InstanceURL:   config.NormalizeInstanceURL(instanceURLFlag),
		}

		if err := config.Save(cfg); err != nil {
			return fmt.Errorf("failed to save configuration: %w", err)
		}

		configPath, _ := config.GetConfigPath()
		fmt.Printf("✓ Authentication configured successfully!\n")
		fmt.Printf("Configuration saved to: %s\n", configPath)
		fmt.Printf("\nRun 'lazyhog live' to start streaming events.\n")
		return nil
	}

	// Interactive mode
	m := loginModel{
		instanceURL: "https://app.posthog.com",
		step:        0,
	}

	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("failed to run login UI: %w", err)
	}

	if finalModel.(loginModel).err != nil {
		return finalModel.(loginModel).err
	}

	return nil
}
